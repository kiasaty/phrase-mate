FROM golang:1.23.3 as builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o phrase-mate .

FROM debian:bookworm-slim

# Install CA certificates
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=builder /app/phrase-mate .
COPY .env .
RUN mkdir -p /app/data

CMD ["./phrase-mate", "fetch-updates"]
