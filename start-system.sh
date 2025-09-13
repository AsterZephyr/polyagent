#!/bin/bash

# PolyAgent 推荐系统启动脚本
echo "🚀 启动 PolyAgent 推荐系统..."

# 检查Go环境
if ! command -v go &> /dev/null; then
    echo "❌ Go 未安装，请先安装 Go 1.21+"
    exit 1
fi

# 检查Node.js环境
if ! command -v node &> /dev/null; then
    echo "❌ Node.js 未安装，请先安装 Node.js 18+"
    exit 1
fi

# 启动后端服务
echo "🔧 启动后端服务..."
cd eino-polyagent
go run cmd/server/main.go &
BACKEND_PID=$!

# 等待后端启动
echo "⏳ 等待后端服务启动..."
sleep 5

# 检查后端是否启动成功
if ! curl -s http://localhost:8080/health > /dev/null; then
    echo "❌ 后端服务启动失败"
    kill $BACKEND_PID 2>/dev/null
    exit 1
fi

echo "✅ 后端服务启动成功 (PID: $BACKEND_PID)"

# 启动前端服务
echo "🎨 启动前端服务..."
cd ../v0-polyagent
npm run dev &
FRONTEND_PID=$!

echo "✅ 前端服务启动成功 (PID: $FRONTEND_PID)"

echo ""
echo "🎉 PolyAgent 推荐系统启动完成！"
echo "📊 后端API: http://localhost:8080"
echo "🎨 前端界面: http://localhost:3000"
echo ""
echo "按 Ctrl+C 停止所有服务"

# 等待用户中断
trap "echo '🛑 正在停止服务...'; kill $BACKEND_PID $FRONTEND_PID 2>/dev/null; exit 0" INT
wait
