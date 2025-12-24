#!/bin/sh
# Alpine Linux nginx安装脚本 - 包含stream模块支持

echo "正在安装支持stream模块的nginx..."

# 方法1: 尝试安装nginx-mod-stream模块（推荐）
if apk add --no-cache nginx-mod-stream 2>/dev/null; then
    echo "✓ 成功安装nginx-mod-stream模块"
    echo "请确保在nginx主配置文件中加载模块："
    echo "load_module /usr/lib/nginx/modules/ngx_stream_module.so;"
    exit 0
fi

# 方法2: 如果方法1失败，使用官方nginx镜像中的nginx二进制文件
echo "方法1失败，尝试使用官方nginx镜像..."

# 创建临时目录
TMP_DIR=$(mktemp -d)
cd "$TMP_DIR"

# 从官方nginx镜像提取nginx二进制文件
echo "从nginx:alpine镜像提取nginx..."
docker create --name nginx-temp nginx:alpine 2>/dev/null || {
    echo "错误: 需要Docker来提取nginx二进制文件"
    echo ""
    echo "或者，您可以："
    echo "1. 使用Docker运行nginx（推荐）"
    echo "2. 从源码编译nginx"
    exit 1
}

docker cp nginx-temp:/usr/sbin/nginx ./nginx
docker cp nginx-temp:/etc/nginx/nginx.conf ./nginx.conf.example
docker rm nginx-temp

# 检查nginx是否支持stream模块
if ./nginx -V 2>&1 | grep -q "stream"; then
    echo "✓ 提取的nginx支持stream模块"
    echo "将nginx复制到/usr/sbin/nginx（需要root权限）"
    sudo cp ./nginx /usr/sbin/nginx || {
        echo "需要root权限来复制nginx"
        echo "请手动执行: sudo cp $TMP_DIR/nginx /usr/sbin/nginx"
    }
else
    echo "警告: 提取的nginx可能不支持stream模块"
fi

# 清理
cd -
rm -rf "$TMP_DIR"

echo ""
echo "安装完成！"
echo "请检查nginx配置: nginx -t"
echo "如果仍有问题，请使用Docker运行nginx"

