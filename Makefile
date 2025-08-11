# Load environment variables from .env
ifneq (,$(wildcard .env))
  include .env
  export
endif

DB_URL := postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)
# Default target shows usage
.PHONY: help
help:
	@echo "Usage:"
	@echo "PHONY: migrate-down:
migrate-down:
	migrate -database $(DB_URL) -path db/migrations down

.PHONY: migrate-down-one
migrate-down-one:
	migrate -database $(DB_URL) -path db/migrations down 1

.PHONY: migrate-force
migrate-force:
	@if [ -z "$(V)" ]; then \
		  make migrate-up         # Run all migrations"
	@echo "  make migrate-down       # Roll back migrations"
	@echo "  make migrate-force V=X  # Force to version X"
	@echo "  make migrate-version    # Show current version"
	@echo "  make migrate-drop       # Drop all migrations"
	@echo "  make migrate-create N=X # Create migration named X"

.PHONY: migrate-up
migrate-up:
	migrate -database $(DB_URL) -path db/migrations up

.echo "Error: Version required. Use make migrate-force V=version_number"; \
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