# AxonHubPlus - 企业级 AI 网关与响应保护系统

<div align="center">

[![Docker Build](https://github.com/ifsherlock/axonhubplus/actions/workflows/docker-build.yml/badge.svg)](https://github.com/ifsherlock/axonhubplus/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go)](https://golang.org)
[![React](https://img.shields.io/badge/React-19-61DAFB?logo=react)](https://react.dev)

**统一管理多个 LLM 提供商 | 智能负载均衡 | 内容安全防护 | 完整的请求追踪**

[快速开始](#-快速开始) • [功能特性](#-核心特性) • [部署文档](./DOCKER_DEPLOYMENT.md) • [API 文档](#-api-使用)

</div>

---

## 📖 项目简介

AxonHubPlus 是一个企业级 AI 网关解决方案，提供统一的 API 入口管理多个 LLM 服务商（OpenAI、Anthropic、Google Gemini 等）。

**核心价值**：
- 🔄 **降低成本**：智能路由选择最优渠道，自动 failover 保证高可用
- 🛡️ **安全合规**：双重内容保护（输入/输出），防止敏感信息泄露
- 📊 **精细管控**：项目级权限、用量统计、完整审计日志
- 🚀 **无缝迁移**：100% 兼容 OpenAI API，零代码改造

---

## ✨ 核心特性

### 🔐 双重内容保护

#### 1️⃣ Prompt Protection（输入保护）
保护用户输入中的敏感信息：
- API 密钥、Token 过滤
- 个人隐私信息脱敏
- 恶意注入检测

#### 2️⃣ Response Protection（响应保护）⭐ **本项目核心功能**

实时过滤 LLM 响应内容，支持三种处理模式：

| 模式 | 说明 | 使用场景 |
|------|------|----------|
| **Mask（脱敏）** | 将匹配内容替换为指定文本 | 过滤广告链接、替换敏感词 |
| **Reject（拒绝）** | 直接拒绝包含违规内容的响应 | 拦截暴力、色情等不当内容 |
| **Failover（切换）** | 检测到特定内容时自动切换渠道 | 应对上游审核拦截，保证服务可用 |

**示例规则**：
```json
{
  "name": "过滤推广域名",
  "pattern": "(?i)dc\\.hhhl\\.cc",
  "action": "mask",
  "replacement": "[已过滤]",
  "scopes": ["text"]
}
```

### 🚀 其他特性

- **多渠道管理**：统一管理 OpenAI、Anthropic、Google、国产大模型等
- **智能负载均衡**：基于权重、延迟、成本的智能路由
- **实时监控**：TPS、成本、成功率可视化看板
- **项目隔离**：多租户、多项目权限管理
- **完整审计**：请求日志、Token 用量、错误追踪

---

## 🚀 快速开始

### 方式一：Docker Compose（推荐）

**1. 克隆仓库**
```bash
git clone https://github.com/ifsherlock/axonhubplus.git
cd axonhubplus
```

**2. 启动服务**
```bash
docker-compose up -d
```

这将启动：
- ✅ AxonHub 后端（端口 8090）
- ✅ PostgreSQL 数据库
- ✅ 前端界面

**3. 访问系统**
- 前端地址：http://localhost:8090
- GraphQL Playground：http://localhost:8090/graphql
- 默认账号：首次访问时设置管理员账号

**4. 配置渠道**
1. 登录后进入 **Channels** 页面
2. 点击 **Add Channel** 添加你的 LLM 渠道
3. 配置 API Key 和参数
4. 启用渠道

**5. 创建响应保护规则**
1. 进入 **Admin > Response Protection**
2. 点击 **Create Rule**
3. 配置规则：名称、正则表达式、处理模式
4. 启用规则

**6. 开始使用**
```bash
curl http://localhost:8090/api/v1/chat/completions \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

📖 详细部署文档：[DOCKER_DEPLOYMENT.md](./DOCKER_DEPLOYMENT.md)

---

### 方式二：使用 GitHub Container Registry 镜像

等待 GitHub Actions 构建完成后：

```bash
# 拉取镜像
docker pull ghcr.io/ifsherlock/axonhubplus:main

# 创建数据库（可选，使用 SQLite）
mkdir -p data

# 启动容器
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

### 方式三：本地开发

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

---

## 📊 系统架构

```
┌──────────────────────────────────────────────────────────────┐
│                         Client                               │
│                     (OpenAI API 格式)                         │
└────────────────────────┬─────────────────────────────────────┘
                         │
                         ▼
┌──────────────────────────────────────────────────────────────┐
│                    AxonHub Gateway                           │
│                                                              │
│  ┌────────────────────────────────────────────────────────┐ │
│  │  1. Authentication & Authorization                     │ │
│  └────────────────────────────────────────────────────────┘ │
│  ┌────────────────────────────────────────────────────────┐ │
│  │  2. Prompt Protection (输入过滤)                        │ │
│  │     • 敏感信息检测    • API Key 过滤                   │ │
│  └────────────────────────────────────────────────────────┘ │
│  ┌────────────────────────────────────────────────────────┐ │
│  │  3. Channel Selection (智能路由)                       │ │
│  │     • 负载均衡        • 成本优化     • 自动 Failover  │ │
│  └────────────────────────────────────────────────────────┘ │
│  ┌────────────────────────────────────────────────────────┐ │
│  │  4. Response Protection (输出过滤) ⭐                   │ │
│  │     • Mask 脱敏       • Reject 拒绝   • Failover 切换  │ │
│  └────────────────────────────────────────────────────────┘ │
│  ┌────────────────────────────────────────────────────────┐ │
│  │  5. Logging & Monitoring                               │ │
│  └────────────────────────────────────────────────────────┘ │
└────────────┬─────────────┬─────────────┬─────────────────────┘
             │             │             │
             ▼             ▼             ▼
       ┌─────────┐   ┌─────────┐   ┌─────────┐
       │ OpenAI  │   │Anthropic│   │  Gemini │
       │Channel 1│   │Channel 2│   │Channel 3│
       └─────────┘   └─────────┘   └─────────┘
```

---

## 💡 响应保护使用示例

### 场景 1：过滤推广链接

**问题**：某些 LLM 响应中包含推广链接或广告域名

**解决方案**：
```json
{
  "name": "过滤广告域名",
  "pattern": "(?i)(dc\\.hhhl\\.cc|promo\\.example\\.com)",
  "action": "mask",
  "replacement": "[推广内容已过滤]",
  "scopes": ["text"],
  "enabled": true
}
```

**效果**：
- 输入：`"访问 dc.hhhl.cc 了解更多"`
- 输出：`"访问 [推广内容已过滤] 了解更多"`

### 场景 2：拦截违规内容

**问题**：需要阻止暴力、色情等不当内容

**解决方案**：
```json
{
  "name": "拦截暴力内容",
  "pattern": "(?i)(violence|kill|attack|weapon)",
  "action": "reject",
  "enabled": true
}
```

**效果**：直接返回错误，拒绝包含违规词的响应

### 场景 3：应对上游审核拦截

**问题**：某些渠道对特定内容进行审核拦截，返回 `REQUEST_BLOCKED`

**解决方案**：
```json
{
  "name": "检测上游拦截自动切换",
  "pattern": "(?i)(REQUEST_BLOCKED|CONTENT_FILTERED)",
  "action": "failover",
  "enabled": true
}
```

**效果**：检测到拦截后自动切换到备用渠道重试，保证服务可用

---

## 🛠️ 技术栈

| 层级 | 技术 | 说明 |
|------|------|------|
| **后端** | Go 1.23+ | 高性能并发处理 |
| | Ent | 类型安全的 ORM |
| | GraphQL (gqlgen) | 灵活的 API 查询 |
| | Gin | HTTP 框架 |
| **前端** | React 19 | 现代化 UI |
| | TypeScript | 类型安全 |
| | TanStack Router & Query | 路由和数据管理 |
| | Tailwind CSS + shadcn/ui | 美观的组件库 |
| **数据库** | PostgreSQL (推荐) | 生产环境 |
| | MySQL / SQLite | 也支持 |
| **部署** | Docker / Docker Compose | 容器化部署 |
| | GitHub Actions | CI/CD 自动化 |

---

## 📖 API 使用

### OpenAI 兼容 API

AxonHub 完全兼容 OpenAI API 格式，无需修改代码：

```bash
# Chat Completions
curl http://localhost:8090/api/v1/chat/completions \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4",
    "messages": [
      {"role": "system", "content": "You are a helpful assistant."},
      {"role": "user", "content": "What is AI?"}
    ],
    "stream": false
  }'

# 流式响应
curl http://localhost:8090/api/v1/chat/completions \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4",
    "messages": [{"role": "user", "content": "Tell me a story"}],
    "stream": true
  }'
```

### GraphQL API

访问 http://localhost:8090/graphql 使用 GraphQL Playground

**查询响应保护规则**：
```graphql
query {
  responseProtectionRules {
    id
    name
    pattern
    action
    enabled
  }
}
```

**创建规则**：
```graphql
mutation {
  createResponseProtectionRule(input: {
    name: "过滤广告"
    pattern: "(?i)promo\\.link"
    action: MASK
    replacement: "[已过滤]"
    scopes: [TEXT]
  }) {
    id
    name
  }
}
```

---

## ⚙️ 配置说明

### 环境变量

| 变量名 | 说明 | 默认值 | 示例 |
|--------|------|--------|------|
| `AXONHUB_DB_DIALECT` | 数据库类型 | `postgres` | `postgres` / `mysql` / `sqlite3` |
| `AXONHUB_DB_DSN` | 数据库连接串 | - | `postgres://user:pass@host:5432/axonhub` |
| `AXONHUB_SERVER_PORT` | 服务端口 | `8090` | `8090` |
| `AXONHUB_LOG_LEVEL` | 日志级别 | `info` | `debug` / `info` / `warn` / `error` |
| `AXONHUB_LOG_FORMAT` | 日志格式 | `json` | `json` / `console` |

### 配置文件

创建 `config.yml`（可选）：

```yaml
server:
  port: 8090
  cors:
    enabled: true
    
database:
  dialect: postgres
  dsn: postgres://axonhub:password@localhost:5432/axonhub
  max_open_conns: 100
  max_idle_conns: 10
  
log:
  level: info
  format: json
```

---

## 🔧 开发指南

### 代码生成

```bash
# 生成 Ent ORM 代码
go generate ./internal/ent

# 生成 GraphQL 代码
cd internal/server/gql
go generate
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

# 本地构建二进制
go build -o axonhub ./cmd/axonhub
```

---

## 🤝 贡献指南

欢迎提交 Issue 和 Pull Request！

1. Fork 本仓库
2. 创建功能分支：`git checkout -b feature/amazing-feature`
3. 提交更改：`git commit -m 'Add amazing feature'`
4. 推送分支：`git push origin feature/amazing-feature`
5. 提交 Pull Request

---

## 📄 开源协议

MIT License - 详见 [LICENSE](./LICENSE) 文件

---

## 🙏 致谢

- [Ent](https://entgo.io/) - 强大的 Go ORM
- [gqlgen](https://gqlgen.com/) - GraphQL 服务器生成器
- [shadcn/ui](https://ui.shadcn.com/) - 优秀的 UI 组件库

---

## 📮 联系方式

- 🐛 Issues: https://github.com/ifsherlock/axonhubplus/issues
- 💬 Discussions: https://github.com/ifsherlock/axonhubplus/discussions

---

<div align="center">

**⭐ 如果这个项目对你有帮助，请给个 Star！**

Made with ❤️ by [ifsherlock](https://github.com/ifsherlock)

</div>

