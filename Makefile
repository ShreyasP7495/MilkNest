.PHONY: help run build tidy fmt vet test docker-up docker-down migrate-up migrate-down migrate-create

APP_NAME ?= milknest
PORT ?= 8080
DB_URL ?= postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)

help:
	@echo "Available targets:"
	@echo "  run             - run the API locally (requires .env)"
	@echo "  build           - build the binary into ./bin/$(APP_NAME)"
	@echo "  tidy            - go mod tidy"
	@echo "  fmt             - gofmt all source files"
	@echo "  vet             - go vet"
	@echo "  docker-up       - docker compose up -d"
	@echo "  docker-down     - docker compose down"
	@echo "  migrate-up      - run all migrations (requires golang-migrate CLI)"
	@echo "  migrate-down    - rollback last migration"
	@echo "  migrate-create  - create a new migration (NAME=add_foo)"

run:
	go run ./cmd/server

build:
	go build -o bin/$(APP_NAME) ./cmd/server

tidy:
	go mod tidy

fmt:
	gofmt -w .

vet:
	go vet ./...

docker-up:
	docker compose up -d

docker-down:
	docker compose down

migrate-up:
	migrate -path migrations -database "$(DB_URL)" up

migrate-down:
	migrate -path migrations -database "$(DB_URL)" down 1

migrate-create:
	migrate create -ext sql -dir migrations -seq $(NAME)
