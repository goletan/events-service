# Base image
FROM golang:1.20-alpine AS builder

# Set environment variables
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

# Set working directory
WORKDIR /app

# Copy the Go modules files
COPY go.mod go.sum ./

# Copy go.work and all referenced directories
COPY go.work .
COPY config ./config
COPY logger ./logger
COPY observability ./observability
COPY resilience ./resilience
COPY security ./security
COPY services ./services

# Sync Go work dependencies
RUN go work sync

# Build the Events Service binary
RUN go build -o events-service cmd/events-service/main.go

# Final runtime image
FROM alpine:latest
WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/events .

# Expose ports (adjust based on your app's requirements)
EXPOSE 9090 2112

# Start the Events Service
ENTRYPOINT ["./events"]
