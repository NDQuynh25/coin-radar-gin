.PHONY: dev-up dev-down migrate-up migrate-down sqlc-gen tidy build run-api run-ingestor run-bot run-aggregator

dev-up:
	docker compose up -d

dev-down:
	docker compose down

migrate-up:
	go run ./cmd/migrate up

migrate-down:
	@echo "Down migrations are intentionally not automated. Restore from backup or add an explicit rollback command."

sqlc-gen:
	sqlc generate

tidy:
	go mod tidy

build:
	@echo "Building binaries..."
	go build -o bin/api cmd/api/main.go
	go build -o bin/ingestor cmd/ingestor/main.go
	go build -o bin/aggregator cmd/aggregator/main.go
	go build -o bin/bot cmd/bot/main.go
	@echo "Binaries built in bin/"

run-api:
	go run cmd/api/main.go

run-ingestor:
	go run cmd/ingestor/main.go

run-bot:
	go run cmd/bot/main.go

run-aggregator:
	go run cmd/aggregator/main.go
