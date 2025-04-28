# /home/bsant/testing/go-grpc-template/makefile.mk

.PHONY: proto tidy build-server build-client run-server run-client all clean \
        db-connect db-create db-setup db-clean-database \
        migrate-create migrate-up migrate-down-last migrate-down-all migrate-force migrate-status \
        docker-build docker-up docker-down docker-logs docker-prune docker-build-nocache docker-dev

.ONESHELL: # Ensures recipes run in a single shell instance

SHELL := /bin/bash
MODULE_PATH=github.com/BSantosCoding/grpc-go-example
PROTO_DIR=proto
GEN_DIR=gen
BIN_DIR=bin
ENV_FILE=.env

# --- Database Configuration (Used primarily for HOST commands like migrate/psql) ---
PGHOST ?= localhost # Default to localhost as DB port is mapped
PGPORT ?= 5432      # Default to 5432 as DB port is mapped
PGUSER ?= postgres
PGPASSWORD ?= postgres
DBNAME ?= users_db
PG_DEFAULT_DB ?= postgres

# Connection args for psql command (connecting from host)
PSQL_CONN_ARGS = -h $(PGHOST) -p $(PGPORT) -U $(PGUSER)

# --- Migration Configuration (Run from HOST against container DB) ---
MIGRATE_CMD = migrate
MIGRATIONS_DIR = db/migrations
MIGRATIONS_PATH = $(MIGRATIONS_DIR)

# Helper to export .env variables for specific commands
define export-env
if [ -f "$(ENV_FILE)" ]; then
    echo "Exporting environment variables from $(ENV_FILE)..."
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

# --- Go Commands (Mainly for local checks/client) ---
tidy:
	go mod tidy
	@echo "Dependencies tidied."

build-server: tidy
	@echo "Building server locally (use 'make docker-build' for container image)..."
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/server ./cmd/server

run-server: build-server
	@echo "Starting server locally (use 'make docker-up' to run in container)..."
	./$(BIN_DIR)/server

build-client: tidy
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/client ./cmd/client
	@echo "Client built."

run-client: build-client
	@echo "Running client (connecting to localhost:50051)..."
	./$(BIN_DIR)/client $(filter-out $@,$(MAKECMDGOALS))

all: proto build-client # Removed build-server from default 'all'

clean:
	@rm -rf $(GEN_DIR) $(BIN_DIR) tmp air_errors.log
	@echo "Cleaned generated files and binaries."


# --- Database Commands (Run from HOST against container DB via mapped port) ---

# Connect to the containerized database using psql
db-connect:
	@echo "Connecting to containerized database '$(DBNAME)' on $(PGHOST):$(PGPORT) as user '$(PGUSER)'..."
	@$(export-env)
	PGPASSWORD=$$PGPASSWORD psql -h $$PGHOST -p $$PGPORT -U $$PGUSER -d $$DBNAME

# Create database (Handled by docker compose on first run via POSTGRES_DB)
db-create:
	@echo "NOTE: Database creation is typically handled by docker-compose using POSTGRES_DB."
	@echo "Attempting to ensure database '$(DBNAME)' exists on $(PGHOST):$(PGPORT)..."
	@$(export-env)
	PGPASSWORD=$$PGPASSWORD psql -h $$PGHOST -p $$PGPORT -U $$PGUSER -d $(PG_DEFAULT_DB) -c "CREATE DATABASE $$DBNAME" || echo "Database '$$DBNAME' likely already exists or creation failed."

# Apply all pending 'up' migrations (Run from host against container DB)
migrate-up:
	@echo "Applying UP migrations from $(MIGRATIONS_DIR) to container DB..."
	@$(export-env)
	if [ -z "$$PGHOST" ] || [ -z "$$PGPORT" ] || [ -z "$$PGUSER" ] || [ -z "$$PGPASSWORD" ] || [ -z "$$DBNAME" ]; then
	    echo "Error: One or more required database environment variables are not set. Cannot proceed."
	    exit 1
	fi
	_DATABASE_URL="postgresql://$$PGUSER:$$PGPASSWORD@$$PGHOST:$$PGPORT/$$DBNAME?sslmode=disable"
	echo "Using DB URL (for host migrate): postgresql://$$PGUSER:****@$$PGHOST:$$PGPORT/$$DBNAME"
	$(MIGRATE_CMD) -database "$$_DATABASE_URL" -path $(MIGRATIONS_PATH) up

# Show current migration status (Run from host against container DB)
migrate-status:
	@echo "Checking migration status on container DB..."
	@$(export-env)
	if [ -z "$$PGHOST" ] || [ -z "$$PGPORT" ] || [ -z "$$PGUSER" ] || [ -z "$$PGPASSWORD" ] || [ -z "$$DBNAME" ]; then
	    echo "Error: One or more required database environment variables are not set. Cannot proceed."
	    exit 1
	fi
	_DATABASE_URL="postgresql://$$PGUSER:$$PGPASSWORD@$$PGHOST:$$PGPORT/$$DBNAME?sslmode=disable"
	echo "Using DB URL (for host migrate): postgresql://$$PGUSER:****@$$PGHOST:$$PGPORT/$$DBNAME"
	echo "Path: $(MIGRATIONS_PATH)"
	$(MIGRATE_CMD) -database "$$_DATABASE_URL" -path $(MIGRATIONS_PATH) version

