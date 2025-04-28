// cmd/server/container.go
package main // Keep in package main as it's used directly by server's main.go

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	// Adjust import paths if necessary
	"github.com/BSantosCoding/grpc-go-example/internal/repository"
	repoInterfaces "github.com/BSantosCoding/grpc-go-example/internal/repository/interfaces"
	"github.com/BSantosCoding/grpc-go-example/internal/server"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq" // Driver
)

// Container holds the application's dependencies.
type Container struct {
	DB         *sql.DB                     // Exported to allow closing in main
	UserRepo   repoInterfaces.UserRepository
	UserServer *server.UserServer
	// Add other dependencies here as the app grows (e.g., other repos, clients, config struct)
}

// AppConfig holds configuration values loaded from environment.
type AppConfig struct {
	DBSource string
	// Add other config fields here
}

// LoadConfig loads configuration from environment variables.
// It prioritizes DATABASE_URL then constructs from PG* vars.
func LoadConfig() (*AppConfig, error) {
	// Load .env file first. Ignore error if file not found.
	err := godotenv.Load()
	if err != nil && !os.IsNotExist(err) {
		// Log error only if it's not a "file not found" error
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	config := &AppConfig{}

	// Prioritize DATABASE_URL if set, otherwise construct from parts
	dbSource := os.Getenv(dbSourceEnvVar)
	if dbSource == "" {
		log.Println("DATABASE_URL not set, constructing from PG* variables.")
		pgHost := getEnv(dbHostEnvVar, "localhost")
		pgPort := getEnv(dbPortEnvVar, "5432")
		pgUser := getEnv(dbUserEnvVar, "postgres")
		pgPassword := getEnv(dbPasswordEnvVar, "mysecretpassword") // Use fallback from example if not set
		dbName := getEnv(dbNameEnvVar, "users_db")

		// Construct the connection string
		dbSource = fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable",
			pgUser, pgPassword, pgHost, pgPort, dbName)
	} else {
		log.Printf("Using DATABASE_URL from environment.")
	}
	config.DBSource = dbSource

	// Load other config values here

	return config, nil
}


// NewContainer creates and wires up the application dependencies.
func NewContainer(config *AppConfig) (*Container, error) {
	log.Println("Building application container...")

	// --- Build Dependencies ---

	// 1. Database Connection
	db, err := buildDBConnection(config.DBSource)
	if err != nil {
		return nil, fmt.Errorf("failed to build database connection: %w", err)
	}

	// 2. Repositories
	userRepo := repository.NewUserRepository(db)

	// 3. gRPC Server Handlers/Services
	userServer := server.NewUserServer(userRepo)

	// --- Create Container ---
	container := &Container{
		DB:         db,
		UserRepo:   userRepo,
		UserServer: userServer,
	}

	log.Println("Application container built successfully.")
	return container, nil
}

// buildDBConnection establishes and pings the database connection.
// (Moved from main.go)
func buildDBConnection(dataSourceName string) (*sql.DB, error) {
	db, err := sql.Open(dbDriver, dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Verify the connection is working
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err = db.PingContext(ctx); err != nil {
		db.Close() // Close the connection if ping fails
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Successfully connected to the database.")
	return db, nil
}

// getEnv helper retrieves environment variable with a fallback.
// (Moved from main.go)
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	log.Printf("Environment variable %s not set, using default: %s", key, fallback)
	return fallback
}

// Close cleans up resources held by the container (like DB connection).
// Note: We export DB so main can call Close, but a Close method on Container is cleaner.
func (c *Container) Close() error {
	log.Println("Closing container resources...")
	if c.DB != nil {
		if err := c.DB.Close(); err != nil {
			return fmt.Errorf("error closing database connection: %w", err)
		}
		log.Println("Database connection closed.")
	}
	// Add closing logic for other resources if needed
	return nil
}

