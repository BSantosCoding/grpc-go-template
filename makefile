.PHONY: proto tidy build-server build-client run-server run-client all clean

MODULE_PATH=github.com/BSantosCoding/grpc-go-example # Keep this updated
PROTO_DIR=proto
GEN_DIR=gen
BIN_DIR=bin

# Generate protobuf Go code for User service
proto:
	@mkdir -p $(GEN_DIR)/proto/user/v1
	protoc --proto_path=$(PROTO_DIR) \
	       --go_out=$(GEN_DIR) --go_opt=paths=source_relative \
	       --go-grpc_out=$(GEN_DIR) --go-grpc_opt=paths=source_relative \
	       $(PROTO_DIR)/user/v1/user.proto # <-- Updated proto file path
	@echo "Protobuf code generated for User service."

# Tidy dependencies (no change needed)
tidy:
	go mod tidy
	@echo "Dependencies tidied."

# Build server binary (no change needed)
build-server: tidy
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/server ./cmd/server
	@echo "Server built."

# Build client binary (no change needed)
build-client: tidy
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/client ./cmd/client
	@echo "Client built."

# Run the server (no change needed)
run-server: build-server
	@echo "Starting server..."
	./$(BIN_DIR)/server

# Run the client (no change needed, arguments handled by client main.go)
run-client: build-client
	@echo "Running client..."
	./$(BIN_DIR)/client $(filter-out $@,$(MAKECMDGOALS))

# Build all (depends on updated proto target)
all: proto build-server build-client

# Clean generated files and binaries (no change needed)
clean:
	@rm -rf $(GEN_DIR)
	@rm -rf $(BIN_DIR)
	@echo "Cleaned generated files and binaries."

# Default target
default: all

# Allow passing arguments to run-client
%:
	@:

