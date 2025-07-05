# api-gateway/Dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY shared/ ./shared/
COPY api-gateway/go.mod api-gateway/go.sum ./
RUN go mod download
COPY api-gateway/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/
COPY --from=builder /app/main .
RUN addgroup -g 1001 -S appgroup && adduser -u 1001 -S appuser -G appgroup
USER appuser
EXPOSE 8080
CMD ["./main"]