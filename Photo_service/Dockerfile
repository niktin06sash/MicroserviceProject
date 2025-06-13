FROM golang 1.23.4 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod.download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o photo_service ./cmd/main.go
FROM alpine:latest
WORKDIR /root/photo_service
COPY --from=builder /app/photo_service
COPY --from=builder /app/photo_service/configs ./internal/configs
CMD ["./photo_service"]