# Build Stage
FROM golang:1.24-alpine AS builder

# Set environment variables
ENV GO111MODULE=on \
    CGO_ENABLED=1 \
    GOOS=linux \
    GOARCH=amd64

# Install build dependencies (git, gcc, musl-dev required for CGO/SQLite)
RUN apk add --no-cache git gcc musl-dev

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN go build -o git-manage-service .

# Runtime Stage
FROM alpine:latest

# Install runtime dependencies
# git: for git operations
# openssh-client: for ssh git access
# ca-certificates: for https git access
# tzdata: for timezone setting
RUN apk add --no-cache \
    git \
    openssh-client \
    ca-certificates \
    tzdata

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/git-manage-service .

# Copy frontend assets
COPY --from=builder /app/public ./public

# Copy swagger docs
COPY --from=builder /app/docs ./docs

# Set environment variables
ENV GIN_MODE=release \
    PORT=8080 \
    DB_PATH=/app/data/git_sync.db

# Expose port
EXPOSE 8080

# Create volume directories
RUN mkdir -p /app/data /root/.ssh && chmod 700 /root/.ssh

# Start command
CMD ["./git-manage-service"]
