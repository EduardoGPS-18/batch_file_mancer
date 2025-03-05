FROM golang:1.20 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /app/bin/performatic-file-processor

FROM alpine:latest AS api

WORKDIR /root/

COPY --from=builder /app/bin/performatic-file-processor .

CMD ["./build/main"]

FROM alpine:latest AS worker

WORKDIR /root/

COPY --from=builder /app/bin/performatic-file-processor .

CMD ["./build/worker"]