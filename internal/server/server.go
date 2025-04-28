// internal/server/server.go
package server

import (
	"context"
	"log"

	// Adjust import paths if necessary
	userpb "github.com/BSantosCoding/grpc-go-example/gen/user/v1"
	// Import the interfaces package
	"github.com/BSantosCoding/grpc-go-example/internal/repository/interfaces"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UserServer now depends on the UserRepository interface from the sub-package.
type UserServer struct {
	userpb.UnimplementedUserServiceServer
	repo interfaces.UserRepository // Use the interface type from the 'interfaces' package
}

// NewUserServer constructor now accepts the interface from the sub-package.
func NewUserServer(repo interfaces.UserRepository) *UserServer {
	return &UserServer{
		repo: repo,
	}
}

// Server methods remain the same, calling the injected repository interface methods.
func (s *UserServer) CreateUser(ctx context.Context, req *userpb.CreateUserRequest) (*userpb.CreateUserResponse, error) {
	log.Printf("Handler: Received CreateUser request: Name=%s, Email=%s", req.GetName(), req.GetEmail())
	if req.GetName() == "" || req.GetEmail() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Name and Email cannot be empty")
	}

	// Prepare data for repository
	userToCreate := &userpb.User{
		Name:  req.GetName(),
		Email: req.GetEmail(),
	}

	createdUser, err := s.repo.Create(ctx, userToCreate)
	if err != nil {
		log.Printf("Handler: Error calling repository Create: %v", err)
		// Assuming repo returns appropriate gRPC status errors
		return nil, err
	}

	log.Printf("Handler: User created via repository with ID: %s", createdUser.Id)
	return &userpb.CreateUserResponse{User: createdUser}, nil
}

func (s *UserServer) GetUser(ctx context.Context, req *userpb.GetUserRequest) (*userpb.GetUserResponse, error) {
	userID := req.GetId()
	log.Printf("Handler: Received GetUser request for ID: %s", userID)
	if userID == "" {
		return nil, status.Errorf(codes.InvalidArgument, "User ID cannot be empty")
	}

	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		log.Printf("Handler: Error calling repository GetByID: %v", err)
		return nil, err // Assuming repo returns appropriate gRPC status errors
	}

	log.Printf("Handler: Found user via repository: ID=%s, Name=%s", user.GetId(), user.GetName())
	return &userpb.GetUserResponse{User: user}, nil
}

func (s *UserServer) ListUsers(ctx context.Context, req *userpb.ListUsersRequest) (*userpb.ListUsersResponse, error) {
	log.Printf("Handler: Received ListUsers request")

	users, err := s.repo.ListAll(ctx)
	if err != nil {
		log.Printf("Handler: Error calling repository ListAll: %v", err)
		return nil, err // Assuming repo returns appropriate gRPC status errors
	}

	log.Printf("Handler: Returning %d users from repository", len(users))
	return &userpb.ListUsersResponse{Users: users}, nil
}
