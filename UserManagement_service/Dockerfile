FROM golang:1.23.4 AS builder

WORKDIR /app


COPY UserManagement_service/go.mod UserManagement_service/go.sum ./
COPY SessionManagement_service/go.mod SessionManagement_service/go.sum ./SessionManagement_service/


RUN go mod edit -replace github.com/niktin06sash/MicroserviceProject/SessionManagement_service=/app/SessionManagement_service


RUN go mod download


COPY . .


WORKDIR /app/UserManagement_service
RUN CGO_ENABLED=0 GOOS=linux go build -o user_service ./cmd/main.go


FROM alpine:latest

WORKDIR /root/
COPY --from=builder /app/UserManagement_service/user_service .
COPY --from=builder /app/UserManagement_service/internal/configs ./internal/configs/

CMD ["./user_service"]