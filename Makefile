.PHONY: run build docker-up docker-down

run:
	go run ./cmd/bot/...

build:
	go build -o bin/bot ./cmd/bot/...

docker-up:
	docker compose up -d postgres redis

docker-down:
	docker compose down

migrate:
	go run ./cmd/bot/... migrate