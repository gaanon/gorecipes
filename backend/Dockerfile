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

# Copy the compiled binary from the builder stage
COPY --from=builder /app/gorecipes-backend /app/gorecipes-backend
# Copy generated Swagger docs
COPY --from=builder /app/docs /app/docs

# Copy schema and seed data SQL files from the builder stage
COPY --from=builder /app/internal/database/schema.sql /app/internal/database/schema.sql

# The application needs a directory for BadgerDB data and uploaded images.
# These paths should correspond to where Docker volumes will be mounted.
# The application currently uses "data/badger" and "uploads/images" relative to its execution path.
# We don't strictly need to create them here if volumes are mounted,
# but it's good practice for clarity and if run without volumes initially.
RUN mkdir -p /app/data/badger /app/uploads/images && \
    chown -R nobody:nogroup /app/data /app/uploads && \
    chmod -R 770 /app/data /app/uploads 
# Consider using a non-root user for added security
# USER nobody:nogroup # Or a dedicated app user

# Expose the port the backend listens on (default 8080)
EXPOSE 8080

# Command to run the application
# The binary is now at /app/gorecipes-backend
CMD ["/app/gorecipes-backend"]
