// internal/server/server_test.go
package server_test // Use _test package for black-box testing

import (
	"context"
	"database/sql"
	"log"
	"net"
	"testing"
	"time"

	// Adjust paths as necessary
	userpb "github.com/BSantosCoding/grpc-go-example/gen/user/v1"
	"github.com/BSantosCoding/grpc-go-example/internal/config" // Use shared config
	"github.com/BSantosCoding/grpc-go-example/internal/repository"
	repoInterfaces "github.com/BSantosCoding/grpc-go-example/internal/repository/interfaces"
	"github.com/BSantosCoding/grpc-go-example/internal/server" // Import the package being tested

	_ "github.com/lib/pq" // Postgres driver
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status" // For checking error codes
)

const (
	testTimeout = 5 * time.Second
	dbDriver    = "postgres" // Keep consistent
)

// UserServerTestSuite defines the suite structure
type UserServerTestSuite struct {
	suite.Suite // Embed testify suite
	client      userpb.UserServiceClient
	conn        *grpc.ClientConn
	db          *sql.DB
	repo        repoInterfaces.UserRepository
	grpcServer  *grpc.Server
	listener    net.Listener
	serverAddr  string
}

// SetupSuite runs once before all tests in the suite
func (s *UserServerTestSuite) SetupSuite() {
	log.Println("Setting up test suite...")
	require := s.Require() // Use require for setup failures

	// 1. Load Config (using shared package)
	cfg, err := config.LoadConfig()
	require.NoError(err, "Failed to load config")

	// 2. Connect to DB
	s.db, err = sql.Open(dbDriver, cfg.DBSource)
	require.NoError(err, "Failed to open DB connection")
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()
	err = s.db.PingContext(ctx)
	require.NoError(err, "Failed to ping DB")
	log.Println("Test DB connection established")

	// 3. Create Repository
	s.repo = repository.NewUserRepository(s.db)

	// 4. Start Test gRPC Server
	s.listener, err = net.Listen("tcp", ":0") // Listen on random available port
	require.NoError(err, "Failed to start listener")
	s.serverAddr = s.listener.Addr().String() // Get the actual address
	log.Printf("Test server listening on %s", s.serverAddr)

	s.grpcServer = grpc.NewServer()
	userSrv := server.NewUserServer(s.repo) // Create instance of server being tested
	userpb.RegisterUserServiceServer(s.grpcServer, userSrv)

	// Run server in a goroutine
	go func() {
		log.Println("Starting test gRPC server...")
		if err := s.grpcServer.Serve(s.listener); err != nil {
			// Log error, but don't fatal during test run if server stops unexpectedly
			log.Printf("Test gRPC server Serve() error: %v", err)
		}
	}()

	// 5. Create Test Client
	// Use NewClient instead of deprecated Dial
	s.conn, err = grpc.NewClient(s.serverAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(), // Block until connected for tests
	)
	require.NoError(err, "Failed to connect to test server")
	s.client = userpb.NewUserServiceClient(s.conn)

	log.Println("Test suite setup complete.")
}

// TearDownSuite runs once after all tests
func (s *UserServerTestSuite) TearDownSuite() {
	log.Println("Tearing down test suite...")
	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
		log.Println("Test gRPC server stopped.")
	}
	if s.conn != nil {
		s.conn.Close()
		log.Println("Test client connection closed.")
	}
	if s.db != nil {
		s.db.Close()
		log.Println("Test DB connection closed.")
	}
	log.Println("Test suite teardown complete.")
}

// BeforeTest runs before each test function
func (s *UserServerTestSuite) BeforeTest(suiteName, testName string) {
	log.Printf("--- Running test: %s ---", testName)
	// Clean the users table before each test
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()
	_, err := s.db.ExecContext(ctx, "TRUNCATE users RESTART IDENTITY CASCADE")
	s.Require().NoError(err, "Failed to truncate users table")
	log.Println("Users table truncated.")
}

// --- Test Functions ---

func (s *UserServerTestSuite) TestCreateUser_Success() {
	require := s.Require()
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	req := &userpb.CreateUserRequest{
		Name:  "Test User One",
		Email: "test.user.one@example.com",
	}
	res, err := s.client.CreateUser(ctx, req)

	require.NoError(err)
	require.NotNil(res)
	require.NotNil(res.User)
	require.NotEmpty(res.User.Id)
	require.Equal(req.Name, res.User.Name)
	require.Equal(req.Email, res.User.Email)
	require.NotNil(res.User.CreatedAt)

	// Optional: Verify in DB directly (though this tests repo implicitly)
	// var count int
	// err = s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE id = $1", res.User.Id).Scan(&count)
	// require.NoError(err)
	// require.Equal(1, count)
}

