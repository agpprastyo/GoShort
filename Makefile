# Database connection information
DB_URL := postgres://postgres:postgres@localhost:5432/goshort?sslmode=disable

# Default target shows usage
.PHONY: help
help:
	@echo "Usage:"
	@echo "PHONY: migrate-down:
migrate-down:
	migrate -database $(DB_URL) -path db/migrations down

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