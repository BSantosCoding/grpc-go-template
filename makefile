# /home/bsant/testing/go-grpc-template/makefile.mk

.PHONY: proto tidy build-server build-client run-server run-client all clean \
        db-connect db-create db-setup db-clean-database \
        migrate-create migrate-up migrate-down-last migrate-down-all migrate-force migrate-status \
        dev

.ONESHELL: # Ensures recipes run in a single shell instance

SHELL :=/bin/bash
MODULE_PATH=github.com/BSantosCoding/grpc-go-example
PROTO_DIR=proto
GEN_DIR=gen
BIN_DIR=bin
ENV_FILE=.env

# --- Database Configuration ---
# Provides FALLBACK defaults if not set in .env or environment. '?=' assigns only if not already set.
PGHOST ?= localhost
PGPORT ?= 5432
PGUSER ?= postgres
PGPASSWORD ?= postgres # Default password, override in .env
DBNAME ?= users_db
PG_DEFAULT_DB ?= postgres # Database to connect to for creating DBNAME

# Connection args for psql command
PSQL_CONN_ARGS = -h $(PGHOST) -p $(PGPORT) -U $(PGUSER)

# --- Migration Configuration ---
MIGRATE_CMD = migrate # Assumes 'migrate' command is in PATH
MIGRATIONS_DIR = db/migrations
MIGRATIONS_PATH = $(MIGRATIONS_DIR) # Use relative path for migrate tool

# Helper to export .env variables for specific commands
# Uses define/endef for better script block handling.
define export-env
if [ -f "$(ENV_FILE)" ]; then
    echo "Exporting environment variables from $(ENV_FILE)..."
    # Use grep to filter comments/empty lines, sed to format for export, then evaluate
    eval $$(grep -v '^#' "$(ENV_FILE)" | grep -v '^$$' | sed -e 's/^/export /')
    _export_exit_code=$$?
    if [ $$_export_exit_code -ne 0 ]; then
         echo "Error: Exporting variables from $(ENV_FILE) failed with exit code $$_export_exit_code."
         exit 1
    fi
else
    echo "Warning: $(ENV_FILE) not found. Using existing environment variables or defaults."
fi
endef

# --- Protobuf Generation ---
proto:
	@mkdir -p $(GEN_DIR)/user/v1
	protoc --proto_path=$(PROTO_DIR) \
	       --go_out=$(GEN_DIR) --go_opt=paths=source_relative \
	       --go-grpc_out=$(GEN_DIR) --go-grpc_opt=paths=source_relative \
	       $(PROTO_DIR)/user/v1/user.proto
	@echo "Protobuf code generated for User service."

# --- Go Commands ---
tidy:
	go mod tidy
	@echo "Dependencies tidied."

build-server: tidy
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/server ./cmd/server
	@echo "Server built."

build-client: tidy
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/client ./cmd/client
	@echo "Client built."

# Go app loads .env itself via godotenv
run-server: build-server
	@echo "Starting server (will load config from .env or environment)..."
	./$(BIN_DIR)/server

run-client: build-client
	@echo "Running client..."
	./$(BIN_DIR)/client $(filter-out $@,$(MAKECMDGOALS))

all: proto build-server build-client

clean:
	@rm -rf $(GEN_DIR) $(BIN_DIR) tmp air_errors.log
	@echo "Cleaned generated files and binaries."


# --- Database Commands ---

# Connect to the target database using psql
db-connect:
	@echo "Connecting to database '$(DBNAME)' on $(PGHOST):$(PGPORT) as user '$(PGUSER)'..."
	@$(export-env)
	PGPASSWORD=$$PGPASSWORD psql -h $$PGHOST -p $$PGPORT -U $$PGUSER -d $$DBNAME

# Create the database if it doesn't exist
db-create:
	@echo "Attempting to create database '$(DBNAME)' on $(PGHOST):$(PGPORT)..."
	@$(export-env)
	# Connect to default DB to create the target DB
	PGPASSWORD=$$PGPASSWORD psql -h $$PGHOST -p $$PGPORT -U $$PGUSER -d $(PG_DEFAULT_DB) -c "CREATE DATABASE $$DBNAME" || echo "Database '$$DBNAME' likely already exists or creation failed."

# Apply all pending 'up' migrations
migrate-up:
	@echo "Applying UP migrations from $(MIGRATIONS_DIR)..."
	@$(export-env)
	if [ -z "$$PGHOST" ] || [ -z "$$PGPORT" ] || [ -z "$$PGUSER" ] || [ -z "$$PGPASSWORD" ] || [ -z "$$DBNAME" ]; then
	    echo "Error: One or more required database environment variables are not set. Cannot proceed."
	    exit 1
	fi
	# Construct URL within the shell after exporting env vars
	_DATABASE_URL="postgresql://$$PGUSER:$$PGPASSWORD@$$PGHOST:$$PGPORT/$$DBNAME?sslmode=disable"
	echo "Using DB: postgresql://$$PGUSER:****@$$PGHOST:$$PGPORT/$$DBNAME"
	$(MIGRATE_CMD) -database "$$_DATABASE_URL" -path $(MIGRATIONS_PATH) up

