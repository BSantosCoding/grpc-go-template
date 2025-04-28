package main

import (
	"bufio" // For reading user input
	"context"
	"fmt" // For printing menu and prompts
	"log"
	"os"
	"strings" // For trimming input
	"time"

	userpb "github.com/BSantosCoding/grpc-go-example/gen/user/v1"
	// "github.com/google/uuid" // No longer needed for generating random IDs here

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

const (
	serverAddr     = "localhost:50051"
	requestTimeout = 15 * time.Second
)

// Helper function to read trimmed string input from the user
func readInput(reader *bufio.Reader, prompt string) (string, error) {
	fmt.Print(prompt)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(input), nil
}

func main() {
	log.Printf("Connecting to gRPC server at %s", serverAddr)

	// Establish connection once at the start
	conn, err := grpc.NewClient(serverAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("Did not connect: %v", err)
	}
	// Defer closing the connection until the main function exits
	defer conn.Close()

	log.Println("Connection established.")

	// Create the client stub once
	c := userpb.NewUserServiceClient(conn)
	// Create a reader for user input
	reader := bufio.NewReader(os.Stdin)

	// --- Interactive Loop ---
	for {
		// Print Menu
		fmt.Println("\n--- User Service Client ---")
		fmt.Println("1. Create User")
		fmt.Println("2. Get User by ID")
		fmt.Println("3. List All Users")
		fmt.Println("q. Quit")
		fmt.Println("---------------------------")

		choice, err := readInput(reader, "Enter choice: ")
		if err != nil {
			log.Printf("Error reading input: %v. Exiting.", err)
			return
		}

		// Create a new context with timeout for each request inside the loop
		ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)

		switch choice {
		case "1":
			// --- Create User ---
			userName, err := readInput(reader, "Enter name: ")
			if err != nil {
				log.Printf("Error reading name: %v", err)
				cancel() // Cancel context if input fails
				continue // Go back to menu
			}
			userEmail, err := readInput(reader, "Enter email: ")
			if err != nil {
				log.Printf("Error reading email: %v", err)
				cancel()
				continue
			}

			if userName == "" || userEmail == "" {
				log.Println("Name and Email cannot be empty.")
				cancel()
				continue
			}

			log.Printf("--- Calling CreateUser ---")
			createRes, err := c.CreateUser(ctx, &userpb.CreateUserRequest{Name: userName, Email: userEmail})
			if err != nil {
				// Use log.Printf instead of log.Fatalf to keep client running
				log.Printf("Could not create user: %v", err)
			} else {
				createdUser := createRes.GetUser()
				log.Printf("User created successfully: ID=%s, Name=%s, Email=%s, CreatedAt=%s",
					createdUser.GetId(),
					createdUser.GetName(),
					createdUser.GetEmail(),
					createdUser.GetCreatedAt().AsTime().Local().Format(time.RFC1123))
			}

		case "2":
			// --- Get User ---
			userID, err := readInput(reader, "Enter user ID: ")
			if err != nil {
				log.Printf("Error reading user ID: %v", err)
				cancel()
				continue
			}
			if userID == "" {
				log.Println("User ID cannot be empty.")
				cancel()
				continue
			}

			log.Printf("\n--- Calling GetUser ---")
			getRes, err := c.GetUser(ctx, &userpb.GetUserRequest{Id: userID})
			if err != nil {
				if st, ok := status.FromError(err); ok {
					log.Printf("GetUser failed with code %s: %s", st.Code(), st.Message())
				} else {
					log.Printf("GetUser failed: %v", err)
				}
			} else {
				retrievedUser := getRes.GetUser()
				log.Printf("User retrieved successfully: ID=%s, Name=%s, Email=%s, CreatedAt=%s",
					retrievedUser.GetId(),
					retrievedUser.GetName(),
					retrievedUser.GetEmail(),
					retrievedUser.GetCreatedAt().AsTime().Local().Format(time.RFC1123))
			}

		case "3":
			// --- List Users ---
			log.Printf("\n--- Calling ListUsers ---")
			listRes, err := c.ListUsers(ctx, &userpb.ListUsersRequest{})
			if err != nil {
				log.Printf("Could not list users: %v", err)
			} else {
				log.Printf("Users list retrieved successfully:")
				if len(listRes.GetUsers()) == 0 {
					log.Println("  (No users found in database)")
				}
				for i, user := range listRes.GetUsers() {
					log.Printf("  %d: ID=%s, Name=%s, Email=%s, CreatedAt=%s",
						i+1,
						user.GetId(),
						user.GetName(),
						user.GetEmail(),
						user.GetCreatedAt().AsTime().Local().Format(time.RFC1123))
				}
			}

		case "q", "Q":
			// --- Quit ---
			log.Println("Exiting client.")
			cancel() // Cancel context before exiting
			return   // Exit the main function

		default:
			log.Println("Invalid choice. Please try again.")
		}

		// Cancel the context for the current request after handling the case
		cancel()
	} // End of infinite loop
}
