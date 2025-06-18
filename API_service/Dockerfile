FROM golang:1.23.4 AS builder

WORKDIR /app

COPY API_service/go.mod API_service/go.sum ./
COPY SessionManagement_service/go.mod SessionManagement_service/go.sum ./SessionManagement_service/
RUN go mod edit -replace github.com/niktin06sash/MicroserviceProject/SessionManagement_service=/app/SessionManagement_service
COPY Photo_service/go.mod Photo_service/go.sum ./Photo_service/
RUN go mod edit -replace github.com/niktin06sash/MicroserviceProject/Photo_service=/app/Photo_service
RUN go mod download

COPY . .

WORKDIR /app/API_service
RUN CGO_ENABLED=0 GOOS=linux go build -o api_service ./cmd/main.go

FROM alpine:latest


WORKDIR /root/
COPY --from=builder /app/API_service/api_service .
COPY --from=builder /app/API_service/internal/configs ./internal/configs/


CMD ["./api_service"]
