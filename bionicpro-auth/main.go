package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

type Session struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

type PKCESession struct {
	CodeVerifier string
	State        string
}

var (
	sessions       = make(map[string]*Session)
	sessionsMux    sync.RWMutex
	pceSessions    = make(map[string]*PKCESession)
	pceSessionsMux sync.RWMutex
	encryptionKey  = []byte("12345678901234567890123456789012")
)

func main() {
	// CORS middleware
	corsMiddleware := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Set CORS headers
			w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Credentials", "true")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next(w, r)
		}
	}

	// Apply CORS middleware to all handlers
	http.HandleFunc("/auth/login", corsMiddleware(handleLogin))
	http.HandleFunc("/auth/callback", corsMiddleware(handleCallback))
	http.HandleFunc("/auth/refresh", corsMiddleware(handleRefresh))
	http.HandleFunc("/reports", corsMiddleware(handleReports))

	fmt.Println("Server starting on :8000")
	http.ListenAndServe(":8000", nil)
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	// Generate PKCE code verifier and challenge
	codeVerifier := generateCodeVerifier()
	codeChallenge := generateCodeChallenge(codeVerifier)
	state := generateState()

	// Store PKCE session
	pceSessionsMux.Lock()
	pceSessions[state] = &PKCESession{
		CodeVerifier: codeVerifier,
		State:        state,
	}
	pceSessionsMux.Unlock()

	authURL := "http://localhost:8080/realms/reports-realm/protocol/openid-connect/auth"
	params := url.Values{}
	params.Add("client_id", "reports-frontend")
	params.Add("response_type", "code")
	params.Add("scope", "openid")
	params.Add("redirect_uri", "http://localhost:8000/auth/callback")
	params.Add("state", state)
	params.Add("code_challenge", codeChallenge)
	params.Add("code_challenge_method", "S256")

	redirectURL := authURL + "?" + params.Encode()
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

func handleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	if code == "" {
		errorDesc := r.URL.Query().Get("error_description")
		http.Error(w, "Missing authorization code: "+errorDesc, http.StatusBadRequest)
		return
	}

	if state == "" {
		http.Error(w, "Missing state parameter", http.StatusBadRequest)
		return
	}

	// Retrieve PKCE session
	pceSessionsMux.RLock()
	pkceSession, exists := pceSessions[state]
	pceSessionsMux.RUnlock()

	if !exists {
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}

	// Clean up PKCE session
	pceSessionsMux.Lock()
	delete(pceSessions, state)
	pceSessionsMux.Unlock()

	tokens, err := exchangeCodeForTokens(code, pkceSession.CodeVerifier)
	if err != nil {
		http.Error(w, "Failed to exchange code: "+err.Error(), http.StatusInternalServerError)
		return
	}

	sessionID := generateSessionID()
	encryptedRefreshToken, err := encrypt(tokens.RefreshToken)
	if err != nil {
		http.Error(w, "Failed to encrypt token", http.StatusInternalServerError)
		return
	}

	sessionsMux.Lock()
	sessions[sessionID] = &Session{
		AccessToken:  tokens.AccessToken,
		RefreshToken: encryptedRefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(tokens.ExpiresIn) * time.Second),
	}
	sessionsMux.Unlock()

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   3600,
	})

	http.Redirect(w, r, "http://localhost:3000", http.StatusFound)
}

func exchangeCodeForTokens(code, codeVerifier string) (*TokenResponse, error) {
	tokenURL := "http://keycloak:8080/realms/reports-realm/protocol/openid-connect/token"

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("client_id", "reports-frontend")
	data.Set("code", code)
	data.Set("redirect_uri", "http://localhost:8000/auth/callback")
	data.Set("code_verifier", codeVerifier)

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange failed: %s", string(body))
	}

	var tokens TokenResponse
	if err := json.Unmarshal(body, &tokens); err != nil {
		return nil, err
	}

	return &tokens, nil
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Report data would be here",
		"status":  "success",
		"user":    "authenticated",
	})
}

func handleRefresh(w http.ResponseWriter, r *http.Request) {
	sessionCookie, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	sessionsMux.RLock()
	session, exists := sessions[sessionCookie.Value]
	sessionsMux.RUnlock()

	if !exists {
		http.Error(w, "Session not found", http.StatusUnauthorized)
		return
	}

	newSessionID, err := refreshTokens(sessionCookie.Value, session)
	if err != nil {
		http.Error(w, "Refresh failed: "+err.Error(), http.StatusUnauthorized)
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

	w.WriteHeader(http.StatusOK)
}

func refreshTokens(oldSessionID string, session *Session) (string, error) {
	refreshToken, err := decrypt(session.RefreshToken)
	if err != nil {
		return "", err
	}

	tokenURL := "http://keycloak:8080/realms/reports-realm/protocol/openid-connect/token"

	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("client_id", "reports-frontend")
	data.Set("refresh_token", refreshToken)

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("refresh failed: %s", string(body))
	}

	var tokens TokenResponse
	if err := json.Unmarshal(body, &tokens); err != nil {
		return "", err
	}

	newSessionID := generateSessionID()
	encryptedRefreshToken, err := encrypt(tokens.RefreshToken)
	if err != nil {
		return "", err
	}

	sessionsMux.Lock()
	defer sessionsMux.Unlock()

	delete(sessions, oldSessionID)
	sessions[newSessionID] = &Session{
		AccessToken:  tokens.AccessToken,
		RefreshToken: encryptedRefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(tokens.ExpiresIn) * time.Second),
	}

	return newSessionID, nil
}

// PKCE functions
func generateCodeVerifier() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func generateCodeChallenge(verifier string) string {
	hash := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}

func generateState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func generateSessionID() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func encrypt(text string) (string, error) {
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(text), nil)
	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

func decrypt(encryptedText string) (string, error) {
	data, err := base64.URLEncoding.DecodeString(encryptedText)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
