# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /
COPY shared/ ./shared/

WORKDIR /app
COPY services/fm-service/ ./

# Build the application
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/server/main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Expose port
EXPOSE 8001

# Run the application
CMD ["./main"]