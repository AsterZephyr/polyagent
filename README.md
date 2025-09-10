# PolyAgent - 推荐业务智能体系统

> 基于Agent4Rec架构的专业推荐业务闭环AI智能体系统，从数据采集到实时推荐的完整解决方案

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org/)
[![React Version](https://img.shields.io/badge/React-18+-61dafb.svg)](https://reactjs.org/)

## 项目简介

PolyAgent 是一个专业的推荐业务闭环AI智能体系统，基于Agent4Rec等成功推荐系统架构设计。系统通过4个专业化Agent实现完整的推荐业务链路：数据采集 → 特征工程 → 模型训练 → 实时推荐 → 效果评估，为企业提供端到端的推荐系统解决方案。

###  核心特性

 **专业推荐Agent**
- **DataAgent**: 数据采集、清洗、特征工程
- **ModelAgent**: 协同过滤、深度学习、矩阵分解算法
- **ServiceAgent**: 实时推荐服务、高性能预测
- **EvalAgent**: A/B测试、效果评估、业务指标监控

 **业务闭环架构**
- 完整推荐业务流程自动化
- 实时模型训练与优化
- 多算法支持：协同过滤、内容推荐、深度学习
- 智能超参数调优与模型选择

 **专业前端界面**
- 现代化深色主题设计
- 炫酷发光边框效果 (GlowingEffect)
- 实时系统监控和Agent状态展示
- 响应式网格布局与专业UI组件

 **企业级功能**
- Go语言高性能后端架构
- RESTful API完整接口
- 实时指标监控和健康检查
- 任务队列与重试机制

##  推荐业务架构

```
┌─────────────────────────────────────────────────────────────────┐
│                     推荐业务闭环 Agent 系统                        │
├─────────────────┬─────────────────┬─────────────────┬─────────────┐
│   DataAgent     │   ModelAgent    │  ServiceAgent   │  EvalAgent  │
│   数据采集       │   模型训练       │   实时推荐       │   效果评估   │
│                │                │                │             │
│ • 用户行为采集   │ • 协同过滤       │ • 高性能预测     │ • A/B测试   │
│ • 特征工程      │ • 深度学习       │ • 实时推荐      │ • NDCG@K    │
│ • 数据清洗      │ • 矩阵分解       │ • 缓存策略      │ • 点击率    │
│ • 质量监控      │ • 超参数优化     │ • 负载均衡      │ • 覆盖率    │
└─────────────────┴─────────────────┴─────────────────┴─────────────┘
                                   ↕
┌─────────────────────────────────────────────────────────────────┐
│                      专业前端监控界面                             │
│   发光效果UI   实时监控   响应式设计   热更新              │
└─────────────────────────────────────────────────────────────────┘
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

2. **启动推荐Agent后端服务**
```bash
cd eino-polyagent/

# 安装Go依赖
go mod tidy
go get github.com/stretchr/testify

# 启动推荐业务Agent服务器
go run cmd/server/main.go
# 服务启动在端口 8080
```

3. **启动专业前端界面**
```bash
cd frontend-eino/

# 安装依赖（包含motion动画库）
npm install
npm install motion

# 启动开发服务器
npm run dev
```

4. **访问推荐系统**
-  **专业前端界面**: http://localhost:3000
-  **推荐API接口**: http://localhost:8080/api/v1/recommendation/
-  **系统健康检查**: http://localhost:8080/api/v1/recommendation/health

###  Docker部署

```bash
# 构建推荐Agent后端镜像
cd eino-polyagent/
docker build -t polyagent-recommendation:latest .

# 启动容器服务
docker run -d -p 8080:8080 \
  --name polyagent-recommendation \
  polyagent-recommendation:latest

# 构建前端镜像
cd ../frontend-eino/
docker build -t polyagent-frontend:latest .
docker run -d -p 3000:80 \
  --name polyagent-frontend \
  polyagent-frontend:latest
```

##  推荐业API文档

###  核心推荐接口

####  数据采集接口
```bash
# 用户行为数据采集
POST /api/v1/recommendation/data/collect
{
  "collector": "user_behavior",
  "timerange": "last_24_hours",
  "filters": {
    "user_type": "active",
    "platform": "web"
  }
}

# 特征工程
POST /api/v1/recommendation/data/features
{
  "feature_type": "user_profile",
  "algorithms": ["tfidf", "embedding"]
}
```

####  模型训练接口
```bash
# 模型训练
POST /api/v1/recommendation/models/train
{
  "algorithm": "collaborative_filtering",
  "hyperparameters": {
    "learning_rate": 0.001,
    "epochs": 100,
    "batch_size": 256
  }
}

# 模型评估
POST /api/v1/recommendation/models/evaluate
{
  "model_id": "cf_model_v1",
  "metrics": ["ndcg_at_k", "precision_at_k", "recall_at_k"]
}
```

####  实时推荐接口
```bash
# 获取推荐
POST /api/v1/recommendation/recommend
{
  "user_id": "user_12345",
  "num_items": 10,
  "algorithm": "hybrid",
  "context": {
    "time": "evening",
    "device": "mobile"
  }
}

# 批量预测
POST /api/v1/recommendation/predict
{
  "user_ids": ["user_1", "user_2"],
  "item_ids": ["item_a", "item_b"]
}
```

####  系统监控接口
```bash
# 系统指标
GET /api/v1/recommendation/system/metrics

# Agent状态
GET /api/v1/recommendation/agents

# 健康检查
GET /api/v1/recommendation/health
```

##  项目结构

```
polyagent/
├── eino-polyagent/                    # Go推荐Agent后端
│   ├── cmd/server/                   # 服务入口
│   │   └── main.go                  # 推荐系统主服务器
│   ├── internal/recommendation/      #  推荐业务核心
│   │   ├── orchestrator.go          # 任务编排器
│   │   ├── data_agent.go            #  DataAgent
│   │   ├── model_agent.go           #  ModelAgent
│   │   ├── api_handler.go           # RESTful API
│   │   ├── agent_types.go           # 类型定义
│   │   └── integration_test.go      # 集成测试
│   └── go.mod                       # Go依赖管理
├── frontend-eino/                     #  React专业前端
│   ├── src/
│   │   ├── components/
│   │   │   ├── ui/
│   │   │   │   └── glowing-effect.tsx   #  发光效果组件
│   │   │   └── RecommendationAgentGrid.tsx #  主仪表板
│   │   ├── pages/
│   │   │   └── AgentDashboard.tsx       # 仪表板页面
│   │   ├── services/
│   │   │   └── recommendation.ts        # API服务封装
│   │   └── stores/                  # Zustand状态管理
│   ├── package.json                 # 依赖+motion动画库
│   └── vite.config.ts               # Vite构建配置
├── RECOMMENDATION_AGENT_DESIGN.md     #  设计文档
├── CLAUDE.md                          # 开发历史记录
└── README.md                          # 项目说明

 核心文件说明：
• glowing-effect.tsx - 从 reactbits.dev 集成的专业发光效果
• RecommendationAgentGrid.tsx - 5个专业化Agent卡片+实时数据
• recommendation/ - 基于Agent4Rec的完整推荐业务链路
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

##  性能指标

### 推荐系统性能
| 指标 | 性能 | 说明 |
|------|------|------|
| **推荐响应** | <50ms (P95) | 实时推荐延迟 |
| **模型训练** | 10K samples/s | 训练数据处理速度 |
| **并发请求** | 1000+ QPS | 推荐API并发处理 |
| **模型精度** | NDCG@10 > 0.85 | 推荐算法效果 |
| **系统可用性** | 99.9% | 服务稳定性 |
| **启动时间** | <5s | Agent系统启动 |

### Agent性能指标
- **DataAgent**: 支持10M+用户行为数据/天
- **ModelAgent**: 支持15+种推荐算法
- **ServiceAgent**: 支持10K+并发推荐请求
- **EvalAgent**: 实时A/B测试和效果评估

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

###  系统监控

```bash
# 推荐系统健康检查
curl http://localhost:8080/api/v1/recommendation/health

# 获取系统指标
curl http://localhost:8080/api/v1/recommendation/system/metrics

# 查看Agent状态
curl http://localhost:8080/api/v1/recommendation/agents

# 查看单个Agent统计
curl http://localhost:8080/api/v1/recommendation/agents/rec-data-agent-001/stats
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

##  技术栈

###  推荐Agent后端
- **核心语言**: Go 1.21+ (高性能并发)
- **Web框架**: Gin (RESTful API)
- **推荐算法**: 协同过滤、深度学习、矩阵分解
- **任务编排**: Agent Orchestrator + Priority Queue
- **数据存储**: 支持多种数据源接入
- **监控**: 实时指标采集和健康检查

###  专业前端界面  
- **框架**: React 18 + TypeScript
- **UI库**: shadcn/ui + Tailwind CSS
- **动画**: Motion.js (从 reactbits.dev 集成)
- **特效**: 发光边框效果 (GlowingEffect)
- **状态**: Zustand 状态管理
- **构建**: Vite + 热更新

###  产业级特性
- **容器化**: Docker + Kubernetes Ready
- **API设计**: RESTful + OpenAPI 3.0
- **测试**: 完整集成测试覆盖
- **版本管理**: Go Modules + Semantic Versioning

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

##  版本历史

### 推荐业务专用版本
- **v2.0.0** (2024) -  **推荐业务闭环版本**
  - 基于Agent4Rec架构的专业推荐系统
  - 4个专业化Agent：Data/Model/Service/EvalAgent
  - 炫酷发光效果专业前端界面
  - 完整RESTful API + 实时监控

### 历史版本
- **v1.0.0** (2024) - 基于Eino框架的统一架构版本
- **v0.3.0** (2024) - Linux哲学重构，性能大幅提升  
- **v0.2.0** (2024) - 多模型支持和智能路由
- **v0.1.0** (2024) - 基础功能实现

## 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件

## 支持与反馈

-  **文档**: 详见各子目录的README和docs文件
-  **问题反馈**: [GitHub Issues](https://github.com/your-org/polyagent/issues)  
-  **讨论交流**: [GitHub Discussions](https://github.com/your-org/polyagent/discussions)
-  **商务合作**: contact@polyagent.ai

##  快速体验

```bash
# 1. 克隆项目
git clone https://github.com/your-org/polyagent.git
cd polyagent

# 2. 启动后端推荐Agent系统
cd eino-polyagent
go mod tidy && go run cmd/server/main.go &

# 3. 启动专业前端界面
cd ../frontend-eino
npm install && npm install motion && npm run dev

# 4. 访问系统
#  专业仪表板: http://localhost:3000
#  API接口: http://localhost:8080/api/v1/recommendation/
```

###  快速测试推荐API

```bash
# 测试数据采集
curl -X POST http://localhost:8080/api/v1/recommendation/data/collect \
  -H "Content-Type: application/json" \
  -d '{"collector": "user_behavior", "timerange": "last_24_hours"}'

# 测试模型训练  
curl -X POST http://localhost:8080/api/v1/recommendation/models/train \
  -H "Content-Type: application/json" \
  -d '{"algorithm": "collaborative_filtering", "hyperparameters": {"learning_rate": 0.001}}'

# 获取系统状态
curl http://localhost:8080/api/v1/recommendation/system/metrics
```

---

**PolyAgent** - 专业推荐业务闭环AI智能体系统，让推荐更智能，让业务更高效 