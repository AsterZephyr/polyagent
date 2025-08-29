"""
Unified AI Adapter
"""

import asyncio
import json
import time
from typing import Dict, List, Optional, Any, AsyncIterator, Union
from dataclasses import dataclass, asdict
from abc import ABC, abstractmethod
import logging

import httpx
import openai
from anthropic import Anthropic, AsyncAnthropic
from openai import AsyncOpenAI

from .models import (
    ModelProvider, ModelConfig, ModelCapability, AVAILABLE_MODELS,
    ModelSelector, get_default_model_config
)
from ..core.logging import LoggerMixin
from ..core.exceptions import AgentException

logger = logging.getLogger(__name__)

@dataclass
class GenerationRequest:
    """Generation request data structure"""
    messages: List[Dict[str, str]]
    model: str
    temperature: float = 0.7
    max_tokens: int = 2000
    stream: bool = False
    tools: Optional[List[Dict[str, Any]]] = None
    tool_choice: Optional[str] = None
    system_message: Optional[str] = None
    stop_sequences: Optional[List[str]] = None
    top_p: float = 0.9
    frequency_penalty: float = 0.0
    presence_penalty: float = 0.0

@dataclass
class GenerationResponse:
    """Generation response data structure"""
    content: str
    model: str
    provider: str
    usage: Dict[str, int]
    finish_reason: str
    tool_calls: Optional[List[Dict[str, Any]]] = None
    response_time: float = 0.0
    cost_estimate: float = 0.0

class BaseAdapter(ABC, LoggerMixin):
    """Base adapter abstract class"""
    
    def __init__(self, model_config: ModelConfig, api_key: str):
        super().__init__()
        self.model_config = model_config
        self.api_key = api_key
        self._client = None
        self.request_count = 0
        self.total_tokens = 0
        
    @abstractmethod
    async def generate(self, request: GenerationRequest) -> GenerationResponse:
        """Generate text response"""
        pass
    
    @abstractmethod
    async def stream_generate(self, request: GenerationRequest) -> AsyncIterator[str]:
        """Stream generate text response"""
        pass
    
    async def health_check(self) -> bool:
        """Perform health check"""
        try:
            test_request = GenerationRequest(
                messages=[{"role": "user", "content": "Hello"}],
                model=self.model_config.model_id,
                max_tokens=10
            )
            await self.generate(test_request)
            return True
        except Exception as e:
            self.logger.error(f"Health check failed for {self.model_config.display_name}: {e}")
            return False
    
    def get_stats(self) -> Dict[str, Any]:
        """Get adapter statistics"""
        return {
            "model": self.model_config.display_name,
            "provider": self.model_config.provider.value,
            "request_count": self.request_count,
            "total_tokens": self.total_tokens,
        }