# Roll back the last 'down' migration (Run from host against container DB)
migrate-down-last:
	@echo "Applying last DOWN migration from $(MIGRATIONS_DIR) to container DB..."
	@$(export-env)
	# ... (rest of migrate-down-last logic using _DATABASE_URL) ...
	_DATABASE_URL="postgresql://$$PGUSER:$$PGPASSWORD@$$PGHOST:$$PGPORT/$$DBNAME?sslmode=disable"
	echo "Using DB URL (for host migrate): postgresql://$$PGUSER:****@$$PGHOST:$$PGPORT/$$DBNAME"
	$(MIGRATE_CMD) -database "$$_DATABASE_URL" -path $(MIGRATIONS_PATH) down 1

# Roll back all 'down' migrations (Run from host against container DB)
migrate-down-all:
	@echo "WARNING: This will apply ALL DOWN migrations from $(MIGRATIONS_DIR) to container DB."
	@read -p "Are you sure? (y/N) " -r
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then
		echo "Applying all DOWN migrations..."
		$(export-env)
		# ... (rest of migrate-down-all logic using _DATABASE_URL) ...
		_DATABASE_URL="postgresql://$$PGUSER:$$PGPASSWORD@$$PGHOST:$$PGPORT/$$DBNAME?sslmode=disable"
		echo "Using DB URL (for host migrate): postgresql://$$PGUSER:****@$$PGHOST:$$PGPORT/$$DBNAME"
		$(MIGRATE_CMD) -database "$$_DATABASE_URL" -path $(MIGRATIONS_PATH) down -all
		echo "All DOWN migrations applied."
	else
		echo "Aborted."
	fi

# Force set migration version (Run from host against container DB)
migrate-force:
	@if [ -z "$(VERSION)" ]; then echo "Error: VERSION variable must be set. Usage: make migrate-force VERSION=<version_number>"; exit 1; fi
	@echo "WARNING: Forcing migration version to $(VERSION) in container database."
	@read -p "Are you sure? (y/N) " -r
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then
		echo "Forcing version..."
		$(export-env)
		# ... (rest of migrate-force logic using _DATABASE_URL) ...
		_DATABASE_URL="postgresql://$$PGUSER:$$PGPASSWORD@$$PGHOST:$$PGPORT/$$DBNAME?sslmode=disable"
		echo "Using DB URL (for host migrate): postgresql://$$PGUSER:****@$$PGHOST:$$PGPORT/$$DBNAME"
		$(MIGRATE_CMD) -database "$$_DATABASE_URL" -path $(MIGRATIONS_PATH) force $(VERSION)
		echo "Version forced to $(VERSION)."
	else
		echo "Aborted."
	fi

# --- Combined Setup (Run from HOST against container DB) ---
# Creates database (if needed via compose) and applies all migrations
db-setup: migrate-up # db-create is handled by compose init
	@echo "Database migrations applied to container DB."

# --- Database Cleanup ---
# Drops the entire database (use with extreme caution!)
# This still works by connecting from host to mapped port
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
		PGPASSWORD=$$PGPASSWORD psql -h $$PGHOST -p $$PGPORT -U $$PGUSER -d $(PG_DEFAULT_DB) -c "DROP DATABASE IF EXISTS $$DBNAME;"
		echo "Database '$$DBNAME' dropped."
	else
		echo "Aborted."
	fi

# --- Docker Compose Commands ---
docker-build:
	@echo "Building Docker images using docker-compose..."
	docker compose build

docker-build-nocache:
	@echo "Building Docker images using docker-compose..."
	docker compose build --no-cache

docker-up:
	@echo "Starting services using docker-compose..."
	docker compose up -d # -d runs in detached mode

docker-down:
	@echo "Stopping services using docker-compose..."
	docker compose down

docker-down-volumes:
	@echo "Stopping services and removing volumes using docker-compose..."
	docker compose down -v # -v removes named volumes

docker-logs:
	@echo "Following logs for app service..."
	docker compose logs -f app

docker-logs-db:
	@echo "Following logs for db service..."
	docker compose logs -f db

docker-prune: docker-down-volumes
	@echo "Pruning unused Docker images, networks, and build cache..."
	docker system prune -af --volumes
	@echo "Docker system pruned."

docker-dev:
	@make docker-build
	@make docker-up
	@make migrate-up

# --- Testing ---
test: tidy
	@echo "Running tests..."
	@$(export-env) # Export env vars in case config loading relies on them
	go test ./... -v -cover -count=1 # Run tests in all subdirs, verbose, show coverage, disable caching

# Default target
default: all

# Allow passing arguments to run-client and migrate-create
%:
	@:
