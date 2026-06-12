# AxonHub - AI Gateway with Response Protection

AxonHub 是一个功能强大的 AI 网关，支持多渠道负载均衡、智能切换、请求保护和响应过滤。

## ✨ 核心特性

- 🚀 **多渠道管理**：统一管理 OpenAI、Anthropic、Google 等多个 LLM 提供商
- 🔄 **智能切换**：支持自动 failover 和负载均衡
- 🛡️ **内容保护**：
  - **Prompt Protection**：保护敏感的输入提示词
  - **Response Protection**：过滤、脱敏或拦截不当响应内容
- 📊 **使用统计**：详细的请求追踪和用量分析
- 🎯 **项目隔离**：多项目、多团队权限管理
- 🔌 **兼容 OpenAI API**：无缝迁移现有应用

## 🚀 快速开始

### Docker 部署（推荐）

```bash
# 克隆仓库
git clone https://github.com/YOUR_USERNAME/axonhub.git
cd axonhub

# 启动服务（PostgreSQL）
docker-compose up -d

# 访问前端
open http://localhost:8090
```

默认账号：`admin` / `admin`（首次启动时设置）

详细部署文档：[DOCKER_DEPLOYMENT.md](./DOCKER_DEPLOYMENT.md)

### 本地开发

**后端**
```bash
# 安装依赖
go mod download

# 运行数据库迁移
go run ./cmd/axonhub migrate up

# 启动后端
go run ./cmd/axonhub serve
```

**前端**
```bash
cd frontend
pnpm install
pnpm dev
```

## 📖 响应保护功能

响应保护（Response Protection）是 AxonHub 的核心安全特性，支持三种处理模式：

### 1. Mask（脱敏）
将匹配的敏感内容替换为指定文本：
```json
{
  "name": "过滤广告域名",
  "pattern": "(?i)dc\.hhhl\.cc",
  "action": "mask",
  "replacement": "[已过滤]"
}
```

### 2. Reject（拒绝）
直接拒绝包含敏感内容的响应：
```json
{
  "name": "拒绝暴力内容",
  "pattern": "(?i)(violence|attack)",
  "action": "reject"
}
```

### 3. Failover（切换渠道）
检测到特定内容时自动切换到备用渠道：
```json
{
  "name": "检测上游拦截",
  "pattern": "(?i)REQUEST_BLOCKED",
  "action": "failover"
}
```

**使用场景**：
- 🚫 过滤广告推广链接
- 🔒 拦截敏感信息泄露
- 🔄 自动处理上游内容审核拦截
- 📝 统一响应格式和内容规范

在前端界面：**Admin > Response Protection** 创建和管理规则。

## 🏗️ 架构

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │
       ▼
┌─────────────────────────────────────┐
│         AxonHub Gateway             │
│  ┌───────────────────────────────┐  │
│  │   Prompt Protection           │  │  ← 输入过滤
│  └───────────────────────────────┘  │
│  ┌───────────────────────────────┐  │
│  │   Channel Selector            │  │  ← 渠道选择
│  └───────────────────────────────┘  │
│  ┌───────────────────────────────┐  │
│  │   Response Protection         │  │  ← 输出过滤
│  └───────────────────────────────┘  │
└─────────────────────────────────────┘
       │         │         │
       ▼         ▼         ▼
   ┌─────┐   ┌─────┐   ┌─────┐
   │ Ch1 │   │ Ch2 │   │ Ch3 │
   └─────┘   └─────┘   └─────┘
```

## 📊 技术栈

**后端**
- Go 1.23+
- Ent (ORM)
- GraphQL (gqlgen)
- Gin (HTTP framework)

**前端**
- React 19
- TypeScript
- TanStack Router & Query
- Tailwind CSS
- shadcn/ui

**数据库**
- PostgreSQL (推荐)
- MySQL
- SQLite

## 🔧 配置

### 环境变量

```bash
# 数据库
AXONHUB_DB_DIALECT=postgres
AXONHUB_DB_DSN=postgres://user:pass@localhost:5432/axonhub

# 服务端口
AXONHUB_SERVER_PORT=8090

# 日志
AXONHUB_LOG_LEVEL=info
```

### 配置文件

创建 `config.yml`：

```yaml
server:
  port: 8090
  
database:
  dialect: postgres
  dsn: postgres://axonhub:password@localhost:5432/axonhub
  
log:
  level: info
  format: json
```

## 🛠️ 开发

### 代码生成

```bash
# 生成 GraphQL 代码
cd internal/server/gql
go generate

# 生成 Ent 代码
go generate ./internal/ent
```

### 运行测试

```bash
# 后端测试
go test ./...

# 前端测试
cd frontend
pnpm test
```

### 构建

```bash
# 构建 Docker 镜像
docker build -t axonhub:latest .

# 本地构建
go build -o axonhub ./cmd/axonhub
```

## 📝 API 文档

AxonHub 兼容 OpenAI API 格式：

```bash
curl http://localhost:8090/api/v1/chat/completions \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

GraphQL Playground: http://localhost:8090/graphql

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

1. Fork 本仓库
2. 创建功能分支：`git checkout -b feature/amazing-feature`
3. 提交更改：`git commit -m 'Add amazing feature'`
4. 推送分支：`git push origin feature/amazing-feature`
5. 提交 Pull Request

## 📄 License

MIT License - 详见 [LICENSE](./LICENSE) 文件

## 🙏 致谢

- [Ent](https://entgo.io/) - 强大的 Go ORM
- [gqlgen](https://gqlgen.com/) - GraphQL 服务器生成器
- [shadcn/ui](https://ui.shadcn.com/) - 优秀的 UI 组件库

## 📮 联系方式

- Issues: https://github.com/YOUR_USERNAME/axonhub/issues
- Discussions: https://github.com/YOUR_USERNAME/axonhub/discussions

---

**⭐ 如果这个项目对你有帮助，请给个 Star！**
