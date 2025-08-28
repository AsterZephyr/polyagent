"""
工具端点
"""

from fastapi import APIRouter, HTTPException, Depends
from app.api.models import ToolListResponse, ToolExecutionRequest, ToolExecutionResult
from app.services.tool_service import ToolService
from app.core.logging import get_logger

router = APIRouter()
logger = get_logger("tools")

async def get_tool_service() -> ToolService:
    """获取工具服务实例"""
    from main import tool_service
    if not tool_service:
        raise HTTPException(status_code=503, detail="Tool service not available")
    return tool_service

@router.get("/list", response_model=ToolListResponse)
async def list_tools(tool_service: ToolService = Depends(get_tool_service)):
    """获取可用工具列表"""
    
    try:
        tools = await tool_service.get_available_tools()
        
        return ToolListResponse(
            tools=tools,
            total=len(tools)
        )
        
    except Exception as e:
        logger.error(f"Failed to list tools: {str(e)}")
        raise HTTPException(status_code=500, detail=str(e))

@router.post("/execute", response_model=ToolExecutionResult)
async def execute_tool(
    request: ToolExecutionRequest,
    tool_service: ToolService = Depends(get_tool_service)
):
    """执行工具"""
    
    try:
        result = await tool_service.execute_tool(
            tool_name=request.name,
            parameters=request.parameters
        )
        
        return result
        
    except Exception as e:
        logger.error(f"Tool execution failed: {str(e)}")
        raise HTTPException(status_code=500, detail=str(e))