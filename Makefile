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
	@set -a; . ./.env; set +a; \
	for f in db/migrations/*.sql; do \
		echo "Applying $$f"; \
		psql "$$DATABASE_URL" -v ON_ERROR_STOP=1 -f "$$f" || exit 1; \
	done