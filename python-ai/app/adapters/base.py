"""
AI模型适配器基类
"""

from abc import ABC, abstractmethod
from typing import Dict, List, Any, Optional, AsyncGenerator
from pydantic import BaseModel
from app.core.logging import LoggerMixin

class ChatMessage(BaseModel):
    """聊天消息"""
    role: str  # system, user, assistant, tool
    content: str
    tool_calls: Optional[List[Dict[str, Any]]] = None
    tool_call_id: Optional[str] = None

class ToolCall(BaseModel):
    """工具调用"""
    id: str
    name: str
    parameters: Dict[str, Any]

class ChatResponse(BaseModel):
    """聊天响应"""
    content: str
    tool_calls: Optional[List[ToolCall]] = None
    usage: Optional[Dict[str, int]] = None
    model: str
    finish_reason: Optional[str] = None

class StreamChunk(BaseModel):
    """流式响应块"""
    content: str = ""
    tool_calls: Optional[List[ToolCall]] = None
    finish_reason: Optional[str] = None
    is_final: bool = False

class BaseAIAdapter(ABC, LoggerMixin):
    """AI模型适配器基类"""
    
    def __init__(self, api_key: str, model: str, **kwargs):
        self.api_key = api_key
        self.model = model
        self.config = kwargs
        
    @property
    @abstractmethod
    def provider_name(self) -> str:
        """提供商名称"""
        pass
    
    @property
    @abstractmethod
    def supported_models(self) -> List[str]:
        """支持的模型列表"""
        pass
    
    @abstractmethod
    async def chat(
        self,
        messages: List[ChatMessage],
        temperature: float = 0.7,
        max_tokens: int = 1000,
        tools: Optional[List[Dict[str, Any]]] = None,
        **kwargs
    ) -> ChatResponse:
        """聊天补全"""
        pass
    
    @abstractmethod
    async def stream_chat(
        self,
        messages: List[ChatMessage],
        temperature: float = 0.7,
        max_tokens: int = 1000,
        tools: Optional[List[Dict[str, Any]]] = None,
        **kwargs
    ) -> AsyncGenerator[StreamChunk, None]:
        """流式聊天补全"""
        pass
    
    async def validate_model(self) -> bool:
        """验证模型是否可用"""
        try:
            test_messages = [
                ChatMessage(role="user", content="Hello")
            ]
            response = await self.chat(test_messages, max_tokens=10)
            return bool(response.content)
        except Exception as e:
            self.logger.error(f"Model validation failed: {str(e)}")
            return False
    
    def format_tools_for_provider(self, tools: List[Dict[str, Any]]) -> List[Dict[str, Any]]:
        """为特定提供商格式化工具定义"""
        return tools
    
    def extract_tool_calls(self, response_data: Any) -> Optional[List[ToolCall]]:
        """从响应中提取工具调用"""
        return None
    
    def calculate_cost(self, usage: Dict[str, int]) -> float:
        """计算API调用成本"""
        return 0.0