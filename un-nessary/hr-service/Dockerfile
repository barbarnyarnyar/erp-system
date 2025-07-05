# services/hr/Dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY shared/ ./shared/
COPY services/hr-service/ ./
RUN go mod tidy && go build -o main cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
EXPOSE 8002
CMD ["./main"]