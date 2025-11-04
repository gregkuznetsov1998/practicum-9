CREATE DATABASE IF NOT EXISTS reports;

-- Таблица для телеметрии
CREATE TABLE IF NOT EXISTS reports.telemetry (
    id Int32,
    user_id String,
    timestamp DateTime,
    device_id String,
    steps_count Int32,
    battery_level Int32,
    activity_level Float32,
    temperature Float32,
    created_at DateTime
) ENGINE = MergeTree()
ORDER BY (user_id, timestamp);

-- Таблица для CRM данных через Kafka
CREATE TABLE IF NOT EXISTS reports.crm (
    id Int32,
    user_id String,
    branch String,
    employee_type String,
    registration_date Date,
    last_activity_date Date,
    status String,
    created_at DateTime
) ENGINE = MergeTree()
ORDER BY (user_id);

-- Таблица Kafka Engine для приема данных из Kafka - используем сырые типы
CREATE TABLE IF NOT EXISTS reports.kafka_crm (
    id Int32,
    user_id String,
    branch String,
    employee_type String,
    registration_date Int32,
    last_activity_date Int32,
    status String,
    created_at Int64
) ENGINE = Kafka()
SETTINGS
    kafka_broker_list = 'kafka:9092',
    kafka_topic_list = 'crm_topic.public.crm_data',
    kafka_group_name = 'clickhouse_group',
    kafka_format = 'JSONEachRow',
    kafka_skip_broken_messages = 100;

-- Materialized View с правильной конвертацией дат
CREATE MATERIALIZED VIEW IF NOT EXISTS reports.crm_consumer TO reports.crm AS
SELECT 
    id,
    user_id,
    branch,
    employee_type,
    toDate(registration_date) as registration_date, -- Конвертация дней в Date
    toDate(last_activity_date) as last_activity_date, -- Конвертация дней в Date
    status,
    fromUnixTimestamp64Micro(created_at) as created_at -- Конвертация микросекунд в DateTime
FROM reports.kafka_crm;

-- Витрина данных с агрегацией
CREATE TABLE IF NOT EXISTS reports.data_mart (
    user_id String,
    branch String,
    employee_type String,
    total_steps Int32,
    avg_activity_level Float32,
    avg_temperature Float32,
    last_activity_date Date,
    report_date Date DEFAULT today()
) ENGINE = MergeTree()
ORDER BY (user_id, report_date);

-- Materialized View для автоматического обновления витрины
CREATE MATERIALIZED VIEW IF NOT EXISTS reports.data_mart_consumer TO reports.data_mart AS
SELECT 
    t.user_id,
    c.branch,
    c.employee_type,
    sum(t.steps_count) as total_steps,
    avg(t.activity_level) as avg_activity_level,
    avg(t.temperature) as avg_temperature,
    max(toDate(t.timestamp)) as last_activity_date
FROM reports.telemetry t
JOIN reports.crm c ON t.user_id = c.user_id
GROUP BY t.user_id, c.branch, c.employee_type;