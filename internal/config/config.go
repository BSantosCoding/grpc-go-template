// internal/config/config.go
package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Environment variable names (keep consistent)
const (
	dbHostEnvVar     = "PGHOST"
	dbPortEnvVar     = "PGPORT"
	dbUserEnvVar     = "PGUSER"
	dbPasswordEnvVar = "PGPASSWORD"
	dbNameEnvVar     = "DBNAME"
	dbSourceEnvVar   = "DATABASE_URL"
)

// AppConfig holds configuration values loaded from environment.
type AppConfig struct {
	DBSource string
}

// LoadConfig loads configuration from environment variables.
func LoadConfig() (*AppConfig, error) {
	// Load .env file first. Ignore error if file not found.
	err := godotenv.Load() 
	if err != nil && !os.IsNotExist(err) {
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
		pgPassword := getEnv(dbPasswordEnvVar, "postgres") // Use fallback
		dbName := getEnv(dbNameEnvVar, "users_db")

		// Construct the connection string
		dbSource = fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable",
			pgUser, pgPassword, pgHost, pgPort, dbName)
	} else {
		log.Printf("Using DATABASE_URL from environment.")
	}
	config.DBSource = dbSource

	return config, nil
}

// getEnv helper retrieves environment variable with a fallback.
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
