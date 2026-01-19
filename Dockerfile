FROM golang:1.23.5-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o crawler cmd/crawler/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/crawler .
COPY .env .env

# CA certificates for HTTPS
RUN apk --no-cache add ca-certificates

CMD ["./crawler"]
