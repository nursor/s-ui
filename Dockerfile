# ============================================
# 阶段1: 前端构建
# ============================================
FROM --platform=linux/amd64 node:20-alpine AS frontend-builder

WORKDIR /frontend

# 先复制依赖文件（利用Docker层缓存）
COPY frontend-src/package*.json ./
COPY frontend-src/tsconfig*.json ./
COPY frontend-src/vite.config.mts ./

# 安装依赖
RUN npm ci

# 复制源码并构建
COPY frontend-src/index.html ./
COPY frontend-src/src ./src
COPY frontend-src/public ./public
COPY frontend-src/media ./media

RUN npm run build:vite && \
    ls -lah dist/ && \
    test -f dist/index.html || (echo "前端构建失败" && exit 1)

# ============================================
# 阶段2: 后端构建
# ============================================
FROM --platform=linux/amd64 golang:1.25-alpine AS backend-builder

WORKDIR /app
ARG TARGETARCH=amd64
ENV CGO_ENABLED=1
ENV CGO_CFLAGS="-D_LARGEFILE64_SOURCE"
ENV GOARCH=amd64

# ���装构建依赖
RUN apk update && apk add --no-cache \
    gcc \
    musl-dev \
    libc-dev \
    make \
    git \
    wget \
    unzip \
    bash

ENV CC=gcc

# Go模块依赖（利用缓存）
COPY go.mod go.sum ./
RUN go mod download

# 复制源码
COPY . .

# 从前端构建器复制构建产物到 web/html/
COPY --from=frontend-builder /frontend/dist /app/web/html

# 验证嵌入内容
RUN ls -lah /app/web/html/ && \
    test -f /app/web/html/index.html || (echo "错误: index.html 不存在" && exit 1)

# 构建 Go 应用（会嵌入 web/html/）
RUN go build -ldflags="-w -s" \
    -tags "with_quic,with_grpc,with_utls,with_acme,with_gvisor" \
    -o sui main.go

FROM --platform=linux/amd64 alpine:latest
LABEL org.opencontainers.image.authors="any@gmail.com"
ENV TZ=Asia/Tehran
ENV SUI_DB_TYPE=mysql
ENV SUI_DB_HOST=172.16.238.2
ENV SUI_DB_PORT=30409
ENV SUI_DB_USER=root
ENV SUI_DB_PASSWORD=asd123456
ENV SUI_DB_NAME=sui


WORKDIR /app

RUN apk add --no-cache --update ca-certificates tzdata bash nginx nginx-mod-stream 2>/dev/null || \
    apk add --no-cache --repository=http://dl-cdn.alpinelinux.org/alpine/edge/main nginx-mod-stream || \
    (echo "错误: 无法安装nginx-mod-stream模块" && exit 1)


COPY --from=backend-builder /app/sui /app/
COPY --from=backend-builder /app/nginx.conf /app/
COPY --from=backend-builder /app/nginx-main.conf /etc/nginx/nginx.conf
COPY entrypoint.sh /app/
ENTRYPOINT [ "./entrypoint.sh" ]