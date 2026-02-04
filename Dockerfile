# ============================================
# 阶段1: 前端构建
# ============================================
FROM node:20-alpine AS frontend-builder

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
FROM golang:1.25-alpine AS backend-builder

WORKDIR /app
ARG TARGETARCH
ENV CGO_ENABLED=1
ENV CGO_CFLAGS="-D_LARGEFILE64_SOURCE"
ENV GOARCH=${TARGETARCH}
ENV GOPROXY=https://goproxy.cn,direct

# 装构建依赖
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.tuna.tsinghua.edu.cn/g' /etc/apk/repositories
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

# 先从前端构建器复制构建产物到临时位置
COPY --from=frontend-builder /frontend/dist /tmp/frontend-dist

# 复制源码（但排除 web/html 目录，因为我们要用前端构建的）
COPY . .

# 将前端构建产物移动到正确位置（在 go build 之前！）
RUN mkdir -p /app/web/html && \
    cp -r /tmp/frontend-dist/* /app/web/html/ && \
    ls -lah /app/web/html/ && \
    test -f /app/web/html/index.html || (echo "错误: index.html 不存在" && exit 1)

# 构建 Go 应用（现在 go:embed 可以正确嵌入 web/html/ 的内容）
RUN go build -v -ldflags="-w -s" \
    -tags "with_quic,with_grpc,with_utls,with_acme,with_gvisor" \
    -o sui main.go

FROM alpine:latest
LABEL org.opencontainers.image.authors="any@gmail.com"
ENV TZ=Asia/Tehran
ENV SUI_DB_TYPE=mysql
ENV SUI_DB_HOST=172.16.238.2
ENV SUI_DB_PORT=30409
ENV SUI_DB_USER=root
ENV SUI_DB_PASSWORD=asd123456
ENV SUI_DB_NAME=sui


WORKDIR /app
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.tuna.tsinghua.edu.cn/g' /etc/apk/repositories
RUN apk add --no-cache --update ca-certificates tzdata bash nginx nginx-mod-stream certbot certbot-nginx openssl 2>/dev/null || \
    apk add --no-cache --repository=http://dl-cdn.alpinelinux.org/alpine/edge/main nginx-mod-stream || \
    (echo "错误: 无法安装nginx-mod-stream模块" && exit 1)

# Set up auto-renewal cron job (every day at 3am)
RUN echo "0 3 * * * certbot renew --quiet --post-hook 'nginx -s reload'" >> /var/spool/cron/crontabs/root



COPY --from=backend-builder /app/sui /app/
COPY --from=backend-builder /app/nginx.conf /app/
COPY --from=backend-builder /app/nginx-main.conf /etc/nginx/nginx.conf
COPY entrypoint.sh /app/
ENTRYPOINT [ "./entrypoint.sh" ]