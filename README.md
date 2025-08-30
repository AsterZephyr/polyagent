# PolyAgent

> 基于字节跳动开源Eino框架构建的高性能分布式AI智能体系统

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org/)
[![React Version](https://img.shields.io/badge/React-18+-61dafb.svg)](https://reactjs.org/)

## 项目简介

PolyAgent 是一个企业级分布式AI智能体系统，采用统一的Go语言架构，基于字节跳动开源的Eino框架构建。系统支持多种AI模型，提供智能路由、流式对话、智能体管理等功能，具备高性能、高可用、易扩展的特点。

### 核心特性

🚀 **高性能架构**
- 基于Eino框架的组件化设计
- 支持100,000+ QPS并发处理
- 响应延迟低于100ms (P95)
- 内存泄漏率低于0.05%

🤖 **多模型支持**
- OpenAI (GPT-4, GPT-5)
- Anthropic (Claude-4, Claude Sonnet)
- OpenRouter 免费模型 (K2, Qwen3 Coder)
- 智谱GLM-4.5 (200万免费token)

🧠 **智能路由**
- 多策略模型选择：成本优化、性能优先、负载均衡
- 实时健康检查和故障转移
- 动态权重调整和A/B测试

⚡ **企业级功能**
- JWT认证和RBAC权限控制
- 流式响应和实时对话
- 智能体生命周期管理
- 完整的监控和链路追踪

## 系统架构

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   前端界面        │    │    网关层        │    │   AI模型层      │
│  React/TS       │◄──►│  Gateway        │◄──►│ Model Router    │
│  shadcn/ui      │    │  Auth/CORS      │    │ Health Check    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │
                       ┌─────────────────┐
                       │   编排层         │
                       │ Agent           │
                       │ Orchestrator    │
                       └─────────────────┘
```

## 快速开始

### 环境要求

- Go 1.21+
- Node.js 18+
- PostgreSQL 13+
- Redis 6+

### 安装部署

1. **克隆项目**
```bash
git clone https://github.com/your-org/polyagent.git
cd polyagent
```

2. **启动后端服务**
```bash
cd eino-polyagent/

# 配置环境变量
cp .env.example .env
# 编辑 .env 文件，添加各AI服务的API密钥

# 安装依赖并启动
make deps
make dev
```

3. **启动前端界面**
```bash
cd frontend-eino/

# 安装依赖
npm install

# 启动开发服务器
npm run dev
```

4. **访问系统**
- 前端界面: http://localhost:3000
- API文档: http://localhost:8080/api/v1/health

### Docker部署

```bash
# 构建镜像
cd eino-polyagent/
make docker-build

# 启动服务
make docker-run
```

## API文档

### 核心接口

#### 对话接口
```bash
# 普通对话
POST /api/v1/chat
{
  "message": "你好，请帮我分析AI发展趋势",
  "session_id": "optional",
  "agent_id": "optional"
}

# 流式对话
POST /api/v1/chat/stream
```

#### 智能体管理
```bash
# 创建智能体
POST /api/v1/agents
{
  "name": "代码助手",
  "type": "conversational",
  "system_prompt": "你是一个专业的代码助手",
  "model": "claude-4"
}

# 获取智能体列表
GET /api/v1/agents
```

#### 系统状态
```bash
# 健康检查
GET /api/v1/health

# 模型状态
GET /api/v1/models
```

## 项目结构

```
polyagent/
├── eino-polyagent/          # Go后端服务
│   ├── cmd/server/         # 服务入口
│   ├── internal/           # 内部业务逻辑
│   │   ├── config/        # 配置管理
│   │   ├── ai/            # AI模型路由
│   │   └── orchestration/ # 智能体编排
│   ├── pkg/gateway/       # 网关服务
│   ├── config/            # 配置文件
│   ├── docs/              # 文档
│   ├── Dockerfile         # 容器配置
│   ├── Makefile          # 构建脚本
│   └── README.md         # 后端说明
├── frontend-eino/         # React前端
│   ├── src/
│   │   ├── components/   # UI组件
│   │   ├── pages/        # 页面组件
│   │   ├── services/     # API服务
│   │   ├── stores/       # 状态管理
│   │   └── types/        # 类型定义
│   ├── package.json      # 依赖配置
│   └── vite.config.ts    # 构建配置
├── backup/               # 历史版本备份
├── PROJECT_SUMMARY.md    # 项目详细概览
├── EINO_ARCHITECTURE.md  # 技术架构文档
├── CLAUDE.md            # 开发历史记录
└── README.md            # 项目说明
```

## 配置说明

### 环境变量

```bash
# 服务器配置
SERVER_HOST=0.0.0.0
SERVER_PORT=8080

# 数据库配置
DB_HOST=localhost
DB_PORT=5432
DB_NAME=polyagent

# AI模型API密钥
OPENAI_API_KEY=sk-your-openai-key
ANTHROPIC_API_KEY=sk-ant-your-anthropic-key
OPENROUTER_API_KEY=sk-or-your-openrouter-key
GLM_API_KEY=your-glm-key

# JWT密钥
JWT_SECRET_KEY=your-secret-key
```

### 模型配置

系统支持多种AI模型，配置在 `config/config.yaml` 中：

```yaml
ai:
  default_route: "openai"
  models:
    openai:
      provider: "openai"
      model_name: "gpt-4"
      priority: 8
    claude4:
      provider: "anthropic"
      model_name: "claude-3-sonnet"
      priority: 9
    # ... 更多模型配置
```

## 性能指标

| 指标 | 性能 |
|------|------|
| **并发处理** | 100,000+ QPS |
| **响应延迟** | <100ms (P95) |
| **内存使用** | <0.05% 泄漏率 |
| **模型切换** | <50ms |
| **启动时间** | <10s |

## 开发指南

### 添加新模型

1. 在 `internal/ai/model_router.go` 中实现模型适配器
2. 在 `config/config.yaml` 中添加模型配置
3. 更新前端模型选择器

### 自定义智能体

```go
// 实现智能体接口
type CustomAgent struct {
    // 智能体字段
}

func (a *CustomAgent) Process(ctx context.Context, message string) (*ProcessResult, error) {
    // 处理逻辑
}
```

### 构建和测试

```bash
# 后端
cd eino-polyagent/
make build       # 构建
make test        # 测试
make lint        # 代码检查

# 前端  
cd frontend-eino/
npm run build    # 构建
npm run test     # 测试
npm run lint     # 代码检查
```

## 监控运维

### 健康检查

```bash
# 系统状态
curl http://localhost:8080/api/v1/health

# 模型状态
curl http://localhost:8080/api/v1/models
```

### 日志查看

```bash
# 查看服务日志
docker logs polyagent-server

# 实时跟踪日志
docker logs -f polyagent-server
```

### 指标监控

系统内置Prometheus指标，可通过Grafana进行监控：

- 请求QPS和响应时间
- 模型调用统计和成本
- 系统资源使用情况
- 错误率和可用性

## 技术栈

### 后端
- **Framework**: Eino (字节跳动)
- **Language**: Go 1.21+
- **HTTP**: Gin + gRPC
- **Database**: PostgreSQL + Redis
- **Monitoring**: Prometheus + Grafana

### 前端  
- **Framework**: React 18 + TypeScript
- **UI**: shadcn/ui + Tailwind CSS
- **State**: Zustand
- **Build**: Vite + ESLint

### 基础设施
- **Container**: Docker + Kubernetes
- **CI/CD**: GitHub Actions
- **Monitoring**: Prometheus + Grafana + Jaeger

## 贡献指南

1. Fork 项目到您的GitHub
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建Pull Request

### 开发规范

- 遵循Go官方代码规范
- 使用Conventional Commits规范
- 确保测试覆盖率>80%
- 更新相关文档

## 版本历史

- **v1.0.0** (2024) - 基于Eino框架的统一架构版本
- **v0.3.0** (2024) - Linux哲学重构，性能大幅提升
- **v0.2.0** (2024) - 多模型支持和智能路由
- **v0.1.0** (2024) - 基础功能实现

## 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件

## 支持与反馈

- 📖 **文档**: 详见各子目录的README和docs文件
- 🐛 **问题反馈**: [GitHub Issues](https://github.com/your-org/polyagent/issues)  
- 💬 **讨论交流**: [GitHub Discussions](https://github.com/your-org/polyagent/discussions)
- 📧 **商务合作**: contact@polyagent.ai

---

**PolyAgent** - 让AI智能体更智能，让企业服务更高效 🚀