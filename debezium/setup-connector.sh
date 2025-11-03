#!/bin/bash

echo "Waiting for Debezium to start..."
until curl -f -H "Accept:application/json" http://debezium:8083/connectors; do
  echo "Waiting for Debezium to start..."
  sleep 5
done

echo "Debezium started. Registering PostgreSQL connector..."

# Регистрация Debezium PostgreSQL connector
curl -i -X POST -H "Accept:application/json" -H "Content-Type:application/json" \
http://debezium:8083/connectors/ -d @- <<'EOF'
{
  "name": "crm-connector",
  "config": {
    "connector.class": "io.debezium.connector.postgresql.PostgresConnector",
    "database.hostname": "postgres_crm",
    "database.port": "5432",
    "database.user": "crm_user",
    "database.password": "crm_password",
    "database.dbname": "crm",
    "database.server.name": "crm_server",
    "table.include.list": "public.crm_data",
    "topic.prefix": "crm_topic",
    "value.converter": "org.apache.kafka.connect.json.JsonConverter",
    "value.converter.schemas.enable": "false",
    "key.converter": "org.apache.kafka.connect.json.JsonConverter",
    "key.converter.schemas.enable": "false",
    "transforms": "unwrap",
    "transforms.unwrap.type": "io.debezium.transforms.ExtractNewRecordState",
    "transforms.unwrap.drop.tombstones": "false",
    "plugin.name": "pgoutput"
  }
}
EOF

echo "Connector registered successfully."

# Проверка статуса коннектора
echo "Checking connector status..."
curl -H "Accept:application/json" http://localhost:8083/connectors/crm-connector/status