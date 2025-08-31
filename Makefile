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