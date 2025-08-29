# Linux Philosophy Refactor Analysis

## 当前架构问题 (Linus批判视角)

### 1. 违反"Do One Thing Well"原则

**问题**：
- `UnifiedAIAdapter` 既管理模型又管理代理还管理健康检查
- `HybridRetriever` 既做BM25又做向量又做融合还做评估
- `EnhancedFunctionCall` 既管工具注册又管重试又管MCP又管日志

**Linus会说**：
> "This is a fucking mess! One class doing everything is not 'unified', it's just lazy design. Break it down!"

### 2. 过度工程化

**问题**：
```
app/
├── adapters/         # AI模型适配
├── oxy/             # OxyGent组件系统  
├── tools/           # 工具调用系统
├── medical/         # 医疗安全模块
├── rag/             # 检索增强生成
├── core/            # 核心基础设施
└── services/        # 业务服务层
```

**Linus会说**：
> "Why do you need 7 different subsystems to call an AI API? This looks like enterprise Java bullshit!"

### 3. 抽象层过多

**问题**：
- `BaseOxy -> OxyLLM -> UnifiedAIAdapter -> OpenAIAdapter -> openai.AsyncOpenAI`
- 用户调用一个AI接口需要经过5层抽象

**Linus会说**：
> "Abstractions should reduce complexity, not create it. This is abstraction masturbation!"

## Linux哲学重构方案

### 核心原则：Everything is a Function Call

就像Linux中"Everything is a File"，我们采用"Everything is a Function Call"：

```
/polyagent/
├── ai.py           # AI调用的核心实现（类似kernel/sched.c）
├── retrieve.py     # 检索功能（类似fs/）  
├── tools.py        # 工具调用（类似drivers/）
└── main.py         # 主程序（类似init/main.c）
```

### 重构实现：

#### 1. ai.py - 核心AI调用模块
```python
"""
AI Core - Simple, fast, reliable
Like Linux system calls - one function, one purpose
"""

import asyncio
import json
from typing import Dict, Any, List, Optional
from dataclasses import dataclass

@dataclass
class AICall:
    model: str
    messages: List[Dict[str, str]] 
    temperature: float = 0.7
    max_tokens: int = 2000

@dataclass  
class AIResponse:
    content: str
    usage: Dict[str, int]
    model: str

# Core function - does one thing well
async def call_ai(request: AICall, api_key: str, base_url: str = None) -> AIResponse:
    """Single function to call any AI model. Period."""
    
    if request.model.startswith('claude'):
        return await _call_claude(request, api_key, base_url)
    elif request.model.startswith('gpt'):
        return await _call_openai(request, api_key, base_url) 
    elif request.model.startswith('qwen'):
        return await _call_openrouter(request, api_key)
    else:
        raise ValueError(f"Unsupported model: {request.model}")

# Implementation functions - private, simple
async def _call_claude(request: AICall, api_key: str, base_url: str) -> AIResponse:
    # Direct Claude API call - no abstractions
    pass

async def _call_openai(request: AICall, api_key: str, base_url: str) -> AIResponse:
    # Direct OpenAI API call - no abstractions  
    pass

async def _call_openrouter(request: AICall, api_key: str) -> AIResponse:
    # Direct OpenRouter API call - no abstractions
    pass

# Model routing - simple function
def get_best_model(task: str, free_only: bool = False) -> str:
    """Route to best model for task"""
    if free_only:
        return "qwen/qwen-2.5-coder-32b-instruct"
    
    if "code" in task.lower():
        return "qwen/qwen-2.5-coder-32b-instruct"
    elif "reason" in task.lower():  
        return "claude-3-5-sonnet"
    else:
        return "gpt-4o"
```

#### 2. retrieve.py - 简单检索模块
```python
"""
Retrieval - Simple search functionality
"""

import asyncio
from typing import List, Dict, Any
from dataclasses import dataclass

@dataclass
class SearchResult:
    text: str
    score: float
    source: str

# One function for search - that's it
async def search(query: str, 
                docs: List[str],
                method: str = "hybrid",
                top_k: int = 5) -> List[SearchResult]:
    """Search documents. Simple."""
    
    if method == "keyword":
        return await _bm25_search(query, docs, top_k)
    elif method == "semantic":
        return await _vector_search(query, docs, top_k)
    elif method == "hybrid":
        # Simple combination - no fancy fusion
        keyword_results = await _bm25_search(query, docs, top_k)
        semantic_results = await _vector_search(query, docs, top_k)
        
        # Dead simple fusion
        all_results = {}
        for result in keyword_results + semantic_results:
            if result.text in all_results:
                all_results[result.text].score += result.score
            else:
                all_results[result.text] = result
        
        return sorted(all_results.values(), key=lambda x: x.score, reverse=True)[:top_k]
    
    else:
        raise ValueError(f"Unknown method: {method}")

async def _bm25_search(query: str, docs: List[str], top_k: int) -> List[SearchResult]:
    """BM25 keyword search - straightforward implementation"""
    # Simple BM25 - no fancy parameters
    pass

async def _vector_search(query: str, docs: List[str], top_k: int) -> List[SearchResult]:
    """Vector semantic search - straightforward implementation"""  
    # Simple cosine similarity - no fancy embeddings
    pass
```

