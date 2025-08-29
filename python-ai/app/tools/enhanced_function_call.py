"""
Enhanced Function Call System with Error Handling and MCP Integration
"""

import asyncio
import json
import traceback
from typing import Dict, List, Any, Optional, Callable, Union
from dataclasses import dataclass, field
from enum import Enum
import time
from functools import wraps
import logging

from ..core.logging import LoggerMixin
from ..core.tracing import get_tracer
from ..core.base_exceptions import ComponentExecutionException

class ToolStatus(Enum):
    """Tool execution status"""
    PENDING = "pending"
    RUNNING = "running"
    SUCCESS = "success"
    FAILED = "failed"
    RETRYING = "retrying"
    TIMEOUT = "timeout"

@dataclass
class ToolCall:
    """Tool call execution record"""
    call_id: str
    tool_name: str
    parameters: Dict[str, Any]
    status: ToolStatus = ToolStatus.PENDING
    result: Any = None
    error: Optional[str] = None
    start_time: float = field(default_factory=time.time)
    end_time: Optional[float] = None
    execution_time: Optional[float] = None
    retry_count: int = 0
    max_retries: int = 3
    context: Dict[str, Any] = field(default_factory=dict)

@dataclass
class ToolMetadata:
    """Tool metadata for registration"""
    name: str
    description: str
    parameters_schema: Dict[str, Any]
    return_schema: Optional[Dict[str, Any]] = None
    timeout: float = 30.0
    max_retries: int = 3
    category: str = "general"
    requires_auth: bool = False
    rate_limit: Optional[Dict[str, Any]] = None

def tool_logger(func):
    """Decorator for automatic tool logging and tracing"""
    
    @wraps(func)
    async def async_wrapper(*args, **kwargs):
        tool_name = func.__name__
        logger = logging.getLogger(f"polyagent.tools.{tool_name}")
        tracer = get_tracer()
        
        # Start span for tool execution
        trace_context = tracer.start_trace(f"tool_call_{tool_name}")
        
        try:
            logger.info(f"Starting tool execution: {tool_name}", extra={
                "tool": tool_name,
                "parameters": kwargs
            })
            
            start_time = time.time()
            result = await func(*args, **kwargs)
            execution_time = time.time() - start_time
            
            logger.info(f"Tool execution completed: {tool_name}", extra={
                "tool": tool_name,
                "execution_time": execution_time,
                "success": True
            })
            
            await tracer.finish_span(trace_context, status="success",
                                   tool=tool_name, execution_time=execution_time)
            
            return result
            
        except Exception as e:
            execution_time = time.time() - start_time
            error_details = {
                "tool": tool_name,
                "error": str(e),
                "error_type": type(e).__name__,
                "execution_time": execution_time,
                "traceback": traceback.format_exc()
            }
            
            logger.error(f"Tool execution failed: {tool_name}", extra=error_details)
            await tracer.finish_span(trace_context, status="error", error=str(e))
            
            raise
    
    @wraps(func)
    def sync_wrapper(*args, **kwargs):
        tool_name = func.__name__
        logger = logging.getLogger(f"polyagent.tools.{tool_name}")
        
        try:
            logger.info(f"Starting tool execution: {tool_name}", extra={
                "tool": tool_name,
                "parameters": kwargs
            })
            
            start_time = time.time()
            result = func(*args, **kwargs)
            execution_time = time.time() - start_time
            
            logger.info(f"Tool execution completed: {tool_name}", extra={
                "tool": tool_name,
                "execution_time": execution_time,
                "success": True
            })
            
            return result
            
        except Exception as e:
            execution_time = time.time() - start_time
            error_details = {
                "tool": tool_name,
                "error": str(e),
                "error_type": type(e).__name__,
                "execution_time": execution_time,
                "traceback": traceback.format_exc()
            }
            
            logger.error(f"Tool execution failed: {tool_name}", extra=error_details)
            raise
    
    # Return appropriate wrapper based on function type
    if asyncio.iscoroutinefunction(func):
        return async_wrapper
    else:
        return sync_wrapper

