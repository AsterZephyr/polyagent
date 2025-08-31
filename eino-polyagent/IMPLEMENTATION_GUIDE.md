# PolyAgent 实施指南

## 系统概述

PolyAgent是基于串行workflow的Multi-Agent系统，采用Go语言实现，使用ByteDance Eino框架。系统通过工具范式实现Agent协作，避免并行Agent间的复杂交互。

## 核心架构

```
HTTP Gateway → Agent Orchestrator → Workflow Engine → Model Router
     ↓              ↓                    ↓              ↓
   REST API    Agent管理           串行执行        AI模型调用
```

## 关键组件

### 1. Agent Orchestrator
**文件**: `internal/orchestration/agent_orchestrator.go`
**功能**: 管理Agent生命周期，协调Agent间通信
**接口**:
```go
func (o *AgentOrchestrator) ProcessMessage(ctx context.Context, agentID, sessionID, message, userID string) (*ProcessResult, error)
```

### 2. Workflow Engine  
**文件**: `internal/orchestration/workflow_engine.go`
**功能**: 执行串行工作流，管理步骤间上下文传递
**接口**:
```go
func (we *WorkflowEngine) ExecuteWorkflow(ctx context.Context, workflowID string, userID string) error
```

### 3. Model Router
**文件**: `internal/ai/model_router.go`
**功能**: 智能路由AI模型请求，支持多种策略
**接口**:
```go
func (r *ModelRouter) Route(ctx context.Context, req *RouteRequest) (*RouteResponse, error)
```

### 4. Gateway Service
**文件**: `pkg/gateway/service.go`
**功能**: HTTP API网关，处理认证、限流、路由
**接口**:
```go
func (s *GatewayService) Start() error
```

## 配置管理

### 环境变量
```bash
# AI模型API密钥
OPENAI_API_KEY=sk-your-key
ANTHROPIC_API_KEY=sk-ant-your-key
OPENROUTER_API_KEY=sk-or-your-key
GLM_API_KEY=your-glm-key

# 数据库配置
DB_HOST=localhost
DB_PORT=5432
DB_NAME=polyagent

# JWT密钥
JWT_SECRET_KEY=your-secret-key
```

### 配置文件
**位置**: `config/config.yaml`
**内容**: AI模型配置、服务器设置、安全配置

## 部署步骤

### 1. 环境准备
```bash
# 安装Go 1.21+
go version

# 安装PostgreSQL
brew install postgresql

# 安装Redis  
brew install redis
```

### 2. 项目构建
```bash
cd eino-polyagent/
make deps    # 安装依赖
make build   # 构建二进制
```

### 3. 配置设置
```bash
cp .env.example .env
# 编辑.env文件设置API密钥
```

### 4. 服务启动
```bash
make run     # 启动服务
# 或
./bin/polyagent-server
```

## API使用

### 基础对话
```bash
curl -X POST http://localhost:8080/api/v1/chat \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "message": "你好",
    "session_id": "session_123"
  }'
```

### 工作流执行
```bash
curl -X POST http://localhost:8080/api/v1/workflow/execute \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "workflow_name": "编程任务",
    "template": "coding",
    "agent_id": "default",
    "user_message": "实现快速排序算法"
  }'
```

### Agent管理
```bash
# 创建Agent
curl -X POST http://localhost:8080/api/v1/agents \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "name": "代码助手",
    "type": "conversational",  
    "system_prompt": "你是专业的编程助手"
  }'

# 获取Agent列表
curl http://localhost:8080/api/v1/agents \
  -H "Authorization: Bearer $TOKEN"
```

## 开发指南

### 添加新的Workflow模板
1. **定义步骤**
```go
func NewCustomWorkflow(agentID string) *WorkflowBuilder {
    return NewWorkflowBuilder().
        Named("自定义工作流", "描述").
        AddAnalysisStep("分析步骤", agentID).
        AddGenerationStep("生成步骤", agentID)
}
```

2. **注册模板**
```go
// 在 workflow_builder.go 中添加
templates["custom"] = WorkflowTemplate{
    Name: "自定义工作流",
    Description: "工作流描述",
    Steps: []StepTemplate{...},
}
```

### 扩展AI模型支持
1. **模型配置**
```yaml
ai:
  models:
    new_model:
      provider: "new_provider"
      model_name: "model-name"
      api_key: "${API_KEY}"
      priority: 7
```

2. **Provider实现**
```go
func (r *ModelRouter) createChatModel(cfg ModelConfig) (model.ChatModel, error) {
    switch cfg.Provider {
    case "new_provider":
        return NewProviderChatModel(&cfg)
    }
}
```

### 自定义Agent类型
1. **实现Agent接口**
```go
type CustomAgent struct {
    id     string
    config *AgentConfig
}

func (ca *CustomAgent) Process(ctx context.Context, req *ProcessRequest) (*ProcessResponse, error) {
    // 实现处理逻辑
}
```

2. **注册Agent类型**
```go
// 在 agent_orchestrator.go 中添加
case AgentTypeCustom:
    agent, err = o.createCustomAgent(config)
```

## 性能调优

### 内存优化
- 限制上下文历史长度
- 定期清理过期会话
- 使用对象池减少GC压力

### 并发优化  
- 合理设置连接池大小
- 使用读写锁优化热点代码
- 避免不必要的锁竞争

### 网络优化
- 启用HTTP/2
- 配置合适的超时时间
- 使用连接复用

## 监控指标

### 关键指标
- 请求QPS和响应时间
- Agent调用次数和成功率
- 模型路由分布和成本
- 工作流执行时间和失败率

### 监控工具
- Prometheus指标收集
- Grafana可视化面板
- Jaeger链路追踪
- 结构化日志分析

## 故障排查

### 常见问题
1. **API密钥配置错误**
   - 检查环境变量设置
   - 验证密钥格式和权限

2. **数据库连接失败**
   - 确认PostgreSQL服务运行
   - 检查连接配置和网络

3. **模型调用超时**
   - 检查网络连接
   - 调整超时配置
   - 验证API配额

### 调试方法
- 查看结构化日志
- 使用健康检查端点
- 监控系统资源使用
- 分析链路追踪数据

## 生产环境注意事项

### 安全配置
- 使用强JWT密钥
- 启用HTTPS
- 配置防火墙规则
- 定期更新依赖

### 高可用部署
- 多实例部署
- 负载均衡配置
- 数据库主从复制
- Redis集群模式

### 备份策略
- 数据库定期备份
- 配置文件版本管理
- 日志归档和轮转
- 监控数据保留

## 测试策略

### 单元测试
```bash
go test ./internal/...
```

### 集成测试
```bash
go test ./tests/integration/...
```

### 性能测试
```bash
go test -bench=. ./internal/...
```

### 端到端测试
```bash
# 使用实际API进行测试
curl测试脚本或Postman集合
```

## 扩展建议

### 短期改进
- 添加更多工作流模板
- 优化上下文压缩算法
- 完善错误处理机制

### 长期规划
- 支持插件系统
- 实现多租户架构
- 添加可视化调试工具
- 集成更多AI模型

这个实施指南提供了系统部署和使用的完整流程，开发者可以按照指南快速搭建和扩展PolyAgent系统。