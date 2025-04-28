# Dockerfile

# ---- Builder Stage ----
# Use a specific Go version. Alpine is smaller.
FROM golang:1.24-alpine AS builder

# Set working directory
WORKDIR /app

# Install build dependencies if needed (e.g., git, gcc for CGO)
# RUN apk add --no-cache git build-base

# Copy go module files
COPY go.mod go.sum ./

# Download dependencies. This leverages Docker cache.
# Dependencies are only re-downloaded if go.mod/go.sum change.
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the application binary
# -o specifies the output file
# ./cmd/server is the package to build
# CGO_ENABLED=0 produces a static binary (good for scratch/alpine)
# -ldflags="-w -s" strips debug info, making the binary smaller
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /app/server ./cmd/server

# ---- Final Stage ----
# Use a minimal base image
FROM alpine:latest

# Set working directory
WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/server /app/server

# Expose the gRPC port the application listens on
EXPOSE 50051

# Command to run the application
CMD ["/app/server"]

# Optional: Add healthcheck if your app supports it
# HEALTHCHECK --interval=5s --timeout=3s CMD curl --fail http://localhost:8081/health || exit 1
