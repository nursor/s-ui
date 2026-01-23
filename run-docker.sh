#!/bin/bash
# Docker一键运行脚本 - 构建并运行s-ui容器

set -e

IMAGE_NAME="${1:-s-ui:1.0.4}"
CONTAINER_NAME="s-ui"
PLATFORM="linux/amd64"

echo "=========================================="
echo "Docker一键构建并运行"
echo "=========================================="
echo "镜像名称: $IMAGE_NAME"
echo "容器名称: $CONTAINER_NAME"
echo "目标平台: $PLATFORM"
echo ""

# 检查Docker是否运行
if ! docker info >/dev/null 2>&1; then
    echo "错误: Docker未运行，请启动Docker Desktop"
    exit 1
fi

# 1. 停止并删除旧容器
echo "步骤 1/3: 检查并清理旧容器..."
if docker ps -a --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
    echo "  - 停止旧容器: ${CONTAINER_NAME}"
    docker stop "${CONTAINER_NAME}" 2>/dev/null || true
    echo "  - 删除旧容器: ${CONTAINER_NAME}"
    docker rm "${CONTAINER_NAME}" 2>/dev/null || true
    echo "  ✓ 旧容器已清理"
else
    echo "  ✓ 未发现旧容器"
fi
echo ""

# 2. 构建新镜像
echo "步骤 2/3: 构建Docker镜像..."
if docker buildx version >/dev/null 2>&1; then
    echo "  使用Docker Buildx构建..."
    docker buildx build \
        --platform $PLATFORM \
        --tag $IMAGE_NAME \
        --load \
        --progress=plain \
        .
else
    echo "  使用标准Docker构建..."
    docker build \
        --platform $PLATFORM \
        --tag $IMAGE_NAME \
        --progress=plain \
        .
fi

if [ $? -ne 0 ]; then
    echo ""
    echo "=========================================="
    echo "✗ 构建失败"
    echo "=========================================="
    exit 1
fi

echo "  ✓ 镜像构建成功"
echo ""

# 3. 运行新容器
echo "步骤 3/3: 启动容器..."
docker run --rm \
    --name "${CONTAINER_NAME}" \
    -p 2095:2095 \
    -p 2096:2096 \
    "${IMAGE_NAME}"

# 容器停止后会自动清理（因为使用了 --rm 参数）
echo ""
echo "=========================================="
echo "✓ 容器已停止并自动清理"
echo "=========================================="