class OpenAIAdapter(BaseAdapter):
    """OpenAI API adapter"""
    
    def __init__(self, model_config: ModelConfig, api_key: str, proxy_url: str = None):
        super().__init__(model_config, api_key)
        client_params = {"api_key": api_key}
        if proxy_url:
            client_params["base_url"] = proxy_url
        self._client = AsyncOpenAI(**client_params)
    
    async def generate(self, request: GenerationRequest) -> GenerationResponse:
        start_time = time.time()
        
        try:
            # Build messages
            messages = request.messages.copy()
            if request.system_message:
                messages.insert(0, {"role": "system", "content": request.system_message})
            
            # Build request parameters
            params = {
                "model": request.model,
                "messages": messages,
                "temperature": request.temperature,
                "max_tokens": request.max_tokens,
                "top_p": request.top_p,
                "frequency_penalty": request.frequency_penalty,
                "presence_penalty": request.presence_penalty,
            }
            
            if request.tools and self.model_config.supports_tools:
                params["tools"] = request.tools
                if request.tool_choice:
                    params["tool_choice"] = request.tool_choice
            
            if request.stop_sequences:
                params["stop"] = request.stop_sequences
            
            response = await self._client.chat.completions.create(**params)
            
            # Process response
            choice = response.choices[0]
            content = choice.message.content or ""
            
            # Process tool calls
            tool_calls = None
            if choice.message.tool_calls:
                tool_calls = [
                    {
                        "id": tc.id,
                        "type": tc.type,
                        "function": {
                            "name": tc.function.name,
                            "arguments": tc.function.arguments
                        }
                    }
                    for tc in choice.message.tool_calls
                ]
            
            # Calculate cost
            usage = response.usage.model_dump()
            cost = ModelSelector.estimate_cost(
                request.model,
                usage.get("prompt_tokens", 0),
                usage.get("completion_tokens", 0)
            )
            
            # Update statistics
            self.request_count += 1
            self.total_tokens += usage.get("total_tokens", 0)
            
            return GenerationResponse(
                content=content,
                model=request.model,
                provider=self.model_config.provider.value,
                usage=usage,
                finish_reason=choice.finish_reason,
                tool_calls=tool_calls,
                response_time=time.time() - start_time,
                cost_estimate=cost
            )
            
        except Exception as e:
            self.logger.error(f"OpenAI generation failed: {e}")
            raise AgentException(f"OpenAI API error: {str(e)}")
    
    async def stream_generate(self, request: GenerationRequest) -> AsyncIterator[str]:
        try:
            messages = request.messages.copy()
            if request.system_message:
                messages.insert(0, {"role": "system", "content": request.system_message})
            
            params = {
                "model": request.model,
                "messages": messages,
                "temperature": request.temperature,
                "max_tokens": request.max_tokens,
                "stream": True,
            }
            
            async for chunk in await self._client.chat.completions.create(**params):
                if chunk.choices and chunk.choices[0].delta.content:
                    yield chunk.choices[0].delta.content
                    
        except Exception as e:
            self.logger.error(f"OpenAI streaming failed: {e}")
            raise AgentException(f"OpenAI streaming error: {str(e)}")

class AnthropicAdapter(BaseAdapter):
    """Anthropic Claude API adapter"""
    
    def __init__(self, model_config: ModelConfig, api_key: str):
        super().__init__(model_config, api_key)
        self._client = AsyncAnthropic(api_key=api_key)
    
    async def generate(self, request: GenerationRequest) -> GenerationResponse:
        start_time = time.time()
        
        try:
            # Process system message
            system_msg = request.system_message or ""
            messages = []
            
            for msg in request.messages:
                if msg["role"] == "system":
                    system_msg += "\n" + msg["content"]
                else:
                    messages.append(msg)
            
            params = {
                "model": request.model,
                "messages": messages,
                "max_tokens": request.max_tokens,
                "temperature": request.temperature,
                "top_p": request.top_p,
            }
            
            if system_msg:
                params["system"] = system_msg
            
            if request.stop_sequences:
                params["stop_sequences"] = request.stop_sequences
            
            # Claude tool calling support
            if request.tools and self.model_config.supports_tools:
                params["tools"] = request.tools
            
            response = await self._client.messages.create(**params)
            
            # Process response
            content = ""
            tool_calls = None
            
            for content_block in response.content:
                if content_block.type == "text":
                    content += content_block.text
                elif content_block.type == "tool_use":
                    if tool_calls is None:
                        tool_calls = []
                    tool_calls.append({
                        "id": content_block.id,
                        "type": "function",
                        "function": {
                            "name": content_block.name,
                            "arguments": json.dumps(content_block.input)
                        }
                    })
            
            # Usage statistics
            usage = {
                "prompt_tokens": response.usage.input_tokens,
                "completion_tokens": response.usage.output_tokens,
                "total_tokens": response.usage.input_tokens + response.usage.output_tokens
            }
            
            cost = ModelSelector.estimate_cost(
                request.model,
                usage["prompt_tokens"],
                usage["completion_tokens"]
            )
            
            self.request_count += 1
            self.total_tokens += usage["total_tokens"]
            
            return GenerationResponse(
                content=content,
                model=request.model,
                provider=self.model_config.provider.value,
                usage=usage,
                finish_reason=response.stop_reason,
                tool_calls=tool_calls,
                response_time=time.time() - start_time,
                cost_estimate=cost
            )
            
        except Exception as e:
            self.logger.error(f"Anthropic generation failed: {e}")
            raise AgentException(f"Anthropic API error: {str(e)}")
    
    async def stream_generate(self, request: GenerationRequest) -> AsyncIterator[str]:
        try:
            system_msg = request.system_message or ""
            messages = []
            
            for msg in request.messages:
                if msg["role"] == "system":
                    system_msg += "\n" + msg["content"]
                else:
                    messages.append(msg)
            
            params = {
                "model": request.model,
                "messages": messages,
                "max_tokens": request.max_tokens,
                "temperature": request.temperature,
                "stream": True,
            }
            
            if system_msg:
                params["system"] = system_msg
            
            async with self._client.messages.stream(**params) as stream:
                async for chunk in stream:
                    if chunk.type == "content_block_delta" and hasattr(chunk.delta, 'text'):
                        yield chunk.delta.text
                        
        except Exception as e:
            self.logger.error(f"Anthropic streaming failed: {e}")
            raise AgentException(f"Anthropic streaming error: {str(e)}")

