#!/bin/sh
# 快速测试nginx配置

echo "=========================================="
echo "Nginx配置测试"
echo "=========================================="
echo ""

# 检查nginx是否安装
if ! command -v nginx >/dev/null 2>&1; then
    echo "✗ nginx未安装"
    exit 1
fi

echo "✓ nginx已安装"
echo ""

# 检查stream模块
echo "检查stream模块..."
if nginx -V 2>&1 | grep -q "stream"; then
    echo "✓ nginx编译时包含stream模块"
else
    echo "✗ nginx编译时不包含stream模块"
fi

# 检查模块文件
if [ -f /usr/lib/nginx/modules/ngx_stream_module.so ]; then
    echo "✓ stream模块文件存在: /usr/lib/nginx/modules/ngx_stream_module.so"
else
    echo "✗ stream模块文件不存在"
    echo "  尝试查找其他位置..."
    find /usr -name "*stream*.so" 2>/dev/null || echo "  未找到stream模块文件"
fi

echo ""

# 测试nginx配置
echo "测试nginx配置文件..."
if nginx -t 2>&1; then
    echo ""
    echo "✓ nginx配置测试通过！"
    exit 0
else
    echo ""
    echo "✗ nginx配置测试失败"
    echo ""
    echo "常见问题："
    echo "1. 检查/etc/nginx/nginx.conf中是否加载了stream模块"
    echo "2. 检查stream模块文件是否存在"
    echo "3. 检查nginx.conf中的路径是否正确"
    exit 1
fi

