include .env

# ==== Protocol Buffer Generation ====

.PHONY: generate-proto
generate-proto::
	buf generate --template buf.gen.yaml

.PHONY: update-proto-deps
update-proto-deps::
	buf dep update

# ==== Generator ====
ADDR  ?= localhost:50051
COUNT ?= 10
DELAY ?= 0


# ===== Tools ====

# Generates and sends audit events to the teals-server via gRPC.
# Signing is optional — omit KEY and KID to send unsigned events.
# Usage: make run-generator
#        make run-generator COUNT=50
#        make run-generator COUNT=100 ADDR=localhost:9090 DELAY=200
#        make run-generator KEY=<b64-private-key> KID=<thumbprint>
.PHONY: run-generator

run-generator:
	@echo "Running generator: $(COUNT) events → $(ADDR)..."
	go run ./tools/generator/cmd \
		--count=$(COUNT) \
		--addr=$(ADDR) \
		--delay=$(DELAY) \
		$(if $(KEY),--key=$(KEY)) \
		$(if $(KID),--kid=$(KID))

.PHONY: run-keygen-tool
 
run-keygen-tool:
	@go run ./tools/keygen

.PHONY: run-verify
# Verifies an inclusion or consistency proof against the ledger.
#
# Inclusion mode (default):
#   make run-verify EVENT_ID=<uuid> PAYLOAD_FILE=path/to/event.json LEDGER_SIZE=<n> TRUSTED_ROOT=<base64>
#
# Consistency mode:
#   make run-verify MODE=consistency FROM_SIZE=<n> TO_SIZE=<m> OLD_ROOT=<base64> NEW_ROOT=<base64>
MODE ?= inclusion
EVENT_ID ?=
PAYLOAD_FILE ?=
LEDGER_SIZE ?=
TRUSTED_ROOT ?=
FROM_SIZE ?=
TO_SIZE ?=
OLD_ROOT ?=
NEW_ROOT ?=

run-verify:
	@if [ "$(MODE)" = "consistency" ]; then \
		if [ -z "$(FROM_SIZE)" ] || [ -z "$(TO_SIZE)" ] || [ -z "$(OLD_ROOT)" ] || [ -z "$(NEW_ROOT)" ]; then \
			echo "Error: FROM_SIZE, TO_SIZE, OLD_ROOT and NEW_ROOT are required for consistency mode."; \
			echo "Usage: make run-verify MODE=consistency FROM_SIZE=<n> TO_SIZE=<m> OLD_ROOT=<base64> NEW_ROOT=<base64>"; \
			exit 1; \
		fi; \
		go run ./tools/verifier \
			--mode=consistency \
			--from-size=$(FROM_SIZE) \
			--to-size=$(TO_SIZE) \
			--old-root=$(OLD_ROOT) \
			--new-root=$(NEW_ROOT) \
			--addr=$(ADDR); \
	else \
		if [ -z "$(EVENT_ID)" ] || [ -z "$(PAYLOAD_FILE)" ]; then \
			echo "Error: EVENT_ID and PAYLOAD_FILE are required for inclusion mode."; \
			echo "Usage: make run-verify EVENT_ID=<uuid> PAYLOAD_FILE=path/to/event.json"; \
			exit 1; \
		fi; \
		go run ./tools/verifier \
			--mode=inclusion \
			--event-id=$(EVENT_ID) \
			--payload-file=$(PAYLOAD_FILE) \
			$(if $(LEDGER_SIZE),--ledger-size=$(LEDGER_SIZE)) \
			$(if $(TRUSTED_ROOT),--trusted-root=$(TRUSTED_ROOT)) \
			--addr=$(ADDR); \
	fi


# ==== TEALS Database Migrations ====
DATABASE_URL ?= "postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:$(POSTGRES_PORT)/$(POSTGRES_DB)"
TEALS_MIGRATIONS_PATH ?= services/teals/internal/infrastructure/repository/sql/migrations

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

# === Master KEK Rotation ===
.PHONY: gen-master-kek
gen-master-kek:
	@openssl rand -base64 32

# ==== Testing ====

# Fast unit tests only (service + pkg)
test:
	go test ./...

# Infrastructure integration tests only
test-integration:
	go test -tags=integration ./...
