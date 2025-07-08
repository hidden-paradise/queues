# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY ./consumer ./consumer
COPY ./internal ./internal

RUN go build -o consumer-bin ./consumer

FROM alpine:latest

COPY --from=builder /app/consumer-bin /usr/local/bin/consumer

EXPOSE 9100

ENTRYPOINT ["/usr/local/bin/consumer"]