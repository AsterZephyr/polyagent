"""
工具服务管理
负责工具的注册、发现和执行
"""

import asyncio
from typing import Dict, List, Any, Optional
from app.core.config import Settings
from app.core.logging import LoggerMixin
from app.core.exceptions import ToolExecutionException
from app.api.models import ToolInfo, ToolExecutionResult

class ToolService(LoggerMixin):
    """工具服务管理器"""
    
    def __init__(self, settings: Settings):
        self.settings = settings
        self.tools: Dict[str, Dict[str, Any]] = {}
        
    async def startup(self):
        """启动工具服务"""
        self.logger.info("Initializing Tool Service...")
        
        # 注册内置工具
        await self._register_builtin_tools()
        
        self.logger.info(f"Tool Service initialized with {len(self.tools)} tools")
    
    async def shutdown(self):
        """关闭工具服务"""
        self.logger.info("Shutting down Tool Service...")
        self.tools.clear()
    
    async def _register_builtin_tools(self):
        """注册内置工具"""
        
        # 计算器工具
        self.tools["calculator"] = {
            "name": "calculator",
            "description": "Perform basic mathematical calculations",
            "parameters": {
                "type": "object",
                "properties": {
                    "expression": {
                        "type": "string",
                        "description": "Mathematical expression to evaluate"
                    }
                },
                "required": ["expression"]
            },
            "category": "math",
            "handler": self._calculator_handler
        }
        
        # 网页搜索工具（如果启用）
        if self.settings.ENABLE_WEB_SEARCH:
            self.tools["web_search"] = {
                "name": "web_search",
                "description": "Search the web for information",
                "parameters": {
                    "type": "object", 
                    "properties": {
                        "query": {
                            "type": "string",
                            "description": "Search query"
                        },
                        "max_results": {
                            "type": "integer",
                            "description": "Maximum number of results to return",
                            "default": 5
                        }
                    },
                    "required": ["query"]
                },
                "category": "search",
                "handler": self._web_search_handler
            }
        
        # 时间工具
        self.tools["get_time"] = {
            "name": "get_time",
            "description": "Get current date and time",
            "parameters": {
                "type": "object",
                "properties": {
                    "timezone": {
                        "type": "string",
                        "description": "Timezone (e.g., 'UTC', 'Asia/Shanghai')",
                        "default": "UTC"
                    }
                }
            },
            "category": "utility",
            "handler": self._time_handler
        }
    
    async def get_available_tools(self) -> List[ToolInfo]:
        """获取可用工具列表"""
        
        tool_list = []
        for tool_name, tool_def in self.tools.items():
            tool_list.append(ToolInfo(
                name=tool_def["name"],
                description=tool_def["description"],
                parameters=tool_def["parameters"],
                category=tool_def["category"]
            ))
        
        return tool_list
    
    async def get_tools_by_names(self, tool_names: List[str]) -> List[Dict[str, Any]]:
        """根据名称获取工具定义"""
        
        tools = []
        for name in tool_names:
            if name in self.tools:
                tool_def = self.tools[name].copy()
                # 移除内部字段
                tool_def.pop("handler", None)
                tools.append(tool_def)
            else:
                self.logger.warning(f"Tool '{name}' not found")
        
        return tools
    
    async def execute_tool(
        self,
        tool_name: str,
        parameters: Dict[str, Any]
    ) -> ToolExecutionResult:
        """执行工具"""
        
        if tool_name not in self.tools:
            raise ToolExecutionException(tool_name, "Tool not found")
        
        tool_def = self.tools[tool_name]
        handler = tool_def.get("handler")
        
        if not handler:
            raise ToolExecutionException(tool_name, "Tool handler not found")
        
        try:
            self.logger.info(f"Executing tool '{tool_name}' with parameters: {parameters}")
            
            # 执行工具处理函数
            result = await handler(parameters)
            
            return ToolExecutionResult(
                success=True,
                result=result
            )
            
        except Exception as e:
            self.logger.error(f"Tool execution failed: {str(e)}")
            return ToolExecutionResult(
                success=False,
                result=None,
                error=str(e)
            )
    
    # 工具处理函数
    
    async def _calculator_handler(self, parameters: Dict[str, Any]) -> Any:
        """计算器工具处理函数"""
        expression = parameters.get("expression", "")
        
        try:
            # 安全的数学表达式求值
            # 注意：在生产环境中应该使用更安全的方法
            allowed_chars = set("0123456789+-*/.() ")
            if not all(c in allowed_chars for c in expression):
                raise ValueError("Invalid characters in expression")
            
            result = eval(expression)
            return f"The result of {expression} is {result}"
            
        except Exception as e:
            raise Exception(f"Calculation error: {str(e)}")
    
    async def _web_search_handler(self, parameters: Dict[str, Any]) -> Any:
        """网页搜索工具处理函数"""
        query = parameters.get("query", "")
        max_results = parameters.get("max_results", 5)
        
        # 基础实现，后续可集成真实的搜索API
        self.logger.info(f"Web search: {query}")
        
        # 模拟搜索结果
        results = [
            {
                "title": f"Search result {i+1} for: {query}",
                "url": f"https://example.com/result{i+1}",
                "snippet": f"This is a sample search result snippet for query: {query}"
            }
            for i in range(min(max_results, 3))
        ]
        
        return {
            "query": query,
            "results": results,
            "total": len(results)
        }
    
    async def _time_handler(self, parameters: Dict[str, Any]) -> Any:
        """时间工具处理函数"""
        from datetime import datetime
        import pytz
        
        timezone_name = parameters.get("timezone", "UTC")
        
        try:
            if timezone_name == "UTC":
                tz = pytz.UTC
            else:
                tz = pytz.timezone(timezone_name)
            
            now = datetime.now(tz)
            
            return {
                "datetime": now.isoformat(),
                "timezone": timezone_name,
                "formatted": now.strftime("%Y-%m-%d %H:%M:%S %Z")
            }
            
        except Exception as e:
            raise Exception(f"Time error: {str(e)}")