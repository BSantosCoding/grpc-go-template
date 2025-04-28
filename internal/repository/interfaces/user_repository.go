package interfaces

import (
	"context"

	userpb "github.com/BSantosCoding/grpc-go-example/gen/user/v1"
)

// UserRepository defines the interface for user data operations.
type UserRepository interface {
	Create(ctx context.Context, user *userpb.User) (*userpb.User, error)
	GetByID(ctx context.Context, id string) (*userpb.User, error)
	ListAll(ctx context.Context) ([]*userpb.User, error)
	// Add other methods like Update, Delete, GetByEmail etc. as needed
}
