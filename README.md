# Go gRPC User Service Example

A sample project demonstrating a gRPC API in Go for managing users. This project features a structured layout, PostgreSQL integration, database migrations using `golang-migrate`, Docker support, configuration via `.env` files, dependency injection, input validation, and integration tests.

## Features

- gRPC API for User CRUD-like operations (Create, Get, List).
- PostgreSQL database persistence.
- Database migrations managed by `golang-migrate`.
- Configuration via `.env` files (`godotenv`).
- Dependency Injection using a simple container pattern.
- Input validation on API requests (server-side).
- Docker and Docker Compose for containerized application and database.
- Makefile (`makefile.mk`) for easy build, test, database, and Docker management tasks.
- Integration tests using `testify/suite`.
- Interactive gRPC client for manual testing.

## Prerequisites

Make sure you have the following installed on your system:

- **Go:** Version 1.21 or later.
- **Docker & Docker Compose:** For running the application and database in containers.
- **Make:** For using the helper commands in `makefile.mk`.
- **Protocol Buffer Compiler (`protoc`):** Version 3.x or later. (Installation Guide)
- **Go Plugins for Protobuf:**
  ```bash
  go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
  go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
  ```
- **`golang-migrate` CLI:** (Installation Guide)
- **`psql` (Optional):** PostgreSQL command-line client, useful for direct database access via `make -f makefile.mk db-connect`.
- **`air` (Optional):** For live reloading during local development (if not using Docker primarily). `go install github.com/cosmtrek/air@latest`

_Ensure `protoc`, `protoc-gen-go`, `protoc-gen-go-grpc`, `migrate`, and `air` (if used) binaries are available in your system's `PATH`._

## Setup

1.  **Clone the Repository:**

    ```bash
    git clone <your-repository-url>
    cd go-grpc-template # Or your project directory name
    ```

2.  **Configure Environment:**

    - Copy the example environment file:
      ```bash
      cp .env.example .env
      ```
    - Edit the `.env` file and set your desired configuration, especially `PGPASSWORD`.
    - **Important:** For the Docker setup, keep `PGHOST=localhost` and `PGPORT=5432` in `.env`. The Makefile commands (like `migrate-up`, `db-connect`) run from your host and connect to the database container via its mapped port. The application container (`app`) uses service discovery (`PGHOST=db`) configured in `docker-compose.yml`.

3.  **Docker Setup (Recommended Workflow):**

    - **Build Images:** Build the application Docker image.
      ```bash
      make -f makefile.mk docker-build
      ```
    - **Start Services:** Start the application and database containers in the background.
      ```bash
      make -f makefile.mk docker-up
      ```
      _(The database will be created automatically on the first run based on `.env` variables)_.
    - **Apply Migrations:** Run database migrations against the containerized database.
      ```bash
      make -f makefile.mk migrate-up
      ```

4.  **Local Setup (Alternative - Requires Local PostgreSQL):**
    - Ensure a local PostgreSQL server is running and accessible.
    - Update `.env` with connection details for your _local_ PostgreSQL instance.
    - **Setup Database & Run Migrations:**
      ```bash
      make -f makefile.mk db-setup
      ```
    - **Generate Protobuf Code:**
      ```bash
      make -f makefile.mk proto
      ```
    - **Install Go Dependencies:**
      ```bash
      make -f makefile.mk tidy
      ```

## Usage

_(Remember to use `make -f makefile.mk <target>` for all commands)_

**Running with Docker (Recommended):**

- **Start Services:** `make -f makefile.mk docker-up`
- **Stop Services:** `make -f makefile.mk docker-down`
- **Stop Services & Remove Data:** `make -f makefile.mk docker-down-volumes` (Deletes the database volume!)
- **View App Logs:** `make -f makefile.mk docker-logs` (Press `Ctrl+C` to stop)
- **View DB Logs:** `make -f makefile.mk docker-logs-db` (Press `Ctrl+C` to stop)
- **Rebuild Image:** `make -f makefile.mk docker-build`

**Running Locally (Alternative):**

- **Run with Hot-Reload (Air):** `make -f makefile.mk dev`
- **Run without Hot-Reload:** `make -f makefile.mk run-server`

**Running the Client:**

- Ensure the server (either local or Docker) is running.
- Run the interactive client:
  ```bash
  make -f makefile.mk run-client
  ```
  Follow the prompts in the client menu.

**Running Tests:**

- Ensure the database container is running (`make -f makefile.mk docker-up`).
- Run the integration tests:
  ```bash
  make -f makefile.mk test
  ```

**Database Migrations:**

- Ensure the database container is running (`make -f makefile.mk docker-up`).
- **Apply Pending Migrations:** `make -f makefile.mk migrate-up`
- **Rollback Last Migration:** `make -f makefile.mk migrate-down-last`
- **Check Migration Status:** `make -f makefile.mk migrate-status`
- **Create New Migration Files:**
  ```bash
  make -f makefile.mk migrate-create NAME=describe_your_change
  ```
  _(Edit the generated `.up.sql` and `.down.sql` files in `db/migrations/`)_

**Other Makefile Commands:**

- **Generate Protobuf Code:** `make -f makefile.mk proto`
- **Tidy Go Modules:** `make -f makefile.mk tidy`
- **Clean Generated Files:** `make -f makefile.mk clean`
- **Connect to DB via psql:** `make -f makefile.mk db-connect`

## API Reference

The gRPC service definition, including request and response messages, can be found in:
`proto/user/v1/user.proto`
