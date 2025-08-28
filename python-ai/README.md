# PolyAgent Python AI Service

PolyAgent 系统的 Python AI 服务层，负责AI模型集成、推理和工具调用。

## 🏗️ 架构设计

```
FastAPI Application
├── Core (配置、日志、异常)
├── Adapters (AI模型适配器)
│   ├── OpenAI
│   ├── Claude
│   └── 其他模型...
├── Services (业务服务)
│   ├── AI Service (模型管理)
│   ├── Tool Service (工具执行)
│   └── RAG Service (知识检索)
└── API (REST端点)
    ├── Tasks (任务执行)
    ├── Chat (聊天接口)
    ├── RAG (文档检索)
    └── Tools (工具管理)
```

## 🚀 快速开始

### 环境要求
- Python 3.11+
- Redis (可选，用于缓存)
- PostgreSQL (可选，用于持久化)

### 开发环境

```bash
# 进入目录
cd python-ai

# 使用开发脚本启动
./scripts/start-dev.sh

# 或手动启动
python3 -m venv venv
source venv/bin/activate
pip install -r requirements.txt
cp .env.example .env
# 编辑 .env 配置文件
python main.py
```

### 配置API密钥

编辑 `.env` 文件：
```bash
OPENAI_API_KEY=your-openai-api-key
ANTHROPIC_API_KEY=your-claude-api-key
```

## 📖 API 使用

### 聊天接口

```bash
curl -X POST http://localhost:8000/api/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "messages": [
      {"role": "user", "content": "Hello, AI!"}
    ],
    "model": "gpt-3.5-turbo"
  }'
```

### 任务执行接口

```bash
curl -X POST http://localhost:8000/api/v1/tasks/execute \
  -H "Content-Type: application/json" \
  -d '{
    "task_id": "task-123",
    "user_id": "user-456",
    "session_id": "session-789",
    "agent_type": "general",
    "input": "What is the weather like today?",
    "tools": ["web_search"]
  }'
```

### 工具执行

```bash
curl -X POST http://localhost:8000/api/v1/tools/execute \
  -H "Content-Type: application/json" \
  -d '{
    "name": "calculator",
    "parameters": {
      "expression": "2 + 2 * 3"
    }
  }'
```

## 🔧 核心组件

### AI 适配器

支持多个AI提供商的统一接口：

- **OpenAI**: GPT-3.5, GPT-4 系列
- **Anthropic**: Claude-3 系列
- **可扩展**: 易于添加新的模型提供商

### 工具系统

内置工具：
- **calculator**: 数学计算
- **web_search**: 网页搜索
- **get_time**: 时间查询

可通过继承基类轻松扩展新工具。

### RAG 系统

- 文档上传和处理
- 向量化存储
- 语义检索
- 上下文生成

## 🔄 与Go服务交互

Python AI服务作为Go服务的下游，处理具体的AI推理任务：

```
Go Gateway → Task Queue → Python AI Service
     ↓              ↓              ↓
   用户请求    → 任务调度    → AI推理执行
     ↓              ↓              ↓
   返回响应    ← 结果回调    ← 完成处理
```

## 🏭 生产部署

### Docker 部署

```bash
# 构建镜像
docker build -t polyagent-ai .

# 运行容器
docker run -d \
  -p 8000:8000 \
  -e OPENAI_API_KEY=your-key \
  -e ANTHROPIC_API_KEY=your-key \
  polyagent-ai
```

### 环境变量

| 变量名 | 描述 | 默认值 |
|--------|------|--------|
| `HOST` | 服务主机 | `0.0.0.0` |
| `PORT` | 服务端口 | `8000` |
| `LOG_LEVEL` | 日志级别 | `INFO` |
| `OPENAI_API_KEY` | OpenAI API密钥 | - |
| `ANTHROPIC_API_KEY` | Anthropic API密钥 | - |

## 🧪 测试

```bash
# 运行测试
pytest

# 覆盖率测试
pytest --cov=app

# 性能测试
locust -f tests/load_test.py
```

## 📊 监控

- **健康检查**: `GET /health`
- **模型状态**: `GET /api/v1/models/health`  
- **工具列表**: `GET /api/v1/tools/list`

## 🔒 安全

- API密钥环境变量管理
- 输入验证和清理
- 错误信息脱敏
- 请求限流

## 📝 开发指南

### 添加新的AI适配器

```python
from app.adapters.base import BaseAIAdapter

class NewAIAdapter(BaseAIAdapter):
    @property
    def provider_name(self) -> str:
        return "new_provider"
    
    async def chat(self, messages, **kwargs):
        # 实现聊天逻辑
        pass
```

### 添加新工具

```python
async def my_tool_handler(parameters: Dict[str, Any]) -> Any:
    # 工具执行逻辑
    return result
```

---

## 📞 支持

- 项目地址: https://github.com/polyagent/polyagent  
- 问题反馈: https://github.com/polyagent/polyagent/issues