#### 3. tools.py - 工具调用模块
```python
"""
Tools - Function calling without the ceremony
"""

import asyncio
import json
from typing import Dict, Any, Callable, Optional

# Global tool registry - simple dictionary
TOOLS: Dict[str, Callable] = {}

def register_tool(name: str, func: Callable):
    """Register a tool. That's it."""
    TOOLS[name] = func

async def call_tool(name: str, params: Dict[str, Any], retries: int = 3) -> Any:
    """Call a tool with retry. Simple."""
    
    if name not in TOOLS:
        raise ValueError(f"Tool not found: {name}")
    
    last_error = None
    for attempt in range(retries):
        try:
            func = TOOLS[name]
            if asyncio.iscoroutinefunction(func):
                return await func(**params)
            else:
                return func(**params)
        except Exception as e:
            last_error = e
            if attempt < retries - 1:
                await asyncio.sleep(2 ** attempt)  # Simple backoff
    
    raise last_error

# Medical safety - separate concern, separate function
def check_medical_safety(text: str) -> bool:
    """Check if medical text is safe. Simple boolean."""
    dangerous_patterns = ['诊断为', '确诊', '建议服用']
    return not any(pattern in text for pattern in dangerous_patterns)

def add_medical_disclaimer(text: str) -> str:
    """Add disclaimer if needed."""
    if any(word in text for word in ['症状', '治疗', '药物']):
        return text + "\n\n⚠️ 此信息仅供参考，请咨询医疗专业人员。"
    return text
```

#### 4. main.py - 主程序
```python
"""
PolyAgent - Simple AI Agent System
Like a Unix utility - does one job well
"""

import asyncio
import os
from typing import Dict, Any

from ai import call_ai, AICall, get_best_model
from retrieve import search
from tools import call_tool, check_medical_safety, add_medical_disclaimer

class PolyAgent:
    """Simple AI agent - no fancy patterns"""
    
    def __init__(self, api_keys: Dict[str, str]):
        self.api_keys = api_keys
        self.docs = []  # Simple document storage
    
    async def chat(self, message: str, context: str = "") -> str:
        """Main chat function - keeps it simple"""
        
        # 1. Search if we have docs
        search_results = []
        if self.docs:
            search_results = await search(message, self.docs, method="hybrid", top_k=3)
            context += "\n".join([r.text for r in search_results])
        
        # 2. Choose model
        model = get_best_model(message, free_only=not self.api_keys.get('OPENAI_API_KEY'))
        
        # 3. Build messages
        messages = []
        if context:
            messages.append({"role": "system", "content": context})
        messages.append({"role": "user", "content": message})
        
        # 4. Call AI
        api_key = self._get_api_key_for_model(model)
        response = await call_ai(
            AICall(model=model, messages=messages),
            api_key=api_key
        )
        
        # 5. Safety check for medical content
        result = response.content
        if not check_medical_safety(result):
            result = "抱歉，我不能提供具体的医疗诊断建议。请咨询医疗专业人员。"
        else:
            result = add_medical_disclaimer(result)
        
        return result
    
    def _get_api_key_for_model(self, model: str) -> str:
        """Get API key for model - simple mapping"""
        if model.startswith('claude'):
            return self.api_keys['ANTHROPIC_API_KEY']
        elif model.startswith('gpt'):
            return self.api_keys['OPENAI_API_KEY'] 
        elif 'openrouter' in model or 'qwen' in model:
            return self.api_keys['OPENROUTER_API_KEY']
        else:
            raise ValueError(f"No API key for model: {model}")

# CLI interface - Unix style
async def main():
    """Main entry point"""
    
    # Get API keys from environment - Unix way
    api_keys = {
        'OPENAI_API_KEY': os.getenv('OPENAI_API_KEY'),
        'ANTHROPIC_API_KEY': os.getenv('ANTHROPIC_API_KEY'),
        'OPENROUTER_API_KEY': os.getenv('OPENROUTER_API_KEY'),
    }
    
    # Remove None values
    api_keys = {k: v for k, v in api_keys.items() if v}
    
    if not api_keys:
        print("Error: No API keys found in environment")
        return 1
    
    agent = PolyAgent(api_keys)
    
    print("PolyAgent ready. Type 'quit' to exit.")
    
    while True:
        try:
            user_input = input("> ").strip()
            if user_input.lower() in ['quit', 'exit']:
                break
            
            if not user_input:
                continue
                
            response = await agent.chat(user_input)
            print(f"Assistant: {response}\n")
            
        except KeyboardInterrupt:
            break
        except Exception as e:
            print(f"Error: {e}")
    
    return 0

if __name__ == "__main__":
    exit(asyncio.run(main()))
```