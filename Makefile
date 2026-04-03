# ==== Protocol Buffer Generation ====

.PHONY: generate_proto update_proto_deps

generate_proto::
	buf generate --template buf.gen.yaml

update_proto_deps::
	buf dep update

# ==== Generator ====
ADDR  ?= localhost:50051
COUNT ?= 10
DELAY ?= 0
.PHONY: run-generator
# Generates and sends audit events to the teals-server via gRPC.
# Usage: make run-generator
#        make run-generator COUNT=50
#        make run-generator COUNT=100 ADDR=localhost:9090 DELAY=200
run-generator:
	@echo "Running generator: $(COUNT) events → $(ADDR)..."
	go run ./services/generator/cmd \
		--count=$(COUNT) \
		--addr=$(ADDR) \
		--delay=$(DELAY)

# ==== TEALS Database Migrations ====

include .env
DATABASE_URL ?= "postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:$(POSTGRES_PORT)/$(POSTGRES_DB)"
TEALS_MIGRATIONS_PATH ?= services/teals-server/internal/infrastructure/repository/sql/migrations

.PHONY: migrate-create-teals migrate-up-teals migrate-down-teals

# Creates a new migration file for the teals server.
# Usage: make migrate-create-teals name=create_users_table
migrate-create-teals:
	@if [ -z "$(name)" ]; then \
		echo "Error: 'name' is a required variable."; \
		echo "Usage: make migrate-create-teals name=<migration_name>"; \
		exit 1; \
	fi
	@echo "Creating teals migration: $(name)..."
	migrate create -ext sql -dir $(TEALS_MIGRATIONS_PATH) $(name)

# Applies all pending up migrations for the teals server.
# Usage: make migrate-up-teals
migrate-up-teals:
	@echo "Applying teals migrations..."
	migrate -database "$(DATABASE_URL)?sslmode=disable" -path $(TEALS_MIGRATIONS_PATH) up

# Reverts the last applied migration for the teals server.
# Usage: make migrate-down-teals
migrate-down-teals:
	@echo "Reverting last teals migration..."
	migrate -database "$(DATABASE_URL)?sslmode=disable" -path $(TEALS_MIGRATIONS_PATH) down 1
