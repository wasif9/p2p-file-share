# Stage 1: Build the Go application
FROM golang:1.24 as builder

# Set the working directory inside the container
WORKDIR /app

# Copy the go.mod and go.sum files to download dependencies
COPY ../go.mod ../go.sum ./

# Download dependencies
RUN go mod download

# Copy the db-manager source code
COPY . .

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux go build -o db-manager ./db-manager

# Stage 2: Create a lightweight image
FROM alpine:latest

# Install ca-certificates for HTTPS requests (if needed)
RUN apk --no-cache add ca-certificates

# Set the working directory
WORKDIR /app/

# Copy the binary from the builder stage
COPY --from=builder /app/db-manager .

# Copy the config.json file into the container
COPY ./db-manager/config.json ./config.json

# Expose any necessary ports (adjust based on your app's needs)
EXPOSE 8082

# Command to run the application
CMD ["./db-manager", "config.json"]