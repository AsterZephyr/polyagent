# PolyAgent - 多语言智能体系统

> 基于 Go + Python 混合架构的企业级智能体平台，支持多 AI 模型集成和个性化 RAG

## 🚀 项目特性

- **🔥 混合架构**: Go 高性能服务层 + Python AI 计算层
- **🤖 多智能体**: 支持对话、RAG、代码、工具等多种智能体
- **🔌 多模型**: 集成 OpenAI、Claude、通义千问等主流 AI 模型  
- **📚 智能 RAG**: 个性化知识库检索增强生成
- **🛠️ 工具生态**: 可扩展的工具调用和插件系统
- **💾 记忆管理**: 长期对话记忆和上下文管理
- **🔄 流式输出**: 实时响应和渐进式结果展示

## 🏗️ 系统架构

```
Frontend (React/CLI)
        ↓
  Go API Gateway (8080)
        ↓
┌─────────────────┬─────────────────┐
│  Task Scheduler │   Data Storage  │
│   (Goroutines)  │ (Redis/Postgres)│
└─────────────────┴─────────────────┘
        ↓
Python AI Core (8000)
        ↓
┌──────────────┬─────────────────┬─────────────────┐
│ Multi-AI API │  RAG Engine     │  Agent System   │
│   Adapter    │ (ChromaDB)      │ (Tools/Memory)  │
└──────────────┴─────────────────┴─────────────────┘
```

## 📁 项目结构

```
polyagent/
├── go-services/           # Go 服务层
│   ├── gateway/          # API 网关服务
│   ├── scheduler/        # 任务调度服务  
│   ├── storage/          # 数据存储服务
│   ├── registry/         # 智能体注册中心
│   └── plugins/          # 插件系统
├── python-ai/            # Python AI 层
│   ├── adapter/          # AI 模型适配器
│   ├── core/             # 智能体核心逻辑
│   ├── rag/              # RAG 检索系统
│   ├── tools/            # 工具调用管理
│   └── memory/           # 记忆管理系统
├── frontend/             # 前端层
│   ├── web/              # Web 管理界面
│   ├── cli/              # 命令行工具
│   └── sdk/              # 客户端 SDK
└── docs/                 # 文档
    ├── api/              # API 接口文档
    ├── architecture/     # 架构设计文档
    └── deployment/       # 部署运维文档
```

## 🛠️ 技术栈

### 后端服务
- **Go**: Gin + gRPC + Redis + PostgreSQL
- **Python**: FastAPI + LangChain + ChromaDB

### AI 集成  
- **模型**: OpenAI GPT、Claude、通义千问等
- **向量数据库**: ChromaDB / Pinecone
- **工具框架**: LangChain + 自定义工具

### 前端交互
- **Web UI**: React + TypeScript + Ant Design
- **CLI**: Go Cobra + 交互式命令行
- **部署**: Docker + Kubernetes

## 🚀 快速开始

### 环境要求
- Go 1.21+
- Python 3.11+
- Docker & Docker Compose
- Redis 7+
- PostgreSQL 15+

### 启动服务
```bash
# 克隆项目
git clone https://github.com/polyagent/polyagent.git
cd polyagent

# 配置环境变量
cp .env.example .env
# 编辑 .env 文件，填入 API Keys

# 启动所有服务
docker-compose up -d

# 访问服务
# Web UI: http://localhost:3000  
# API: http://localhost:8080
# Python AI: http://localhost:8000
```

### 开发模式
```bash
# 启动 Go 服务
cd go-services
go run main.go

# 启动 Python AI 服务  
cd python-ai
python -m uvicorn main:app --reload --port 8000

# 启动前端
cd frontend/web
npm install && npm start
```

## 📖 API 文档

### 对话接口
```bash
curl -X POST http://localhost:8080/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{
    "message": "你好，帮我分析一下今天的股市行情",
    "agent_type": "rag",
    "tools": ["web_search", "stock_analyzer"]
  }'
```

### 文档上传
```bash  
curl -X POST http://localhost:8080/api/v1/documents/upload \
  -F "file=@document.pdf" \
  -F "user_id=user123"
```

详细 API 文档: [docs/api/README.md](docs/api/README.md)

## 🔧 配置说明

### AI 模型配置
```yaml
ai_models:
  openai:
    api_key: "your-openai-key"
    base_url: "https://api.openai.com/v1"
    models: ["gpt-4", "gpt-3.5-turbo"]
  
  claude:
    api_key: "your-claude-key" 
    models: ["claude-3-sonnet", "claude-3-haiku"]
```

### RAG 系统配置
```yaml
rag:
  vector_db: "chromadb"  # chromadb / pinecone
  chunk_size: 1000
  overlap: 200
  top_k: 5
```

## 🤝 贡献指南

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

## 📄 许可证

本项目采用 [MIT](LICENSE) 许可证

## 📞 联系方式

- 项目主页: https://github.com/polyagent/polyagent
- 文档网站: https://docs.polyagent.dev
- 讨论社区: https://discord.gg/polyagent

---

**PolyAgent** - 让每个人都能拥有自己的智能体助手 🤖✨