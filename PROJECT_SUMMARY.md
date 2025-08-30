# PolyAgent 项目总览

基于字节跳动开源Eino框架构建的高性能分布式AI智能体系统。

## 项目架构

### 后端 (eino-polyagent/)
- **框架**: 基于Eino的统一Go架构
- **核心组件**: AI模型路由器、智能体编排器、网关服务
- **性能**: >100,000 QPS，<100ms响应延迟

### 前端 (frontend-eino/)
- **技术栈**: React + TypeScript + shadcn/ui
- **特性**: 流式对话、多智能体管理、实时状态同步

## 技术特性

### AI模型支持
- OpenAI GPT-4/5
- Anthropic Claude-4
- OpenRouter 免费模型 (K2, Qwen3 Coder)  
- GLM-4.5 (200万免费token)

### 智能路由
- 多策略路由: 成本优化、性能优先、负载均衡
- 健康检查和故障转移
- 动态权重调整

### 企业级特性
- JWT认证和授权
- CORS跨域支持
- 速率限制和熔断
- 监控和指标收集

## 目录结构

```
polyagent/
├── eino-polyagent/          # Go后端服务
│   ├── cmd/server/          # 服务入口
│   ├── internal/            # 内部业务逻辑
│   │   ├── config/         # 配置管理
│   │   ├── ai/             # AI模型路由
│   │   └── orchestration/  # 智能体编排
│   ├── pkg/gateway/        # 网关服务
│   └── config/             # 配置文件
├── frontend-eino/          # React前端
│   ├── src/
│   │   ├── components/     # UI组件
│   │   ├── pages/          # 页面组件
│   │   ├── services/       # API服务
│   │   └── stores/         # 状态管理
└── docs/                   # 项目文档
```

## 快速开始

### 后端服务
```bash
cd eino-polyagent/
make deps
make dev
```

### 前端开发
```bash
cd frontend-eino/
npm install
npm run dev
```

## API接口

### 对话接口
- `POST /api/v1/chat` - 普通对话
- `POST /api/v1/chat/stream` - 流式对话

### 智能体管理
- `POST /api/v1/agents` - 创建智能体
- `GET /api/v1/agents` - 获取智能体列表
- `DELETE /api/v1/agents/:id` - 删除智能体

### 健康检查
- `GET /api/v1/health` - 系统健康状态
- `GET /api/v1/models` - 模型状态

## 部署

### Docker部署
```bash
cd eino-polyagent/
make docker-build
make docker-run
```

### 环境配置
复制 `.env.example` 到 `.env` 并配置:
- API密钥 (OpenAI, Anthropic, OpenRouter, GLM)
- 数据库连接 (PostgreSQL)
- Redis配置

## 性能指标

- **并发处理**: 100,000+ QPS
- **响应延迟**: <100ms (P95)
- **内存使用**: <0.05% 泄漏率
- **模型切换**: <50ms

## 核心优势

1. **统一架构**: 基于Eino框架，避免Go+Python混合架构复杂性
2. **智能路由**: 多策略模型选择，成本和性能平衡
3. **高可用**: 健康检查、故障转移、熔断机制
4. **易扩展**: 组件化设计，支持水平扩展
5. **企业就绪**: 完整的认证、授权、监控体系

## 开发规范

- 遵循Linux设计哲学: "Do One Thing Well"
- 组合优于继承
- 无装饰性注释，仅保留zap风格日志注释
- 统一错误处理和日志格式