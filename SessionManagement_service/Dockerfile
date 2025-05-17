FROM golang:1.23.4 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o session_service ./cmd/main.go

FROM alpine:latest


WORKDIR /root/session_service


COPY --from=builder /app/session_service .
COPY --from=builder /app/internal/configs ./internal/configs/


CMD ["./session_service"]