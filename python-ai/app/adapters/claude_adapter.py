"""
Anthropic Claude模型适配器
"""

import json
from typing import Dict, List, Any, Optional, AsyncGenerator
import anthropic
from anthropic import AsyncAnthropic

from app.adapters.base import BaseAIAdapter, ChatMessage, ChatResponse, ToolCall, StreamChunk
from app.core.exceptions import APIKeyMissingException, ModelNotAvailableException

class ClaudeAdapter(BaseAIAdapter):
    """Claude模型适配器"""
    
    def __init__(self, api_key: str, model: str, **kwargs):
        if not api_key:
            raise APIKeyMissingException("anthropic")
            
        super().__init__(api_key, model, **kwargs)
        
        self.client = AsyncAnthropic(api_key=api_key)
        
        # 模型定价（每1000 tokens）
        self.pricing = {
            "claude-3-sonnet-20240229": {"input": 0.003, "output": 0.015},
            "claude-3-haiku-20240307": {"input": 0.00025, "output": 0.00125},
            "claude-3-opus-20240229": {"input": 0.015, "output": 0.075}
        }
    
    @property
    def provider_name(self) -> str:
        return "anthropic"
    
    @property
    def supported_models(self) -> List[str]:
        return [
            "claude-3-sonnet-20240229",
            "claude-3-haiku-20240307",
            "claude-3-opus-20240229"
        ]
    
    async def chat(
        self,
        messages: List[ChatMessage],
        temperature: float = 0.7,
        max_tokens: int = 1000,
        tools: Optional[List[Dict[str, Any]]] = None,
        **kwargs
    ) -> ChatResponse:
        """Claude聊天补全"""
        
        if self.model not in self.supported_models:
            raise ModelNotAvailableException(self.model)
        
        # 转换消息格式
        claude_messages = self._convert_messages(messages)
        system_message = self._extract_system_message(messages)
        
        # 构建请求参数
        request_params = {
            "model": self.model,
            "messages": claude_messages,
            "temperature": temperature,
            "max_tokens": max_tokens,
            **kwargs
        }
        
        if system_message:
            request_params["system"] = system_message
        
        # 添加工具定义
        if tools:
            request_params["tools"] = self.format_tools_for_provider(tools)
        
        try:
            self.logger.info(f"Calling Claude {self.model} with {len(messages)} messages")
            
            response = await self.client.messages.create(**request_params)
            
            # 提取响应内容
            content = ""
            tool_calls = []
            
            for content_block in response.content:
                if content_block.type == "text":
                    content += content_block.text
                elif content_block.type == "tool_use":
                    tool_calls.append(ToolCall(
                        id=content_block.id,
                        name=content_block.name,
                        parameters=content_block.input
                    ))
            
            return ChatResponse(
                content=content,
                tool_calls=tool_calls if tool_calls else None,
                usage=response.usage.dict() if response.usage else None,
                model=response.model,
                finish_reason=response.stop_reason
            )
            
        except anthropic.APIError as e:
            self.logger.error(f"Claude API error: {str(e)}")
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
        """Claude流式聊天补全"""
        
        if self.model not in self.supported_models:
            raise ModelNotAvailableException(self.model)
        
        # 转换消息格式
        claude_messages = self._convert_messages(messages)
        system_message = self._extract_system_message(messages)
        
        # 构建请求参数
        request_params = {
            "model": self.model,
            "messages": claude_messages,
            "temperature": temperature,
            "max_tokens": max_tokens,
            "stream": True,
            **kwargs
        }
        
        if system_message:
            request_params["system"] = system_message
        
        # 添加工具定义
        if tools:
            request_params["tools"] = self.format_tools_for_provider(tools)
        
        try:
            self.logger.info(f"Starting streaming chat with Claude {self.model}")
            
            stream = await self.client.messages.create(**request_params)
            
            current_tool_call = None
            
            async for event in stream:
                if event.type == "message_start":
                    continue
                elif event.type == "content_block_start":
                    if event.content_block.type == "tool_use":
                        current_tool_call = {
                            "id": event.content_block.id,
                            "name": event.content_block.name,
                            "parameters": {}
                        }
                elif event.type == "content_block_delta":
                    if event.delta.type == "text_delta":
                        yield StreamChunk(
                            content=event.delta.text,
                            finish_reason=None,
                            is_final=False
                        )
                    elif event.delta.type == "input_json_delta" and current_tool_call:
                        # 累积工具参数
                        pass
                elif event.type == "content_block_stop":
                    if current_tool_call:
                        yield StreamChunk(
                            content="",
                            tool_calls=[ToolCall(
                                id=current_tool_call["id"],
                                name=current_tool_call["name"],
                                parameters=current_tool_call["parameters"]
                            )],
                            finish_reason=None,
                            is_final=False
                        )
                        current_tool_call = None
                elif event.type == "message_stop":
                    yield StreamChunk(
                        content="",
                        finish_reason="stop",
                        is_final=True
                    )
                    break
                
        except anthropic.APIError as e:
            self.logger.error(f"Claude streaming error: {str(e)}")
            raise
        except Exception as e:
            self.logger.error(f"Unexpected streaming error: {str(e)}")
            raise
    
    def _convert_messages(self, messages: List[ChatMessage]) -> List[Dict[str, Any]]:
        """转换消息格式为Claude格式"""
        claude_messages = []
        
        for msg in messages:
            if msg.role == "system":
                continue  # system消息单独处理
            
            # 基础消息结构
            claude_msg = {
                "role": "user" if msg.role == "user" else "assistant",
                "content": []
            }
            
            # 添加文本内容
            if msg.content:
                claude_msg["content"].append({
                    "type": "text",
                    "text": msg.content
                })
            
            # 添加工具调用（assistant消息）
            if msg.tool_calls and msg.role == "assistant":
                for tc in msg.tool_calls:
                    claude_msg["content"].append({
                        "type": "tool_use",
                        "id": tc.get("id"),
                        "name": tc.get("name"),
                        "input": tc.get("parameters", {})
                    })
            
            # 添加工具响应（tool消息转为user消息）
            if msg.role == "tool":
                claude_msg = {
                    "role": "user",
                    "content": [{
                        "type": "tool_result",
                        "tool_use_id": msg.tool_call_id,
                        "content": msg.content
                    }]
                }
            
            claude_messages.append(claude_msg)
        
        return claude_messages
    
    def _extract_system_message(self, messages: List[ChatMessage]) -> Optional[str]:
        """提取系统消息"""
        for msg in messages:
            if msg.role == "system":
                return msg.content
        return None
    
    def format_tools_for_provider(self, tools: List[Dict[str, Any]]) -> List[Dict[str, Any]]:
        """为Claude格式化工具定义"""
        claude_tools = []
        
        for tool in tools:
            claude_tool = {
                "name": tool["name"],
                "description": tool["description"],
                "input_schema": tool.get("parameters", {})
            }
            claude_tools.append(claude_tool)
        
        return claude_tools
    
    def calculate_cost(self, usage: Dict[str, int]) -> float:
        """计算Claude API调用成本"""
        if self.model not in self.pricing:
            return 0.0
        
        pricing = self.pricing[self.model]
        input_cost = (usage.get("input_tokens", 0) / 1000) * pricing["input"]
        output_cost = (usage.get("output_tokens", 0) / 1000) * pricing["output"]
        
        return input_cost + output_cost