class OpenRouterAdapter(BaseAdapter):
    """OpenRouter API adapter"""
    
    def __init__(self, model_config: ModelConfig, api_key: str):
        super().__init__(model_config, api_key)
        self._client = AsyncOpenAI(
            api_key=api_key,
            base_url=model_config.base_url or "https://openrouter.ai/api/v1"
        )
    
    async def generate(self, request: GenerationRequest) -> GenerationResponse:
        start_time = time.time()
        
        try:
            messages = request.messages.copy()
            if request.system_message:
                messages.insert(0, {"role": "system", "content": request.system_message})
            
            params = {
                "model": request.model,
                "messages": messages,
                "temperature": request.temperature,
                "max_tokens": request.max_tokens,
                "top_p": request.top_p,
            }
            
            response = await self._client.chat.completions.create(**params)
            
            choice = response.choices[0]
            content = choice.message.content or ""
            
            # OpenRouter may not return detailed usage info
            usage = getattr(response, 'usage', None)
            if usage:
                usage = usage.model_dump()
            else:
                # Estimate token usage
                estimated_input = sum(len(m["content"]) // 4 for m in messages)
                estimated_output = len(content) // 4
                usage = {
                    "prompt_tokens": estimated_input,
                    "completion_tokens": estimated_output,
                    "total_tokens": estimated_input + estimated_output
                }
            
            cost = ModelSelector.estimate_cost(
                request.model,
                usage.get("prompt_tokens", 0),
                usage.get("completion_tokens", 0)
            )
            
            self.request_count += 1
            self.total_tokens += usage.get("total_tokens", 0)
            
            return GenerationResponse(
                content=content,
                model=request.model,
                provider=self.model_config.provider.value,
                usage=usage,
                finish_reason=choice.finish_reason or "stop",
                response_time=time.time() - start_time,
                cost_estimate=cost
            )
            
        except Exception as e:
            self.logger.error(f"OpenRouter generation failed: {e}")
            raise AgentException(f"OpenRouter API error: {str(e)}")
    
    async def stream_generate(self, request: GenerationRequest) -> AsyncIterator[str]:
        try:
            messages = request.messages.copy()
            if request.system_message:
                messages.insert(0, {"role": "system", "content": request.system_message})
            
            params = {
                "model": request.model,
                "messages": messages,
                "temperature": request.temperature,
                "max_tokens": request.max_tokens,
                "stream": True,
            }
            
            async for chunk in await self._client.chat.completions.create(**params):
                if chunk.choices and chunk.choices[0].delta.content:
                    yield chunk.choices[0].delta.content
                    
        except Exception as e:
            self.logger.error(f"OpenRouter streaming failed: {e}")
            raise AgentException(f"OpenRouter streaming error: {str(e)}")

class GLMAdapter(BaseAdapter):
    """Zhipu AI GLM adapter"""
    
    def __init__(self, model_config: ModelConfig, api_key: str):
        super().__init__(model_config, api_key)
        self.base_url = model_config.base_url or "https://open.bigmodel.cn/api/paas/v4"
        
    async def generate(self, request: GenerationRequest) -> GenerationResponse:
        start_time = time.time()
        
        try:
            # GLM API format
            messages = request.messages.copy()
            if request.system_message:
                messages.insert(0, {"role": "system", "content": request.system_message})
            
            headers = {
                "Authorization": f"Bearer {self.api_key}",
                "Content-Type": "application/json"
            }
            
            data = {
                "model": request.model,
                "messages": messages,
                "temperature": request.temperature,
                "max_tokens": request.max_tokens,
                "top_p": request.top_p,
            }
            
            if request.tools and self.model_config.supports_tools:
                data["tools"] = request.tools
            
            async with httpx.AsyncClient() as client:
                response = await client.post(
                    f"{self.base_url}/chat/completions",
                    json=data,
                    headers=headers,
                    timeout=60.0
                )
                response.raise_for_status()
                result = response.json()
            
            if "error" in result:
                raise Exception(result["error"]["message"])
            
            choice = result["choices"][0]
            content = choice["message"]["content"]
            
            # Process tool calls
            tool_calls = None
            if "tool_calls" in choice["message"]:
                tool_calls = choice["message"]["tool_calls"]
            
            usage = result.get("usage", {
                "prompt_tokens": 0,
                "completion_tokens": 0,
                "total_tokens": 0
            })
            
            cost = ModelSelector.estimate_cost(
                request.model,
                usage.get("prompt_tokens", 0),
                usage.get("completion_tokens", 0)
            )
            
            self.request_count += 1
            self.total_tokens += usage.get("total_tokens", 0)
            
            return GenerationResponse(
                content=content,
                model=request.model,
                provider=self.model_config.provider.value,
                usage=usage,
                finish_reason=choice.get("finish_reason", "stop"),
                tool_calls=tool_calls,
                response_time=time.time() - start_time,
                cost_estimate=cost
            )
            
        except Exception as e:
            self.logger.error(f"GLM generation failed: {e}")
            raise AgentException(f"GLM API error: {str(e)}")
    
    async def stream_generate(self, request: GenerationRequest) -> AsyncIterator[str]:
        try:
            messages = request.messages.copy()
            if request.system_message:
                messages.insert(0, {"role": "system", "content": request.system_message})
            
            headers = {
                "Authorization": f"Bearer {self.api_key}",
                "Content-Type": "application/json"
            }
            
            data = {
                "model": request.model,
                "messages": messages,
                "temperature": request.temperature,
                "max_tokens": request.max_tokens,
                "stream": True,
            }
            
            async with httpx.AsyncClient() as client:
                async with client.stream(
                    "POST",
                    f"{self.base_url}/chat/completions",
                    json=data,
                    headers=headers,
                    timeout=60.0
                ) as response:
                    response.raise_for_status()
                    
                    async for line in response.aiter_lines():
                        if line.startswith("data: "):
                            data_str = line[6:]
                            if data_str.strip() == "[DONE]":
                                break
                            try:
                                chunk = json.loads(data_str)
                                if "choices" in chunk and chunk["choices"]:
                                    delta = chunk["choices"][0].get("delta", {})
                                    if "content" in delta:
                                        yield delta["content"]
                            except json.JSONDecodeError:
                                continue
                                
        except Exception as e:
            self.logger.error(f"GLM streaming failed: {e}")
            raise AgentException(f"GLM streaming error: {str(e)}")

class UnifiedAIAdapter(LoggerMixin):
    """Unified AI adapter manager"""
    
    def __init__(self, api_keys: Dict[str, str] = None, proxy_config: Dict[str, str] = None):
        super().__init__()
        self.api_keys = api_keys or {}
        self.proxy_config = proxy_config or {}
        self.adapters: Dict[str, BaseAdapter] = {}
        self.model_selector = ModelSelector()
        self._initialize_adapters()
    
    def _initialize_adapters(self):
        """Initialize all available adapters"""
        
        for model_id, model_config in AVAILABLE_MODELS.items():
            api_key = self.api_keys.get(model_config.api_key_env)
            if not api_key:
                continue
            
            try:
                proxy_url = self.proxy_config.get(model_config.provider.value)
                
                if model_config.provider == ModelProvider.OPENAI:
                    adapter = OpenAIAdapter(model_config, api_key, proxy_url)
                elif model_config.provider == ModelProvider.ANTHROPIC:
                    adapter = AnthropicAdapter(model_config, api_key)
                elif model_config.provider == ModelProvider.OPENROUTER:
                    adapter = OpenRouterAdapter(model_config, api_key)
                elif model_config.provider == ModelProvider.GLM:
                    adapter = GLMAdapter(model_config, api_key)
                else:
                    continue
                
                self.adapters[model_id] = adapter
                self.logger.info(f"Initialized adapter for {model_config.display_name}")
                
            except Exception as e:
                self.logger.warning(f"Failed to initialize {model_config.display_name}: {e}")
    
    async def generate(
        self,
        messages: List[Dict[str, str]],
        model: str = None,
        **kwargs
    ) -> GenerationResponse:
        """Generate text response"""
        
        # Auto select model
        if not model:
            model = self.model_selector.get_model_for_task("general")
        
        if model not in self.adapters:
            # Try to find alternative model
            available_models = list(self.adapters.keys())
            if not available_models:
                raise AgentException("No AI models available")
            model = available_models[0]
            self.logger.warning(f"Requested model not available, using {model}")
        
        request = GenerationRequest(
            messages=messages,
            model=model,
            **kwargs
        )
        
        return await self.adapters[model].generate(request)
    
    async def stream_generate(
        self,
        messages: List[Dict[str, str]],
        model: str = None,
        **kwargs
    ) -> AsyncIterator[str]:
        """Stream generate text"""
        
        if not model:
            model = self.model_selector.get_model_for_task("general")
        
        if model not in self.adapters:
            available_models = list(self.adapters.keys())
            if not available_models:
                raise AgentException("No AI models available")
            model = available_models[0]
        
        request = GenerationRequest(
            messages=messages,
            model=model,
            stream=True,
            **kwargs
        )
        
        async for chunk in self.adapters[model].stream_generate(request):
            yield chunk
    
    def get_available_models(self) -> List[str]:
        """Get available models list"""
        return list(self.adapters.keys())
    
    async def health_check_all(self) -> Dict[str, bool]:
        """Check all models health status"""
        results = {}
        
        for model_id, adapter in self.adapters.items():
            try:
                results[model_id] = await adapter.health_check()
            except Exception as e:
                self.logger.error(f"Health check failed for {model_id}: {e}")
                results[model_id] = False
        
        return results
    
    def get_model_stats(self) -> Dict[str, Any]:
        """Get all models statistics"""
        return {
            model_id: adapter.get_stats()
            for model_id, adapter in self.adapters.items()
        }
    
    def estimate_total_cost(self) -> float:
        """Estimate total cost"""
        total_cost = 0.0
        
        for adapter in self.adapters.values():
            stats = adapter.get_stats()
            # TODO: Calculate based on actual usage
            
        return total_cost