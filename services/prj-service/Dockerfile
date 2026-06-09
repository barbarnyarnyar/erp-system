# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /
COPY shared/ ./shared/

WORKDIR /app
COPY services/pm-service/ ./

# Build the application
RUN go mod tidy && go build -o main ./cmd

# Final stage
FROM alpine:latest

WORKDIR /app

# Copy the binary
COPY --from=builder /app/main .

# Expose port
EXPOSE 8006

# Run the application
CMD ["./main"]