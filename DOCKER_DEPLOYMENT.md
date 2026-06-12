# Docker 部署完全指南

本文档提供 AxonHubPlus 的完整 Docker 部署方案，包括开发、测试和生产环境的最佳实践。

---

## 📋 目录

- [前置要求](#-前置要求)
- [快速开始](#-快速开始)
- [部署方案](#-部署方案)
- [数据库配置](#-数据库配置)
- [环境变量](#-环境变量)
- [生产环境部署](#-生产环境部署)
- [运维操作](#-运维操作)
- [常见问题](#-常见问题)

---

## 📦 前置要求

| 软件 | 最低版本 | 推荐版本 |
|------|----------|----------|
| Docker | 20.10+ | 24.0+ |
| Docker Compose | 2.0+ | 2.20+ |
| 可用内存 | 2GB | 4GB+ |
| 可用磁盘 | 5GB | 10GB+ |

**检查安装**：
```bash
docker --version
docker-compose --version
```

---

## 🚀 快速开始

### 30 秒启动

```bash
# 1. 克隆仓库
git clone https://github.com/ifsherlock/axonhubplus.git
cd axonhubplus

# 2. 启动服务
docker-compose up -d

# 3. 等待服务就绪（约 30-60 秒）
docker-compose logs -f axonhub

# 4. 访问系统
open http://localhost:8090
```

**预期输出**：
```
axonhub-1  | Starting AxonHub server on :8090
axonhub-1  | Database migration completed successfully
axonhub-1  | Server is ready to accept connections
```

✅ 访问 http://localhost:8090 开始使用！

---

## 📚 部署方案

### 方案一：Docker Compose + PostgreSQL（推荐生产环境）

**适用场景**：生产环境、多用户、大流量

#### 1. 使用默认配置

```bash
# 启动 PostgreSQL + AxonHub
docker-compose up -d

# 查看服务状态
docker-compose ps
```

输出示例：
```
NAME                    IMAGE                    STATUS
axonhubplus-axonhub-1   axonhub:latest          Up (healthy)
axonhubplus-postgres-1  postgres:17-alpine      Up (healthy)
```

#### 2. 自定义配置

创建 `.env` 文件：
```bash
# 数据库密码
DB_PASSWORD=your_secure_password_here

# 服务端口（可选）
AXONHUB_PORT=8090
```

修改 `docker-compose.yml` 中的数据库配置：
```yaml
environment:
  POSTGRES_PASSWORD: ${DB_PASSWORD}
  POSTGRES_USER: axonhub
  POSTGRES_DB: axonhub
```

#### 3. 查看日志

```bash
# 所有服务日志
docker-compose logs -f

# 只看 AxonHub 日志
docker-compose logs -f axonhub

# 只看最近 100 行
docker-compose logs --tail=100 axonhub
```

#### 4. 管理服务

```bash
# 停止服务（数据保留）
docker-compose stop

# 重启服务
docker-compose restart

# 停止并删除容器（数据保留在 volume）
docker-compose down

# 完全清理（删除所有数据）
docker-compose down -v
```

---

### 方案二：Docker Compose + SQLite（适合开发/小型部署）

**适用场景**：个人使用、开发测试、轻量部署

#### 1. 修改配置

编辑 `docker-compose.yml`，修改 axonhub 服务的环境变量：
```yaml
services:
  axonhub:
    environment:
      - AXONHUB_DB_DIALECT=sqlite3
      - AXONHUB_DB_DSN=file:/data/axonhub.db?cache=shared&_fk=1
    volumes:
      - ./data:/data
```

#### 2. 启动服务

```bash
# 创建数据目录
mkdir -p data

# 启动（无需 PostgreSQL）
docker-compose up -d axonhub

# 查看日志
docker-compose logs -f axonhub
```

---

### 方案三：使用 GitHub Container Registry 镜像

**适用场景**：快速部署、生产环境

等待 GitHub Actions 构建完成后（约 10-15 分钟）：

#### PostgreSQL 版本

```bash
# 1. 启动 PostgreSQL
docker run -d \
  --name axonhub-postgres \
  -e POSTGRES_USER=axonhub \
  -e POSTGRES_PASSWORD=your_password \
  -e POSTGRES_DB=axonhub \
  -v axonhub-postgres-data:/var/lib/postgresql/data \
  postgres:17-alpine

# 2. 启动 AxonHub
docker run -d \
  --name axonhub \
  -p 8090:8090 \
  -e AXONHUB_DB_DIALECT=postgres \
  -e AXONHUB_DB_DSN="postgres://axonhub:your_password@axonhub-postgres:5432/axonhub?sslmode=disable" \
  --link axonhub-postgres:postgres \
  ghcr.io/ifsherlock/axonhubplus:main serve

# 3. 查看日志
docker logs -f axonhub
```

#### SQLite 版本（最简单）

```bash
# 创建数据目录
mkdir -p data

# 一行命令启动
docker run -d \
  --name axonhub \
  -p 8090:8090 \
  -e AXONHUB_DB_DIALECT=sqlite3 \
  -e AXONHUB_DB_DSN="file:/data/axonhub.db?cache=shared&_fk=1" \
  -v $(pwd)/data:/data \
  ghcr.io/ifsherlock/axonhubplus:main serve

# 查看日志
docker logs -f axonhub
```

---

### 方案四：从源码构建（开发环境）

```bash
# 1. 构建镜像
docker build -t axonhub:dev .

# 2. 运行
docker run -d \
  --name axonhub-dev \
  -p 8090:8090 \
  -e AXONHUB_DB_DIALECT=sqlite3 \
  -e AXONHUB_DB_DSN="file:/data/axonhub.db?cache=shared&_fk=1" \
  -v $(pwd)/data:/data \
  axonhub:dev serve
```

---

## 🗄️ 数据库配置

### PostgreSQL（推荐生产环境）

**优点**：性能强、并发好、功能完整

**DSN 格式**：
```
postgres://username:password@host:port/database?sslmode=disable
```

**Docker Compose 配置**：
```yaml
services:
  postgres:
    image: postgres:17-alpine
    environment:
      POSTGRES_USER: axonhub
      POSTGRES_PASSWORD: ${DB_PASSWORD}
      POSTGRES_DB: axonhub
    volumes:
      - postgres-data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U axonhub"]
      interval: 10s
      timeout: 5s
      retries: 5

volumes:
  postgres-data:
```

**连接参数**：
- `sslmode=disable` - 开发环境（生产建议 `sslmode=require`）
- `pool_max_conns=100` - 最大连接数
- `pool_min_conns=10` - 最小连接数

---

### MySQL

**优点**：兼容性好、生态丰富

**DSN 格式**：
```
username:password@tcp(host:3306)/database?charset=utf8mb4&parseTime=True&loc=Local
```

**Docker Compose 配置**：
```yaml
services:
  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: ${DB_PASSWORD}
      MYSQL_DATABASE: axonhub
      MYSQL_USER: axonhub
      MYSQL_PASSWORD: ${DB_PASSWORD}
    volumes:
      - mysql-data:/var/lib/mysql
    command: --default-authentication-plugin=mysql_native_password

  axonhub:
    environment:
      - AXONHUB_DB_DIALECT=mysql
      - AXONHUB_DB_DSN=axonhub:${DB_PASSWORD}@tcp(mysql:3306)/axonhub?charset=utf8mb4&parseTime=True

volumes:
  mysql-data:
```

---

### SQLite（开发/小型部署）

**优点**：零配置、单文件、轻量

**DSN 格式**：
```
file:/data/axonhub.db?cache=shared&_fk=1&_pragma=journal_mode(WAL)
```

**Docker 配置**：
```yaml
services:
  axonhub:
    environment:
      - AXONHUB_DB_DIALECT=sqlite3
      - AXONHUB_DB_DSN=file:/data/axonhub.db?cache=shared&_fk=1
    volumes:
      - ./data:/data
```

**参数说明**：
- `cache=shared` - 多连接共享缓存
- `_fk=1` - 启用外键约束
- `_pragma=journal_mode(WAL)` - 启用 WAL 模式（提升并发）

---

## ⚙️ 环境变量

### 核心配置

| 变量名 | 说明 | 默认值 | 示例 |
|--------|------|--------|------|
| `AXONHUB_DB_DIALECT` | 数据库类型 | `postgres` | `postgres` / `mysql` / `sqlite3` |
| `AXONHUB_DB_DSN` | 数据库连接串 | - | `postgres://user:pass@host:5432/db` |
| `AXONHUB_SERVER_PORT` | HTTP 端口 | `8090` | `8090` |
| `AXONHUB_SERVER_HOST` | 监听地址 | `0.0.0.0` | `0.0.0.0` / `127.0.0.1` |

### 日志配置

| 变量名 | 说明 | 默认值 | 可选值 |
|--------|------|--------|--------|
| `AXONHUB_LOG_LEVEL` | 日志级别 | `info` | `debug` / `info` / `warn` / `error` |
| `AXONHUB_LOG_FORMAT` | 日志格式 | `json` | `json` / `console` |

### 性能配置

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| `AXONHUB_DB_MAX_OPEN_CONNS` | 最大数据库连接数 | `100` |
| `AXONHUB_DB_MAX_IDLE_CONNS` | 最大空闲连接数 | `10` |
| `AXONHUB_DB_CONN_MAX_LIFETIME` | 连接最大存活时间 | `1h` |

### 完整示例

```bash
# .env 文件
DB_PASSWORD=your_secure_password
AXONHUB_DB_DIALECT=postgres
AXONHUB_DB_DSN=postgres://axonhub:your_secure_password@postgres:5432/axonhub?sslmode=disable
AXONHUB_SERVER_PORT=8090
AXONHUB_LOG_LEVEL=info
AXONHUB_LOG_FORMAT=json
```

---
