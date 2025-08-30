"""
AI Core - Simple, reliable AI calls (no external dependencies version)
Following Linux philosophy: Do one thing and do it well
"""

import asyncio
import json
import urllib.request
import urllib.parse
import urllib.error
from typing import Dict, Any, List, Optional
from dataclasses import dataclass

@dataclass
class AICall:
    """AI request - simple data structure"""
    model: str
    messages: List[Dict[str, str]]
    temperature: float = 0.7
    max_tokens: int = 2000

@dataclass
class AIResponse:
    """AI response - simple data structure"""
    content: str
    usage: Dict[str, int]
    model: str
    cost: float = 0.0

# Fallback AI call using only standard library
async def call_ai_fallback(request: AICall, api_key: str, base_url: str = None) -> AIResponse:
    """
    Fallback AI calling using only standard library
    This is a basic implementation for demonstration
    """
    
    # In a real implementation without external dependencies,
    # you would implement HTTP calls using urllib
    # For now, return a mock response
    
    mock_response = f"Mock AI response to: {request.messages[-1]['content'][:50]}..."
    
    return AIResponse(
        content=mock_response,
        usage={"prompt_tokens": 20, "completion_tokens": 30, "total_tokens": 50},
        model=request.model,
        cost=0.001
    )

# Simple model selection
def get_best_model_simple(query: str, free_only: bool = True) -> str:
    """Simple model selection without external config"""
    
    if free_only:
        return "mock-free-model"
    else:
        return "mock-premium-model"

# Test function
async def test_simple_ai():
    """Test the simple AI system"""
    
    test_call = AICall(
        model="mock-model",
        messages=[{"role": "user", "content": "Hello, this is a test"}]
    )
    
    response = await call_ai_fallback(test_call, "mock-api-key")
    
    print(f"Test response: {response.content}")
    print(f"Usage: {response.usage}")
    print(f"Cost: ${response.cost}")
    
    return True

if __name__ == "__main__":
    # Self-test
    asyncio.run(test_simple_ai())