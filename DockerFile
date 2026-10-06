# Build Stage
FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o producer ./cmd/server/main.go

# Runtime Stage
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/producer .

RUN apk add --no-cache ca-certificates

EXPOSE 8080

CMD ["./producer"]