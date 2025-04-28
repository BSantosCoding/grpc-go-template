package server

import (
	"context"
	"fmt"
	"log"

	// Import the generated protobuf code
	userpb "github.com/BSantosCoding/grpc-go-example/gen/proto/user/v1"
)

// UserServer implements the generated UserServiceServer interface.
type UserServer struct {
	// Embed the UnimplementedUserServiceServer
	// This is required for forward compatibility.
	userpb.UnimplementedUserServiceServer
}

// NewUserServer creates a new instance of our UserServer.
func NewUserServer() *UserServer {
	return &UserServer{}
}

// SayHello implements the SayHello RPC method.
func (s *UserServer) SayHello(ctx context.Context, req *userpb.SayHelloRequest) (*userpb.SayHelloResponse, error) {
	log.Printf("Received SayHello request with name: %s", req.GetName())

	if req.GetName() == "" {
		log.Println("Received empty name")
		// You might want to return an error for invalid input
		// return nil, status.Errorf(codes.InvalidArgument, "Name cannot be empty")
		// For simplicity, we'll just use a default name
		req.Name = "Anonymous"
	}

	message := fmt.Sprintf("Hello, %s!", req.GetName())

	return &userpb.SayHelloResponse{Message: message}, nil
}
