FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY ./producer ./producer
COPY ./internal ./internal

RUN go build -o producer-bin ./producer

FROM alpine:latest

COPY --from=builder /app/producer-bin /usr/local/bin/producer

EXPOSE 9100

ENTRYPOINT ["/usr/local/bin/producer"]