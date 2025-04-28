// cmd/server/container.go
package main // Keep in package main as it's used directly by server's main.go

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	// Adjust import paths if necessary
	"github.com/BSantosCoding/grpc-go-example/internal/config"
	"github.com/BSantosCoding/grpc-go-example/internal/repository"
	repoInterfaces "github.com/BSantosCoding/grpc-go-example/internal/repository/interfaces"
	"github.com/BSantosCoding/grpc-go-example/internal/server"
	_ "github.com/lib/pq" // Driver
)

// Container holds the application's dependencies.
type Container struct {
	DB         *sql.DB                     
	UserRepo   repoInterfaces.UserRepository
	UserServer *server.UserServer
}

// NewContainer creates and wires up the application dependencies.
func NewContainer(config *config.AppConfig) (*Container, error) {
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

// Close cleans up resources held by the container (like DB connection).
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

