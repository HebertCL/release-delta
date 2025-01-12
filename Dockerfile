# Use a lightweight Go image
FROM golang:1.23-alpine as builder

# Set the working directory
WORKDIR /app

# Copy the Go modules manifest files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the application source code
COPY . .

# Build the application
RUN go build -o release-delta

# Use a minimal base image for running the app
FROM alpine:latest

# Set the working directory
WORKDIR /app

# Copy the built application from the builder
COPY --from=builder /app .

# Expose the port the application listens on
EXPOSE 8080

# Command to run the application
CMD ["./release-delta"]
