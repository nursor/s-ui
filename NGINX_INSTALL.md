# Nginx Stream模块安装指南 (Alpine Linux)

## 问题
Alpine Linux默认的nginx包不包含stream模块，导致出现错误：
```
nginx: [emerg] unknown directive "stream" in /etc/nginx/nginx.conf:10
```

## 解决方案

### 方案1: 使用Docker运行nginx（推荐）⭐

这是最简单可靠的方法，官方nginx镜像已包含stream模块：

```bash
# 运行nginx容器
docker run -d \
  --name nginx \
  --restart=unless-stopped \
  -p 443:443 \
  -v /path/to/nginx.conf:/etc/nginx/nginx.conf:ro \
  nginx:alpine

# 测试配置
docker exec nginx nginx -t

# 重载配置
docker exec nginx nginx -s reload
```

### 方案2: 在Alpine系统上安装nginx-mod-stream

```bash
# 更新包索引
apk update

# 安装nginx和stream模块
apk add nginx nginx-mod-stream

# 如果nginx-mod-stream不可用，尝试从edge仓库安装
apk add nginx nginx-mod-stream --repository=http://dl-cdn.alpinelinux.org/alpine/edge/main

# 在主配置文件中加载stream模块
# 编辑 /etc/nginx/nginx.conf，在文件开头添加：
# load_module /usr/lib/nginx/modules/ngx_stream_module.so;
```

### 方案3: 从源码编译nginx

如果以上方法都不可用，可以从源码编译：

```bash
# 安装编译依赖
apk add --no-cache \
    gcc \
    make \
    musl-dev \
    pcre-dev \
    zlib-dev \
    openssl-dev \
    linux-headers

# 下载nginx源码
cd /tmp
NGINX_VERSION=1.25.3
wget http://nginx.org/download/nginx-${NGINX_VERSION}.tar.gz
tar -xzf nginx-${NGINX_VERSION}.tar.gz
cd nginx-${NGINX_VERSION}

# 配置编译（包含stream模块）
./configure \
    --prefix=/etc/nginx \
    --sbin-path=/usr/sbin/nginx \
    --modules-path=/usr/lib/nginx/modules \
    --conf-path=/etc/nginx/nginx.conf \
    --error-log-path=/var/log/nginx/error.log \
    --http-log-path=/var/log/nginx/access.log \
    --pid-path=/var/run/nginx.pid \
    --lock-path=/var/run/nginx.lock \
    --http-client-body-temp-path=/var/cache/nginx/client_temp \
    --http-proxy-temp-path=/var/cache/nginx/proxy_temp \
    --http-fastcgi-temp-path=/var/cache/nginx/fastcgi_temp \
    --http-uwsgi-temp-path=/var/cache/nginx/uwsgi_temp \
    --http-scgi-temp-path=/var/cache/nginx/scgi_temp \
    --with-perl_modules_path=/usr/lib/perl5/vendor_perl \
    --user=nginx \
    --group=nginx \
    --with-compat \
    --with-file-aio \
    --with-threads \
    --with-http_addition_module \
    --with-http_auth_request_module \
    --with-http_dav_module \
    --with-http_flv_module \
    --with-http_gunzip_module \
    --with-http_gzip_static_module \
    --with-http_mp4_module \
    --with-http_random_index_module \
    --with-http_realip_module \
    --with-http_secure_link_module \
    --with-http_slice_module \
    --with-http_ssl_module \
    --with-http_stub_status_module \
    --with-http_sub_module \
    --with-http_v2_module \
    --with-mail \
    --with-mail_ssl_module \
    --with-stream \
    --with-stream_ssl_module \
    --with-stream_ssl_preread_module \
    --with-stream_realip_module \
    --with-stream_geoip_module \
    --with-stream_geoip_module=dynamic

# 编译和安装
make
make install

# 创建nginx用户
addgroup -g 101 -S nginx 2>/dev/null || true
adduser -S -D -H -u 101 -h /var/cache/nginx -s /sbin/nologin -G nginx -g nginx nginx 2>/dev/null || true
```

## 验证安装

安装完成后，验证nginx是否支持stream模块：

```bash
# 检查nginx版本和编译选项
nginx -V 2>&1 | grep -i stream

# 应该看到类似输出：
# --with-stream
# --with-stream_ssl_preread_module
```

## 配置nginx

1. 确保主配置文件 `/etc/nginx/nginx.conf` 包含stream模块加载（如果需要）：
```nginx
load_module /usr/lib/nginx/modules/ngx_stream_module.so;

events {
    worker_connections 1024;
}

# 包含你的stream配置
include /path/to/nginx.conf;  # 或者直接在这里添加stream块
```

2. 测试配置：
```bash
nginx -t
```

3. 启动/重载nginx：
```bash
# 启动
nginx

# 或使用systemd
systemctl start nginx
systemctl enable nginx

# 重载配置（不中断服务）
nginx -s reload
# 或
systemctl reload nginx
```

## 快速检查脚本

运行以下命令检查当前nginx是否支持stream：

```bash
nginx -V 2>&1 | grep -q "stream" && echo "✓ nginx支持stream模块" || echo "✗ nginx不支持stream模块"
```

