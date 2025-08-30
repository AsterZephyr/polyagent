#!/usr/bin/env python3
"""
PolyAgent - Simple AI Agent System
Following Linux philosophy: Simple, reliable, does one job well

Like a Unix utility:
- Read from stdin, write to stdout
- Configurable via environment variables  
- Does one thing (AI chat) and does it well
- Composable with other tools
"""

import asyncio
import os
import sys
import json
from typing import Dict, Any, List, Optional

from ai import call_ai, AICall, get_best_model, test_model
from retrieve import search, load_documents, SearchResult
from tools import (
    call_tool, extract_and_execute_tools, check_medical_safety, 
    add_medical_disclaimer, list_tools, get_system_info
)

class PolyAgent:
    """
    Simple AI agent - no fancy patterns, just works
    
    Like a Unix daemon - simple, reliable, focused
    """
    
    def __init__(self, api_keys: Dict[str, str], document_paths: List[str] = None):
        self.api_keys = {k: v for k, v in api_keys.items() if v}  # Remove None values
        self.documents = []
        
        # Load documents if provided
        if document_paths:
            print(f"Loading documents from {len(document_paths)} paths...")
            self.documents = load_documents(document_paths)
            print(f"Loaded {len(self.documents)} document chunks")
    
    async def chat(self, message: str, context: str = "", use_tools: bool = True) -> str:
        """
        Main chat function - keeps it simple
        
        Unix philosophy: Do one thing well
        """
        
        try:
            # 1. Document search if we have docs
            search_context = ""
            if self.documents:
                search_results = await search(
                    query=message, 
                    documents=self.documents, 
                    method="hybrid", 
                    top_k=3
                )
                
                if search_results:
                    search_context = "\n相关信息：\n" + "\n".join([
                        f"- {result.text[:200]}..." if len(result.text) > 200 else f"- {result.text}"
                        for result in search_results
                    ])
            
            # 2. Choose best model
            model = get_best_model(message, self.api_keys, free_only=False)
            
            # 3. Build conversation
            messages = []
            
            # System context
            system_context = "你是PolyAgent，一个有用的AI助手。"
            if context:
                system_context += f"\n\n背景信息：{context}"
            if search_context:
                system_context += f"\n\n{search_context}"
            if use_tools:
                available_tools = list_tools()
                system_context += f"\n\n可用工具：{', '.join(available_tools)}"
                system_context += "\n如需使用工具，请在回答中包含：tool_name(param=\"value\")"
            
            messages.append({"role": "system", "content": system_context})
            messages.append({"role": "user", "content": message})
            
            # 4. Call AI model
            api_key = self._get_api_key_for_model(model)
            if not api_key:
                return "错误：没有找到合适的API密钥"
            
            ai_call = AICall(
                model=model,
                messages=messages,
                temperature=0.7,
                max_tokens=2000
            )
            
            response = await call_ai(ai_call, api_key)
            result = response.content
            
            # 5. Execute tools if requested
            tool_calls = []
            if use_tools:
                result, tool_calls = await extract_and_execute_tools(result)
            
            # 6. Medical safety check
            if not check_medical_safety(result):
                result = "抱歉，我不能提供具体的医疗诊断或治疗建议。请咨询合格的医疗专业人员。"
            else:
                result = add_medical_disclaimer(result)
            
            # 7. Add usage info if requested
            if os.getenv('POLYAGENT_VERBOSE'):
                usage_info = f"\n\n[模型: {model}, 消耗: {response.usage.get('total_tokens', 0)} tokens"
                if response.cost > 0:
                    usage_info += f", 成本: ${response.cost:.4f}"
                if tool_calls:
                    usage_info += f", 工具调用: {len(tool_calls)}"
                usage_info += "]"
                result += usage_info
            
            return result
            
        except Exception as e:
            return f"错误：{str(e)}"
    
    def _get_api_key_for_model(self, model: str) -> Optional[str]:
        """Get API key for model - simple mapping"""
        
        if 'claude' in model:
            return self.api_keys.get('anthropic') or self.api_keys.get('ANTHROPIC_API_KEY')
        elif 'gpt' in model:
            return self.api_keys.get('openai') or self.api_keys.get('OPENAI_API_KEY')
        elif 'qwen' in model or 'openrouter' in model or 'wizardlm' in model:
            return self.api_keys.get('openrouter') or self.api_keys.get('OPENROUTER_API_KEY')
        elif 'glm' in model:
            return self.api_keys.get('glm') or self.api_keys.get('GLM_API_KEY')
        else:
            # Try any available key
            return next(iter(self.api_keys.values())) if self.api_keys else None
    
    async def health_check(self) -> Dict[str, Any]:
        """Check system health"""
        
        health = {
            "status": "healthy",
            "api_keys": list(self.api_keys.keys()),
            "documents_loaded": len(self.documents),
            "available_tools": len(list_tools()),
            "models": {}
        }
        
        # Test each model
        for key_name, api_key in self.api_keys.items():
            if key_name == 'OPENAI_API_KEY':
                model = 'gpt-4o'
            elif key_name == 'ANTHROPIC_API_KEY':
                model = 'claude-3-5-sonnet-20241022'
            elif key_name == 'OPENROUTER_API_KEY':
                model = 'qwen/qwen-2.5-coder-32b-instruct'
            elif key_name == 'GLM_API_KEY':
                model = 'glm-4-plus'
            else:
                continue
            
            try:
                is_working = await test_model(model, api_key)
                health["models"][model] = "working" if is_working else "failed"
            except Exception as e:
                health["models"][model] = f"error: {str(e)}"
        
        if not any(status == "working" for status in health["models"].values()):
            health["status"] = "unhealthy"
        
        return health

