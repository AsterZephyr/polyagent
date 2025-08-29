"""
Tools - Function calling without the ceremony
Simple tool system following Unix philosophy
"""

import asyncio
import json
import re
from typing import Dict, Any, Callable, Optional, List

# Global tool registry - simple dictionary
# Like /proc filesystem in Linux - simple, accessible
TOOLS: Dict[str, Callable] = {}

def register_tool(name: str):
    """
    Register a tool decorator - simple as possible
    
    Usage:
        @register_tool("search_web")
        def search_web(query: str) -> str:
            return f"Searching for: {query}"
    """
    def decorator(func: Callable):
        TOOLS[name] = func
        return func
    return decorator

async def call_tool(name: str, params: Dict[str, Any], retries: int = 2) -> Any:
    """
    Call a tool with retry. Simple.
    
    Like system calls in Linux - simple interface, robust implementation
    """
    
    if name not in TOOLS:
        raise ValueError(f"Tool not found: {name}")
    
    last_error = None
    
    for attempt in range(retries + 1):
        try:
            func = TOOLS[name]
            
            # Handle both sync and async functions
            if asyncio.iscoroutinefunction(func):
                return await func(**params)
            else:
                # Run sync function in thread pool to avoid blocking
                loop = asyncio.get_event_loop()
                return await loop.run_in_executor(None, lambda: func(**params))
                
        except Exception as e:
            last_error = e
            
            # Simple retry logic - only for network-like errors
            if _is_retryable_error(e) and attempt < retries:
                await asyncio.sleep(min(2 ** attempt, 5))  # Simple backoff
                continue
            else:
                break
    
    raise last_error

def _is_retryable_error(error: Exception) -> bool:
    """Check if error is worth retrying"""
    error_str = str(error).lower()
    retryable_keywords = ['timeout', 'connection', 'network', '503', '502', '429']
    return any(keyword in error_str for keyword in retryable_keywords)

def list_tools() -> List[str]:
    """List available tools"""
    return list(TOOLS.keys())

# Medical safety - separate concern, separate function  
def check_medical_safety(text: str) -> bool:
    """
    Check if medical text is safe. Simple boolean.
    
    Returns False if text contains dangerous medical advice
    """
    dangerous_patterns = [
        r'诊断为|确诊为',           # Diagnosis claims
        r'建议.*服用.*药|推荐.*药物',  # Medication recommendations  
        r'不需要看医生|无需就医',      # Discouraging medical care
        r'立即手术|需要手术',         # Surgery recommendations
        r'停止.*药物|停药'           # Stop medication advice
    ]
    
    text_lower = text.lower()
    return not any(re.search(pattern, text_lower) for pattern in dangerous_patterns)

def add_medical_disclaimer(text: str) -> str:
    """Add medical disclaimer if text contains medical content"""
    
    medical_keywords = ['症状', '治疗', '药物', '疾病', '诊断', '血压', '心率', '体温', '疼痛']
    
    if any(keyword in text for keyword in medical_keywords):
        disclaimer = "\n\n⚠️ 医疗提醒：此信息仅供参考，不能替代专业医疗建议。如有健康问题，请咨询合格的医疗专业人员。"
        return text + disclaimer
    
    return text

# Built-in tools - keep it simple

@register_tool("get_time")
def get_current_time() -> str:
    """Get current time"""
    import datetime
    return datetime.datetime.now().strftime("%Y-%m-%d %H:%M:%S")

@register_tool("calculate")  
def calculate(expression: str) -> str:
    """Simple calculator - safe evaluation"""
    try:
        # Only allow safe mathematical operations
        allowed_chars = set('0123456789+-*/.() ')
        if not all(c in allowed_chars for c in expression):
            return "错误：只允许基本数学运算"
        
        # Evaluate safely
        result = eval(expression, {"__builtins__": {}})
        return str(result)
    except Exception as e:
        return f"计算错误：{str(e)}"

@register_tool("search_web")
async def search_web(query: str) -> str:
    """
    Simple web search - placeholder implementation
    In production, you'd integrate with actual search APIs
    """
    # This is a placeholder - integrate with real search API
    return f"搜索结果：'{query}' 的相关信息（这是一个示例搜索结果）"

@register_tool("weather")
async def get_weather(location: str) -> str:
    """
    Get weather info - placeholder implementation  
    In production, you'd integrate with weather APIs
    """
    # Placeholder - integrate with real weather API
    return f"{location} 的天气：晴朗，温度 22°C（这是一个示例天气信息）"

@register_tool("translate")
def translate_text(text: str, target_language: str = "English") -> str:
    """
    Simple translation - placeholder implementation
    In production, you'd integrate with translation APIs
    """
    # Placeholder - integrate with real translation API
    return f"翻译结果（{target_language}）：{text} (This is a placeholder translation)"

# Tool execution with function calling
async def extract_and_execute_tools(text: str, available_tools: List[str] = None) -> Tuple[str, List[Dict[str, Any]]]:
    """
    Extract tool calls from AI response and execute them
    Simple pattern matching - no fancy parsing
    """
    
    if available_tools is None:
        available_tools = list_tools()
    
    tool_calls = []
    modified_text = text
    
    # Look for tool call patterns: tool_name(param1="value1", param2="value2")
    tool_pattern = r'(\w+)\(([^)]*)\)'
    
    matches = re.finditer(tool_pattern, text)
    
    for match in matches:
        tool_name = match.group(1)
        params_str = match.group(2)
        
        if tool_name in available_tools:
            try:
                # Parse parameters - simple key=value parsing
                params = _parse_tool_params(params_str)
                
                # Execute tool
                result = await call_tool(tool_name, params)
                
                tool_calls.append({
                    "tool": tool_name,
                    "params": params,
                    "result": result
                })
                
                # Replace tool call with result in text
                tool_call_text = match.group(0)
                modified_text = modified_text.replace(tool_call_text, str(result))
                
            except Exception as e:
                error_msg = f"工具调用失败：{str(e)}"
                modified_text = modified_text.replace(match.group(0), error_msg)
    
    return modified_text, tool_calls

def _parse_tool_params(params_str: str) -> Dict[str, Any]:
    """Parse tool parameters from string"""
    params = {}
    
    if not params_str.strip():
        return params
    
    # Simple parsing - split by comma, then by =
    for param in params_str.split(','):
        param = param.strip()
        if '=' in param:
            key, value = param.split('=', 1)
            key = key.strip().strip('"\'')
            value = value.strip().strip('"\'')
            
            # Try to convert to appropriate type
            if value.lower() in ['true', 'false']:
                params[key] = value.lower() == 'true'
            elif value.isdigit():
                params[key] = int(value)
            elif value.replace('.', '').isdigit():
                params[key] = float(value)
            else:
                params[key] = value
    
    return params

# System information - Unix style
def get_system_info() -> Dict[str, Any]:
    """Get system information - like uname in Linux"""
    import platform
    import sys
    
    return {
        "platform": platform.system(),
        "version": platform.version(), 
        "python_version": sys.version,
        "available_tools": len(TOOLS),
        "tools": list(TOOLS.keys())
    }