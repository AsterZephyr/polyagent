"""
聊天端点
提供直接的聊天API接口
"""

import json
from datetime import datetime
from fastapi import APIRouter, HTTPException, Depends
from fastapi.responses import StreamingResponse
from fastapi.responses import JSONResponse

from app.api.models import ChatRequest, ChatResponse
from app.services.ai_service import AIService
from app.core.logging import get_logger

router = APIRouter()
logger = get_logger("chat")

async def get_ai_service() -> AIService:
    """获取AI服务实例"""
    from main import ai_service
    if not ai_service:
        raise HTTPException(status_code=503, detail="AI service not available")
    return ai_service

@router.post("/completions", response_model=ChatResponse)
async def chat_completions(
    request: ChatRequest,
    ai_service: AIService = Depends(get_ai_service)
):
    """聊天补全接口"""
    
    logger.info(f"Processing chat request with {request.model}")
    
    try:
        # 转换消息格式
        messages = [msg.dict() for msg in request.messages]
        
        # 调用AI服务
        response = await ai_service.chat(
            model=request.model,
            messages=messages,
            temperature=request.temperature,
            max_tokens=request.max_tokens,
            tools=request.tools
        )
        
        return response
        
    except Exception as e:
        logger.error(f"Chat completion failed: {str(e)}")
        raise HTTPException(status_code=500, detail=str(e))

@router.post("/stream")
async def stream_chat(
    request: ChatRequest,
    ai_service: AIService = Depends(get_ai_service)
):
    """流式聊天补全接口"""
    
    if not request.stream:
        request.stream = True
    
    logger.info(f"Processing streaming chat request with {request.model}")
    
    try:
        # 转换消息格式
        messages = [msg.dict() for msg in request.messages]
        
        async def generate():
            try:
                async for chunk in ai_service.stream_chat(
                    model=request.model,
                    messages=messages,
                    temperature=request.temperature,
                    max_tokens=request.max_tokens,
                    tools=request.tools
                ):
                    # 转换为SSE格式
                    chunk_data = {
                        "content": chunk.content,
                        "tool_calls": [tc.dict() for tc in chunk.tool_calls] if chunk.tool_calls else None,
                        "finish_reason": chunk.finish_reason,
                        "is_final": chunk.is_final
                    }
                    
                    yield f"data: {json.dumps(chunk_data)}\n\n"
                    
                    if chunk.is_final:
                        break
                
                # 发送结束标记
                yield "data: [DONE]\n\n"
                
            except Exception as e:
                error_data = {
                    "error": str(e),
                    "is_final": True
                }
                yield f"data: {json.dumps(error_data)}\n\n"
        
        return StreamingResponse(
            generate(),
            media_type="text/plain",
            headers={
                "Cache-Control": "no-cache",
                "Connection": "keep-alive",
                "Content-Type": "text/event-stream"
            }
        )
        
    except Exception as e:
        logger.error(f"Streaming chat failed: {str(e)}")
        raise HTTPException(status_code=500, detail=str(e))