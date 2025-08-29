"""
Oxy Core - Modular component system based on OxyGent design
"""

import asyncio
import uuid
import time
from typing import Dict, List, Any, Optional, Union, Callable, Type
from dataclasses import dataclass, field, asdict
from enum import Enum
from abc import ABC, abstractmethod
import json
import inspect

from ..core.logging import LoggerMixin
from ..core.base_exceptions import AgentException

class OxyType(Enum):
    """Oxy component type enum"""
    AGENT = "agent"
    TOOL = "tool"
    LLM = "llm"
    FUNCTION = "function"
    MEMORY = "memory"
    PLANNER = "planner"
    EVALUATOR = "evaluator"
    FILTER = "filter"
    TRANSFORMER = "transformer"
    ROUTER = "router"

class OxyStatus(Enum):
    """Oxy component status enum"""
    IDLE = "idle"
    RUNNING = "running"
    COMPLETED = "completed"
    ERROR = "error"
    PAUSED = "paused"

@dataclass
class OxyMessage:
    """Oxy message data structure"""
    id: str = field(default_factory=lambda: str(uuid.uuid4()))
    type: str = "text"
    content: Any = None
    metadata: Dict[str, Any] = field(default_factory=dict)
    timestamp: float = field(default_factory=time.time)
    sender_id: Optional[str] = None
    recipient_id: Optional[str] = None
    thread_id: Optional[str] = None

@dataclass
class OxyContext:
    """Oxy execution context"""
    session_id: str
    user_id: str
    thread_id: Optional[str] = None
    variables: Dict[str, Any] = field(default_factory=dict)
    history: List[OxyMessage] = field(default_factory=list)
    metadata: Dict[str, Any] = field(default_factory=dict)
    
    def add_message(self, message: OxyMessage):
        """Add message to history"""
        message.thread_id = self.thread_id
        self.history.append(message)
    
    def get_variable(self, key: str, default: Any = None) -> Any:
        """Get context variable"""
        return self.variables.get(key, default)
    
    def set_variable(self, key: str, value: Any):
        """Set context variable"""
        self.variables[key] = value

