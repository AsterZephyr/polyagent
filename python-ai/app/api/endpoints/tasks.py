"""
任务执行端点
处理来自Go服务的任务执行请求
"""

from datetime import datetime
from fastapi import APIRouter, HTTPException, Depends
from fastapi.responses import StreamingResponse

from app.api.models import TaskRequest, TaskResponse
from app.services.ai_service import AIService
from app.services.tool_service import ToolService
from app.core.logging import get_logger

router = APIRouter()
logger = get_logger("tasks")

async def get_ai_service() -> AIService:
    """获取AI服务实例"""
    from main import ai_service
    if not ai_service:
        raise HTTPException(status_code=503, detail="AI service not available")
    return ai_service

async def get_tool_service() -> ToolService:
    """获取工具服务实例"""
    from main import tool_service
    if not tool_service:
        raise HTTPException(status_code=503, detail="Tool service not available")
    return tool_service

@router.post("/execute", response_model=TaskResponse)
async def execute_task(
    request: TaskRequest,
    ai_service: AIService = Depends(get_ai_service),
    tool_service: ToolService = Depends(get_tool_service)
):
    """执行智能体任务"""
    
    logger.info(f"Executing task {request.task_id} for user {request.user_id}")
    
    try:
        # 构建消息历史
        messages = []
        
        # 添加系统消息
        system_prompt = _build_system_prompt(request.agent_type, request.tools or [])
        if system_prompt:
            messages.append({
                "role": "system",
                "content": system_prompt
            })
        
        # 添加历史消息
        if request.memory and "messages" in request.memory:
            messages.extend(request.memory["messages"])
        
        # 添加当前用户输入
        messages.append({
            "role": "user",
            "content": request.input
        })
        
        # 获取可用工具
        available_tools = []
        if request.tools:
            available_tools = await tool_service.get_tools_by_names(request.tools)
        
        # 选择模型（可以根据agent_type或用户偏好选择）
        model = _select_model_for_agent(request.agent_type)
        
        # 执行AI推理
        response = await ai_service.chat(
            model=model,
            messages=messages,
            tools=available_tools,
            temperature=0.7,
            max_tokens=2000
        )
        
        # 处理工具调用
        final_content = response.content
        tool_calls = []
        
        if response.tool_calls:
            logger.info(f"Processing {len(response.tool_calls)} tool calls")
            
            for tool_call in response.tool_calls:
                try:
                    # 执行工具
                    tool_result = await tool_service.execute_tool(
                        tool_call.name,
                        tool_call.parameters
                    )
                    
                    tool_calls.append({
                        "id": tool_call.id,
                        "name": tool_call.name,
                        "parameters": tool_call.parameters,
                        "result": tool_result.result if tool_result.success else None,
                        "error": tool_result.error
                    })
                    
                    # 如果工具调用成功，可能需要继续对话
                    if tool_result.success:
                        # 添加工具结果到消息历史，继续推理
                        messages.append({
                            "role": "assistant",
                            "content": response.content or "",
                            "tool_calls": [{"id": tool_call.id, "name": tool_call.name, "parameters": tool_call.parameters}]
                        })
                        messages.append({
                            "role": "tool",
                            "content": str(tool_result.result),
                            "tool_call_id": tool_call.id
                        })
                        
                        # 再次调用AI获取最终回复
                        final_response = await ai_service.chat(
                            model=model,
                            messages=messages,
                            temperature=0.7,
                            max_tokens=2000
                        )
                        final_content = final_response.content
                
                except Exception as e:
                    logger.error(f"Tool execution failed: {str(e)}")
                    tool_calls.append({
                        "id": tool_call.id,
                        "name": tool_call.name,
                        "parameters": tool_call.parameters,
                        "result": None,
                        "error": str(e)
                    })
        
        # 更新对话记忆
        updated_memory = _update_memory(request.memory, messages, final_content)
        
        return TaskResponse(
            task_id=request.task_id,
            status="success",
            output=final_content,
            tool_calls=tool_calls if tool_calls else None,
            memory=updated_memory,
            metadata={
                "model": model,
                "usage": response.usage,
                "finish_reason": response.finish_reason
            },
            timestamp=datetime.utcnow().isoformat()
        )
        
    except Exception as e:
        logger.error(f"Task execution failed: {str(e)}")
        return TaskResponse(
            task_id=request.task_id,
            status="error",
            output="",
            error=str(e),
            timestamp=datetime.utcnow().isoformat()
        )

def _build_system_prompt(agent_type: str, tools: list) -> str:
    """构建系统提示词"""
    
    base_prompts = {
        "general": "You are a helpful AI assistant. Answer questions clearly and concisely.",
        "code": "You are an expert code assistant. Help with programming tasks, debugging, and code review.",
        "rag": "You are a knowledgeable assistant with access to a knowledge base. Use the provided context to answer questions accurately.",
        "chat": "You are a friendly conversational AI. Engage in natural, helpful dialogue."
    }
    
    system_prompt = base_prompts.get(agent_type, base_prompts["general"])
    
    if tools:
        tool_names = ", ".join(tools)
        system_prompt += f"\n\nYou have access to the following tools: {tool_names}. Use them when appropriate to help the user."
    
    return system_prompt

def _select_model_for_agent(agent_type: str) -> str:
    """根据智能体类型选择合适的模型"""
    
    model_mapping = {
        "general": "gpt-3.5-turbo",
        "code": "gpt-4",
        "rag": "gpt-3.5-turbo",
        "chat": "gpt-3.5-turbo"
    }
    
    return model_mapping.get(agent_type, "gpt-3.5-turbo")

def _update_memory(current_memory: dict, messages: list, final_output: str) -> dict:
    """更新对话记忆"""
    
    if not current_memory:
        current_memory = {"messages": []}
    
    # 添加最新的对话轮次
    current_memory["messages"].extend([
        messages[-1],  # 用户输入
        {"role": "assistant", "content": final_output}  # AI回复
    ])
    
    # 保持消息历史在合理长度内（比如最近20条消息）
    if len(current_memory["messages"]) > 20:
        current_memory["messages"] = current_memory["messages"][-20:]
    
    # 更新摘要（如果消息过多）
    if len(current_memory["messages"]) > 10:
        current_memory["summary"] = _generate_conversation_summary(current_memory["messages"])
    
    return current_memory

def _generate_conversation_summary(messages: list) -> str:
    """生成对话摘要（简化实现）"""
    # 这里可以调用AI模型生成更智能的摘要
    return f"Conversation with {len(messages)} messages"