def load_config() -> Dict[str, Any]:
    """Load configuration from environment - Unix way"""
    
    return {
        'api_keys': {
            'OPENAI_API_KEY': os.getenv('OPENAI_API_KEY'),
            'ANTHROPIC_API_KEY': os.getenv('ANTHROPIC_API_KEY'),
            'OPENROUTER_API_KEY': os.getenv('OPENROUTER_API_KEY'),
            'GLM_API_KEY': os.getenv('GLM_API_KEY'),
        },
        'document_paths': os.getenv('POLYAGENT_DOCS', '').split(',') if os.getenv('POLYAGENT_DOCS') else [],
        'verbose': os.getenv('POLYAGENT_VERBOSE', '').lower() in ['1', 'true', 'yes'],
        'tools_enabled': os.getenv('POLYAGENT_TOOLS', 'true').lower() in ['1', 'true', 'yes'],
        'log_level': os.getenv('POLYAGENT_LOG_LEVEL', 'INFO').upper()
    }

async def cli_mode():
    """Interactive CLI mode"""
    
    config = load_config()
    
    # Filter out empty API keys  
    api_keys = {k: v for k, v in config['api_keys'].items() if v}
    
    if not api_keys:
        print("Error: No API keys found in environment variables:")
        print("  OPENAI_API_KEY - for GPT models")
        print("  ANTHROPIC_API_KEY - for Claude models") 
        print("  OPENROUTER_API_KEY - for OpenRouter models")
        print("  GLM_API_KEY - for GLM models")
        print("\nAt least one API key is required.")
        return 1
    
    print("PolyAgent - Simple AI Assistant")
    print(f"Available API keys: {list(api_keys.keys())}")
    
    # Initialize agent
    agent = PolyAgent(api_keys, config['document_paths'])
    
    # Health check
    if config['verbose']:
        print("\nPerforming health check...")
        health = await agent.health_check()
        print(f"Status: {health['status']}")
        print(f"Working models: {[k for k, v in health['models'].items() if v == 'working']}")
        print(f"Documents: {health['documents_loaded']}, Tools: {health['available_tools']}")
    
    print("\nReady! Type 'quit' to exit, 'help' for commands.")
    print("=" * 50)
    
    while True:
        try:
            user_input = input("\n> ").strip()
            
            if not user_input:
                continue
            
            if user_input.lower() in ['quit', 'exit', 'q']:
                print("Goodbye!")
                break
            
            if user_input.lower() in ['help', 'h']:
                print_help()
                continue
            
            if user_input.lower() in ['health', 'status']:
                health = await agent.health_check()
                print(json.dumps(health, indent=2, ensure_ascii=False))
                continue
            
            if user_input.lower().startswith('tools'):
                tools = list_tools()
                print(f"Available tools ({len(tools)}): {', '.join(tools)}")
                continue
            
            # Get response
            response = await agent.chat(
                message=user_input, 
                use_tools=config['tools_enabled']
            )
            
            print(f"\nAssistant: {response}")
            
        except KeyboardInterrupt:
            print("\n\nGoodbye!")
            break
        except Exception as e:
            print(f"Error: {e}")

def print_help():
    """Print help information"""
    help_text = """
PolyAgent Commands:
  help, h      - Show this help
  quit, q      - Exit the program  
  health       - Show system health
  tools        - List available tools
  
Environment Variables:
  OPENAI_API_KEY      - OpenAI API key
  ANTHROPIC_API_KEY   - Anthropic API key
  OPENROUTER_API_KEY  - OpenRouter API key  
  GLM_API_KEY         - GLM API key
  POLYAGENT_DOCS      - Comma-separated document paths
  POLYAGENT_VERBOSE   - Enable verbose output (1/true)
  POLYAGENT_TOOLS     - Enable tool calling (1/true, default)
  
Examples:
  > Hello, how are you?
  > What's the weather like?
  > calculate(2 + 2 * 3)
  > Search for information about Python
"""
    print(help_text)

async def pipe_mode():
    """Pipe mode - read from stdin, write to stdout"""
    
    config = load_config()
    api_keys = {k: v for k, v in config['api_keys'].items() if v}
    
    if not api_keys:
        print("Error: No API keys configured", file=sys.stderr)
        return 1
    
    agent = PolyAgent(api_keys, config['document_paths'])
    
    # Read all input
    input_text = sys.stdin.read().strip()
    if not input_text:
        return 1
    
    # Process and output
    response = await agent.chat(input_text, use_tools=config['tools_enabled'])
    print(response)
    
    return 0

async def main():
    """Main entry point - Unix style"""
    
    # Check if running in pipe mode (no TTY)
    if not sys.stdin.isatty():
        return await pipe_mode()
    else:
        return await cli_mode()

if __name__ == "__main__":
    try:
        exit_code = asyncio.run(main())
        sys.exit(exit_code or 0)
    except KeyboardInterrupt:
        print("\nInterrupted")
        sys.exit(130)  # Standard Unix exit code for Ctrl+C
    except Exception as e:
        print(f"Fatal error: {e}", file=sys.stderr)
        sys.exit(1)