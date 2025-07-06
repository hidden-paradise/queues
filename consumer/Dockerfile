# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Кэшируем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходники
COPY ./consumer ./consumer
COPY ./internal ./internal

# Собираем бинарник
RUN go build -o consumer-bin ./consumer

# Final stage
FROM alpine:latest


COPY --from=builder /app/consumer-bin /usr/local/bin/consumer

# Открываем порт для Prometheus
EXPOSE 9100

ENTRYPOINT ["/usr/local/bin/consumer"]