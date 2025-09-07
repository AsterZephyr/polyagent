# PolyAgent 架构说明

## 系统定位转变

本系统已从**通用AI Agent系统**转变为**专业化推荐业务Agent系统**。

## 架构组件分类

### 🎯 推荐业务核心组件 (Production Ready)

```
internal/recommendation/          # 推荐业务专用Agent系统 ⭐
├── orchestrator.go              # 推荐任务编排器
├── data_agent.go               # 数据采集和特征工程Agent
├── model_agent.go              # 模型训练和优化Agent
├── api_handler.go              # HTTP API接口
├── agent_types.go              # Agent类型定义
└── integration_test.go         # 集成测试

cmd/server/
├── recommendation_server.go     # 推荐业务专用服务器 ⭐
└── main.go                     # 通用Agent服务器 (遗留)
```

### 🔧 共享基础组件 (Used by Recommendation System)

```
internal/config/                 # 配置管理 (共享)
internal/ai/                     # AI模型路由 (共享)
```

### 📦 通用Agent组件 (Legacy - Not Used in Recommendation Mode)

```
internal/orchestration/          # 通用智能体编排 (遗留)
├── agent_orchestrator.go       # 通用Agent编排器
├── workflow_engine.go           # 工作流引擎
├── workflow_builder.go          # 工作流构建器
├── tools.go                     # 通用工具
├── context_manager.go           # 上下文管理
└── core_interfaces.go           # 通用接口定义
```

## 启动方式

### 推荐业务专用服务器 (推荐)
```bash
# 方式1: 使用Makefile
make run-rec

# 方式2: 直接运行
go run cmd/server/recommendation_server.go

# 方式3: 构建后运行
make build-rec
./bin/recommendation-agent-server
```

### 通用Agent服务器 (遗留功能)
```bash
# 包含对话、工作流等通用功能
go run cmd/server/main.go
```

## 测试命令

```bash
# 推荐业务专用测试
make test-rec

# 全部测试 (包含遗留功能)
make test
```

## API端点

### 推荐业务API (生产就绪)
- `/api/v1/recommendation/*` - 推荐业务专用API
- `/health` - 健康检查

### 通用Agent API (遗留)
- `/api/v1/chat` - 对话功能
- `/api/v1/agents` - 通用Agent管理
- `/api/v1/workflows/*` - 工作流功能

## 清理建议

如果完全专注于推荐业务，可以考虑移除以下遗留组件：
1. `internal/orchestration/` 目录
2. `cmd/server/main.go` 中的通用Agent功能
3. 通用工作流相关代码

## 迁移路径

1. **当前状态**: 推荐业务功能完整，通用功能保留
2. **建议路径**: 创建专门分支保存通用功能，主分支专注推荐业务
3. **完全专业化**: 移除所有通用Agent遗留代码