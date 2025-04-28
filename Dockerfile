# Dockerfile

# ---- Builder Stage ----
FROM golang:1.24-alpine AS builder
WORKDIR /app
# RUN apk add --no-cache git build-base # Uncomment if CGO is needed
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /app/server ./cmd/server

# ---- Development Stage ----
FROM golang:1.24-alpine AS development
WORKDIR /app
# Install Air
RUN go install github.com/air-verse/air@latest
# Copy go module files first for caching
COPY go.mod go.sum ./
RUN go mod download
# Copy air config
COPY .air.toml .
# Copy the rest of the application source code
COPY . .
# Expose the gRPC port
EXPOSE 50051
# Command to run Air (will use .air.toml by default)
CMD ["air"]

# ---- Final Stage ----
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/server /app/server
EXPOSE 50051
CMD ["/app/server"]

    