# Show current migration status/version
migrate-status:
	@echo "Checking migration status..."
	@$(export-env)
	if [ -z "$$PGHOST" ] || [ -z "$$PGPORT" ] || [ -z "$$PGUSER" ] || [ -z "$$PGPASSWORD" ] || [ -z "$$DBNAME" ]; then
	    echo "Error: One or more required database environment variables are not set. Cannot proceed."
	    exit 1
	fi
	_DATABASE_URL="postgresql://$$PGUSER:$$PGPASSWORD@$$PGHOST:$$PGPORT/$$DBNAME?sslmode=disable"
	echo "Using DB: postgresql://$$PGUSER:****@$$PGHOST:$$PGPORT/$$DBNAME"
	echo "Path: $(MIGRATIONS_PATH)"
	$(MIGRATE_CMD) -database "$$_DATABASE_URL" -path $(MIGRATIONS_PATH) version

# Roll back the last 'down' migration
migrate-down-last:
	@echo "Applying last DOWN migration from $(MIGRATIONS_DIR)..."
	@$(export-env)
	if [ -z "$$PGHOST" ] || [ -z "$$PGPORT" ] || [ -z "$$PGUSER" ] || [ -z "$$PGPASSWORD" ] || [ -z "$$DBNAME" ]; then
	    echo "Error: One or more required database environment variables are not set. Cannot proceed."
	    exit 1
	fi
	_DATABASE_URL="postgresql://$$PGUSER:$$PGPASSWORD@$$PGHOST:$$PGPORT/$$DBNAME?sslmode=disable"
	echo "Using DB: postgresql://$$PGUSER:****@$$PGHOST:$$PGPORT/$$DBNAME"
	$(MIGRATE_CMD) -database "$$_DATABASE_URL" -path $(MIGRATIONS_PATH) down 1

# Roll back all 'down' migrations (use with caution!)
migrate-down-all:
	@echo "WARNING: This will apply ALL DOWN migrations from $(MIGRATIONS_DIR)."
	@read -p "Are you sure? (y/N) " -r
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then
		echo "Applying all DOWN migrations..."
		$(export-env)
		if [ -z "$$PGHOST" ] || [ -z "$$PGPORT" ] || [ -z "$$PGUSER" ] || [ -z "$$PGPASSWORD" ] || [ -z "$$DBNAME" ]; then
		    echo "Error: One or more required database environment variables are not set. Cannot proceed."
		    exit 1
		fi
		_DATABASE_URL="postgresql://$$PGUSER:$$PGPASSWORD@$$PGHOST:$$PGPORT/$$DBNAME?sslmode=disable"
		echo "Using DB: postgresql://$$PGUSER:****@$$PGHOST:$$PGPORT/$$DBNAME"
		$(MIGRATE_CMD) -database "$$_DATABASE_URL" -path $(MIGRATIONS_PATH) down -all
		echo "All DOWN migrations applied."
	else
		echo "Aborted."
	fi

# Force set migration version (useful for fixing dirty states, use with caution!)
migrate-force:
	@if [ -z "$(VERSION)" ]; then echo "Error: VERSION variable must be set. Usage: make migrate-force VERSION=<version_number>"; exit 1; fi
	@echo "WARNING: Forcing migration version to $(VERSION) in database. This does NOT run migrations."
	@read -p "Are you sure? (y/N) " -r
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then
		echo "Forcing version..."
		$(export-env)
		if [ -z "$$PGHOST" ] || [ -z "$$PGPORT" ] || [ -z "$$PGUSER" ] || [ -z "$$PGPASSWORD" ] || [ -z "$$DBNAME" ]; then
		    echo "Error: One or more required database environment variables are not set. Cannot proceed."
		    exit 1
		fi
		_DATABASE_URL="postgresql://$$PGUSER:$$PGPASSWORD@$$PGHOST:$$PGPORT/$$DBNAME?sslmode=disable"
		echo "Using DB: postgresql://$$PGUSER:****@$$PGHOST:$$PGPORT/$$DBNAME"
		$(MIGRATE_CMD) -database "$$_DATABASE_URL" -path $(MIGRATIONS_PATH) force $(VERSION)
		echo "Version forced to $(VERSION)."
	else
		echo "Aborted."
	fi

# --- Combined Setup ---
# Creates database and applies all migrations
db-setup: db-create migrate-up
	@echo "Database setup complete for '$(DBNAME)'."

# --- Database Cleanup ---
# Drops the entire database (use with extreme caution!)
db-clean-database:
	@echo "WARNING: This will permanently drop the database '$(DBNAME)' on $(PGHOST):$(PGPORT)."
	@read -p "Are you sure? (y/N) " -r
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then
		echo "Dropping database '$$DBNAME'..."
		$(export-env)
		if [ -z "$$PGHOST" ] || [ -z "$$PGPORT" ] || [ -z "$$PGUSER" ] || [ -z "$$PGPASSWORD" ] || [ -z "$$DBNAME" ]; then
		    echo "Error: One or more required database environment variables are not set. Cannot proceed."
		    exit 1
		fi
		# Connect to default DB to drop the target DB
		PGPASSWORD=$$PGPASSWORD psql -h $$PGHOST -p $$PGPORT -U $$PGUSER -d $(PG_DEFAULT_DB) -c "DROP DATABASE IF EXISTS $$DBNAME;"
		echo "Database '$$DBNAME' dropped."
	else
		echo "Aborted."
	fi

# --- Development ---
# Run server with Air hot-reloading
dev:
	@echo "Starting server with Air hot-reload (will load config from .env or environment)..."
	@air

# Default target
default: all

# Allow passing arguments to run-client and migrate-create
# Example: make run-client NAME=Bob EMAIL=bob@example.com
# Example: make migrate-create NAME=add_new_table
%:
	@: # No-op target to prevent errors for undefined targets used for arguments
