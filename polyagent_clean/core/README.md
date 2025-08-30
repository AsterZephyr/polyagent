# Core Module - PolyAgent AI Engine

## Overview

Core模块是PolyAgent的核心AI引擎，遵循Linux设计哲学，实现了"Do One Thing Well"的原则。包含4个核心文件，每个文件负责一个明确的职责。

## Architecture

```mermaid
graph TB
    subgraph "PolyAgent Core Architecture"
        User[👤 User] --> CLI[main.py - CLI Interface]
        CLI --> Agent[PolyAgent Class]
        
        Agent --> AI[ai.py - Model Calls]
        Agent --> Search[retrieve.py - Document Search]
        Agent --> Tools[tools.py - Function Calling]
        
        AI --> Claude[Claude API]
        AI --> OpenAI[OpenAI API]
        AI --> OpenRouter[OpenRouter API]
        AI --> GLM[GLM API]
        
        Search --> BM25[BM25 Search]
        Search --> Semantic[Semantic Search]
        
        Tools --> BuiltIn[Built-in Tools]
        Tools --> Custom[Custom Tools]
        Tools --> Medical[Medical Safety]
    end
    
    subgraph "External Services"
        Claude
        OpenAI
        OpenRouter
        GLM
    end
```

## Data Flow

```mermaid
sequenceDiagram
    participant U as User
    participant M as main.py
    participant A as PolyAgent
    participant R as retrieve.py
    participant AI as ai.py
    participant T as tools.py
    
    U->>M: Input message
    M->>A: chat(message)
    
    A->>R: search(query, documents)
    R->>A: search_results
    
    A->>AI: get_best_model(query, api_keys)
    AI->>A: model_name
    
    A->>AI: call_ai(request, api_key)
    AI->>A: ai_response
    
    A->>T: extract_and_execute_tools(response)
    T->>A: processed_response
    
    A->>T: check_medical_safety(response)
    T->>A: safety_checked_response
    
    A->>M: final_response
    M->>U: Output
```

## Module Breakdown

### 1. main.py - CLI Interface & Orchestration

**职责**: 命令行接口和系统协调

**核心类**:
```python
class PolyAgent:
    def __init__(self, api_keys: Dict[str, str], document_paths: List[str] = None)
    async def chat(self, message: str, context: str = "", use_tools: bool = True) -> str
    async def health_check(self) -> Dict[str, Any]
```

**特性**:
- Unix风格CLI接口（支持管道、环境变量）
- 交互模式和管道模式
- 优雅的错误处理
- 健康检查和监控

### 2. ai.py - AI Model Integration

**职责**: AI模型调用和路由

**核心函数**:
```python
async def call_ai(request: AICall, api_key: str, base_url: str = None) -> AIResponse
def get_best_model(query: str, api_keys: Dict[str, str], free_only: bool = False) -> str
async def test_model(model: str, api_key: str, base_url: str = None) -> bool
```

**支持的模型**:
- **Claude**: claude-3-5-sonnet-20241022, claude-4-opus, claude-4-sonnet
- **OpenAI**: gpt-4o, gpt-5, gpt-4-turbo
- **OpenRouter**: qwen/qwen-2.5-coder-32b-instruct, openrouter/k2-free, qwen/qwen-3-coder-free
- **GLM**: glm-4-plus, glm-4.5-turbo

**模型路由逻辑**:
```mermaid
flowchart TD
    Query[User Query] --> Check{Check Query Type}
    
    Check -->|Contains 'code'| Code[qwen/qwen-2.5-coder-32b-instruct]
    Check -->|Contains 'reason/analyze'| Reason[claude-3-5-sonnet-20241022]
    Check -->|Contains 'image'| Image[gpt-4o]
    Check -->|Default| Default[claude-3-5-sonnet-20241022]
    
    Code --> APICheck{API Key Available?}
    Reason --> APICheck
    Image --> APICheck
    Default --> APICheck
    
    APICheck -->|Yes| UseModel[Use Selected Model]
    APICheck -->|No| Fallback[Use Free Model]
    
    Fallback --> FreeModel[microsoft/wizardlm-2-8x22b]
```

### 3. retrieve.py - Document Search & RAG

**职责**: 文档检索和相关性匹配

**核心函数**:
```python
async def search(query: str, documents: List[str], method: str = "hybrid", top_k: int = 5) -> List[SearchResult]
def load_documents(paths: List[str]) -> List[str]
```

**搜索方法**:
```mermaid
graph LR
    Query[Search Query] --> Method{Search Method}
    
    Method -->|keyword| BM25[BM25 Algorithm]
    Method -->|semantic| Semantic[Semantic Similarity]
    Method -->|hybrid| Hybrid[BM25 + Semantic]
    
    BM25 --> BM25Score[TF-IDF + BM25 Score]
    Semantic --> EmbedScore[Cosine Similarity]
    Hybrid --> CombineScore[Combined Score]
    
    BM25Score --> Rank[Ranked Results]
    EmbedScore --> Rank
    CombineScore --> Rank
    
    Rank --> TopK[Top-K Results]
```

