# Build stage
FROM golang:1.24-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-s -w" -o rummage ./cmd/rummage

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Set working directory
WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/rummage .

# Expose the port
EXPOSE 8080

# Set environment variables
ENV RUMMAGE_SERVER_PORT=8080
ENV RUMMAGE_REDIS_URL=redis://redis:6379
ENV RUMMAGE_SERVER_BASEURL=http://localhost:8080

# Create config directory
RUN mkdir -p /app/config
WORKDIR /app

# Copy config file
COPY config/ /app/config/

# Command to run the executable
CMD ["/root/rummage"]
