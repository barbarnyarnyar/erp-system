# services/auth-service/Dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy the entire project context first
COPY . .

# Change to the auth service directory
WORKDIR /app/services/auth-service

# Initialize and download dependencies
RUN go mod tidy
RUN go build -o main cmd/main.go

# Final stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/services/auth-service/main .
EXPOSE 8000
CMD ["./main"]