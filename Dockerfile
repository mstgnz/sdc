# Build aşaması
FROM golang:1.23-bullseye AS builder

# Install the necessary build tools
RUN apt-get update && apt-get install -y git

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o sqlporter .

# Run stage
FROM debian:bullseye-slim

# Install necessary packages for SSL certificates
RUN apt-get update && apt-get install -y ca-certificates

WORKDIR /root/

# Copy the built application from the builder stage
COPY --from=builder /app/sqlporter .

# Run the application
CMD ["./sqlporter"] 