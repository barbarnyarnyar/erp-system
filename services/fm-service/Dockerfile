# services/fm-service/Dockerfile  
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY . .
WORKDIR /app/services/fm-service

RUN go mod tidy
RUN go build -o main cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/services/fm-service/main .
EXPOSE 8001
CMD ["./main"]