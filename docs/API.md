# API

### GET /health/alive

Проверка, что сервис запущен.

Ответ:
```json
{ "status": "im ok" }
```

### GET /health/ready

Проверка готовности сервиса принимать запросы.

Ответ:
```json
{ "status": "im ready" }
```