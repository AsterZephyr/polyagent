"""
API路由配置
"""

from fastapi import APIRouter
from app.api.endpoints import tasks, chat, rag, tools, models

# 创建主路由器
api_router = APIRouter()

# 包含各个端点路由
api_router.include_router(tasks.router, prefix="/tasks", tags=["tasks"])
api_router.include_router(chat.router, prefix="/chat", tags=["chat"]) 
api_router.include_router(rag.router, prefix="/rag", tags=["rag"])
api_router.include_router(tools.router, prefix="/tools", tags=["tools"])
api_router.include_router(models.router, prefix="/models", tags=["models"])