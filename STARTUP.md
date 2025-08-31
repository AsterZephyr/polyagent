# PolyAgent 完整启动指南

## 🚀 快速启动

### 环境要求

- **后端**: Go 1.21+
- **前端**: Node.js 18+ / npm
- **操作系统**: macOS/Linux/Windows

### 1. 启动后端服务

```bash
# 切换到后端目录
cd /Users/hxz/code/polyagent/eino-polyagent

# 方式1: 直接运行 (推荐)
PORT=8082 go run cmd/server/main.go

# 方式2: 先构建再运行
go build -o bin/server cmd/server/main.go
PORT=8082 ./bin/server

# 方式3: 使用Makefile
make build && PORT=8082 ./bin/polyagent-server
```

**后端服务将在 http://localhost:8082 启动**

### 2. 启动前端服务

```bash
# 切换到前端目录 (新终端窗口)
cd /Users/hxz/code/polyagent/frontend-eino

# 安装依赖 (仅首次需要)
npm install

# 启动开发服务器
npm run dev
```

**前端服务将在 http://localhost:3000 启动**

### 3. 验证服务

```bash
# 测试后端健康检查
curl http://localhost:8082/api/v1/health

# 测试前端代理
curl http://localhost:3000/api/v1/health

# 测试聊天API
curl -X POST http://localhost:3000/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{"message":"Hello PolyAgent"}'
```

## 📋 API 端点

### 后端直接访问 (端口8082)

| 端点 | 方法 | 说明 |
|------|------|------|
| `/health` | GET | 服务健康检查 |
| `/api/v1/health` | GET | API健康检查 |
| `/api/v1/chat` | POST | 聊天对话 |
| `/api/v1/agents` | GET/POST | Agent管理 |
| `/api/v1/workflows/:type/execute` | POST | 工作流执行 |

### 前端代理访问 (端口3000)

- 前端自动将 `/api` 请求代理到后端8082端口
- 直接访问 http://localhost:3000 使用Web界面

## 🛠️ 常用命令

### 后端开发

```bash
cd /Users/hxz/code/polyagent/eino-polyagent

# 运行测试
go test ./tests/ -v

# 构建项目
go build ./...

# 清理构建产物
rm -f bin/* server

# 查看进程
ps aux | grep "go run"
```

### 前端开发

```bash
cd /Users/hxz/code/polyagent/frontend-eino

# 类型检查
npm run type-check

# 代码检查
npm run lint

# 构建生产版本
npm run build

# 预览生产版本
npm run preview
```

## 🔧 故障排除

### 端口占用问题

```bash
# 查看端口占用
lsof -i:8080,8081,8082,3000

# 杀死占用进程
kill -9 <PID>
```

### 停止所有服务

```bash
# 杀死Go进程
pkill -f "go run cmd/server/main.go"

# 杀死npm进程
pkill -f "npm run dev"

# 或使用Ctrl+C在对应终端停止
```

### 清理环境

```bash
# 清理Go构建产物
cd /Users/hxz/code/polyagent/eino-polyagent
rm -f bin/* server

# 清理前端缓存 (如有问题)
cd /Users/hxz/code/polyagent/frontend-eino
rm -rf node_modules/.cache
```

## 🎯 开发工作流

1. **启动后端**: `PORT=8082 go run cmd/server/main.go`
2. **启动前端**: `npm run dev` 
3. **访问应用**: http://localhost:3000
4. **API测试**: 使用curl或Postman测试API
5. **停止服务**: Ctrl+C或kill进程

## ✅ 验证清单

- [ ] 后端服务在8082端口响应
- [ ] 前端服务在3000端口响应  
- [ ] 前端可以代理API请求
- [ ] 聊天接口正常工作
- [ ] Agent创建功能正常
- [ ] 工作流执行正常

---

> 💡 **提示**: 建议使用两个终端窗口，一个运行后端，一个运行前端，方便查看日志和调试。