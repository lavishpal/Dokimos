# ---- Build Stage ----
# Use an official Go runtime as a parent image
FROM golang:1.24.1-alpine AS builder

# Set the working directory in the container
WORKDIR /app
# Copy go.mod and go.sum files to download dependencies
COPY go.mod go.sum ./

RUN go mod download && go mod verify

# Copy the rest of the application's source code
COPY . .

# Build the Go application
# CGO_ENABLED=0 disables Cgo, producing a static binary (good for Alpine)
# GOOS=linux ensures the binary is built for Linux
# -o /app/main specifies the output file path and name

RUN CGO_ENABLED=0 GOOS=linux go build -v -o /tmp/api  ./cmd/api/main.go
# If your main.go is in a subdirectory, e.g., cmd/server/main.go:
# RUN CGO_ENABLED=0 GOOS=linux go build -v -o /app/main ./cmd/server/

# ---- Final Stage ----
# Use a minimal base image for a small footprint
FROM alpine:latest

# Set the working directory
WORKDIR /app

# Copy the pre-built binary from the builder stage
COPY --from=builder /tmp/api  /app/main

# (Optional) If your application needs any static assets or config files, copy them here
# For example, if you have a .env file that needs to be present (though for Docker, env vars are preferred)
# COPY .env .env
# COPY ./config /app/config
# COPY ./static /app/static

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
# The entrypoint allows you to pass arguments to your application if needed
ENTRYPOINT ["/app/main"]