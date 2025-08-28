"""
模型管理端点
"""

from fastapi import APIRouter, HTTPException, Depends
from app.api.models import ModelListResponse, HealthResponse
from app.services.ai_service import AIService
from app.core.logging import get_logger

router = APIRouter()
logger = get_logger("models")

async def get_ai_service() -> AIService:
    """获取AI服务实例"""
    from main import ai_service
    if not ai_service:
        raise HTTPException(status_code=503, detail="AI service not available")
    return ai_service

@router.get("/list", response_model=ModelListResponse)
async def list_models(ai_service: AIService = Depends(get_ai_service)):
    """获取可用模型列表"""
    
    try:
        models = ai_service.get_available_models()
        
        return ModelListResponse(
            models=models,
            total=len(models)
        )
        
    except Exception as e:
        logger.error(f"Failed to list models: {str(e)}")
        raise HTTPException(status_code=500, detail=str(e))

@router.get("/health", response_model=HealthResponse) 
async def health_check(ai_service: AIService = Depends(get_ai_service)):
    """AI服务健康检查"""
    
    try:
        health_status = await ai_service.health_check()
        
        return HealthResponse(**health_status)
        
    except Exception as e:
        logger.error(f"Health check failed: {str(e)}")
        raise HTTPException(status_code=500, detail=str(e))

@router.post("/validate/{model_name}")
async def validate_model(
    model_name: str,
    ai_service: AIService = Depends(get_ai_service)
):
    """验证特定模型是否可用"""
    
    try:
        is_valid = await ai_service.validate_model(model_name)
        
        return {
            "model": model_name,
            "valid": is_valid,
            "timestamp": "2024-01-01T00:00:00Z"  # 应该使用实际时间戳
        }
        
    except Exception as e:
        logger.error(f"Model validation failed: {str(e)}")
        raise HTTPException(status_code=500, detail=str(e))