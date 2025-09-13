# PolyAgent 推荐系统

基于Agent编排模式的智能推荐系统，包含Go后端和Next.js前端。

## 🏗️ 系统架构

```
polyagent/
├── eino-polyagent/          # Go后端服务
│   ├── cmd/server/          # 服务器入口
│   ├── internal/recommendation/  # 推荐业务核心
│   └── config/              # 配置文件
├── v0-polyagent/            # Next.js前端界面
│   ├── app/                 # 页面组件
│   ├── components/          # UI组件
│   └── lib/                 # API集成
└── start-system.sh          # 一键启动脚本
```

## 🚀 快速开始

### 方式一：一键启动（推荐）

```bash
# 克隆项目
git clone <repository-url>
cd polyagent

# 一键启动前后端
./start-system.sh
```

### 方式二：分别启动

#### 启动后端服务

```bash
cd eino-polyagent

# 安装依赖
go mod tidy

# 启动服务
go run cmd/server/main.go
```

#### 启动前端服务

```bash
cd v0-polyagent

# 安装依赖
npm install

# 启动开发服务器
npm run dev
```

## 📊 访问地址

- **前端界面**: http://localhost:3000
- **后端API**: http://localhost:8080
- **API文档**: http://localhost:8080/api/v1/recommendation/health

## 🎯 功能特性

### 后端功能
- **Agent编排**: 数据Agent、模型Agent、服务Agent、评估Agent
- **推荐算法**: 协同过滤、内容推荐、矩阵分解、深度学习
- **实时监控**: 系统指标、性能监控、健康检查
- **RESTful API**: 完整的推荐业务API接口

### 前端功能
- **系统概览**: 实时监控系统状态和性能指标
- **Agent管理**: 监控和管理所有Agent的状态
- **模型管理**: 模型训练、评估、部署管理
- **数据管理**: 数据采集、特征工程、验证
- **推荐服务**: 推荐预测和测试界面

## 🔧 技术栈

### 后端
- **语言**: Go 1.21
- **框架**: Gin (HTTP框架)
- **日志**: Logrus
- **配置**: Viper

### 前端
- **框架**: Next.js 14
- **语言**: TypeScript
- **UI库**: Tailwind CSS + shadcn/ui
- **图表**: Recharts
- **状态管理**: React Hooks

## 📡 API接口

### 系统监控
- `GET /api/v1/recommendation/system/metrics` - 系统指标
- `GET /api/v1/recommendation/health` - 健康检查

### Agent管理
- `GET /api/v1/recommendation/agents` - 获取所有Agent
- `GET /api/v1/recommendation/agents/:id/stats` - 获取Agent统计

### 模型管理
- `GET /api/v1/recommendation/models` - 模型列表
- `POST /api/v1/recommendation/models/train` - 模型训练
- `POST /api/v1/recommendation/models/evaluate` - 模型评估
- `POST /api/v1/recommendation/models/deploy` - 模型部署

### 数据操作
- `POST /api/v1/recommendation/data/collect` - 数据采集
- `POST /api/v1/recommendation/data/features` - 特征工程
- `POST /api/v1/recommendation/data/validate` - 数据验证

### 推荐服务
- `POST /api/v1/recommendation/recommend` - 获取推荐
- `POST /api/v1/recommendation/predict` - 单物品预测

## 🛠️ 开发指南

### 后端开发

```bash
cd eino-polyagent

# 运行测试
make test

# 构建
make build

# 运行
make run
```

### 前端开发

```bash
cd v0-polyagent

# 安装依赖
npm install

# 开发模式
npm run dev

# 构建
npm run build

# 生产模式
npm start
```

## 📝 配置说明

### 后端配置
编辑 `eino-polyagent/config/config.yaml` 文件：

```yaml
server:
  host: "0.0.0.0"
  port: 8080

database:
  driver: "postgres"
  host: "localhost"
  port: 5432
  # ... 其他配置
```

### 前端配置
编辑 `v0-polyagent/next.config.mjs` 文件：

```javascript
const nextConfig = {
  env: {
    NEXT_PUBLIC_API_URL: 'http://localhost:8080',
  },
}
```

## 🐛 故障排除

### 常见问题

1. **后端启动失败**
   - 检查Go版本是否为1.21+
   - 检查端口8080是否被占用
   - 查看日志输出

2. **前端无法连接后端**
   - 确认后端服务已启动
   - 检查API URL配置
   - 查看浏览器控制台错误

3. **Agent状态异常**
   - 检查Agent健康状态
   - 查看系统日志
   - 重启相关服务

### 日志查看

```bash
# 后端日志
cd eino-polyagent
go run cmd/server/main.go

# 前端日志
cd v0-polyagent
npm run dev
```

## 🤝 贡献指南

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 📞 支持

如有问题或建议，请：
- 提交 Issue
- 发送邮件至 [your-email@example.com]
- 查看项目文档

---

**PolyAgent 推荐系统** - 让推荐更智能，让业务更高效 🚀