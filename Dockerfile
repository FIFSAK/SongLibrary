# Use an official Go runtime as a parent image
FROM golang:1.23-alpine AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod tidy

COPY . .
COPY .env .env

# Build the Go app
RUN go build -o myapp ./cmd

# Start a new stage from scratch
FROM alpine:latest

# Install necessary dependencies for running the Go app (if any)
RUN apk --no-cache add ca-certificates

# Set the Current Working Directory inside the container
WORKDIR /root/

# Copy the compiled binary from the builder image
COPY --from=builder /app/myapp .

# Копируем .env в финальный контейнер
COPY --from=builder /app/.env .

COPY --from=builder /app/migrations ./migrations

# Expose the port the app runs on
EXPOSE 8080

# Command to run the executable
CMD ["./myapp"]
