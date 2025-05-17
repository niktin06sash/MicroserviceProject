FROM golang:1.23.4 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o logs_service ./cmd/main.go

FROM alpine:latest


WORKDIR /root/logs_service


COPY --from=builder /app/logs_service .
COPY --from=builder /app/internal/configs ./internal/configs/

RUN mkdir -p internal/logs/api-service && \
    mkdir -p internal/logs/user-service && \
    mkdir -p internal/logs/session-service && \
    touch internal/logs/api-service/info.log && \
    touch internal/logs/api-service/error.log && \
    touch internal/logs/api-service/warn.log && \
    touch internal/logs/user-service/info.log && \
    touch internal/logs/user-service/error.log && \
    touch internal/logs/user-service/warn.log && \
    touch internal/logs/session-service/info.log && \
    touch internal/logs/session-service/error.log && \
    touch internal/logs/session-service/warn.log

VOLUME /root/logs_service/internal/logs
CMD ["./logs_service"]