# Build stage
FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/api ./cmd/api

# Runtime stage
FROM alpine:3.18

RUN apk add --no-cache ca-certificates curl

WORKDIR /app

COPY --from=builder /app/bin/api .

EXPOSE 8080

HEALTHCHECK --interval=10s --timeout=5s --retries=5 \
  CMD curl -f http://localhost:8080/health || exit 1

CMD ["./api"]
