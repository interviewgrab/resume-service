# Use the official Golang image as the base image
FROM golang:1.20.2 as builder

# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum files to the workspace
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

FROM alpine:latest

# Set the working directory
WORKDIR /app

COPY --from=builder /app /app/

# Expose the port used by the application
EXPOSE 8080

# Run the application
CMD ["/app/main"]
