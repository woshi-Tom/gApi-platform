FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB=off

COPY go.mod go.sum* ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o gapi-server ./cmd/server

FROM alpine:3.19

RUN apk add --no-cache ca-certificates curl tzdata

WORKDIR /app

COPY --from=builder /app/gapi-server .
COPY --from=builder /app/docs ./docs

RUN mkdir -p /app/config /app/logs

RUN echo 'database:\n  host: ${DB_HOST:-localhost}\n  port: ${DB_PORT:-5432}\n  user: ${DB_USER:-gapi}\n  password: ${DB_PASSWORD:-}\n  name: ${DB_NAME:-gapi}\n\nredis:\n  host: ${REDIS_HOST:-localhost}\n  port: ${REDIS_PORT:-6379}\n  password: ${REDIS_PASSWORD:-}\n\nrabbitmq:\n  host: ${RABBITMQ_HOST:-localhost}\n  port: ${RABBITMQ_PORT:-5672}\n  user: ${RABBITMQ_USER:-guest}\n  password: ${RABBITMQ_PASSWORD:-guest}\n\nserver:\n  port: ${SERVER_PORT:-8080}\n  frontend: ${FRONTEND_URL:-http://localhost:5173}\n  admin_frontend: ${ADMIN_FRONTEND_URL:-http://localhost:5174}' > /app/config/config.yaml

EXPOSE 8080

ENV TZ=Asia/Shanghai

HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

ENTRYPOINT ["./gapi-server"]
