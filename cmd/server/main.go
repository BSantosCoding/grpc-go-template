package main

import (
	"log"
	"net"

	// Import your internal server implementation
	"github.com/BSantosCoding/grpc-go-example/internal/server"

	// Import the generated protobuf code
	userpb "github.com/BSantosCoding/grpc-go-example/gen/proto/user/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection" // Optional: for gRPC reflection
)

const (
	port = ":50051" // Port the server will listen on
)

func main() {
	log.Printf("Starting gRPC server on port %s", port)

	// Create a TCP listener on the specified port
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// Create a new gRPC server instance
	s := grpc.NewServer(
	// Add server options here if needed (e.g., interceptors, TLS credentials)
	)

	// Create an instance of your User server implementation
	userSrv := server.NewUserServer()

	// Register your service implementation with the gRPC server
	userpb.RegisterUserServiceServer(s, userSrv)

	// Optional: Register reflection service on gRPC server.
	// This allows tools like grpcurl to query the server's services.
	reflection.Register(s)
	log.Println("gRPC reflection registered.")

	// Start serving requests
	log.Printf("Server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
