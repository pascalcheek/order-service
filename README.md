# Order Service (Go + PostgreSQL + Kafka-compatible Redpanda)

Демонстрационный микросервис:
- Консюмит заказы из Kafka (Redpanda), валидирует и сохраняет в PostgreSQL.
- Кеширует заказы в памяти для быстрых ответов.
- HTTP API: `GET /order/{order_uid}` возвращает JSON заказа.
- Простой веб-интерфейс на `/`.

## Быстрый старт (Docker)

```bash
docker compose up --build
```

Это поднимет:
- Postgres (с миграциями)
- Redpanda (совместим с Kafka API)
- order-service (порт `8081`)
- seed-продюсер (однократно отправит пример `model.json` в топик `orders`)

Зайдите в браузер: http://localhost:8081Введите: `b563feb7b2b84b6test` и получите заказ.

## Локальный запуск без Docker

1. Поднимите Postgres и Kafka/Redpanda вручную.
2. Выполните миграции из `migrations.sql`.
3. Запустите сервис:
   ```bash
   export HTTP_ADDR=:8081
   export DB_DSN="postgres://orderuser:orderpass@localhost:5432/ordersdb?sslmode=disable"
   export KAFKA_BROKERS="localhost:9092"
   export KAFKA_TOPIC="orders"
   export KAFKA_GROUP="order-consumer-group"
   export CACHE_PRELOAD_LIMIT=0
   go run ./cmd/order-service
   ```
4. Отправьте тест:
   ```bash
   go run ./cmd/seed
   ```

## API

- `GET /order/{order_uid}` — JSON заказа или 404.
- `GET /healthz` — healthcheck.
- `/` — простая HTML страница.

## Особенности надежности

- Обработка каждого сообщения — транзакция в БД.
- Идемпотентность: `orders`, `delivery`, `payment` — upsert; `items` — полная замена в транзакции.
- Коммит смещений Kafka происходит **только после** успешной записи в БД и обновления кеша.
- При старте сервис загружает все заказы в кеш (или ограниченное кол-вом через `CACHE_PRELOAD_LIMIT`).

## Структура

```
cmd/
  order-service/
    main.go
  seed/
    main.go
internal/
  api.go
  cache.go
  db.go
  kafka.go
  model.go
  ui.go
static/
  index.html
migrations.sql
model.json
docker-compose.yml
Dockerfile
Dockerfile.seed
```
