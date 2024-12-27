# Build stage
FROM golang:1.23-bullseye AS builder

# Install necessary build tools
RUN apt-get update && apt-get install -y \
    git \
    gcc \
    g++ \
    sqlite3 \
    && rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -o sqlmapper ./cmd/sqlmapper

# Final stage
FROM debian:bullseye-slim

# Install runtime dependencies
RUN apt-get update && apt-get install -y \
    ca-certificates \
    sqlite3 \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy the built application from builder stage
COPY --from=builder /app/sqlmapper .
COPY --from=builder /app/examples ./examples

# Set environment variables
ENV PATH="/app:${PATH}"
ENV SQLMAPPER_LOG_LEVEL=info

# Run the application
ENTRYPOINT ["./sqlmapper"]
CMD ["--help"] 