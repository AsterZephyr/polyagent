#!/bin/bash

# PolyAgent Go Services 开发环境启动脚本

set -e

echo "Starting PolyAgent Go Services in development mode..."

# 检查依赖
echo "Checking dependencies..."

# 检查 PostgreSQL
if ! command -v pg_isready &> /dev/null; then
    echo "WARNING: PostgreSQL not found. Please install PostgreSQL and ensure it's running."
else
    echo "PostgreSQL found"
fi

# 检查 Redis
if ! command -v redis-cli &> /dev/null; then
    echo "WARNING: Redis not found. Please install Redis and ensure it's running."
else
    echo "Redis found"
fi

# 创建必要的目录
mkdir -p build logs

# 构建服务
echo "Building services..."
go build -o build/gateway ./gateway
go build -o build/scheduler ./scheduler

# 设置环境变量
export LOG_LEVEL=debug
export GIN_MODE=debug

# 启动 PostgreSQL (如果使用 Docker)
echo "Starting PostgreSQL..."
if command -v docker &> /dev/null; then
    docker run -d \
        --name polyagent-postgres \
        -p 5432:5432 \
        -e POSTGRES_DB=polyagent \
        -e POSTGRES_USER=user \
        -e POSTGRES_PASSWORD=pass \
        postgres:15-alpine 2>/dev/null || echo "PostgreSQL container already running"
fi

# 启动 Redis (如果使用 Docker)
echo "Starting Redis..."
if command -v docker &> /dev/null; then
    docker run -d \
        --name polyagent-redis \
        -p 6379:6379 \
        redis:7-alpine 2>/dev/null || echo "Redis container already running"
fi

# 等待服务启动
echo "Waiting for services to be ready..."
sleep 3

# 启动调度器 (后台运行)
echo "Starting Task Scheduler..."
./build/scheduler > logs/scheduler.log 2>&1 &
SCHEDULER_PID=$!

# 启动网关
echo "Starting API Gateway..."
echo "Gateway logs:"
./build/gateway

# 清理函数
cleanup() {
    echo "Cleaning up..."
    if [ ! -z "$SCHEDULER_PID" ]; then
        kill $SCHEDULER_PID 2>/dev/null || true
    fi
    
    if command -v docker &> /dev/null; then
        docker stop polyagent-postgres polyagent-redis 2>/dev/null || true
        docker rm polyagent-postgres polyagent-redis 2>/dev/null || true
    fi
}

# 注册清理函数
trap cleanup EXIT

# 等待中断信号
wait