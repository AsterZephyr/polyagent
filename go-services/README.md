# PolyAgent Go Services

PolyAgent 系统的 Go 服务层，提供高性能的 API 网关、任务调度和数据存储服务。

## 📁 项目结构

```
go-services/
├── gateway/                    # API 网关服务
│   ├── main.go                # 网关主程序
│   ├── handlers/              # HTTP 处理器
│   │   ├── chat.go           # 聊天相关接口
│   │   ├── agent.go          # 智能体管理
│   │   ├── document.go       # 文档管理
│   │   ├── user.go           # 用户管理
│   │   └── health.go         # 健康检查
│   └── middleware/            # 中间件
│       └── middleware.go     # 认证、CORS、限流等
├── scheduler/                 # 任务调度服务
│   └── main.go               # 调度器主程序
├── internal/                  # 内部包
│   ├── config/               # 配置管理
│   │   └── config.go
│   ├── models/               # 数据模型
│   │   └── types.go
│   ├── storage/              # 存储层
│   │   ├── postgres.go       # PostgreSQL 操作
│   │   └── redis.go          # Redis 操作
│   ├── scheduler/            # 任务调度器
│   │   └── scheduler.go
│   └── ai/                   # AI 客户端
│       └── client.go         # Python AI 服务客户端
├── configs/                   # 配置文件
│   └── config.yaml
├── Makefile                   # 构建脚本
├── Dockerfile.gateway         # 网关服务 Docker 文件
├── go.mod                     # Go 模块定义
└── README.md
```

## 🚀 快速开始

### 环境要求

- Go 1.21+
- PostgreSQL 15+
- Redis 7+

### 安装依赖

```bash
make deps
```

### 配置环境

```bash
# 复制配置文件
cp configs/config.yaml.example configs/config.yaml

# 或使用环境变量
export DATABASE_URL="postgres://user:pass@localhost:5432/polyagent"
export REDIS_URL="redis://localhost:6379/0"
export PYTHON_AI_URL="http://localhost:8000"
export JWT_SECRET="your-secret-key"
```

### 构建项目

```bash
make build
```

### 运行服务

```bash
# 运行 API 网关
make run-gateway

# 运行任务调度器
make run-scheduler

# 或同时运行所有服务
make run-all
```

## 🔧 开发工具

### 代码格式化

```bash
make fmt
```

### 代码检查

```bash
make lint
```

### 运行测试

```bash
make test
```

### 生成测试覆盖率

```bash
make test-coverage
```

### 性能测试

```bash
make bench
```

## 🐳 Docker 部署

### 构建 Docker 镜像

```bash
make docker-build
```

### 运行 Docker 容器

```bash
make docker-run
```

## 📖 API 接口

### 聊天接口

```bash
# 发送聊天消息
curl -X POST http://localhost:8080/api/v1/chat \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "message": "Hello, AI!",
    "agent_type": "general",
    "tools": ["web_search"]
  }'

# 流式聊天
curl -X POST http://localhost:8080/api/v1/chat/stream \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "message": "Tell me about AI",
    "stream": true
  }'
```

### 智能体管理

```bash
# 获取智能体列表
curl -X GET http://localhost:8080/api/v1/agents \
  -H "Authorization: Bearer <token>"

# 创建智能体
curl -X POST http://localhost:8080/api/v1/agents \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "name": "Code Assistant",
    "type": "code",
    "description": "AI assistant for coding tasks",
    "tools": ["code_analyzer", "git_helper"]
  }'
```

### 文档管理

```bash
# 上传文档
curl -X POST http://localhost:8080/api/v1/documents/upload \
  -H "Authorization: Bearer <token>" \
  -F "file=@document.pdf"

# 获取文档列表
curl -X GET http://localhost:8080/api/v1/documents \
  -H "Authorization: Bearer <token>"
```

### 健康检查

```bash
# 检查服务健康状态
curl -X GET http://localhost:8080/health

# 获取系统指标
curl -X GET http://localhost:8080/metrics
```

## 🏗️ 架构设计

### 分层架构

```
┌─────────────────────────────────────┐
│           API Gateway               │ ← HTTP/WebSocket 接口
├─────────────────────────────────────┤
│          Middleware                 │ ← 认证、限流、日志
├─────────────────────────────────────┤
│          Handlers                   │ ← 业务逻辑处理
├─────────────────────────────────────┤
│       Task Scheduler                │ ← 异步任务调度
├─────────────────────────────────────┤
│        Storage Layer                │ ← 数据持久化
├─────────────────────────────────────┤
│      External Services              │ ← Python AI / 第三方服务
└─────────────────────────────────────┘
```

### 核心组件

1. **API 网关**: 统一入口，处理认证、限流、路由
2. **任务调度器**: 异步任务队列，支持优先级和重试
3. **存储层**: PostgreSQL + Redis 双重存储
4. **中间件**: 提供横切关注点处理
5. **AI 客户端**: 与 Python AI 服务通信

### 数据流

```
Client Request → Gateway → Middleware → Handler → Scheduler → Python AI
     ↓              ↓          ↓           ↓           ↓           ↓
  Response ← JSON ← Process ← Business ← Queue ← AI Response
```

## 🔒 安全特性

- JWT 认证和授权
- API 请求限流
- CORS 跨域保护
- SQL 注入防护
- XSS 攻击防护
- 敏感数据加密

## 📊 监控和日志

- 结构化日志记录
- 性能指标收集
- 健康检查端点
- 错误追踪和报警
- 请求链路追踪

## 🧪 测试策略

- 单元测试覆盖核心逻辑
- 集成测试验证服务交互
- 性能测试确保系统吞吐量
- 端到端测试验证完整流程

## 📈 性能优化

- 连接池管理
- Redis 缓存策略
- 异步任务处理
- 数据库查询优化
- 内存使用优化

## 🚀 生产部署

### 配置优化

```yaml
# 生产环境配置
server:
  read_timeout: 30
  write_timeout: 30

database:
  max_open_conns: 100
  max_idle_conns: 20

redis:
  pool_size: 50
  max_retries: 3

log:
  level: "warn"
  format: "json"
```

### 扩展部署

- 支持水平扩展
- 负载均衡配置
- 数据库读写分离
- Redis 集群模式

## 🤝 贡献指南

1. Fork 项目
2. 创建特性分支
3. 编写测试
4. 提交更改
5. 创建 Pull Request

## 📝 更新日志

### v1.0.0
- 初始版本发布
- 实现 API 网关服务
- 实现任务调度系统
- 完成基础存储层

---

## 📞 联系方式

- 项目地址: https://github.com/polyagent/polyagent
- 问题反馈: https://github.com/polyagent/polyagent/issues
- 文档网站: https://docs.polyagent.dev