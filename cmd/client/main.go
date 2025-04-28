package main

import (
	"context"
	"log"
	"os"
	"time"

	// Import the generated protobuf code
	userpb "github.com/BSantosCoding/grpc-go-example/gen/proto/user/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure" // For insecure connections (testing only)
)

const (
	serverAddr    = "localhost:50051" // Address of the gRPC server
	defaultName   = "World"
	requestTimeout = 5 * time.Second
)

func main() {
	// Determine the name to greet from command-line arguments
	name := defaultName
	if len(os.Args) > 1 {
		name = os.Args[1]
	}

	log.Printf("Connecting to gRPC server at %s", serverAddr)

	// Set up a connection to the server.
	// grpc.WithTransportCredentials(insecure.NewCredentials()) is used for simplicity.
	// In production, use proper TLS credentials!
	// grpc.WithBlock() makes the connection attempt synchronous and fail fast if unavailable.
	conn, err := grpc.Dial(serverAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(), // Block until connection is established or fails
	)
	if err != nil {
		log.Fatalf("Did not connect: %v", err)
	}
	defer conn.Close() // Ensure the connection is closed when main function exits

	log.Println("Connection established.")

	// Create a client stub for the User service
	c := userpb.NewUserServiceClient(conn)

	// Prepare the request
	req := &userpb.SayHelloRequest{Name: name}

	// Set a timeout for the RPC call
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel() // Ensure the context resources are released

	// Contact the server and print out its response.
	log.Printf("Sending SayHello request with name: %s", name)
	res, err := c.SayHello(ctx, req)
	if err != nil {
		log.Fatalf("Could not greet: %v", err)
	}

	log.Printf("Greeting received: %s", res.GetMessage())
}

