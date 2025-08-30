# PolyAgent 分布式AI智能体系统

高性能、可扩展的分布式AI智能体平台，基于微服务架构设计，支持多AI提供商、智能路由、工作流编排和企业级部署。

## 系统架构

PolyAgent采用分层微服务架构，确保各组件职责清晰、边界明确：

```
接入层 (Gateway Layer)     │ 应用层 (Application Layer)   │ AI层 (AI Layer)
─────────────────────────   │ ─────────────────────────────  │ ──────────────────
• API网关 (Go)              │ • 智能体服务 (Python)          │ • 模型路由器
• 负载均衡                  │ • 工作流引擎 (Python)          │ • AI适配器
• 认证授权                  │ • 会话管理                     │ • 上下文管理器
• 限流熔断                  │ • 工具编排器                   │ • 安全过滤器
```

## 核心特性

### 🚀 高性能分布式架构
- 微服务设计，各组件独立扩展
- 智能负载均衡和故障转移
- 分布式追踪和监控
- 服务网格支持

### 🤖 多AI提供商支持  
- OpenAI (GPT-4, GPT-5)
- Anthropic (Claude-4, Claude-3.5)
- OpenRouter (开源模型)
- GLM (中文模型)
- 统一API接口，便于扩展

### 🧠 智能模型路由
- 基于任务类型自动选择最优模型
- 成本优化和性能平衡
- A/B测试和流量分流
- 实时健康监控

### ⚡ 高级功能
- 流式响应 (Server-Sent Events)
- 工具调用和函数执行
- 多轮对话和上下文管理
- 工作流编排和状态机
- 安全过滤和医疗安全检查

## 快速开始

### 开发环境搭建

```bash
# 克隆代码
git clone https://github.com/your-org/polyagent.git
cd polyagent

# 安装 Go 依赖 (网关服务)
cd pkg/gateway && go mod tidy

# 安装 Python 依赖 (核心服务)
cd ../services && pip install -r requirements.txt

# 配置环境变量
cp config/env.example config/.env
# 编辑 .env 文件添加 API Keys
```

### 本地运行

```bash
# 启动网关服务 (Go)
cd cmd/gateway && go run main.go

# 启动智能体服务 (Python)  
cd cmd/agent-service && python main.py

# 启动工作流引擎 (Python)
cd cmd/workflow-engine && python main.py
```

### Docker 部署

```bash
# 构建镜像
docker-compose build

# 启动所有服务
docker-compose up -d

# 检查服务状态
docker-compose ps
```

### Kubernetes 部署

```bash
# 部署到 K8s 集群
kubectl apply -f deployments/k8s/

# 检查部署状态
kubectl get pods -l app=polyagent
```

## API 使用示例

### 基本对话

```bash
curl -X POST "http://localhost:8080/v1/chat" \
  -H "Authorization: Bearer $API_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "message": "你好，请帮我分析一下机器学习的发展趋势",
    "use_tools": true
  }'
```

### 流式对话

```bash
curl -X POST "http://localhost:8080/v1/chat/stream" \
  -H "Authorization: Bearer $API_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "message": "写一个Python排序算法的实现",
    "stream_mode": true
  }'
```

### 创建智能体

```bash  
curl -X POST "http://localhost:8080/v1/agents" \
  -H "Authorization: Bearer $API_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "代码审查助手",
    "agent_type": "tool_calling",
    "system_prompt": "你是一个专业的代码审查助手...",
    "tools_enabled": true
  }'
```

## 项目结构

### 目录组织

```
polyagent/
├── cmd/                    # 应用程序入口点
│   ├── gateway/           # API网关启动器
│   ├── agent-service/     # 智能体服务启动器  
│   └── workflow-engine/   # 工作流引擎启动器
├── pkg/                   # 核心业务包
│   ├── gateway/           # 网关层实现
│   ├── services/          # 应用服务层
│   ├── ai/               # AI处理层
│   ├── data/             # 数据访问层
│   └── infrastructure/    # 基础设施层
├── internal/             # 内部共享代码
│   ├── config/           # 配置管理
│   ├── middleware/       # 中间件
│   └── utils/           # 工具函数
├── api/                  # API定义
│   ├── openapi/         # OpenAPI 规范
│   └── proto/           # gRPC 协议定义
├── deployments/          # 部署配置
│   ├── k8s/             # Kubernetes 清单
│   ├── docker/          # Docker 配置
│   └── helm/            # Helm Charts
├── docs/                # 技术文档
└── tests/               # 测试代码
```

### 服务边界

