"""
AI服务核心模块
负责管理多个AI模型适配器和统一的对话接口
"""

import asyncio
from typing import Dict, List, Any, Optional, AsyncGenerator
from app.core.config import Settings
from app.core.logging import LoggerMixin
from app.core.exceptions import ModelNotAvailableException, APIKeyMissingException

from app.adapters.base import BaseAIAdapter, ChatMessage, ChatResponse, StreamChunk
from app.adapters.openai_adapter import OpenAIAdapter
from app.adapters.claude_adapter import ClaudeAdapter

class AIService(LoggerMixin):
    """AI服务管理器"""
    
    def __init__(self, settings: Settings):
        self.settings = settings
        self.adapters: Dict[str, BaseAIAdapter] = {}
        self.model_to_adapter: Dict[str, str] = {}
        
    async def startup(self):
        """启动AI服务"""
        self.logger.info("Initializing AI Service...")
        
        # 初始化OpenAI适配器
        if self.settings.OPENAI_API_KEY:
            try:
                openai_adapter = OpenAIAdapter(
                    api_key=self.settings.OPENAI_API_KEY,
                    model=self.settings.OPENAI_MODEL,
                    base_url=self.settings.OPENAI_BASE_URL
                )
                
                self.adapters["openai"] = openai_adapter
                
                # 映射支持的模型
                for model in openai_adapter.supported_models:
                    self.model_to_adapter[model] = "openai"
                
                self.logger.info(f"OpenAI adapter initialized with models: {openai_adapter.supported_models}")
                
            except Exception as e:
                self.logger.error(f"Failed to initialize OpenAI adapter: {str(e)}")
        
        # 初始化Claude适配器
        if self.settings.ANTHROPIC_API_KEY:
            try:
                claude_adapter = ClaudeAdapter(
                    api_key=self.settings.ANTHROPIC_API_KEY,
                    model=self.settings.ANTHROPIC_MODEL
                )
                
                self.adapters["anthropic"] = claude_adapter
                
                # 映射支持的模型
                for model in claude_adapter.supported_models:
                    self.model_to_adapter[model] = "anthropic"
                
                self.logger.info(f"Claude adapter initialized with models: {claude_adapter.supported_models}")
                
            except Exception as e:
                self.logger.error(f"Failed to initialize Claude adapter: {str(e)}")
        
        if not self.adapters:
            self.logger.warning("No AI adapters initialized. Please check your API keys.")
        else:
            self.logger.info(f"AI Service initialized with {len(self.adapters)} adapters")
    
    async def shutdown(self):
        """关闭AI服务"""
        self.logger.info("Shutting down AI Service...")
        self.adapters.clear()
        self.model_to_adapter.clear()
    
    def get_available_models(self) -> List[str]:
        """获取可用模型列表"""
        return list(self.model_to_adapter.keys())
    
    def get_adapter_for_model(self, model: str) -> BaseAIAdapter:
        """根据模型名获取适配器"""
        if model not in self.model_to_adapter:
            raise ModelNotAvailableException(model)
        
        adapter_name = self.model_to_adapter[model]
        adapter = self.adapters.get(adapter_name)
        
        if not adapter:
            raise ModelNotAvailableException(model)
        
        # 更新适配器的当前模型
        adapter.model = model
        return adapter
    
    async def chat(
        self,
        model: str,
        messages: List[Dict[str, Any]],
        temperature: float = 0.7,
        max_tokens: int = 1000,
        tools: Optional[List[Dict[str, Any]]] = None,
        **kwargs
    ) -> ChatResponse:
        """统一聊天接口"""
        
        # 转换消息格式
        chat_messages = [ChatMessage(**msg) for msg in messages]
        
        # 获取适配器
        adapter = self.get_adapter_for_model(model)
        
        self.logger.info(f"Processing chat request with {model}")
        
        # 调用适配器
        response = await adapter.chat(
            messages=chat_messages,
            temperature=temperature,
            max_tokens=max_tokens,
            tools=tools,
            **kwargs
        )
        
        # 记录使用情况
        if response.usage:
            cost = adapter.calculate_cost(response.usage)
            self.logger.info(f"API call completed. Usage: {response.usage}, Cost: ${cost:.4f}")
        
        return response
    
    async def stream_chat(
        self,
        model: str,
        messages: List[Dict[str, Any]],
        temperature: float = 0.7,
        max_tokens: int = 1000,
        tools: Optional[List[Dict[str, Any]]] = None,
        **kwargs
    ) -> AsyncGenerator[StreamChunk, None]:
        """统一流式聊天接口"""
        
        # 转换消息格式
        chat_messages = [ChatMessage(**msg) for msg in messages]
        
        # 获取适配器
        adapter = self.get_adapter_for_model(model)
        
        self.logger.info(f"Processing streaming chat request with {model}")
        
        # 调用适配器
        async for chunk in adapter.stream_chat(
            messages=chat_messages,
            temperature=temperature,
            max_tokens=max_tokens,
            tools=tools,
            **kwargs
        ):
            yield chunk
    
    async def validate_model(self, model: str) -> bool:
        """验证模型是否可用"""
        try:
            adapter = self.get_adapter_for_model(model)
            return await adapter.validate_model()
        except Exception:
            return False
    
    async def health_check(self) -> Dict[str, Any]:
        """健康检查"""
        status = {
            "adapters": {},
            "models": self.get_available_models(),
            "total_adapters": len(self.adapters)
        }
        
        # 检查每个适配器的健康状态
        for name, adapter in self.adapters.items():
            try:
                is_healthy = await adapter.validate_model()
                status["adapters"][name] = {
                    "status": "healthy" if is_healthy else "unhealthy",
                    "provider": adapter.provider_name,
                    "models": adapter.supported_models
                }
            except Exception as e:
                status["adapters"][name] = {
                    "status": "error",
                    "error": str(e)
                }
        
        return status