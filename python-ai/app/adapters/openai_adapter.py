"""
OpenAI模型适配器
"""

import json
from typing import Dict, List, Any, Optional, AsyncGenerator
import openai
from openai import AsyncOpenAI

from app.adapters.base import BaseAIAdapter, ChatMessage, ChatResponse, ToolCall, StreamChunk
from app.core.exceptions import APIKeyMissingException, ModelNotAvailableException

class OpenAIAdapter(BaseAIAdapter):
    """OpenAI模型适配器"""
    
    def __init__(self, api_key: str, model: str, base_url: Optional[str] = None, **kwargs):
        if not api_key:
            raise APIKeyMissingException("openai")
            
        super().__init__(api_key, model, **kwargs)
        
        self.client = AsyncOpenAI(
            api_key=api_key,
            base_url=base_url or "https://api.openai.com/v1"
        )
        
        # 模型定价（每1000 tokens）
        self.pricing = {
            "gpt-4": {"input": 0.03, "output": 0.06},
            "gpt-4-turbo": {"input": 0.01, "output": 0.03},
            "gpt-3.5-turbo": {"input": 0.0015, "output": 0.002},
            "gpt-3.5-turbo-16k": {"input": 0.003, "output": 0.004}
        }
    
    @property
    def provider_name(self) -> str:
        return "openai"
    
    @property
    def supported_models(self) -> List[str]:
        return [
            "gpt-4",
            "gpt-4-turbo",
            "gpt-4-turbo-preview", 
            "gpt-3.5-turbo",
            "gpt-3.5-turbo-16k"
        ]
    
    async def chat(
        self,
        messages: List[ChatMessage],
        temperature: float = 0.7,
        max_tokens: int = 1000,
        tools: Optional[List[Dict[str, Any]]] = None,
        **kwargs
    ) -> ChatResponse:
        """OpenAI聊天补全"""
        
        if self.model not in self.supported_models:
            raise ModelNotAvailableException(self.model)
        
        # 转换消息格式
        openai_messages = self._convert_messages(messages)
        
        # 构建请求参数
        request_params = {
            "model": self.model,
            "messages": openai_messages,
            "temperature": temperature,
            "max_tokens": max_tokens,
            **kwargs
        }
        
        # 添加工具定义
        if tools:
            request_params["tools"] = self.format_tools_for_provider(tools)
            request_params["tool_choice"] = "auto"
        
        try:
            self.logger.info(f"Calling OpenAI {self.model} with {len(messages)} messages")
            
            response = await self.client.chat.completions.create(**request_params)
            
            # 提取响应内容
            choice = response.choices[0]
            content = choice.message.content or ""
            
            # 提取工具调用
            tool_calls = None
            if choice.message.tool_calls:
                tool_calls = []
                for tc in choice.message.tool_calls:
                    tool_calls.append(ToolCall(
                        id=tc.id,
                        name=tc.function.name,
                        parameters=json.loads(tc.function.arguments)
                    ))
            
            return ChatResponse(
                content=content,
                tool_calls=tool_calls,
                usage=response.usage.dict() if response.usage else None,
                model=response.model,
                finish_reason=choice.finish_reason
            )
            
        except openai.APIError as e:
            self.logger.error(f"OpenAI API error: {str(e)}")
            raise
        except Exception as e:
            self.logger.error(f"Unexpected error: {str(e)}")
            raise
    
    async def stream_chat(
        self,
        messages: List[ChatMessage],
        temperature: float = 0.7,
        max_tokens: int = 1000,
        tools: Optional[List[Dict[str, Any]]] = None,
        **kwargs
    ) -> AsyncGenerator[StreamChunk, None]:
        """OpenAI流式聊天补全"""
        
        if self.model not in self.supported_models:
            raise ModelNotAvailableException(self.model)
        
        # 转换消息格式
        openai_messages = self._convert_messages(messages)
        
        # 构建请求参数
        request_params = {
            "model": self.model,
            "messages": openai_messages,
            "temperature": temperature,
            "max_tokens": max_tokens,
            "stream": True,
            **kwargs
        }
        
        # 添加工具定义
        if tools:
            request_params["tools"] = self.format_tools_for_provider(tools)
            request_params["tool_choice"] = "auto"
        
        try:
            self.logger.info(f"Starting streaming chat with OpenAI {self.model}")
            
            stream = await self.client.chat.completions.create(**request_params)
            
            async for chunk in stream:
                if not chunk.choices:
                    continue
                    
                choice = chunk.choices[0]
                delta = choice.delta
                
                content = delta.content or ""
                tool_calls = None
                
                # 处理工具调用
                if delta.tool_calls:
                    tool_calls = []
                    for tc in delta.tool_calls:
                        if tc.function:
                            tool_calls.append(ToolCall(
                                id=tc.id or "",
                                name=tc.function.name or "",
                                parameters=json.loads(tc.function.arguments or "{}")
                            ))
                
                yield StreamChunk(
                    content=content,
                    tool_calls=tool_calls,
                    finish_reason=choice.finish_reason,
                    is_final=choice.finish_reason is not None
                )
                
        except openai.APIError as e:
            self.logger.error(f"OpenAI streaming error: {str(e)}")
            raise
        except Exception as e:
            self.logger.error(f"Unexpected streaming error: {str(e)}")
            raise
    
    def _convert_messages(self, messages: List[ChatMessage]) -> List[Dict[str, Any]]:
        """转换消息格式为OpenAI格式"""
        openai_messages = []
        
        for msg in messages:
            openai_msg = {
                "role": msg.role,
                "content": msg.content
            }
            
            # 添加工具调用
            if msg.tool_calls:
                openai_msg["tool_calls"] = []
                for tc in msg.tool_calls:
                    openai_msg["tool_calls"].append({
                        "id": tc.get("id"),
                        "type": "function",
                        "function": {
                            "name": tc.get("name"),
                            "arguments": json.dumps(tc.get("parameters", {}))
                        }
                    })
            
            # 添加工具响应
            if msg.tool_call_id:
                openai_msg["tool_call_id"] = msg.tool_call_id
            
            openai_messages.append(openai_msg)
        
        return openai_messages
    
    def format_tools_for_provider(self, tools: List[Dict[str, Any]]) -> List[Dict[str, Any]]:
        """为OpenAI格式化工具定义"""
        openai_tools = []
        
        for tool in tools:
            openai_tool = {
                "type": "function",
                "function": {
                    "name": tool["name"],
                    "description": tool["description"],
                    "parameters": tool.get("parameters", {})
                }
            }
            openai_tools.append(openai_tool)
        
        return openai_tools
    
    def calculate_cost(self, usage: Dict[str, int]) -> float:
        """计算OpenAI API调用成本"""
        if self.model not in self.pricing:
            return 0.0
        
        pricing = self.pricing[self.model]
        input_cost = (usage.get("prompt_tokens", 0) / 1000) * pricing["input"]
        output_cost = (usage.get("completion_tokens", 0) / 1000) * pricing["output"]
        
        return input_cost + output_cost