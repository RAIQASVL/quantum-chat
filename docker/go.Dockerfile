FROM golang:1.21-alpine

RUN apk add --no-cache git wget curl

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o main ./cmd/server

EXPOSE 8080

# Health check preparation
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --spider http://localhost:8080/health || exit 1

CMD ["./main"]