class EnhancedFunctionCallSystem(LoggerMixin):
    """Enhanced function call system with retry, error handling, and MCP support"""
    
    def __init__(self):
        super().__init__()
        self.tools: Dict[str, Callable] = {}
        self.tool_metadata: Dict[str, ToolMetadata] = {}
        self.call_history: List[ToolCall] = []
        self.mcp_connections: Dict[str, Any] = {}  # MCP server connections
        
    def register_tool(self, 
                     func: Callable,
                     metadata: ToolMetadata,
                     auto_log: bool = True) -> None:
        """Register a tool with metadata"""
        
        if auto_log and not hasattr(func, '__wrapped__'):
            func = tool_logger(func)
        
        self.tools[metadata.name] = func
        self.tool_metadata[metadata.name] = metadata
        
        self.logger.info(f"Registered tool: {metadata.name}")
    
    def register_mcp_server(self, 
                           server_name: str, 
                           connection_config: Dict[str, Any]) -> None:
        """Register MCP server connection"""
        # This would integrate with actual MCP protocol
        self.mcp_connections[server_name] = connection_config
        self.logger.info(f"Registered MCP server: {server_name}")
    
    async def execute_tool_call(self, 
                               call_id: str,
                               tool_name: str, 
                               parameters: Dict[str, Any],
                               context: Dict[str, Any] = None) -> ToolCall:
        """Execute tool call with enhanced error handling and retry logic"""
        
        tool_call = ToolCall(
            call_id=call_id,
            tool_name=tool_name,
            parameters=parameters,
            context=context or {}
        )
        
        self.call_history.append(tool_call)
        
        if tool_name not in self.tools:
            # Check if it's an MCP tool
            mcp_result = await self._try_mcp_call(tool_call)
            if mcp_result:
                return mcp_result
            
            tool_call.status = ToolStatus.FAILED
            tool_call.error = f"Tool '{tool_name}' not found"
            return tool_call
        
        metadata = self.tool_metadata.get(tool_name)
        max_retries = metadata.max_retries if metadata else 3
        timeout = metadata.timeout if metadata else 30.0
        
        for attempt in range(max_retries + 1):
            tool_call.retry_count = attempt
            tool_call.status = ToolStatus.RETRYING if attempt > 0 else ToolStatus.RUNNING
            
            try:
                # Validate parameters if schema available
                if metadata and metadata.parameters_schema:
                    self._validate_parameters(parameters, metadata.parameters_schema)
                
                # Execute with timeout
                result = await asyncio.wait_for(
                    self._execute_with_context(tool_name, parameters, tool_call.context),
                    timeout=timeout
                )
                
                tool_call.status = ToolStatus.SUCCESS
                tool_call.result = result
                tool_call.end_time = time.time()
                tool_call.execution_time = tool_call.end_time - tool_call.start_time
                
                self.logger.info(f"Tool call succeeded: {tool_name} (attempt {attempt + 1})")
                return tool_call
                
            except asyncio.TimeoutError:
                error_msg = f"Tool call timeout after {timeout}s"
                tool_call.error = error_msg
                tool_call.status = ToolStatus.TIMEOUT
                
                if attempt < max_retries:
                    self.logger.warning(f"Tool call timeout: {tool_name} (attempt {attempt + 1}), retrying...")
                    await asyncio.sleep(min(2 ** attempt, 10))  # Exponential backoff
                    continue
                else:
                    self.logger.error(f"Tool call failed after {max_retries + 1} attempts: {tool_name}")
                    break
                    
            except Exception as e:
                error_msg = f"Tool execution error: {str(e)}"
                tool_call.error = error_msg
                tool_call.status = ToolStatus.FAILED
                
                # Add error context for AI model
                tool_call.context["error_details"] = {
                    "error_type": type(e).__name__,
                    "error_message": str(e),
                    "traceback": traceback.format_exc(),
                    "parameters_used": parameters,
                    "attempt_number": attempt + 1
                }
                
                if attempt < max_retries:
                    # Determine if error is retryable
                    if self._is_retryable_error(e):
                        self.logger.warning(f"Retryable error in tool call: {tool_name} (attempt {attempt + 1}), retrying...")
                        await asyncio.sleep(min(2 ** attempt, 10))
                        continue
                    else:
                        self.logger.error(f"Non-retryable error in tool call: {tool_name}")
                        break
                else:
                    self.logger.error(f"Tool call failed after {max_retries + 1} attempts: {tool_name}")
                    break
        
        tool_call.end_time = time.time()
        tool_call.execution_time = tool_call.end_time - tool_call.start_time
        return tool_call
    
    async def _execute_with_context(self, 
                                   tool_name: str, 
                                   parameters: Dict[str, Any],
                                   context: Dict[str, Any]) -> Any:
        """Execute tool with context injection"""
        func = self.tools[tool_name]
        
        # Inject context if function accepts it
        import inspect
        sig = inspect.signature(func)
        if 'context' in sig.parameters:
            parameters['context'] = context
        
        if asyncio.iscoroutinefunction(func):
            return await func(**parameters)
        else:
            # Run sync function in thread pool
            loop = asyncio.get_event_loop()
            return await loop.run_in_executor(None, lambda: func(**parameters))
    
    async def _try_mcp_call(self, tool_call: ToolCall) -> Optional[ToolCall]:
        """Try to execute tool call via MCP protocol"""
        # This would implement actual MCP protocol communication
        # For now, just a placeholder
        
        for server_name, connection in self.mcp_connections.items():
            # Check if server has this tool
            if await self._mcp_has_tool(server_name, tool_call.tool_name):
                try:
                    result = await self._mcp_execute_tool(
                        server_name, 
                        tool_call.tool_name, 
                        tool_call.parameters
                    )
                    
                    tool_call.status = ToolStatus.SUCCESS
                    tool_call.result = result
                    tool_call.end_time = time.time()
                    tool_call.execution_time = tool_call.end_time - tool_call.start_time
                    
                    return tool_call
                    
                except Exception as e:
                    self.logger.warning(f"MCP tool call failed on {server_name}: {e}")
                    continue
        
        return None
    
    async def _mcp_has_tool(self, server_name: str, tool_name: str) -> bool:
        """Check if MCP server has tool"""
        # Implement MCP tool discovery
        return False
    
    async def _mcp_execute_tool(self, 
                               server_name: str,
                               tool_name: str, 
                               parameters: Dict[str, Any]) -> Any:
        """Execute tool via MCP"""
        # Implement MCP tool execution
        raise NotImplementedError("MCP integration pending")
    
    def _validate_parameters(self, 
                           parameters: Dict[str, Any], 
                           schema: Dict[str, Any]) -> None:
        """Validate parameters against JSON schema"""
        # Simple validation - could use jsonschema library
        required = schema.get('required', [])
        for field in required:
            if field not in parameters:
                raise ValueError(f"Required parameter '{field}' missing")
    
    def _is_retryable_error(self, error: Exception) -> bool:
        """Determine if error is retryable"""
        retryable_errors = (
            ConnectionError,
            TimeoutError,
            OSError,
        )
        
        # Network-related errors
        error_str = str(error).lower()
        retryable_keywords = [
            'connection', 'timeout', 'network', 'temporary', 
            'rate limit', 'service unavailable', '503', '502', '429'
        ]
        
        return (isinstance(error, retryable_errors) or 
                any(keyword in error_str for keyword in retryable_keywords))
    
    def get_call_history(self, 
                        tool_name: Optional[str] = None,
                        status: Optional[ToolStatus] = None,
                        limit: int = 100) -> List[ToolCall]:
        """Get filtered call history"""
        filtered = self.call_history
        
        if tool_name:
            filtered = [call for call in filtered if call.tool_name == tool_name]
        
        if status:
            filtered = [call for call in filtered if call.status == status]
        
        return filtered[-limit:] if limit else filtered
    
    def get_tool_stats(self) -> Dict[str, Any]:
        """Get tool execution statistics"""
        stats = {
            'total_calls': len(self.call_history),
            'success_rate': 0.0,
            'average_execution_time': 0.0,
            'tool_breakdown': {},
            'error_breakdown': {}
        }
        
        if not self.call_history:
            return stats
        
        successful = [call for call in self.call_history if call.status == ToolStatus.SUCCESS]
        stats['success_rate'] = len(successful) / len(self.call_history) * 100
        
        if successful:
            total_time = sum(call.execution_time or 0 for call in successful)
            stats['average_execution_time'] = total_time / len(successful)
        
        # Tool breakdown
        for call in self.call_history:
            tool = call.tool_name
            if tool not in stats['tool_breakdown']:
                stats['tool_breakdown'][tool] = {'calls': 0, 'successes': 0, 'failures': 0}
            
            stats['tool_breakdown'][tool]['calls'] += 1
            if call.status == ToolStatus.SUCCESS:
                stats['tool_breakdown'][tool]['successes'] += 1
            else:
                stats['tool_breakdown'][tool]['failures'] += 1
        
        # Error breakdown
        failed_calls = [call for call in self.call_history 
                       if call.status in [ToolStatus.FAILED, ToolStatus.TIMEOUT]]
        
        for call in failed_calls:
            if call.error:
                error_type = call.error.split(':')[0] if ':' in call.error else call.error
                stats['error_breakdown'][error_type] = stats['error_breakdown'].get(error_type, 0) + 1
        
        return stats
    
    async def create_langchain_react_agent(self, llm_model: str):
        """Create LangChain ReAct agent for non-function-calling models"""
        try:
            from langchain.agents import create_react_agent, AgentExecutor
            from langchain.tools import Tool
            from langchain_core.prompts import PromptTemplate
            
            # Convert registered tools to LangChain tools
            lc_tools = []
            for name, metadata in self.tool_metadata.items():
                
                async def tool_wrapper(query: str, tool_name=name):
                    # Parse query as JSON if possible, otherwise use as-is
                    try:
                        params = json.loads(query)
                    except:
                        params = {"input": query}
                    
                    call_id = f"lc_{int(time.time())}"
                    result = await self.execute_tool_call(call_id, tool_name, params)
                    
                    if result.status == ToolStatus.SUCCESS:
                        return json.dumps(result.result) if isinstance(result.result, dict) else str(result.result)
                    else:
                        return f"Error: {result.error}"
                
                lc_tool = Tool(
                    name=name,
                    description=metadata.description,
                    func=tool_wrapper
                )
                lc_tools.append(lc_tool)
            
            # Create ReAct agent
            prompt = PromptTemplate.from_template("""
Answer the following questions as best you can. You have access to the following tools:

{tools}

Use the following format:

Question: the input question you must answer
Thought: you should always think about what to do
Action: the action to take, should be one of [{tool_names}]
Action Input: the input to the action
Observation: the result of the action
... (this Thought/Action/Action Input/Observation can repeat N times)
Thought: I now know the final answer
Final Answer: the final answer to the original input question

Begin!

Question: {input}
Thought: {agent_scratchpad}
""")
            
            # Get LLM instance (would need to be implemented based on your LLM setup)
            # llm = self._get_llm_instance(llm_model)
            
            # agent = create_react_agent(llm, lc_tools, prompt)
            # agent_executor = AgentExecutor(agent=agent, tools=lc_tools, verbose=True)
            
            self.logger.info(f"Created LangChain ReAct agent with {len(lc_tools)} tools")
            # return agent_executor
            
        except ImportError:
            self.logger.warning("LangChain not available for ReAct agent creation")
            return None