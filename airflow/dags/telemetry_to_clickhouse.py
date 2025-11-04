from airflow import DAG
from airflow.operators.python_operator import PythonOperator
from datetime import datetime, timedelta
import logging
import psycopg2
import requests

default_args = {
    'owner': 'airflow',
    'depends_on_past': False,
    'start_date': datetime(2024, 1, 1),
    'email_on_failure': False,
    'email_on_retry': False,
    'retries': 1,
    'retry_delay': timedelta(minutes=5),
}

def transfer_telemetry_data():
    logging.info("Starting telemetry data transfer (fixed datetime format)...")
    
    try:
        # Прямое подключение к PostgreSQL
        conn = psycopg2.connect(
            host="postgres_telemetry",
            database="telemetry", 
            user="telemetry_user",
            password="telemetry_password",
            port="5432"
        )
        cursor = conn.cursor()
        
        # Выборка данных
        query = "SELECT * FROM telemetry_data"
        cursor.execute(query)
        rows = cursor.fetchall()
        
        logging.info(f"Found {len(rows)} records to transfer")
        
        if not rows:
            logging.info("No records found")
            return
        
        # Вставка в ClickHouse с правильным форматом дат
        for row in rows:
            # Преобразуем даты в формат без микросекунд
            timestamp_str = row[2].strftime('%Y-%m-%d %H:%M:%S') if row[2] else '1970-01-01 00:00:00'
            created_at_str = row[8].strftime('%Y-%m-%d %H:%M:%S') if row[8] else '1970-01-01 00:00:00'
            
            insert_query = f"""
            INSERT INTO reports.telemetry VALUES (
                {row[0]}, '{row[1]}', '{timestamp_str}', '{row[3]}', 
                {row[4]}, {row[5]}, {row[6]}, {row[7]}, '{created_at_str}'
            )
            """
            
            response = requests.post(
                'http://clickhouse:8123/',
                data=insert_query,
                params={
                    'user': 'clickhouse_user',
                    'password': 'clickhouse_password', 
                    'database': 'reports'
                }
            )
            
            if response.status_code != 200:
                logging.error(f"ClickHouse error: {response.text}")
            else:
                logging.info(f"✓ Successfully inserted record {row[0]}")
        
        logging.info("Data transfer completed successfully")
        cursor.close()
        conn.close()
        
    except Exception as e:
        logging.error(f"Error in transfer: {str(e)}")
        raise

with DAG(
    'telemetry_to_clickhouse',
    default_args=default_args,
    description='Telemetry data transfer with fixed datetime format',
    schedule_interval=timedelta(hours=1),
    catchup=False,
) as dag:

    transfer_task = PythonOperator(
        task_id='transfer_telemetry_data',
        python_callable=transfer_telemetry_data,
    )