**BM25参数**:
- k1 = 1.2 (term frequency saturation)
- b = 0.75 (field length normalization)

### 4. tools.py - Function Calling & Safety

**职责**: 工具调用和安全检查

**核心功能**:
```python
def register_tool(name: str) -> Decorator
async def call_tool(name: str, params: Dict[str, Any], retries: int = 2) -> Any
def check_medical_safety(text: str) -> bool
def add_medical_disclaimer(text: str) -> str
```

**工具执行流程**:
```mermaid
sequenceDiagram
    participant AI as AI Response
    participant P as Pattern Matcher
    participant E as Tool Executor
    participant R as Retry Logic
    participant S as Safety Check
    
    AI->>P: tool_name(param="value")
    P->>E: Extract tool calls
    E->>R: Execute with retry
    
    loop Retry Logic
        R->>R: Attempt execution
        R->>R: Handle failures
        R->>R: Backoff delay
    end
    
    R->>S: Tool result
    S->>S: Medical safety check
    S->>AI: Safe result
```

**内置工具**:
- `get_time`: 获取当前时间
- `calculate`: 安全数学计算
- `search_web`: 网络搜索（占位符）
- `weather`: 天气查询（占位符）
- `translate`: 文本翻译（占位符）

**医疗安全模式**:
```python
dangerous_patterns = [
    r'诊断为|确诊为',           # Diagnosis claims
    r'建议.*服用.*药|推荐.*药物',  # Medication recommendations  
    r'不需要看医生|无需就医',      # Discouraging medical care
    r'立即手术|需要手术',         # Surgery recommendations
    r'停止.*药物|停药'           # Stop medication advice
]
```

## Configuration

### Environment Variables

```bash
# API Keys (至少需要一个)
OPENAI_API_KEY=sk-your-key
ANTHROPIC_API_KEY=sk-ant-your-key  
OPENROUTER_API_KEY=sk-or-your-key
GLM_API_KEY=your-glm-key

# Behavior Configuration
POLYAGENT_VERBOSE=true          # 启用详细输出
POLYAGENT_TOOLS=true            # 启用工具调用
POLYAGENT_DOCS=./docs/medical,./docs/tech  # 文档路径
POLYAGENT_LOG_LEVEL=INFO        # 日志级别
```

## Performance Metrics

| 操作 | 延迟 | 吞吐量 |
|-----|------|--------|
| 启动时间 | ~0.5s | N/A |
| AI调用 | 1-3s | 取决于模型 |
| 文档搜索 | 50-200ms | 1000 docs/s |
| 工具调用 | 10-100ms | 100 calls/s |
| 内存使用 | ~50MB | N/A |

## Error Handling

```mermaid
flowchart TD
    Error[Error Occurs] --> Type{Error Type}
    
    Type -->|API Error| APIHandle[API Error Handler]
    Type -->|Network Error| NetHandle[Network Retry]
    Type -->|Tool Error| ToolHandle[Tool Error Handler]
    Type -->|Config Error| ConfigHandle[Config Error Handler]
    
    APIHandle --> Log[Log Error]
    NetHandle --> Retry[Retry with Backoff]
    ToolHandle --> Fallback[Fallback Response]
    ConfigHandle --> UserMsg[User-Friendly Message]
    
    Log --> UserResponse[User Response]
    Retry --> UserResponse
    Fallback --> UserResponse
    UserMsg --> UserResponse
```

## Testing

```bash
# 基础功能测试
python3 test_simple.py

# 集成测试
python3 ../test_integration_fixed.py

# 模型路由测试
python3 ../test_model_routing.py
```

## Usage Examples

### 基本对话
```bash
python3 main.py
> Hello, how are you?
Assistant: I'm doing well, thank you! How can I help you today?
```

### 代码相关查询
```bash
> Write a Python function to sort a list
Assistant: [Uses qwen/qwen-2.5-coder-32b-instruct automatically]
```

### 工具调用
```bash
> What time is it?
Assistant: get_time()
2024-08-30 14:30:25

> Calculate 15 * 32 + 7
Assistant: calculate(15 * 32 + 7)
487
```

### 管道模式
```bash
echo "Explain quantum computing" | python3 main.py
# 输出解释内容
```

## Extension Points

### 添加新模型
```python
# 在 ai.py 中添加新的 _call_newmodel 函数
# 在 call_ai() 中添加模型路由逻辑
```

### 添加自定义工具
```python
from tools import register_tool

@register_tool("my_tool")
def my_custom_tool(param: str) -> str:
    return f"Processed: {param}"
```

### 自定义搜索方法
```python
# 在 retrieve.py 中添加新的搜索方法
async def search_custom(query: str, documents: List[str]) -> List[SearchResult]:
    # 实现自定义搜索逻辑
    pass
```

---

*Core模块体现了Unix哲学：简单、可靠、可组合。每个文件做好一件事，组合起来形成强大的AI系统。*