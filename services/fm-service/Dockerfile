FROM golang:1.21-alpine AS builder

# Install git (needed for go mod download)
RUN apk add --no-cache git

WORKDIR /app

# Copy go mod files
COPY go.mod ./
COPY go.sum* ./

# Install swag CLI for generating docs
RUN go install github.com/swaggo/swag/cmd/swag@latest

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Generate Swagger docs
RUN swag init -g cmd/main.go -o docs/

# Download dependencies again after docs generation (in case docs import new packages)
RUN go mod tidy

# Build the application
RUN go build -o main ./cmd

FROM alpine:latest
WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Expose port
EXPOSE 8001

# Run the application
CMD ["./main"]