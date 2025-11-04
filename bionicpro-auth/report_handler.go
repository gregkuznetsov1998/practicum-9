package main

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	_ "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type ReportData struct {
	UserID           string  `json:"user_id"`
	Branch           string  `json:"branch"`
	EmployeeType     string  `json:"employee_type"`
	TotalSteps       int     `json:"total_steps"`
	AvgActivityLevel float64 `json:"avg_activity_level"`
	AvgTemperature   float64 `json:"avg_temperature"`
	LastActivity     string  `json:"last_activity"`
}

func initStorage() error {
	var err error

	dsn := "clickhouse://clickhouse_user:clickhouse_password@clickhouse:9000/reports?dial_timeout=10s&compress=true"

	clickhouseDB, err = sql.Open("clickhouse", dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to ClickHouse: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := clickhouseDB.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping ClickHouse: %v", err)
	}

	minioClient, err = minio.New("minio:9000", &minio.Options{
		Creds:  credentials.NewStaticV4("minioadmin", "minioadmin", ""),
		Secure: false,
	})
	if err != nil {
		return fmt.Errorf("failed to create MinIO client: %v", err)
	}

	bucketCtx, bucketCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer bucketCancel()

	exists, err := minioClient.BucketExists(bucketCtx, "reports")
	if err != nil {
		return fmt.Errorf("failed to check bucket existence: %v", err)
	}

	if !exists {
		err = minioClient.MakeBucket(bucketCtx, "reports", minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %v", err)
		}

		policy := `{
			"Version": "2012-10-17",
			"Statement": [
				{
				"Sid": "PublicReadGetObject",
				"Effect": "Allow",
				"Principal": "*",
				"Action": [
					"s3:GetObject"
				],
				"Resource": [
					"arn:aws:s3:::reports/*"
				]
				}
			]
		}`

		err = minioClient.SetBucketPolicy(bucketCtx, "reports", policy)
		if err != nil {
			fmt.Printf("Warning: Could not set bucket policy: %v\n", err)
		}

		fmt.Println("Bucket 'reports' created successfully with public read policy")
	} else {
		fmt.Println("Bucket 'reports' already exists")
	}

	fmt.Println("Storage initialized successfully")
	return nil
}

func handleReports(w http.ResponseWriter, r *http.Request) {
	sessionCookie, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "Unauthorized - no session", http.StatusUnauthorized)
		return
	}

	sessionsMux.RLock()
	session, exists := sessions[sessionCookie.Value]
	sessionsMux.RUnlock()

	if !exists {
		http.Error(w, "Session not found", http.StatusUnauthorized)
		return
	}

	if time.Now().After(session.ExpiresAt) {
		newSessionID, err := refreshTokens(sessionCookie.Value, session)
		if err != nil {
			http.Error(w, "Token refresh failed: "+err.Error(), http.StatusUnauthorized)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "session_id",
			Value:    newSessionID,
			Path:     "/",
			HttpOnly: true,
			Secure:   false,
			SameSite: http.SameSiteLaxMode,
			MaxAge:   3600,
		})

		sessionsMux.RLock()
		session = sessions[newSessionID]
		sessionsMux.RUnlock()
	}

	userInfo, err := decodeJWTToken(session.AccessToken)
	if err != nil {
		http.Error(w, "Failed to decode user token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	ctx := r.Context()

	objectName := fmt.Sprintf("%s/report.csv", userInfo.Username)
	existsInMinIO, err := checkReportExists(ctx, objectName)
	if err != nil {
		http.Error(w, "Error checking report: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if existsInMinIO {
		reportURL := fmt.Sprintf("http://localhost/reports/%s/report.csv", userInfo.Username)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"report_url": reportURL,
			"status":     "cached",
			"user":       userInfo.Username,
		})
		return
	}

	reportData, err := generateReport(ctx, userInfo.Username)
	if err != nil {
		http.Error(w, "Error generating report: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = saveReportToMinIO(ctx, objectName, reportData)
	if err != nil {
		http.Error(w, "Error saving report: "+err.Error(), http.StatusInternalServerError)
		return
	}

	reportURL := fmt.Sprintf("http://localhost/reports/%s/report.csv", userInfo.Username)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"report_url": reportURL,
		"status":     "generated",
		"user":       userInfo.Username,
	})
}

