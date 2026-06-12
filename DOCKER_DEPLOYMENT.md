# Docker 部署指南

## 快速启动

### 使用 PostgreSQL（推荐）

1. **克隆仓库**
```bash
git clone https://github.com/YOUR_USERNAME/axonhub.git
cd axonhub
```

2. **配置环境变量**
```bash
# 创建 .env 文件
cat > .env << EOF
DB_PASSWORD=your_secure_password
EOF
```

3. **启动服务**
```bash
docker-compose up -d
```

4. **查看日志**
```bash
docker-compose logs -f axonhub
```

5. **访问服务**
- 前端界面：http://localhost:8090
- GraphQL API：http://localhost:8090/graphql
- 健康检查：http://localhost:8090/health

### 使用 SQLite（简单部署）

1. **创建数据目录**
```bash
mkdir -p data
chmod 777 data
```

2. **修改 docker-compose.yml**
```bash
# 注释掉 postgres 和 axonhub 服务
# 取消注释 axonhub-sqlite 服务
```

3. **启动服务**
```bash
docker-compose up -d axonhub-sqlite
```

---

## 从源码构建

### 方式一：使用 Docker

```bash
# 构建镜像
docker build -t axonhub:latest .

# 运行容器
docker run -d \
  -p 8090:8090 \
  -e AXONHUB_DB_DIALECT=sqlite3 \
  -e AXONHUB_DB_DSN="file:/data/axonhub.db?cache=shared&_fk=1" \
  -v $(pwd)/data:/data \
  --name axonhub \
  axonhub:latest serve
```

### 方式二：使用 GitHub Container Registry

推送到 GitHub 后，Actions 会自动构建并推送镜像：

```bash
# 拉取镜像
docker pull ghcr.io/YOUR_USERNAME/axonhub:main

# 使用镜像
docker run -d \
  -p 8090:8090 \
  -e AXONHUB_DB_DIALECT=sqlite3 \
  -e AXONHUB_DB_DSN="file:/data/axonhub.db?cache=shared&_fk=1" \
  -v $(pwd)/data:/data \
  ghcr.io/YOUR_USERNAME/axonhub:main serve
```

---

## 配置说明

### 环境变量

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| `AXONHUB_DB_DIALECT` | 数据库类型：`sqlite3`/`postgres`/`mysql` | `sqlite3` |
| `AXONHUB_DB_DSN` | 数据库连接字符串 | - |
| `AXONHUB_SERVER_PORT` | 服务端口 | `8090` |
| `AXONHUB_LOG_LEVEL` | 日志级别：`debug`/`info`/`warn`/`error` | `info` |

### PostgreSQL DSN 示例
```
postgres://username:password@host:5432/database?sslmode=disable
```

### MySQL DSN 示例
```
username:password@tcp(host:3306)/database?charset=utf8mb4&parseTime=True&loc=Local
```

### SQLite DSN 示例
```
file:/data/axonhub.db?cache=shared&_fk=1&_pragma=journal_mode(WAL)
```

---

## 高级配置

### 使用配置文件

创建 `config.yml`：

```yaml
server:
  port: 8090
  
database:
  dialect: postgres
  dsn: postgres://axonhub:password@postgres:5432/axonhub?sslmode=disable
  
log:
  level: info
  format: json
```

挂载到容器：

```yaml
volumes:
  - ./config.yml:/app/config.yml:ro
```

---

## 数据迁移

### 自动迁移（推荐）

容器启动时会自动运行 `migrate up`，无需手动干预。

### 手动迁移

```bash
# 升级到最新版本
docker exec axonhub-app /app/axonhub migrate up

# 回滚一个版本
docker exec axonhub-app /app/axonhub migrate down 1

# 查看迁移状态
docker exec axonhub-app /app/axonhub migrate status
```

---

## 生产环境部署

### 1. 使用 HTTPS

```nginx
server {
    listen 443 ssl http2;
    server_name axonhub.example.com;
    
    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;
    
    location / {
        proxy_pass http://127.0.0.1:8090;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### 2. 备份数据库

**PostgreSQL**
```bash
docker exec axonhub-postgres pg_dump -U axonhub axonhub > backup.sql
```

**SQLite**
```bash
docker exec axonhub-sqlite-app sqlite3 /data/axonhub.db .dump > backup.sql
```

### 3. 恢复数据库

**PostgreSQL**
```bash
cat backup.sql | docker exec -i axonhub-postgres psql -U axonhub axonhub
```

**SQLite**
```bash
cat backup.sql | docker exec -i axonhub-sqlite-app sqlite3 /data/axonhub.db
```

### 4. 监控和日志

```bash
# 查看容器状态
docker-compose ps

# 查看实时日志
docker-compose logs -f

# 查看资源使用
docker stats axonhub-app

# 健康检查
curl http://localhost:8090/health
```

---

## 常见问题

### 1. 容器启动失败

检查日志：
```bash
docker-compose logs axonhub
```

### 2. 数据库连接失败

确认数据库容器已启动并健康：
```bash
docker-compose ps
docker-compose logs postgres
```

### 3. 前端页面无法访问

检查端口映射：
```bash
docker-compose ps | grep 8090
```

### 4. 修改配置后不生效

重启容器：
```bash
docker-compose restart axonhub
```

---

## 更新升级

```bash
# 拉取最新镜像
docker-compose pull

# 重启服务
docker-compose up -d

# 清理旧镜像
docker image prune -f
```

---

## 卸载

```bash
# 停止并删除容器
docker-compose down

# 删除数据卷（会丢失所有数据）
docker-compose down -v

# 删除镜像
docker rmi looplj/axonhub:latest
```
