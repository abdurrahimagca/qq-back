.PHONY: dev-up dev-up-d dev-build dev-down dev-logs dev-shell dev-restart build build-docker migrate-up migrate-down migrate-create migrate-force migrate-version db-shell db-reset sqlc-generate fmt test test-cover  clean mod-tidy mod-download fmt-check lint

# Development commands (Docker only)
dev-up:
	cd docker && docker compose up

dev-up-d:
	cd docker && docker compose up -d

dev-build:
	cd docker && docker compose build

dev-down:
	cd docker && docker compose down

dev-logs:
	cd docker && docker compose logs -f

dev-shell:
	cd docker && docker compose exec api sh

# Restart just the API service for quick reload
dev-restart:
	cd docker && docker compose restart api

# Build commands
build:
	go build -o bin/main cmd/main.go

build-docker:
	docker build -f docker/Dockerfile -t qq-back .

# Database migrations (via Docker)
migrate-up:
	cd docker && docker compose exec api migrate -path db/migrations -database "${DATABASE_URL}" up

migrate-down:
	cd docker && docker compose exec api migrate -path db/migrations -database "${DATABASE_URL}" down

migrate-create:
	@read -p "Enter migration name: " name; \
	cd docker && docker compose exec api migrate create -ext sql -dir db/migrations $$name

migrate-force:
	@read -p "Enter version to force: " version; \
	cd docker && docker compose exec api migrate -path db/migrations -database "${DATABASE_URL}" force $$version

migrate-version:
	cd docker && docker compose exec api migrate -path db/migrations -database "${DATABASE_URL}" version

# Database operations via Docker
db-shell:
	cd docker && docker compose exec postgres psql -U ${POSTGRES_USER} -d ${POSTGRES_DB}

db-reset:
	cd docker && docker compose exec postgres psql -U ${POSTGRES_USER} -d ${POSTGRES_DB} -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"

# Code generation via Docker
sqlc-generate:
	cd docker && docker compose exec api sqlc generate

# Formatting via Docker
fmt:
	cd docker && docker compose exec api go fmt ./...

# Testing via Docker
test:
	cd docker && docker compose exec api go test -v ./...

test-cover:
	cd docker && docker compose exec api go test -v -cover ./...

# Clean
clean:
	rm -rf bin/
	rm -f coverage.out
	docker system prune -f

# Go mod
mod-tidy:
	go mod tidy

mod-download:
	go mod download

# Code formatting (Docker versions)
fmt-check:
	cd docker && docker compose exec api sh -c 'if [ -n "$$(gofmt -l .)" ]; then echo "Go files need formatting"; exit 1; fi'

lint:
	cd docker && docker compose exec api golangci-lint run || echo "golangci-lint not available in container"
