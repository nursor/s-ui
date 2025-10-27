# CI/CD 配置说明

本项目包含三个 GitHub Actions 工作流文件，用于不同的构建和部署需求：

## 1. `.github/workflows/build.yml` - 二进制文件构建

**用途**: 编译多平台二进制可执行文件

**触发条件**:
- 推送到 `feat/liang` 分支
- 创建标签 (如 `v1.0.0`)
- 手动触发 (`workflow_dispatch`)

**构建平台**:
- Linux: amd64, arm64, arm/v7
- Windows: amd64
- macOS: amd64, arm64

**输出**:
- 各平台的二进制文件
- 校验和文件
- 自动创建 GitHub Release (仅标签触发时)

## 2. `.github/workflows/docker.yml` - Docker 镜像构建

**用途**: 构建和推送 Docker 镜像到 GitHub Container Registry

**触发条件**:
- 推送到 `feat/liang` 分支
- 创建标签 (如 `v1.0.0`)

**构建平台**:
- linux/amd64
- linux/arm64
- linux/arm/v7

**输出**:
- 多架构 Docker 镜像
- 推送到 `ghcr.io/用户名/仓库名`

## 3. `.github/workflows/release.yml` - 发布构建

**用途**: 专门用于发布版本的完整构建流程

**触发条件**:
- 创建标签 (如 `v1.0.0`)
- 手动触发 (`workflow_dispatch`)

**功能**:
- 构建前端
- 编译多平台二进制文件
- 创建 GitHub Release

## 使用方法

### 1. 自动构建
- 推送代码到 `feat/liang` 分支会触发构建
- 在 `feat/liang` 分支上创建标签 (如 `v1.0.0`) 会触发构建和发布

### 2. 创建发布
```bash
git tag v1.0.0
git push origin v1.0.0
```

### 3. 手动触发
在 GitHub Actions 页面点击 "Run workflow" 按钮。

## 环境变量

### 数据库配置 (Docker 环境)
```bash
SUI_DB_TYPE=mysql
SUI_DB_HOST=mysql
SUI_DB_PORT=3306
SUI_DB_USER=sui
SUI_DB_PASSWORD=sui123456
SUI_DB_NAME=sui
```

### 本地开发
```bash
# 使用 SQLite (默认)
export SUI_DB_TYPE=sqlite

# 使用 MySQL
export SUI_DB_TYPE=mysql
export SUI_DB_HOST=localhost
export SUI_DB_PORT=3306
export SUI_DB_USER=root
export SUI_DB_PASSWORD=password
export SUI_DB_NAME=sui
```

## 本地测试

### 使用 Docker Compose
```bash
docker-compose up -d
```

### 手动构建
```bash
# 构建前端
cd frontend
npm install
npm run build
cd ..
mv frontend/dist web/html

# 构建后端
go build -ldflags="-w -s" -tags "with_quic,with_grpc,with_utls,with_acme,with_gvisor" -o sui main.go
```

## 注意事项

1. **前端构建**: 所有工作流都会先构建前端，然后将其嵌入到 Go 二进制文件中
2. **交叉编译**: Windows 和 macOS 的交叉编译需要特殊的工具链
3. **Docker 镜像**: 使用多阶段构建，最终镜像基于 Alpine Linux
4. **表前缀**: 所有数据库表都使用 `sui_` 前缀
5. **MySQL 支持**: 支持 MySQL 和 SQLite 两种数据库