func decodeJWTToken(accessToken string) (*UserInfo, error) {
	parts := strings.Split(accessToken, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid JWT token format, got %d parts", len(parts))
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode JWT payload: %v", err)
	}

	var tokenPayload JWTTokenPayload
	if err := json.Unmarshal(payload, &tokenPayload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JWT payload: %v", err)
	}

	userInfo := &UserInfo{
		Username: tokenPayload.PreferredUsername,
		Email:    tokenPayload.Email,
		UserID:   tokenPayload.Sub,
		Name:     tokenPayload.Name,
	}

	fmt.Printf("Decoded user info: %s (preferred_username), %s (email), %s (sub)\n",
		userInfo.Username, userInfo.Email, userInfo.UserID)

	return userInfo, nil
}

func checkReportExists(ctx context.Context, objectName string) (bool, error) {
	_, err := minioClient.StatObject(ctx, "reports", objectName, minio.StatObjectOptions{})
	if err != nil {
		if errResp, ok := err.(minio.ErrorResponse); ok && errResp.Code == "NoSuchKey" {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func generateReport(ctx context.Context, username string) (string, error) {
	query := `
		SELECT 
			user_id,
			branch,
			employee_type,
			total_steps,
			avg_activity_level,
			avg_temperature,
			last_activity_date
		FROM reports.data_mart 
		WHERE user_id = ?
	`

	var report ReportData
	err := clickhouseDB.QueryRowContext(ctx, query, username).Scan(
		&report.UserID,
		&report.Branch,
		&report.EmployeeType,
		&report.TotalSteps,
		&report.AvgActivityLevel,
		&report.AvgTemperature,
		&report.LastActivity,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Printf("No data found for user: %s, creating empty report\n", username)
			report = ReportData{
				UserID:       username,
				Branch:       "Unknown",
				EmployeeType: "Unknown",
				TotalSteps:   0,
			}
		} else {
			return "", fmt.Errorf("query error: %v", err)
		}
	}

	var csvBuilder strings.Builder
	writer := csv.NewWriter(&csvBuilder)

	headers := []string{"User ID", "Branch", "Employee Type", "Total Steps", "Avg Activity Level", "Avg Temperature", "Last Activity"}
	data := []string{
		report.UserID,
		report.Branch,
		report.EmployeeType,
		fmt.Sprintf("%d", report.TotalSteps),
		fmt.Sprintf("%.2f", report.AvgActivityLevel),
		fmt.Sprintf("%.2f", report.AvgTemperature),
		report.LastActivity,
	}

	if err := writer.Write(headers); err != nil {
		return "", fmt.Errorf("failed to write CSV headers: %v", err)
	}
	if err := writer.Write(data); err != nil {
		return "", fmt.Errorf("failed to write CSV data: %v", err)
	}
	writer.Flush()

	if err := writer.Error(); err != nil {
		return "", fmt.Errorf("CSV writer error: %v", err)
	}

	return csvBuilder.String(), nil
}

func saveReportToMinIO(ctx context.Context, objectName, reportData string) error {
	reader := strings.NewReader(reportData)

	_, err := minioClient.PutObject(ctx, "reports", objectName, reader, int64(len(reportData)), minio.PutObjectOptions{
		ContentType: "text/csv",
	})

	if err != nil {
		return fmt.Errorf("failed to save report to MinIO: %v", err)
	}

	fmt.Printf("Report saved to MinIO: reports/%s\n", objectName)
	return nil
}
