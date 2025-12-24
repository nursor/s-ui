#!/bin/bash
# Docker构建脚本 - 支持x64 Linux架构

set -e

IMAGE_NAME="${1:-register.liang.home/library/anydoor:v1.0.0}"
PLATFORM="linux/amd64"

echo "=========================================="
echo "Docker镜像构建脚本"
echo "=========================================="
echo "镜像名称: $IMAGE_NAME"
echo "目标平台: $PLATFORM (x64 Linux)"
echo ""

# 检查Docker是否运行
if ! docker info >/dev/null 2>&1; then
    echo "错误: Docker未运行，请启动Docker Desktop"
    exit 1
fi

# 检查是否使用buildx（推荐用于多平台构建）
if docker buildx version >/dev/null 2>&1; then
    echo "✓ 检测到Docker Buildx"
    USE_BUILDX=true
else
    echo "使用标准Docker构建"
    USE_BUILDX=false
fi

# 网络问题处理建议
echo ""
echo "如果遇到网络超时问题，可以尝试："
echo "1. 配置Docker镜像加速器（推荐）"
echo "2. 使用代理"
echo "3. 重试构建（网络可能是临时问题）"
echo ""

# 构建命令
if [ "$USE_BUILDX" = true ]; then
    echo "使用Docker Buildx构建..."
    docker buildx build \
        --platform $PLATFORM \
        --tag $IMAGE_NAME \
        --load \
        --progress=plain \
        .
else
    echo "使用标准Docker构建..."
    docker build \
        --platform $PLATFORM \
        --tag $IMAGE_NAME \
        --progress=plain \
        .
fi

if [ $? -eq 0 ]; then
    echo ""
    echo "=========================================="
    echo "✓ 构建成功！"
    echo "=========================================="
    echo ""
    echo "镜像信息:"
    docker images $IMAGE_NAME
    echo ""
    echo "运行镜像:"
    echo "  docker run -it $IMAGE_NAME"
else
    echo ""
    echo "=========================================="
    echo "✗ 构建失败"
    echo "=========================================="
    echo ""
    echo "常见问题解决方案："
    echo ""
    echo "1. 网络超时问题："
    echo "   配置Docker镜像加速器（见下方说明）"
    echo ""
    echo "2. 配置镜像加速器（macOS Docker Desktop）："
    echo "   - 打开 Docker Desktop"
    echo "   - Settings -> Docker Engine"
    echo "   - 添加以下配置："
    echo "     {"
    echo "       \"registry-mirrors\": ["
    echo "         \"https://docker.mirrors.ustc.edu.cn\","
    echo "         \"https://hub-mirror.c.163.com\""
    echo "       ]"
    echo "     }"
    echo "   - 点击 Apply & Restart"
    echo ""
    echo "3. 使用代理："
    echo "   export HTTP_PROXY=http://your-proxy:port"
    echo "   export HTTPS_PROXY=http://your-proxy:port"
    echo ""
    exit 1
fi

