# Makefile for Order Service

.PHONY: up build send-test send-test-docker down create-topic help

# Start all services (без продюсера)
up:
	docker-compose up -d postgres zookeeper kafka order-service

# Build all services
build:
	docker-compose build

# Send test messages from host machine
send-test:
	KAFKA_BROKER=localhost:9093 go run cmd/producer/main.go -n 3

# Send test messages from Docker container
send-test-docker:
	docker-compose run --rm kafka-producer

# Stop all services
down:
	docker-compose down

# Create Kafka topic
create-topic:
	docker-compose exec kafka kafka-topics --create \
		--topic orders \
		--partitions 1 \
		--replication-factor 1 \
		--bootstrap-server kafka:9092


# Test cache performance
test-cache:
	@echo "Testing cache performance..."
	@echo "Sending test data..."
	@docker-compose run --rm kafka-producer > /dev/null 2>&1
	@sleep 5
	@ORDER_UID=$$(docker-compose exec postgres psql -U postgres -d order_service -t -c "SELECT order_uid FROM orders ORDER BY date_created DESC LIMIT 1;" 2>/dev/null | tr -d '[:space:]') && \
	if [ -z "$$ORDER_UID" ]; then \
		echo "Ошибка: не удалось получить order_uid из базы данных"; \
		exit 1; \
	fi && \
	echo "Тестируем заказ: $$ORDER_UID" && \
	echo "" && \
	echo "=== ПЕРВЫЙ ЗАПРОС (БД) ===" && \
	curl -s -w "Время: %{time_total} сек\nHTTP код: %{http_code}\n" -o /dev/null "http://localhost:8080/api/order/$$ORDER_UID" || true && \
	echo "" && \
	echo "=== ВТОРОЙ ЗАПРОС (КЭШ) ===" && \
	curl -s -w "Время: %{time_total} сек\nHTTP код: %{http_code}\n" -o /dev/null "http://localhost:8080/api/order/$$ORDER_UID" || true && \
	echo "" && \
	echo "=== ТРЕТИЙ ЗАПРОС (КЭШ) ===" && \
	curl -s -w "Время: %{time_total} сек\nHTTP код: %{http_code}\n" -o /dev/null "http://localhost:8080/api/order/$$ORDER_UID" || true

# Clean database (remove all data)
clean-db:
	@echo "Cleaning database..."
	docker-compose exec postgres psql -U postgres -d order_service -c "\
	TRUNCATE TABLE items, payments, deliveries, orders RESTART IDENTITY CASCADE;"
	@echo "Database cleaned successfully!"

# Show help
help:
	@echo "Available commands:"
	@echo "  make up               - Start main services"
	@echo "  make build            - Build all services"
	@echo "  make send-test        - Send test messages from host"
	@echo "  make send-test-docker - Send test messages from Docker"
	@echo "  make down             - Stop all services"
	@echo "  make create-topic     - Create Kafka topic"
	@echo "  make help             - Show this help"

.DEFAULT_GOAL := help
