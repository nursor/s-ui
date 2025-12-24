#!/bin/sh
# 快速修复nginx stream模块问题

set -e

echo "=========================================="
echo "Nginx Stream模块快速修复脚本"
echo "=========================================="
echo ""

# 检查是否为root
if [ "$(id -u)" -ne 0 ]; then
    echo "错误: 此脚本需要root权限"
    echo "请使用: sudo $0"
    exit 1
fi

# 检测操作系统
if [ -f /etc/alpine-release ]; then
    echo "检测到Alpine Linux"
    OS="alpine"
elif [ -f /etc/debian_version ]; then
    echo "检测到Debian/Ubuntu"
    OS="debian"
elif [ -f /etc/redhat-release ]; then
    echo "检测到RedHat/CentOS"
    OS="redhat"
else
    echo "警告: 未识别的操作系统，尝试通用方法"
    OS="unknown"
fi

# 检查nginx是否已安装
if ! command -v nginx >/dev/null 2>&1; then
    echo "nginx未安装，正在安装..."
    case $OS in
        alpine)
            apk update
            apk add nginx
            ;;
        debian)
            apt-get update
            apt-get install -y nginx
            ;;
        redhat)
            yum install -y nginx
            ;;
    esac
fi

# 检查nginx是否支持stream模块
echo ""
echo "检查nginx是否支持stream模块..."
if nginx -V 2>&1 | grep -q "stream"; then
    echo "✓ nginx已支持stream模块！"
    echo ""
    echo "请确保nginx配置文件正确。"
    echo "测试配置: nginx -t"
    exit 0
fi

echo "✗ nginx不支持stream模块，正在尝试修复..."
echo ""

# 根据操作系统安装stream模块
case $OS in
    alpine)
        echo "尝试安装nginx-mod-stream..."
        if apk add --no-cache nginx-mod-stream 2>/dev/null; then
            echo "✓ 成功安装nginx-mod-stream"
            echo ""
            echo "请在/etc/nginx/nginx.conf文件开头添加："
            echo "load_module /usr/lib/nginx/modules/ngx_stream_module.so;"
            echo ""
            read -p "是否自动添加此配置？(y/n) " -n 1 -r
            echo
            if [[ $REPLY =~ ^[Yy]$ ]]; then
                if [ -f /etc/nginx/nginx.conf ]; then
                    # 检查是否已存在
                    if ! grep -q "load_module.*ngx_stream_module" /etc/nginx/nginx.conf; then
                        sed -i '1i load_module /usr/lib/nginx/modules/ngx_stream_module.so;' /etc/nginx/nginx.conf
                        echo "✓ 已自动添加模块加载配置"
                    else
                        echo "模块加载配置已存在"
                    fi
                else
                    echo "警告: /etc/nginx/nginx.conf 不存在"
                fi
            fi
        else
            echo "✗ 无法从包管理器安装nginx-mod-stream"
            echo ""
            echo "请尝试以下方法："
            echo "1. 使用Docker运行nginx（推荐）"
            echo "2. 从edge仓库安装: apk add nginx nginx-mod-stream --repository=http://dl-cdn.alpinelinux.org/alpine/edge/main"
            echo "3. 从源码编译nginx（见NGINX_INSTALL.md）"
            exit 1
        fi
        ;;
    debian)
        echo "Debian/Ubuntu的nginx通常包含stream模块"
        echo "如果仍然不支持，请安装nginx-full:"
        echo "  apt-get install nginx-full"
        ;;
    redhat)
        echo "RedHat/CentOS的nginx通常包含stream模块"
        echo "如果仍然不支持，请检查nginx版本:"
        echo "  nginx -v"
        echo "建议升级到最新版本"
        ;;
    *)
        echo "无法自动修复，请参考NGINX_INSTALL.md手动安装"
        exit 1
        ;;
esac

echo ""
echo "=========================================="
echo "修复完成！"
echo "=========================================="
echo ""
echo "下一步："
echo "1. 测试nginx配置: nginx -t"
echo "2. 如果测试通过，重载nginx: nginx -s reload"
echo "   或使用systemd: systemctl reload nginx"
echo ""