| 服务 | 技术栈 | 职责 | 接口 |
|------|--------|------|------|
| **Gateway Service** | Go + Gin | HTTP接入、负载均衡、认证限流 | REST API |
| **Agent Service** | Python + FastAPI | 智能体管理、会话处理、上下文维护 | gRPC + REST |  
| **Workflow Engine** | Python + Celery | 工作流编排、任务调度、状态管理 | gRPC |
| **Model Router** | Python + AsyncIO | 模型路由、健康监控、成本优化 | gRPC |
| **Tool Orchestrator** | Python | 工具调用、安全检查、结果聚合 | gRPC |

## 核心组件

### 1. Gateway Service (pkg/gateway/)

```go
type GatewayService interface {
    HandleChatRequest(ctx context.Context, req *ChatRequest) (*ChatResponse, error)
    Authenticate(ctx context.Context, token string) (*UserContext, error)
    CheckRateLimit(ctx context.Context, userID string) error
}
```

**关键特性:**
- HTTP/2 和 gRPC 支持
- JWT 认证和 RBAC 授权
- 令牌桶算法限流
- 熔断器和故障转移
- 分布式追踪集成

### 2. Agent Service (pkg/services/)

```python
class AgentService:
    async def create_agent(self, config: AgentConfig) -> str
    async def process_message(self, session_id: str, message: str) -> ProcessResult
    async def stream_response(self, session_id: str, message: str) -> AsyncGenerator[str, None]
```

**关键特性:**
- 智能体生命周期管理
- 多轮对话上下文维护
- 流式响应支持
- 工具调用编排
- 记忆和个性化

### 3. Model Router (pkg/ai/)

```python
class ModelRouter:
    async def route_request(self, request: RouteRequest) -> RouteResponse
    async def get_model_health(self) -> Dict[str, ModelHealth]
    async def update_model_weights(self, performance_data: Dict) -> bool
```

**关键特性:**
- 智能模型选择算法
- 成本和性能优化
- 健康监控和故障转移
- A/B 测试框架
- 动态权重调整

## API 文档

完整的 OpenAPI 3.0 规范: [api/openapi/polyagent-api.yaml](api/openapi/polyagent-api.yaml)

主要 API 端点:

- **POST** `/v1/chat` - 发送对话消息
- **POST** `/v1/chat/stream` - 流式对话  
- **POST** `/v1/agents` - 创建智能体
- **GET** `/v1/models` - 获取可用模型
- **GET** `/v1/health` - 系统健康检查

## 配置管理

### 环境变量

```bash
# API Keys
OPENAI_API_KEY=sk-...
ANTHROPIC_API_KEY=sk-ant-...
OPENROUTER_API_KEY=sk-or-...
GLM_API_KEY=...

# 服务配置
GATEWAY_PORT=8080
AGENT_SERVICE_URL=http://localhost:8001
MODEL_ROUTER_URL=http://localhost:8002

# 数据库
POSTGRES_URL=postgresql://localhost:5432/polyagent
REDIS_URL=redis://localhost:6379

# 监控
JAEGER_ENDPOINT=http://localhost:14268/api/traces
PROMETHEUS_ENDPOINT=http://localhost:9090
```

### 配置文件

```yaml
# config/gateway.yaml
gateway:
  port: 8080
  timeout: 30s
  rate_limit:
    requests_per_minute: 60
    burst: 10

models:
  routing_strategy: "balanced"
  cost_optimization: true
  health_check_interval: "30s"

security:
  jwt_secret: "${JWT_SECRET}"
  cors_origins: ["*"]
  require_auth: true
```

## 监控和运维

### 健康检查

```bash
# 系统整体健康状态
curl http://localhost:8080/v1/health

# 特定模型健康状态  
curl http://localhost:8080/v1/models/gpt-4/health
```

### 指标监控

系统集成 Prometheus 和 Grafana，提供丰富的监控指标:

- **请求指标**: QPS、响应时间、错误率
- **业务指标**: 活跃会话数、token消费、成本统计
- **系统指标**: CPU、内存、网络、存储
- **AI指标**: 模型性能、路由效率、成本优化

### 分布式追踪

集成 Jaeger 进行分布式追踪:

```go
// Go 服务中的追踪
span, ctx := opentracing.StartSpanFromContext(ctx, "gateway.handleChat")
defer span.Finish()
```

```python
# Python 服务中的追踪  
@trace_async("agent.process_message")
async def process_message(self, message: str) -> str:
    # 处理逻辑
```

## 性能指标

### 系统容量

| 指标 | 性能 |
|------|------|
| **并发连接** | 10,000+ |
| **QPS** | 1,000+ |  
| **响应时间 P95** | < 2s (含AI调用) |
| **网关延迟** | < 10ms |
| **内存占用** | 512MB (单服务) |
| **启动时间** | < 30s |

### 扩展能力

- **水平扩展**: 支持 Kubernetes HPA
- **垂直扩展**: 支持 CPU/内存动态调整  
- **异地多活**: 支持多区域部署
- **弹性伸缩**: 根据负载自动扩容

