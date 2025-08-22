# Stage 1: Build the Go application
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum files to download dependencies
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy the rest of the backend source code
COPY . .

# Install swag CLI
RUN go install github.com/swaggo/swag/cmd/swag@latest

# Generate Swagger documentation
RUN swag init -dir . -generalInfo ./cmd/server/main.go

# Build the application
# CGO_ENABLED=0 for a static binary, GOOS=linux for Linux compatibility
# -a installs all packages that are dependencies of the build
# -installsuffix cgo to prevent issues with cgo if it were enabled
# Output binary is named gorecipes-backend
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/gorecipes-backend ./cmd/server/main.go

# Stage 2: Create the runtime image
# Using alpine for a small image size
FROM alpine:latest

WORKDIR /app

# Install CA certificates for HTTPS requests (needed for external API calls)
RUN apk --no-cache add ca-certificates

# Copy the binary from the builder stage
COPY --from=builder /app/gorecipes-backend .

# Create necessary directories and set proper permissions
RUN mkdir -p /app/uploads/images && \
    mkdir -p /app/data && \
    chown -R 1000:1000 /app/uploads /app/data && \
    chmod -R 775 /app/uploads /app/data

# Switch to a non-root user for security
USER 1000

# Expose the port the backend listens on (default 8080)
EXPOSE 8080

# Command to run the application
# The binary is now at /app/gorecipes-backend
CMD ["/app/gorecipes-backend"]
