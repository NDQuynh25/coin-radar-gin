.PHONY: dev-up dev-down sqlc-gen prisma-gen prisma-push tidy build run-api run-ingestor run-bot run-aggregator

DB_URL=postgres://postgres:postgres_password@localhost:5432/coin_radar?sslmode=disable

dev-up:
	docker compose up -d

dev-down:
	docker compose down

sqlc-gen:
	sqlc generate

prisma-gen:
	go run github.com/steebchen/prisma-client-go generate

prisma-push:
	go run github.com/steebchen/prisma-client-go db push

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
