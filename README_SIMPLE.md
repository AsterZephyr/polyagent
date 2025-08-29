# PolyAgent

Simple, reliable AI agent system following Linux philosophy.

## Philosophy

**Do One Thing and Do It Well** - PolyAgent focuses on AI conversation with document retrieval and tool calling. No unnecessary abstractions.

**Everything is a Function** - Like Linux treats everything as a file, PolyAgent treats every operation as a simple function call.

**Composition over Complexity** - Simple components that work together, not monolithic frameworks.

## Architecture

```
polyagent/
├── agent/      # Core AI agent (Python)
├── gateway/    # HTTP API gateway (Go, optional)  
├── docs/       # Document knowledge base
├── tools/      # External tool integrations
├── config/     # Configuration files
└── scripts/    # Setup and utility scripts
```

## Quick Start

### 1. Setup

```bash
# Clone and setup
git clone <repo> polyagent
cd polyagent
./scripts/setup.sh
```

### 2. Configure API Keys

```bash
# Edit configuration
cp config/env.example config/.env
vim config/.env  # Add your API keys
```

### 3. Run

```bash
# Interactive mode
cd agent
source ../config/.env
python3 main.py

# Or pipe mode
echo "Hello, how are you?" | python3 main.py

# HTTP API (optional)
cd ../gateway && go run main.go
curl -X POST http://localhost:8080/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Hello!"}'
```

## Features

### ✅ What It Does Well

- **AI Chat**: Supports Claude, GPT, OpenRouter, GLM models
- **Smart Routing**: Automatically selects best model for each query
- **Document Search**: Hybrid BM25 + semantic search
- **Tool Calling**: Simple function calling with retry logic
- **Medical Safety**: Built-in safety checks for medical content
- **Cost Control**: Intelligent model routing to minimize costs
- **Unix-Style**: Works with pipes, environment variables, exit codes

### ❌ What It Doesn't Do

- Complex multi-agent orchestration (use specialized frameworks)
- Real-time streaming (simple request-response model)
- Database ORM (direct SQL/NoSQL preferred)
- Authentication (add reverse proxy if needed)
- Web UI (it's a CLI tool, compose with web frameworks)

## Configuration

### Models (config/models.yaml)

```yaml
models:
  production:
    reasoning: "claude-3-5-sonnet-20241022"  # Best reasoning
    coding: "qwen/qwen-2.5-coder-32b-instruct"  # Best coding, free
    multimodal: "gpt-4o"  # Images/vision
    
  free:
    general: "qwen/qwen-2.5-coder-32b-instruct"  # OpenRouter free
    chinese: "glm-4-plus"  # 2M free tokens
```

### Environment (.env)

```bash
# Required: At least one API key
OPENAI_API_KEY=sk-your-key
ANTHROPIC_API_KEY=sk-ant-your-key  
OPENROUTER_API_KEY=sk-or-your-key
GLM_API_KEY=your-glm-key

# Optional: Behavior
POLYAGENT_VERBOSE=true
POLYAGENT_TOOLS=true
POLYAGENT_DOCS=./docs/medical,./docs/tech
```

## Usage Examples

### Basic Chat

```bash
> Hello, how are you?
Assistant: I'm doing well, thank you! How can I help you today?

> What's 2+2*3?
Assistant: calculate(2 + 2 * 3)
8

> Search for information about Python
Assistant: [Searches loaded documents and provides relevant information]
```

### Tool Usage

```python
# Register custom tools
from agent.tools import register_tool

@register_tool("get_weather")
async def get_weather(location: str) -> str:
    # Your implementation
    return f"Weather in {location}: Sunny, 22°C"
```

### Document Search

```bash
# Add documents
export POLYAGENT_DOCS="./docs/medical,./docs/tech"
python3 main.py

> What are the symptoms of hypertension?
Assistant: [Searches medical documents and provides answer with disclaimer]
⚠️ Medical reminder: This information is for reference only...
```

### Medical Safety

Medical queries automatically get safety checks:

```bash
> I have chest pain, what should I do?
Assistant: 🚨 For chest pain or any medical emergency, please seek immediate medical attention or call emergency services...
```

## API Reference

### Core Functions

```python
# AI calling
from agent.ai import call_ai, AICall
response = await call_ai(AICall(model="claude-3-5-sonnet", messages=[...]))

# Document search  
from agent.retrieve import search
results = await search("query", documents, method="hybrid")

# Tool calling
from agent.tools import call_tool, register_tool
result = await call_tool("tool_name", {"param": "value"})
```

### HTTP API (via Gateway)

```bash
# Chat
POST /chat
{
  "message": "Hello",
  "context": "optional context",
  "use_tools": true
}

# Health check
GET /health
```

## Performance

- **Latency**: ~1-3s (model dependent)
- **Throughput**: ~100 requests/min (API limits)  
- **Memory**: ~50MB baseline
- **Search**: ~50-200ms for 10k documents

## Deployment

### Single Instance (Recommended)

```bash
# Direct execution
cd agent && python3 main.py

# Or with systemd
sudo systemctl enable polyagent
sudo systemctl start polyagent
```

### Load Balanced (If Needed)

```bash
# Multiple agents behind Go gateway
cd gateway && go build -o gateway main.go
./gateway  # Proxies to Python agents
```

## Why This Design?

### Python for AI
- **Ecosystem**: PyTorch, Transformers, best AI libraries
- **Development Speed**: Faster iteration for AI logic
- **Community**: Largest AI developer community

### Go for Gateway (Optional)  
- **Performance**: Better concurrency for HTTP requests
- **Simplicity**: Single binary deployment
- **Reliability**: Better error handling for production

### Simple Architecture
- **Maintainable**: Easy to understand and modify
- **Debuggable**: Clear separation of concerns
- **Testable**: Each component independently testable

## Troubleshooting

### Common Issues

```bash
# No API keys
Error: No API keys found
Solution: Add API keys to config/.env

# Module import errors
Error: ModuleNotFoundError
Solution: pip3 install httpx

# Permission errors
Error: Permission denied
Solution: chmod +x scripts/setup.sh
```

### Debug Mode

```bash
export POLYAGENT_VERBOSE=true
export POLYAGENT_LOG_LEVEL=DEBUG
python3 main.py
```

## Contributing

1. **Keep It Simple** - Follow Linux philosophy
2. **Test Changes** - Run `python3 test.py`
3. **Document Updates** - Update README for API changes
4. **Follow Conventions** - Unix-style commands and configs

## License

MIT License - Use freely, modify as needed.

---

*"Simplicity is the ultimate sophistication." - Leonardo da Vinci*

*Built with Linux philosophy: Simple, reliable, composable.*