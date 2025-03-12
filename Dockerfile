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
ENV PORT=8080
ENV REDIS_URL=redis:6379
ENV BASE_URL=http://localhost:8080

# Command to run the executable
CMD ["./rummage"]
