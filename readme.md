## Запуск
Для запуска debezium надо дать права на файл инициализации

```bash
chmod +x debezium/setup-connector.sh

docker compose up -d --build
```

Так же зайти в keycloak и руками пробросить секреты, они не экспортируются

## Файлы для сдачи задания:

Диаграмма: 

![diagram](diagram.png)

[Код сервиса Авторизации](./bionicpro-auth/main.go)

[Код ручки получения отчетов](./bionicpro-auth/report_handler.go)

[Фронтенд, который ходит в новую Авторизацию](./frontend/src/components/ReportPage.tsx)

[Экспорт реалмов из keycloak](./keycloak/realm-export-full.json)

[DAG для Airflow](./airflow/dags/telemetry_to_clickhouse.py)

[Конфиг nginx](./nginx/nginx.conf)

[Инициализация debezium](./debezium/setup-connector.sh)

[Инициализация Clickhouse](./init/clickhouse_init.sql)