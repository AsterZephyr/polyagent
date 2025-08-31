# PolyAgent - Eino Framework Based Distributed AI System

高性能分布式AI智能体系统，基于字节跳动开源Eino框架构建。

## 架构特性

- **统一Go架构**：基于Eino框架的组件化设计
- **智能模型路由**：多策略路由，支持负载均衡和成本优化
- **流式响应**：实时流式对话体验
- **多模型支持**：OpenAI GPT-4/5、Claude-4、OpenRouter免费模型、GLM-4.5
- **分布式部署**：容器化部署，支持水平扩展
- **企业级安全**：JWT认证、CORS、速率限制

## 快速开始

> 📋 **完整启动指南**: [STARTUP.md](../STARTUP.md)

### 最简启动方式

**后端服务** (终端1):
```bash
cd /Users/hxz/code/polyagent/eino-polyagent
PORT=8082 go run cmd/server/main.go
```

**前端服务** (终端2):
```bash
cd /Users/hxz/code/polyagent/frontend-eino
npm install  # 仅首次需要
npm run dev
```

**访问应用**: http://localhost:3000

### 传统方式 (可选)

```bash
# 安装依赖
make deps

# 安装开发工具 (可选)
make install-tools

# 启动开发服务器 (需要air工具)
make dev
```

### 构建和部署

```bash
# 构建应用
make build

# 运行生产服务器
make run

# Docker部署
make docker-build
make docker-run
```

## API文档

### 健康检查
```
GET /api/v1/health
```

### 对话接口
```bash
# 普通对话
POST /api/v1/chat
{
  "message": "Hello",
  "session_id": "optional",
  "agent_id": "optional"
}

# 流式对话
POST /api/v1/chat/stream
```

### 智能体管理
```bash
# 创建智能体
POST /api/v1/agents
{
  "name": "Assistant",
  "type": "conversational",
  "system_prompt": "You are a helpful assistant"
}

# 获取智能体列表
GET /api/v1/agents
```

## 性能指标

- **QPS**: >100,000
- **响应延迟**: <100ms
- **内存泄漏率**: <0.05%
- **模型切换**: <50ms

## 项目结构

```
eino-polyagent/
├── cmd/server/          # 服务入口
├── internal/            # 内部包
│   ├── config/         # 配置管理
│   ├── ai/             # AI模型路由
│   └── orchestration/  # 智能体编排
├── pkg/gateway/        # 网关服务
├── config/             # 配置文件
└── Makefile           # 构建脚本
```

## 开发命令

```bash
make build          # 构建应用
make run            # 运行应用
make dev            # 开发模式
make test           # 运行测试
make lint           # 代码检查
make fmt            # 代码格式化
make clean          # 清理构建
```