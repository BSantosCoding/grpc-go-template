// cmd/server/main.go
package main

import (
	// "database/sql" // No longer needed directly here
	// "context" // No longer needed directly here
	// "fmt" // No longer needed directly here
	"log"
	"net"

	// "os" // No longer needed directly here
	// "time" // No longer needed directly here

	// Adjust import paths if necessary
	userpb "github.com/BSantosCoding/grpc-go-example/gen/user/v1"
	"github.com/BSantosCoding/grpc-go-example/internal/config"

	// "github.com/BSantosCoding/grpc-go-example/internal/repository" // No longer needed directly here
	// "github.com/BSantosCoding/grpc-go-example/internal/server" // No longer needed directly here

	// "github.com/joho/godotenv" // No longer needed directly here
	_ "github.com/lib/pq" // Driver (still needed for side effect)
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Constants moved to container.go or loaded into AppConfig
const (
	port = ":50051"
	dbDriver = "postgres" // Keep driver name if needed by buildDBConnection
	dbHostEnvVar   = "PGHOST"
	dbPortEnvVar   = "PGPORT"
	dbUserEnvVar   = "PGUSER"
	dbPasswordEnvVar = "PGPASSWORD"
	dbNameEnvVar   = "DBNAME"
	dbSourceEnvVar = "DATABASE_URL"
)

// getEnv moved to container.go
// connectDB moved to container.go and renamed buildDBConnection

func main() {
	// --- Load Configuration ---
	config, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// --- Build Container ---
	container, err := NewContainer(config)
	if err != nil {
		log.Fatalf("Failed to build application container: %v", err)
	}
	// Ensure container resources are closed on exit
	defer func() {
		if err := container.Close(); err != nil {
			log.Printf("Error closing container resources: %v", err)
		}
	}()


	// --- Setup gRPC Server ---
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", port, err)
	}

	// Create the main gRPC server instance
	grpcServer := grpc.NewServer(
		// Add server options like interceptors here if needed
	)

	// Retrieve the fully built UserServer from the container
	userSrv := container.UserServer

	// Register the service implementation with the gRPC server
	userpb.RegisterUserServiceServer(grpcServer, userSrv)

	reflection.Register(grpcServer)
	log.Println("gRPC reflection registered.")

	// --- Start Server ---
	log.Printf("gRPC server listening at %v", lis.Addr())
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC: %v", err)
	}
}
