#!/bin/bash

# PolyAgent Python AI Service 开发环境启动脚本

set -e

echo "Starting PolyAgent Python AI Service in development mode..."

# 检查Python环境
if ! command -v python3 &> /dev/null; then
    echo "ERROR: Python 3 not found. Please install Python 3.11+"
    exit 1
fi

echo "Python version: $(python3 --version)"

# 创建虚拟环境（如果不存在）
if [ ! -d "venv" ]; then
    echo "Creating virtual environment..."
    python3 -m venv venv
fi

# 激活虚拟环境
echo "Activating virtual environment..."
source venv/bin/activate

# 安装依赖
echo "Installing dependencies..."
pip install --upgrade pip
pip install -r requirements.txt

# 检查环境配置
if [ ! -f ".env" ]; then
    echo "Creating .env file from example..."
    cp .env.example .env
    echo "Please edit .env file with your API keys and configurations"
fi

# 设置开发环境变量
export DEBUG=true
export LOG_LEVEL=DEBUG

echo "Starting FastAPI server..."
echo "Access the service at: http://localhost:8000"
echo "API documentation at: http://localhost:8000/docs"

# 启动服务
python main.py