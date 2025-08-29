"""
AI Core - Simple, fast, reliable AI calls
Following Linux philosophy: Do one thing and do it well
"""

import asyncio
import json
import httpx
from typing import Dict, Any, List, Optional
from dataclasses import dataclass

@dataclass
class AICall:
    """AI request - simple data structure"""
    model: str
    messages: List[Dict[str, str]]
    temperature: float = 0.7
    max_tokens: int = 2000
    stream: bool = False

@dataclass
class AIResponse:
    """AI response - simple data structure"""
    content: str
    usage: Dict[str, int]
    model: str
    cost: float = 0.0

# Core function - does one thing well
async def call_ai(request: AICall, api_key: str, base_url: str = None) -> AIResponse:
    """
    Single function to call any AI model. Period.
    
    Like open() in Linux - one interface for all files
    Like call_ai() in PolyAgent - one interface for all models
    """
    
    if 'claude' in request.model:
        return await _call_claude(request, api_key, base_url)
    elif 'gpt' in request.model:
        return await _call_openai(request, api_key, base_url)
    elif 'qwen' in request.model or 'openrouter' in request.model:
        return await _call_openrouter(request, api_key)
    elif 'glm' in request.model:
        return await _call_glm(request, api_key)
    else:
        raise ValueError(f"Unsupported model: {request.model}")

async def _call_claude(request: AICall, api_key: str, base_url: str = None) -> AIResponse:
    """Claude API call - direct implementation"""
    
    base_url = base_url or "https://api.anthropic.com"
    
    # Separate system message
    system_msg = ""
    messages = []
    for msg in request.messages:
        if msg["role"] == "system":
            system_msg += msg["content"] + "\n"
        else:
            messages.append(msg)
    
    data = {
        "model": request.model,
        "messages": messages,
        "max_tokens": request.max_tokens,
        "temperature": request.temperature,
    }
    
    if system_msg:
        data["system"] = system_msg.strip()
    
    headers = {
        "Authorization": f"Bearer {api_key}",
        "Content-Type": "application/json",
        "anthropic-version": "2023-06-01"
    }
    
    async with httpx.AsyncClient(timeout=30.0) as client:
        response = await client.post(
            f"{base_url}/v1/messages",
            json=data,
            headers=headers
        )
        response.raise_for_status()
        result = response.json()
    
    content = ""
    for block in result.get("content", []):
        if block.get("type") == "text":
            content += block.get("text", "")
    
    usage = result.get("usage", {})
    
    return AIResponse(
        content=content,
        usage={
            "prompt_tokens": usage.get("input_tokens", 0),
            "completion_tokens": usage.get("output_tokens", 0),
            "total_tokens": usage.get("input_tokens", 0) + usage.get("output_tokens", 0)
        },
        model=request.model,
        cost=_estimate_cost(request.model, usage.get("input_tokens", 0), usage.get("output_tokens", 0))
    )

async def _call_openai(request: AICall, api_key: str, base_url: str = None) -> AIResponse:
    """OpenAI API call - direct implementation"""
    
    base_url = base_url or "https://api.openai.com"
    
    data = {
        "model": request.model,
        "messages": request.messages,
        "temperature": request.temperature,
        "max_tokens": request.max_tokens,
    }
    
    headers = {
        "Authorization": f"Bearer {api_key}",
        "Content-Type": "application/json"
    }
    
    async with httpx.AsyncClient(timeout=30.0) as client:
        response = await client.post(
            f"{base_url}/v1/chat/completions",
            json=data,
            headers=headers
        )
        response.raise_for_status()
        result = response.json()
    
    choice = result["choices"][0]
    content = choice["message"]["content"] or ""
    usage = result.get("usage", {})
    
    return AIResponse(
        content=content,
        usage=usage,
        model=request.model,
        cost=_estimate_cost(request.model, usage.get("prompt_tokens", 0), usage.get("completion_tokens", 0))
    )