@dataclass
class OxyResult:
    """Oxy execution result"""
    success: bool
    data: Any = None
    message: str = ""
    metadata: Dict[str, Any] = field(default_factory=dict)
    execution_time: float = 0.0
    error: Optional[Exception] = None
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary"""
        result = asdict(self)
        if self.error:
            result["error"] = str(self.error)
        return result

class BaseOxy(ABC, LoggerMixin):
    """Base Oxy component abstract class"""
    
    def __init__(
        self,
        oxy_id: str = None,
        name: str = None,
        description: str = "",
        oxy_type: OxyType = OxyType.FUNCTION,
        config: Dict[str, Any] = None
    ):
        super().__init__()
        
        self.oxy_id = oxy_id or f"{oxy_type.value}_{uuid.uuid4().hex[:8]}"
        self.name = name or self.__class__.__name__
        self.description = description
        self.oxy_type = oxy_type
        self.config = config or {}
        
        self.status = OxyStatus.IDLE
        self.created_at = time.time()
        self.last_executed = None
        self.execution_count = 0
        self.total_execution_time = 0.0
        
        # Dependencies
        self.dependencies: List[str] = []
        self.dependents: List[str] = []
        
        # Input/output schema
        self.input_schema: Dict[str, Any] = {}
        self.output_schema: Dict[str, Any] = {}
    
    @abstractmethod
    async def execute(self, context: OxyContext, **kwargs) -> OxyResult:
        """Execute Oxy component"""
        pass
    
    async def validate_input(self, **kwargs) -> bool:
        """Validate input parameters"""
        # Basic validation, can be overridden by subclasses
        return True
    
    async def pre_execute(self, context: OxyContext, **kwargs):
        """Pre-execution hook"""
        self.status = OxyStatus.RUNNING
        self.execution_count += 1
    
    async def post_execute(self, context: OxyContext, result: OxyResult):
        """Post-execution hook"""
        self.status = OxyStatus.COMPLETED if result.success else OxyStatus.ERROR
        self.last_executed = time.time()
        self.total_execution_time += result.execution_time
    
    def add_dependency(self, oxy_id: str):
        """Add dependency"""
        if oxy_id not in self.dependencies:
            self.dependencies.append(oxy_id)
    
    def remove_dependency(self, oxy_id: str):
        """Remove dependency"""
        if oxy_id in self.dependencies:
            self.dependencies.remove(oxy_id)
    
    def get_info(self) -> Dict[str, Any]:
        """Get component information"""
        return {
            "oxy_id": self.oxy_id,
            "name": self.name,
            "description": self.description,
            "type": self.oxy_type.value,
            "status": self.status.value,
            "config": self.config,
            "created_at": self.created_at,
            "last_executed": self.last_executed,
            "execution_count": self.execution_count,
            "total_execution_time": self.total_execution_time,
            "dependencies": self.dependencies,
            "dependents": self.dependents,
            "input_schema": self.input_schema,
            "output_schema": self.output_schema,
        }
    
    def __repr__(self) -> str:
        return f"<{self.__class__.__name__}(id={self.oxy_id}, name={self.name})>"

class OxyLLM(BaseOxy):
    """LLM Oxy component"""
    
    def __init__(
        self,
        model: str,
        adapter,
        **kwargs
    ):
        super().__init__(oxy_type=OxyType.LLM, **kwargs)
        self.model = model
        self.adapter = adapter
        
        self.input_schema = {
            "messages": {"type": "array", "required": True},
            "temperature": {"type": "number", "default": 0.7},
            "max_tokens": {"type": "integer", "default": 2000},
            "stream": {"type": "boolean", "default": False}
        }
        
        self.output_schema = {
            "content": {"type": "string"},
            "usage": {"type": "object"},
            "model": {"type": "string"}
        }
    
    async def execute(self, context: OxyContext, **kwargs) -> OxyResult:
        """执行LLM调用"""
        start_time = time.time()
        
        try:
            await self.pre_execute(context, **kwargs)
            
            # 验证输入
            if not await self.validate_input(**kwargs):
                return OxyResult(
                    success=False,
                    message="Invalid input parameters",
                    execution_time=time.time() - start_time
                )
            
            # 调用AI模型
            response = await self.adapter.generate(**kwargs)
            
            # 创建结果消息
            result_msg = OxyMessage(
                type="llm_response",
                content=response.content,
                metadata={
                    "model": response.model,
                    "usage": response.usage,
                    "cost": response.cost_estimate
                },
                sender_id=self.oxy_id
            )
            context.add_message(result_msg)
            
            result = OxyResult(
                success=True,
                data=response,
                message="LLM execution completed",
                execution_time=time.time() - start_time
            )
            
            await self.post_execute(context, result)
            return result
            
        except Exception as e:
            self.logger.error(f"LLM execution failed: {e}")
            result = OxyResult(
                success=False,
                message=f"LLM execution error: {str(e)}",
                error=e,
                execution_time=time.time() - start_time
            )
            await self.post_execute(context, result)
            return result

class OxyTool(BaseOxy):
    """工具Oxy组件"""
    
    def __init__(
        self,
        tool_func: Callable,
        **kwargs
    ):
        super().__init__(oxy_type=OxyType.TOOL, **kwargs)
        self.tool_func = tool_func
        self.signature = inspect.signature(tool_func)
        
        # 自动生成输入schema
        self.input_schema = self._generate_input_schema()
    
    def _generate_input_schema(self) -> Dict[str, Any]:
        """自动生成输入schema"""
        schema = {}
        
        for param_name, param in self.signature.parameters.items():
            param_info = {"type": "any"}
            
            if param.annotation != inspect.Parameter.empty:
                if param.annotation == str:
                    param_info["type"] = "string"
                elif param.annotation == int:
                    param_info["type"] = "integer"
                elif param.annotation == float:
                    param_info["type"] = "number"
                elif param.annotation == bool:
                    param_info["type"] = "boolean"
                elif param.annotation == list:
                    param_info["type"] = "array"
                elif param.annotation == dict:
                    param_info["type"] = "object"
            
            if param.default != inspect.Parameter.empty:
                param_info["default"] = param.default
            else:
                param_info["required"] = True
            
            schema[param_name] = param_info
        
        return schema
    
    async def execute(self, context: OxyContext, **kwargs) -> OxyResult:
        """执行工具函数"""
        start_time = time.time()
        
        try:
            await self.pre_execute(context, **kwargs)
            
            # 验证输入参数
            if not await self.validate_input(**kwargs):
                return OxyResult(
                    success=False,
                    message="Invalid input parameters",
                    execution_time=time.time() - start_time
                )
            
            # 过滤参数，只传递函数需要的参数
            func_params = {}
            for param_name in self.signature.parameters.keys():
                if param_name in kwargs:
                    func_params[param_name] = kwargs[param_name]
            
            # 执行工具函数
            if asyncio.iscoroutinefunction(self.tool_func):
                tool_result = await self.tool_func(**func_params)
            else:
                tool_result = self.tool_func(**func_params)
            
            # 创建结果消息
            result_msg = OxyMessage(
                type="tool_result",
                content=tool_result,
                metadata={"tool_name": self.name},
                sender_id=self.oxy_id
            )
            context.add_message(result_msg)
            
            result = OxyResult(
                success=True,
                data=tool_result,
                message="Tool execution completed",
                execution_time=time.time() - start_time
            )
            
            await self.post_execute(context, result)
            return result
            
        except Exception as e:
            self.logger.error(f"Tool execution failed: {e}")
            result = OxyResult(
                success=False,
                message=f"Tool execution error: {str(e)}",
                error=e,
                execution_time=time.time() - start_time
            )
            await self.post_execute(context, result)
            return result

class OxyFunction(BaseOxy):
    """函数Oxy组件"""
    
    def __init__(
        self,
        func: Callable,
        **kwargs
    ):
        super().__init__(oxy_type=OxyType.FUNCTION, **kwargs)
        self.func = func
        self.signature = inspect.signature(func)
        self.input_schema = self._generate_input_schema()
    
    def _generate_input_schema(self) -> Dict[str, Any]:
        """自动生成输入schema"""
        schema = {}
        
        for param_name, param in self.signature.parameters.items():
            if param_name in ['context']:  # 跳过特殊参数
                continue
                
            param_info = {"type": "any"}
            
            if param.annotation != inspect.Parameter.empty:
                if param.annotation == str:
                    param_info["type"] = "string"
                elif param.annotation == int:
                    param_info["type"] = "integer"
                elif param.annotation == float:
                    param_info["type"] = "number"
                elif param.annotation == bool:
                    param_info["type"] = "boolean"
            
            if param.default != inspect.Parameter.empty:
                param_info["default"] = param.default
            else:
                param_info["required"] = True
            
            schema[param_name] = param_info
        
        return schema
    
    async def execute(self, context: OxyContext, **kwargs) -> OxyResult:
        """执行函数"""
        start_time = time.time()
        
        try:
            await self.pre_execute(context, **kwargs)
            
            # 准备函数参数
            func_params = {}
            
            # 如果函数需要context参数，传入context
            if 'context' in self.signature.parameters:
                func_params['context'] = context
            
            # 添加其他参数
            for param_name in self.signature.parameters.keys():
                if param_name != 'context' and param_name in kwargs:
                    func_params[param_name] = kwargs[param_name]
            
            # 执行函数
            if asyncio.iscoroutinefunction(self.func):
                func_result = await self.func(**func_params)
            else:
                func_result = self.func(**func_params)
            
            result = OxyResult(
                success=True,
                data=func_result,
                message="Function execution completed",
                execution_time=time.time() - start_time
            )
            
            await self.post_execute(context, result)
            return result
            
        except Exception as e:
            self.logger.error(f"Function execution failed: {e}")
            result = OxyResult(
                success=False,
                message=f"Function execution error: {str(e)}",
                error=e,
                execution_time=time.time() - start_time
            )
            await self.post_execute(context, result)
            return result

class OxyRouter(BaseOxy):
    """路由Oxy组件"""
    
    def __init__(
        self,
        routes: Dict[str, str],  # 条件 -> 目标Oxy ID
        **kwargs
    ):
        super().__init__(oxy_type=OxyType.ROUTER, **kwargs)
        self.routes = routes
        
        self.input_schema = {
            "input": {"type": "any", "required": True},
            "route_key": {"type": "string", "required": False}
        }
    
    async def execute(self, context: OxyContext, **kwargs) -> OxyResult:
        """执行路由逻辑"""
        start_time = time.time()
        
        try:
            await self.pre_execute(context, **kwargs)
            
            route_key = kwargs.get('route_key')
            input_data = kwargs.get('input')
            
            # 路由逻辑
            target_oxy_id = None
            
            if route_key and route_key in self.routes:
                target_oxy_id = self.routes[route_key]
            else:
                # 默认路由逻辑（可以根据input_data内容决定）
                # 这里简化处理，取第一个路由
                if self.routes:
                    target_oxy_id = list(self.routes.values())[0]
            
            if not target_oxy_id:
                return OxyResult(
                    success=False,
                    message="No route found",
                    execution_time=time.time() - start_time
                )
            
            result = OxyResult(
                success=True,
                data={"target_oxy_id": target_oxy_id, "input": input_data},
                message="Route determined",
                execution_time=time.time() - start_time
            )
            
            await self.post_execute(context, result)
            return result
            
        except Exception as e:
            self.logger.error(f"Router execution failed: {e}")
            result = OxyResult(
                success=False,
                message=f"Router execution error: {str(e)}",
                error=e,
                execution_time=time.time() - start_time
            )
            await self.post_execute(context, result)
            return result

class OxyRegistry:
    """Oxy组件注册中心"""
    
    def __init__(self):
        self.components: Dict[str, BaseOxy] = {}
        self.dependencies_graph: Dict[str, List[str]] = {}
    
    def register(self, oxy: BaseOxy) -> str:
        """注册Oxy组件"""
        self.components[oxy.oxy_id] = oxy
        self.dependencies_graph[oxy.oxy_id] = oxy.dependencies.copy()
        
        # 更新依赖组件的dependents
        for dep_id in oxy.dependencies:
            if dep_id in self.components:
                if oxy.oxy_id not in self.components[dep_id].dependents:
                    self.components[dep_id].dependents.append(oxy.oxy_id)
        
        return oxy.oxy_id
    
    def unregister(self, oxy_id: str) -> bool:
        """注销Oxy组件"""
        if oxy_id not in self.components:
            return False
        
        oxy = self.components[oxy_id]
        
        # 清理依赖关系
        for dep_id in oxy.dependencies:
            if dep_id in self.components:
                if oxy_id in self.components[dep_id].dependents:
                    self.components[dep_id].dependents.remove(oxy_id)
        
        # 清理依赖此组件的其他组件
        for dependent_id in oxy.dependents:
            if dependent_id in self.components:
                if oxy_id in self.components[dependent_id].dependencies:
                    self.components[dependent_id].dependencies.remove(oxy_id)
        
        del self.components[oxy_id]
        del self.dependencies_graph[oxy_id]
        return True
    
    def get(self, oxy_id: str) -> Optional[BaseOxy]:
        """获取Oxy组件"""
        return self.components.get(oxy_id)
    
    def list_components(self, oxy_type: OxyType = None) -> List[BaseOxy]:
        """列出组件"""
        components = list(self.components.values())
        
        if oxy_type:
            components = [c for c in components if c.oxy_type == oxy_type]
        
        return components
    
    def get_execution_order(self, start_oxy_id: str) -> List[str]:
        """获取执行顺序（拓扑排序）"""
        visited = set()
        temp_visited = set()
        result = []
        
        def dfs(oxy_id: str):
            if oxy_id in temp_visited:
                raise Exception(f"Circular dependency detected involving {oxy_id}")
            
            if oxy_id in visited:
                return
            
            temp_visited.add(oxy_id)
            
            # 访问依赖
            for dep_id in self.dependencies_graph.get(oxy_id, []):
                if dep_id in self.components:
                    dfs(dep_id)
            
            temp_visited.remove(oxy_id)
            visited.add(oxy_id)
            result.append(oxy_id)
        
        dfs(start_oxy_id)
        return result
    
    def validate_dependencies(self) -> List[str]:
        """验证依赖关系"""
        errors = []
        
        for oxy_id, dependencies in self.dependencies_graph.items():
            for dep_id in dependencies:
                if dep_id not in self.components:
                    errors.append(f"Component {oxy_id} depends on non-existent component {dep_id}")
        
        return errors

# 全局注册中心实例
oxy_registry = OxyRegistry()