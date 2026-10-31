.PHONY: build run test tidy sqlc up down migrate

build:
	go build ./...

run:
	go run ./cmd/api

test:
	go test ./...

tidy:
	go mod tidy

sqlc:
	sqlc generate

up:
	docker compose up -d

down:
	docker compose down

migrate:
	@echo "Apply db/migrations/*.sql with your migration tool (e.g. goose)"