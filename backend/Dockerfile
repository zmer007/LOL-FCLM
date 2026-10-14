# syntax=docker/dockerfile:1

# ---------- 构建阶段：同时编译 api 与 worker ----------
FROM golang:1.26-alpine AS build
WORKDIR /app

# 先拷依赖清单，利用层缓存
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api \
 && CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/worker ./cmd/worker

# ---------- api 运行镜像 ----------
FROM alpine:3.20 AS api
RUN apk add --no-cache ca-certificates tzdata \
 && adduser -D -u 10001 appuser \
 && mkdir -p /data/media && chown appuser:appuser /data/media
WORKDIR /app
COPY --from=build /out/api /app/api
USER appuser
EXPOSE 9090
ENTRYPOINT ["/app/api"]

# ---------- worker 运行镜像（需要 ffmpeg） ----------
FROM alpine:3.20 AS worker
RUN apk add --no-cache ca-certificates tzdata ffmpeg \
 && adduser -D -u 10001 appuser \
 && mkdir -p /data/media && chown appuser:appuser /data/media
WORKDIR /app
COPY --from=build /out/worker /app/worker
USER appuser
ENTRYPOINT ["/app/worker"]
