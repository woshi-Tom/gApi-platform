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

RUN mkdir -p /app/logs

EXPOSE 8080

ENV TZ=Asia/Shanghai

HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

ENTRYPOINT ["./gapi-server", "-config", "/app/config/config.yaml"]

