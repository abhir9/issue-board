# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
# CGO_ENABLED=1 is required for go-sqlite3
RUN CGO_ENABLED=1 GOOS=linux go build -o main ./cmd/api/main.go
RUN CGO_ENABLED=1 GOOS=linux go build -o seed ./cmd/seed/main.go

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies for sqlite
RUN apk add --no-cache sqlite-libs

# Copy binary from builder
COPY --from=builder /app/main .
COPY --from=builder /app/seed .

# Copy migrations
COPY --from=builder /app/migrations ./migrations

# Expose port
EXPOSE 8080

# Run the binary
CMD ["./main"]