async def _call_openrouter(request: AICall, api_key: str) -> AIResponse:
    """OpenRouter API call - direct implementation"""
    
    data = {
        "model": request.model,
        "messages": request.messages,
        "temperature": request.temperature,
        "max_tokens": request.max_tokens,
    }
    
    headers = {
        "Authorization": f"Bearer {api_key}",
        "Content-Type": "application/json",
        "HTTP-Referer": "https://polyagent.local",
        "X-Title": "PolyAgent"
    }
    
    async with httpx.AsyncClient(timeout=30.0) as client:
        response = await client.post(
            "https://openrouter.ai/api/v1/chat/completions",
            json=data,
            headers=headers
        )
        response.raise_for_status()
        result = response.json()
    
    choice = result["choices"][0]
    content = choice["message"]["content"] or ""
    usage = result.get("usage", {})
    
    return AIResponse(
        content=content,
        usage=usage,
        model=request.model,
        cost=0.0  # Most OpenRouter free models
    )

async def _call_glm(request: AICall, api_key: str) -> AIResponse:
    """GLM API call - direct implementation"""
    
    data = {
        "model": request.model,
        "messages": request.messages,
        "temperature": request.temperature,
        "max_tokens": request.max_tokens,
    }
    
    headers = {
        "Authorization": f"Bearer {api_key}",
        "Content-Type": "application/json"
    }
    
    async with httpx.AsyncClient(timeout=30.0) as client:
        response = await client.post(
            "https://open.bigmodel.cn/api/paas/v4/chat/completions",
            json=data,
            headers=headers
        )
        response.raise_for_status()
        result = response.json()
    
    choice = result["choices"][0]
    content = choice["message"]["content"] or ""
    usage = result.get("usage", {})
    
    return AIResponse(
        content=content,
        usage=usage,
        model=request.model,
        cost=0.0  # Free tier
    )

def get_best_model(query: str, api_keys: Dict[str, str], free_only: bool = False) -> str:
    """
    Route to best model for query
    Simple heuristics - no complex logic
    """
    
    query_lower = query.lower()
    
    # Free models first if requested
    if free_only or not api_keys.get('OPENAI_API_KEY'):
        if 'code' in query_lower or 'python' in query_lower or 'javascript' in query_lower:
            return "qwen/qwen-2.5-coder-32b-instruct"
        elif api_keys.get('GLM_API_KEY'):
            return "glm-4-plus"
        else:
            return "microsoft/wizardlm-2-8x22b"  # OpenRouter free
    
    # Paid models
    if 'code' in query_lower:
        return "qwen/qwen-2.5-coder-32b-instruct"  # Still best for code
    elif any(word in query_lower for word in ['reason', 'think', 'analyze', 'complex']):
        return "claude-3-5-sonnet-20241022"
    elif 'image' in query_lower or 'photo' in query_lower:
        return "gpt-4o"  # Best multimodal
    else:
        return "claude-3-5-sonnet-20241022"  # Default

def _estimate_cost(model: str, input_tokens: int, output_tokens: int) -> float:
    """Simple cost estimation"""
    
    # Cost per 1K tokens (input, output)
    costs = {
        "gpt-4o": (0.005, 0.015),
        "gpt-4": (0.03, 0.06),
        "claude-3-5-sonnet-20241022": (0.003, 0.015),
        "claude-3-haiku": (0.00025, 0.00125),
    }
    
    if model not in costs:
        return 0.0  # Assume free
    
    input_cost, output_cost = costs[model]
    return (input_tokens / 1000 * input_cost) + (output_tokens / 1000 * output_cost)

# Health check - simple function
async def test_model(model: str, api_key: str, base_url: str = None) -> bool:
    """Test if model is working"""
    try:
        test_call = AICall(
            model=model,
            messages=[{"role": "user", "content": "Hi"}],
            max_tokens=10
        )
        
        await call_ai(test_call, api_key, base_url)
        return True
    except:
        return False