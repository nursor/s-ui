# Docker构建指南 - x64 Linux镜像

## 问题说明

构建时遇到 `TLS handshake timeout` 错误，通常是网络问题导致无法从Docker Hub拉取镜像。

## 解决方案

### 方案1: 配置Docker镜像加速器（推荐）⭐

#### macOS Docker Desktop

1. 打开 **Docker Desktop**
2. 点击右上角 **设置图标** (Settings)
3. 选择 **Docker Engine**
4. 在JSON配置中添加以下内容：

```json
{
  "registry-mirrors": [
    "https://docker.mirrors.ustc.edu.cn",
    "https://hub-mirror.c.163.com",
    "https://mirror.baidubce.com"
  ]
}
```

5. 点击 **Apply & Restart** 重启Docker

#### Linux系统

编辑 `/etc/docker/daemon.json`（如果不存在则创建）：

```json
{
  "registry-mirrors": [
    "https://docker.mirrors.ustc.edu.cn",
    "https://hub-mirror.c.163.com",
    "https://mirror.baidubce.com"
  ]
}
```

然后重启Docker服务：

```bash
sudo systemctl daemon-reload
sudo systemctl restart docker
```

### 方案2: 使用构建脚本

使用提供的构建脚本，它会自动处理平台设置：

```bash
./build-docker.sh register.liang.home/library/anydoor:v1.0.0
```

### 方案3: 手动构建（明确指定平台）

```bash
# 使用标准docker build
docker build \
  --platform linux/amd64 \
  -t register.liang.home/library/anydoor:v1.0.0 \
  .

# 或使用buildx（推荐）
docker buildx build \
  --platform linux/amd64 \
  -t register.liang.home/library/anydoor:v1.0.0 \
  --load \
  .
```

### 方案4: 使用代理

如果公司或网络环境需要代理：

```bash
export HTTP_PROXY=http://your-proxy:port
export HTTPS_PROXY=http://your-proxy:port
export NO_PROXY=localhost,127.0.0.1

docker build -t register.liang.home/library/anydoor:v1.0.0 .
```

## 验证配置

配置镜像加速器后，验证是否生效：

```bash
docker info | grep -A 10 "Registry Mirrors"
```

应该能看到配置的镜像源。

## 构建x64 Linux镜像

Dockerfile已经更新，明确指定了 `linux/amd64` 平台：

- ✅ 所有FROM指令都指定了 `--platform=linux/amd64`
- ✅ GOARCH设置为 `amd64`
- ✅ 移除了冗余的平台变量

直接构建即可：

```bash
docker build -t register.liang.home/library/anydoor:v1.0.0 .
```

## 常见问题

### Q: 仍然超时怎么办？

1. **检查网络连接**：确保能访问互联网
2. **尝试其他镜像源**：更换不同的registry-mirrors
3. **使用VPN/代理**：如果在中国大陆，可能需要科学上网
4. **重试**：有时是临时网络问题，多试几次

### Q: 如何验证镜像架构？

```bash
docker inspect register.liang.home/library/anydoor:v1.0.0 | grep Architecture
```

应该显示：`"Architecture": "amd64"`

### Q: 构建后如何测试？

```bash
# 运行容器
docker run -it register.liang.home/library/anydoor:v1.0.0 sh

# 在容器内测试nginx
nginx -t
```

## 国内常用镜像源

- 中科大：`https://docker.mirrors.ustc.edu.cn`
- 网易：`https://hub-mirror.c.163.com`
- 百度云：`https://mirror.baidubce.com`
- 阿里云：需要登录阿里云获取专属加速地址

