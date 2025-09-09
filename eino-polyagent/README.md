# PolyAgent - 推荐业务智能体系统

**专业化推荐业务闭环Agent系统**，基于Agent4Rec等成功案例构建的生产级推荐服务平台。

## 🎯 系统定位

专门针对**推荐业务场景**设计的智能体系统，支持从数据采集到模型部署的完整推荐业务闭环：

```
数据采集 → 特征工程 → 模型训练 → 评估优化 → 部署服务 → 监控反馈
   ↑                                                               ↓
API接口 ←←←←←←←←←←←← 业务闭环监控 ←←←←←←←←←←←← 实时推荐服务
```

## 🚀 核心特性

### 推荐业务专用Agent
- **DataAgent** - 数据采集、清洗、特征工程
- **ModelAgent** - 模型训练、评估、超参优化、部署
- **智能调度** - 任务编排、重试机制、性能监控

### 推荐算法支持
- **协同过滤** (Collaborative Filtering)
- **内容推荐** (Content-based)  
- **矩阵分解** (Matrix Factorization)
- **深度学习** (Deep Learning)

### 生产级特性
- **高性能架构** - Go语言，支持高并发
- **完整API** - RESTful接口，易于集成
- **实时监控** - 业务指标、系统监控
- **容器化部署** - Docker支持，易于扩展

## 📖 快速开始

### 1. 启动推荐Agent服务器

```bash
cd /Users/hxz/code/polyagent/eino-polyagent

# 方式1: 使用Makefile (推荐)
make run

# 方式2: 直接运行推荐专用服务器
go run cmd/server/main.go

# 方式3: 构建后运行
make build
./bin/recommendation-agent-server
```

### 2. 启动现代化前端界面

```bash
cd /Users/hxz/code/polyagent/frontend-eino
npm install  # 仅首次需要
npm run dev
```

**访问应用**: 
- 前端界面: http://localhost:3000
- 后端API: http://localhost:8080
- Claude Analytics: http://localhost:3333 (可选)

### 3. 测试推荐业务API

**数据采集**:
```bash
curl -X POST http://localhost:8080/api/v1/recommendation/data/collect \
  -H "Content-Type: application/json" \
  -d '{"collector": "user_behavior", "timerange": "last_7_days"}'
```

**模型训练**:
```bash
curl -X POST http://localhost:8080/api/v1/recommendation/models/train \
  -H "Content-Type: application/json" \
  -d '{"algorithm": "collaborative_filtering", "hyperparameters": {"learning_rate": 0.001}}'
```

**推荐预测**:
```bash
curl -X POST http://localhost:8080/api/v1/recommendation/predict \
  -H "Content-Type: application/json" \
  -d '{"user_id": "user123", "top_k": 10}'
```

### 3. 系统监控

```bash
# 系统指标
curl http://localhost:8080/api/v1/recommendation/system/metrics

# Agent状态
curl http://localhost:8080/api/v1/recommendation/agents
```

## 🔧 推荐业务API

### 数据操作
- `POST /api/v1/recommendation/data/collect` - 数据采集
- `POST /api/v1/recommendation/data/features` - 特征工程
- `POST /api/v1/recommendation/data/validate` - 数据验证

### 模型操作
- `POST /api/v1/recommendation/models/train` - 模型训练
- `POST /api/v1/recommendation/models/evaluate` - 模型评估
- `POST /api/v1/recommendation/models/optimize` - 超参优化
- `POST /api/v1/recommendation/models/deploy` - 模型部署
- `GET /api/v1/recommendation/models` - 模型列表

### 推荐服务
- `POST /api/v1/recommendation/predict` - 推荐预测
- `POST /api/v1/recommendation/recommend` - 推荐接口

### 系统监控
- `GET /api/v1/recommendation/agents` - Agent状态
- `GET /api/v1/recommendation/system/metrics` - 系统指标
- `GET /api/v1/recommendation/health` - 健康检查

## 🏗️ 项目结构

```
eino-polyagent/
├── cmd/server/                    # 服务入口
├── internal/
│   ├── recommendation/           # 推荐业务Agent系统 ⭐
│   │   ├── orchestrator.go      # 推荐任务编排器
│   │   ├── data_agent.go        # 数据Agent
│   │   ├── model_agent.go       # 模型Agent
│   │   ├── api_handler.go       # HTTP API接口
│   │   ├── agent_types.go       # Agent类型定义
│   │   └── integration_test.go  # 集成测试
│   ├── ai/                      # AI模型路由 (通用)
│   ├── orchestration/           # 通用智能体编排
│   └── config/                  # 配置管理
├── config/                      # 配置文件
└── Makefile                    # 构建脚本
```

## 🧪 测试验证

```bash
# 推荐系统专用测试 (推荐)
make test-rec

# 构建推荐专用服务器
make build-rec

# 启动推荐专用服务器并测试API
make run-rec &
curl http://localhost:8080/api/v1/recommendation/health
```

## 📋 架构文档

详细的系统架构和组件说明请参考：[ARCHITECTURE.md](ARCHITECTURE.md)

## 📊 业务指标

推荐系统支持监控以下关键指标：

- **数据质量**: 完整性、准确性、一致性
- **模型性能**: RMSE、MAE、Precision@K、Recall@K、NDCG@K
- **系统性能**: QPS、延迟、成功率、资源使用
- **业务效果**: 点击率、转化率、覆盖率、多样性

## 🔄 推荐业务闭环

1. **数据采集** - 用户行为、物品特征、交互日志
2. **特征工程** - 用户画像、物品相似度、行为序列
3. **模型训练** - 多算法支持、自动调参、增量学习
4. **模型评估** - 离线评估、A/B测试、效果分析
5. **模型部署** - 热更新、版本管理、服务治理
6. **监控反馈** - 实时监控、异常告警、效果跟踪

## 📈 性能指标

- **并发处理**: >10,000 QPS
- **响应延迟**: <100ms (预测接口)
- **训练效率**: 支持大规模数据训练
- **模型精度**: 支持多种评估指标

## 🚢 部署方式

### Docker部署
```bash
make docker-build
make docker-run
```

### 生产部署
```bash
# 构建
make build

# 运行
./main
```

## 🤝 基于开源实践

本系统基于以下成功案例和框架设计：
- **Agent4Rec** (SIGIR 2024) - 生成式推荐Agent
- **Microsoft Recommenders** - 工业级推荐框架
- **字节跳动Eino框架** - Go语言AI框架

---

**注**: 本系统专门针对推荐业务场景优化，如需通用AI Agent功能，请参考其他分支或版本。