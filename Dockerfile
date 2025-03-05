FROM golang:1.24.0 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN make build

FROM alpine:latest AS api

ENV PORT=8080
ENV APP_ENV="local"
ENV DB_HOST="file-processor-db"
ENV DB_PORT=5432
ENV DB_DATABASE="fileprocessor"
ENV DB_PASSWORD="postgres"
ENV DB_USERNAME="postgres"
ENV DB_SCHEMA="public"
ENV KAFKA_BOOTSTRAP_SERVERS="landoop-kafka:9092"

WORKDIR /root/

COPY --from=builder /app/bin/main .
COPY --from=builder /app/bin/worker .

RUN chmod +x /root/main

RUN apk add --no-cache ca-certificates
RUN apk add --no-cache libc6-compat


RUN wget https://github.com/jwilder/dockerize/releases/download/v0.6.1/dockerize-linux-amd64-v0.6.1.tar.gz && \
  tar -xzvf dockerize-linux-amd64-v0.6.1.tar.gz && \
  mv dockerize /usr/local/bin/


EXPOSE 8080
# CMD in docker compose