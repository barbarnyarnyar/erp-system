# api-gateway/Dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY shared/ ./shared/
COPY api-gateway/ ./
RUN go mod tidy && go build -o main cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
EXPOSE 8080
CMD ["./main"]