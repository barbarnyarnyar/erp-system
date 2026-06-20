# api-gateway/Dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /
COPY shared/ ./shared/
WORKDIR /app
COPY api-gateway/ ./
RUN go mod tidy && go build -o main cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
COPY --from=builder /app/openapi.yaml .
EXPOSE 8080
CMD ["./main"]