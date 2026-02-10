# Golang Short Link Service

Микросервис для сокращения URL с использованием Go, PostgreSQL и Redis.

## Особенности

- **Надежность** - данные сохраняются в PostgreSQL
- **Кэширование** - Redis для быстрого доступа
- **Логирование** - структурированные логи с Zap
- **Мониторинг** - Prometheus метрики
- **Контейнеризация** - Docker и Docker Compose

## Архитектура

REST Handlers                     | Redis 
(internal/transport/rest/*.go)     (Кэш)

Service         │ Logger 
(бизнес-логика) │ (Zap) 

Repository │ PostgreSQL
(CRUD)

## API Документация

### 1. Создание короткой ссылки
Endpoint: POST `http://localhost:8080/oneLink`

Request:
Content-Type: application/json

{
  "url": "https://example.com/very/long/path/to/resource"
}

### 2. Редирект по короткой ссылке
Endpoint: GET `http://localhost:8080/oneLink/{short_code}`

### 3. Получение статистики
Endpoint: GET `http://localhost:8080/oneLink`
