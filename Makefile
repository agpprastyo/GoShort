# Load environment variables from .env
ifneq (,$(wildcard .env.local))
  include .env.local
  export
endif

DB_URL := postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)

# Default target shows usage
.PHONY: help
help:
	@echo "Usage:"
	@echo "  make migrate-up           # Run all migrations"
	@echo "  make migrate-down         # Roll back all migrations"
	@echo "  make migrate-down-one     # Roll back the last migration"
	@echo "  make migrate-force V=X    # Force to version X"
	@echo "  make migrate-version      # Show current migration version"
	@echo "  make migrate-drop         # Drop all migrations"
	@echo "  make migrate-create N=X   # Create a new migration named X"
	@echo "  make build-backend        # Build the Go backend"
	@echo "  make run-backend          # Run the Go backend"
	@echo "  make build-frontend       # Build the React frontend"
	@echo "  make serve-frontend       # Serve the React frontend"
	@echo "  make sqlc-generate        # Generate SQL code with sqlc"

.PHONY: migrate-up
migrate-up:
	migrate -database $(DB_URL) -path db/migrations up

.PHONY: migrate-down
migrate-down:
	migrate -database $(DB_URL) -path db/migrations down

.PHONY: migrate-down-one
migrate-down-one:
	migrate -database $(DB_URL) -path db/migrations down 1

.PHONY: migrate-force
migrate-force:
	@if [ -z "$(V)" ]; then \
		echo "Error: Version required. Use make migrate-force V=version_number"; \
	else \
		migrate -database $(DB_URL) -path db/migrations force $(V); \
	fi

.PHONY: migrate-version
migrate-version:
	migrate -database $(DB_URL) -path db/migrations version

.PHONY: migrate-drop
migrate-drop:
	migrate -database $(DB_URL) -path db/migrations drop

.PHONY: migrate-create
migrate-create:
	@if [ -z "$(N)" ]; then \
		echo "Error: Migration name required. Use make migrate-create N=migration_name"; \
	else \
		migrate create -ext sql -dir db/migrations -seq $(N); \
	fi

.PHONY: build-backend
build-backend:
	go build -o goshort ./cmd/app

.PHONY: run-backend
run-backend:
	./goshort

.PHONY: build-frontend
build-frontend:
	cd web && npm install && npm run build

.PHONY: serve-frontend
serve-frontend:
	cd web && npm run preview

.PHONY: sqlc-generate
sqlc-generate:
	sqlc generate -f db/sqlc.yaml