## 安全特性

### 认证授权
- JWT Token 认证
- RBAC 权限控制
- API Key 管理
- 多租户隔离

### 数据安全
- API Key 加密存储
- 请求响应脱敏
- 审计日志记录
- 敏感数据标记

### 网络安全
- HTTPS/TLS 加密
- CORS 跨域控制
- 请求签名验证
- IP 白名单限制

## 开发指南

### 添加新的AI提供商

1. 实现 AIProvider 接口:

```python
class NewProviderAdapter(AIProvider):
    async def call_model(self, request: AIRequest) -> AIResponse:
        # 实现具体的API调用逻辑
        pass
```

2. 注册到模型路由器:

```python
router.register_provider("new_provider", NewProviderAdapter())
```

3. 添加模型配置:

```yaml
models:
  new_model:
    provider: "new_provider"
    capabilities: ["text_generation"]
    cost_per_1k_tokens: 0.002
```

### 添加自定义工具

```python
@register_tool("custom_tool")
async def custom_tool(param1: str, param2: int) -> Dict[str, Any]:
    """自定义工具实现"""
    # 工具逻辑
    return {"result": "success"}
```

### 创建工作流

```python
workflow = WorkflowBuilder() \
    .add_step("analyze", AnalyzeStep()) \
    .add_step("generate", GenerateStep()) \
    .add_condition("should_review", lambda ctx: ctx.complexity > 0.8) \
    .add_step("review", ReviewStep(), condition="should_review") \
    .build()
```

## 测试策略

### 单元测试

```bash
# Go 服务测试
cd pkg/gateway && go test -v ./...

# Python 服务测试
cd pkg/services && python -m pytest -v
```

### 集成测试

```bash
# 端到端测试
cd tests/integration && python -m pytest -v

# 负载测试
cd tests/performance && go test -bench=.
```

### API 测试

```bash
# 使用 Newman 运行 Postman 集合
newman run tests/api/polyagent-api-tests.json
```

## 部署指南

### 本地开发

```bash
# 使用 Docker Compose
docker-compose -f docker-compose.dev.yml up -d
```

### 生产部署

```bash
# Kubernetes 部署
kubectl apply -f deployments/k8s/namespace.yaml
kubectl apply -f deployments/k8s/configmap.yaml  
kubectl apply -f deployments/k8s/secret.yaml
kubectl apply -f deployments/k8s/deployment.yaml
kubectl apply -f deployments/k8s/service.yaml
kubectl apply -f deployments/k8s/ingress.yaml
```

### Helm 部署

```bash
# 添加 Helm 仓库
helm repo add polyagent https://charts.polyagent.ai

# 安装
helm install polyagent polyagent/polyagent \
  --set config.apiKeys.openai="sk-..." \
  --set ingress.enabled=true
```

## 技术栈

### 后端服务
- **Go**: 网关服务 (Gin, gRPC, OpenTelemetry)
- **Python**: 核心服务 (FastAPI, AsyncIO, Celery)
- **PostgreSQL**: 主数据库
- **Redis**: 缓存和会话存储
- **Elasticsearch**: 日志搜索和分析

### 基础设施
- **Kubernetes**: 容器编排
- **Istio**: 服务网格  
- **Prometheus**: 指标监控
- **Grafana**: 监控面板
- **Jaeger**: 分布式追踪
- **ELK Stack**: 日志聚合

### CI/CD
- **GitHub Actions**: 持续集成
- **ArgoCD**: 持续部署
- **Helm**: 包管理
- **Terraform**: 基础设施即代码

## 贡献指南

### 开发流程

1. Fork 项目并创建特性分支
2. 遵循代码规范和提交规范
3. 编写测试并确保测试通过  
4. 提交 Pull Request

### 代码规范

- **Go**: 遵循 `gofmt` 和 `golint` 规范
- **Python**: 遵循 PEP 8 和 Black 格式化
- **提交消息**: 遵循 Conventional Commits 规范
- **API**: 遵循 REST 和 OpenAPI 3.0 规范

### Code Review 检查项

- [ ] 代码符合团队规范
- [ ] 单元测试覆盖率 > 80%
- [ ] 集成测试通过
- [ ] 性能测试无回归
- [ ] 安全扫描无高危漏洞
- [ ] API 文档已更新
- [ ] 部署脚本已验证

## 许可证

MIT License - 详见 [LICENSE](LICENSE) 文件

## 支持

- **文档**: [docs/](docs/)
- **Issues**: [GitHub Issues](https://github.com/your-org/polyagent/issues)
- **讨论**: [GitHub Discussions](https://github.com/your-org/polyagent/discussions)
- **邮件**: support@polyagent.ai

---

PolyAgent - 企业级分布式AI智能体平台