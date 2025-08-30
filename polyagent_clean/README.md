# PolyAgent - Simple AI Agent System

Simple, reliable AI agent following Linux philosophy.

## Directory Structure

```
polyagent/
├── core/         # Python AI core engine (4 files)
│   ├── ai.py         # AI model calls
│   ├── retrieve.py   # Document search  
│   ├── tools.py      # Function calling
│   └── main.py       # CLI interface
├── gateway/      # Optional HTTP gateway (Go)
├── config/       # Configuration files
├── docs/         # Documentation
├── tools/        # External tool integrations
└── scripts/      # Setup and utilities
```

## Quick Start

```bash
# Setup
cd core
python3 -m venv venv
source venv/bin/activate
pip install httpx

# Configure
cp ../config/env.example ../config/.env
# Edit .env with your API keys

# Run
python3 main.py
```

## Features

- **Multi-Model Support**: Claude, GPT, OpenRouter, GLM
- **Smart Routing**: Auto-select best model for task
- **Document Search**: Hybrid BM25 + semantic search
- **Tool Calling**: Simple function registration
- **Medical Safety**: Built-in safety checks
- **Unix Style**: Pipes, environment vars, exit codes

## Philosophy

Following Linux kernel design principles:
- Do one thing and do it well
- Everything is a function
- Simple composition over complex inheritance
- Clear separation of concerns

## Performance

- **Startup**: ~0.5s
- **Memory**: ~50MB
- **Dependencies**: 1 (httpx only)
- **Files**: 4 core files