func (s *UserServerTestSuite) TestCreateUser_DuplicateEmail() {
	require := s.Require()
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	// Create first user
	req1 := &userpb.CreateUserRequest{
		Name:  "Test User Two",
		Email: "duplicate.email@example.com",
	}
	_, err := s.client.CreateUser(ctx, req1)
	require.NoError(err, "Setup: Failed to create first user")

	// Attempt to create second user with same email
	req2 := &userpb.CreateUserRequest{
		Name:  "Test User Three",
		Email: "duplicate.email@example.com", // Same email
	}
	res, err := s.client.CreateUser(ctx, req2)

	require.Error(err) // Expect an error
	require.Nil(res)   // Expect nil response

	// Check gRPC status code
	st, ok := status.FromError(err)
	require.True(ok, "Error should be a gRPC status error")
	require.Equal(codes.AlreadyExists, st.Code(), "Expected AlreadyExists error code")
	require.Contains(st.Message(), "already exists", "Error message should indicate duplication")
}

func (s *UserServerTestSuite) TestGetUser_Success() {
	require := s.Require()
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	// Setup: Create a user first
	createReq := &userpb.CreateUserRequest{
		Name:  "Get Me User",
		Email: "get.me@example.com",
	}
	createRes, err := s.client.CreateUser(ctx, createReq)
	require.NoError(err, "Setup: Failed to create user")
	require.NotNil(createRes.User)
	createdID := createRes.User.Id

	// Test: Get the created user
	getReq := &userpb.GetUserRequest{Id: createdID}
	getRes, err := s.client.GetUser(ctx, getReq)

	require.NoError(err)
	require.NotNil(getRes)
	require.NotNil(getRes.User)
	require.Equal(createdID, getRes.User.Id)
	require.Equal(createReq.Name, getRes.User.Name)
	require.Equal(createReq.Email, getRes.User.Email)
}

func (s *UserServerTestSuite) TestGetUser_NotFound() {
	require := s.Require()
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	// Generate a valid UUID that doesn't exist
	nonExistentID := "123e4567-e89b-12d3-a456-426614174000" // Example valid UUID

	getReq := &userpb.GetUserRequest{Id: nonExistentID}
	res, err := s.client.GetUser(ctx, getReq)

	require.Error(err)
	require.Nil(res)
	st, ok := status.FromError(err)
	require.True(ok)
	require.Equal(codes.NotFound, st.Code())
}

func (s *UserServerTestSuite) TestGetUser_InvalidID() {
	require := s.Require()
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	getReq := &userpb.GetUserRequest{Id: "not-a-valid-uuid"}
	res, err := s.client.GetUser(ctx, getReq)

	require.Error(err)
	require.Nil(res)
	st, ok := status.FromError(err)
	require.True(ok)
	require.Equal(codes.InvalidArgument, st.Code())
	require.Contains(st.Message(), "Invalid User ID format")
}

func (s *UserServerTestSuite) TestListUsers_Empty() {
	require := s.Require()
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	res, err := s.client.ListUsers(ctx, &userpb.ListUsersRequest{})

	require.NoError(err)
	require.NotNil(res)
	require.Empty(res.Users) // Expect an empty slice, not nil
}

func (s *UserServerTestSuite) TestListUsers_WithData() {
	require := s.Require()
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	// Setup: Create multiple users
	usersToCreate := []*userpb.CreateUserRequest{
		{Name: "List User 1", Email: "list1@example.com"},
		{Name: "List User 2", Email: "list2@example.com"},
	}
	for _, req := range usersToCreate {
		_, err := s.client.CreateUser(ctx, req)
		require.NoError(err, "Setup: Failed to create user %s", req.Name)
	}

	// Test: List users
	res, err := s.client.ListUsers(ctx, &userpb.ListUsersRequest{})

	require.NoError(err)
	require.NotNil(res)
	require.Len(res.Users, len(usersToCreate))

	// Optional: Check if names/emails match (order might vary depending on DB)
	foundNames := make(map[string]bool)
	for _, u := range res.Users {
		foundNames[u.Name] = true
	}
	require.True(foundNames["List User 1"])
	require.True(foundNames["List User 2"])
}

// --- Test Suite Runner ---

// TestUserServerTestSuite runs the testify suite
func TestUserServerTestSuite(t *testing.T) {
	// Check if DB is available, skip if not (useful for CI without DB)
	// You might want more sophisticated checks or build tags
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config for test check: %v", err)
	}
	db, err := sql.Open(dbDriver, cfg.DBSource)
	if err == nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second) // Short timeout for ping
		defer cancel()
		err = db.PingContext(ctx)
		db.Close() // Close the temporary connection
	}
	if err != nil {
		t.Skipf("Skipping server tests: Cannot connect to database (%v)", err)
		return
	}

	// Run the suite
	suite.Run(t, new(UserServerTestSuite))
}

