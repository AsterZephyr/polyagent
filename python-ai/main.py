"""
PolyAgent Python AI Service
主要负责AI模型集成、推理和工具调用
"""

import asyncio
import uvicorn
from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from contextlib import asynccontextmanager

from app.core.config import get_settings
from app.core.logging import setup_logging
from app.api.routes import api_router
from app.core.exceptions import setup_exception_handlers
from app.services.ai_service import AIService
from app.services.tool_service import ToolService
from app.services.rag_service import RAGService

# 全局服务实例
ai_service = None
tool_service = None  
rag_service = None

@asynccontextmanager
async def lifespan(app: FastAPI):
    """应用生命周期管理"""
    global ai_service, tool_service, rag_service
    
    settings = get_settings()
    logger = setup_logging(settings.LOG_LEVEL)
    
    logger.info("Starting PolyAgent AI Service...")
    
    # 初始化服务
    ai_service = AIService(settings)
    tool_service = ToolService(settings)
    rag_service = RAGService(settings)
    
    # 启动服务
    await ai_service.startup()
    await tool_service.startup() 
    await rag_service.startup()
    
    logger.info("AI Service started successfully")
    
    yield
    
    # 关闭服务
    logger.info("Shutting down AI Service...")
    await ai_service.shutdown()
    await tool_service.shutdown()
    await rag_service.shutdown()

# 创建FastAPI应用
app = FastAPI(
    title="PolyAgent AI Service",
    description="Multi-AI integration service for PolyAgent system",
    version="1.0.0",
    lifespan=lifespan
)

# 添加CORS中间件
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],  # 生产环境应该更严格
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# 设置异常处理器
setup_exception_handlers(app)

# 注册路由
app.include_router(api_router, prefix="/api/v1")

# 健康检查
@app.get("/health")
async def health_check():
    """健康检查"""
    return {
        "status": "healthy",
        "service": "polyagent-ai",
        "version": "1.0.0"
    }

# 根路径
@app.get("/")
async def root():
    """根路径"""
    return {
        "message": "PolyAgent AI Service",
        "version": "1.0.0",
        "docs_url": "/docs"
    }

if __name__ == "__main__":
    settings = get_settings()
    
    uvicorn.run(
        "main:app",
        host=settings.HOST,
        port=settings.PORT,
        reload=settings.DEBUG,
        log_level=settings.LOG_LEVEL.lower()
    )