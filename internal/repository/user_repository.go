// internal/repository/user_repository.go
package repository // Package remains 'repository'

import (
	"context"
	"database/sql"
	"log" // Or a proper logger

	// Adjust import paths if necessary
	userpb "github.com/BSantosCoding/grpc-go-example/gen/user/v1"
	// Import the interfaces package
	"github.com/BSantosCoding/grpc-go-example/internal/repository/interfaces"

	"github.com/google/uuid"
	"github.com/lib/pq" // Keep the driver import needed for implementation details (like error checking)
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// userRepository implements the interfaces.UserRepository interface using PostgreSQL.
// The struct name itself can still indicate the underlying tech if desired, or be generic.
// Let's keep it generic here as requested.
type userRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new repository instance.
// It returns the interface type defined in the 'interfaces' sub-package.
// The function name is now generic.
func NewUserRepository(db *sql.DB) interfaces.UserRepository {
	return &userRepository{db: db}
}

// Create implements the interfaces.UserRepository interface.
func (r *userRepository) Create(ctx context.Context, userReq *userpb.User) (*userpb.User, error) {
	userID := uuid.New()
	newUser := &userpb.User{
		Id:    userID.String(),
		Name:  userReq.GetName(),
		Email: userReq.GetEmail(),
	}
	query := `INSERT INTO users (id, name, email) VALUES ($1, $2, $3) RETURNING created_at`
	var createdAt sql.NullTime
	err := r.db.QueryRowContext(ctx, query, newUser.Id, newUser.Name, newUser.Email).Scan(&createdAt)
	if err != nil {
		// Check for unique constraint violation using the imported driver's error type
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			log.Printf("Repo: Email %s already exists", newUser.Email)
			// Return gRPC status errors from repo is one option, custom errors is another
			return nil, status.Errorf(codes.AlreadyExists, "Email '%s' already exists", newUser.Email)
		}
		log.Printf("Repo: Error creating user in database: %v", err)
		return nil, status.Errorf(codes.Internal, "Failed to create user")
	}

	if createdAt.Valid {
		newUser.CreatedAt = timestamppb.New(createdAt.Time)
	}
	log.Printf("Repo: Created user with ID: %s", newUser.Id)
	return newUser, nil
}

// GetByID implements the interfaces.UserRepository interface.
func (r *userRepository) GetByID(ctx context.Context, id string) (*userpb.User, error) {
	if _, err := uuid.Parse(id); err != nil {
		log.Printf("Repo: Invalid User ID format: %s", id)
		return nil, status.Errorf(codes.InvalidArgument, "Invalid User ID format")
	}

	query := `SELECT id, name, email, created_at FROM users WHERE id = $1`
	user := &userpb.User{}
	var createdAt sql.NullTime
	row := r.db.QueryRowContext(ctx, query, id)
	err := row.Scan(&user.Id, &user.Name, &user.Email, &createdAt)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Repo: User not found with ID: %s", id)
			return nil, status.Errorf(codes.NotFound, "User with ID '%s' not found", id)
		}
		log.Printf("Repo: Error retrieving user from database: %v", err)
		return nil, status.Errorf(codes.Internal, "Failed to retrieve user")
	}

	if createdAt.Valid {
		user.CreatedAt = timestamppb.New(createdAt.Time)
	}
	log.Printf("Repo: Found user: ID=%s, Name=%s", user.GetId(), user.GetName())
	return user, nil
}

// ListAll implements the interfaces.UserRepository interface.
func (r *userRepository) ListAll(ctx context.Context) ([]*userpb.User, error) {
	log.Printf("Repo: Received ListAll request")
	query := `SELECT id, name, email, created_at FROM users ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		log.Printf("Repo: Error querying users from database: %v", err)
		return nil, status.Errorf(codes.Internal, "Failed to list users")
	}
	defer rows.Close()

	var userList []*userpb.User
	for rows.Next() {
		user := &userpb.User{}
		var createdAt sql.NullTime
		if err := rows.Scan(&user.Id, &user.Name, &user.Email, &createdAt); err != nil {
			log.Printf("Repo: Error scanning user row: %v", err)
			continue // Skip this user
		}
		if createdAt.Valid {
			user.CreatedAt = timestamppb.New(createdAt.Time)
		}
		userList = append(userList, user)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Repo: Error iterating user rows: %v", err)
		return nil, status.Errorf(codes.Internal, "Failed to process user list")
	}

	log.Printf("Repo: Returning %d users", len(userList))
	return userList, nil
}

