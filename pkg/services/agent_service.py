"""
Agent Service - 智能体服务核心实现

职责:
- 智能体生命周期管理
- 会话状态管理  
- 上下文维护和检索
- 多轮对话处理
"""

from abc import ABC, abstractmethod
from dataclasses import dataclass
from typing import Dict, List, Optional, Any, AsyncGenerator
from enum import Enum
import uuid
from datetime import datetime
import asyncio


class AgentType(Enum):
    """智能体类型"""
    CONVERSATIONAL = "conversational"
    TASK_ORIENTED = "task_oriented"  
    WORKFLOW_BASED = "workflow_based"
    TOOL_CALLING = "tool_calling"


class SessionStatus(Enum):
    """会话状态"""
    ACTIVE = "active"
    PAUSED = "paused"
    COMPLETED = "completed"
    EXPIRED = "expired"


@dataclass
class AgentConfig:
    """智能体配置"""
    agent_id: str
    agent_type: AgentType
    name: str
    description: str
    system_prompt: str
    max_context_length: int = 8000
    temperature: float = 0.7
    tools_enabled: bool = True
    memory_enabled: bool = True
    safety_filters: List[str] = None
    custom_instructions: Dict[str, Any] = None


@dataclass  
class SessionConfig:
    """会话配置"""
    session_id: str
    agent_id: str  
    user_id: str
    max_turns: int = 100
    context_window: int = 4000
    auto_expire_minutes: int = 30
    metadata: Dict[str, Any] = None


@dataclass
class Message:
    """消息结构"""
    message_id: str
    session_id: str
    role: str  # user, assistant, system, tool
    content: str
    timestamp: datetime
    metadata: Dict[str, Any] = None
    tool_calls: List[Dict] = None
    attachments: List[str] = None


@dataclass
class ProcessResult:
    """处理结果"""
    response: str
    model_used: str
    usage: Dict[str, int]
    cost: float
    tools_called: List[str]
    processing_time_ms: int
    session_updated: bool
    context_length: int


class ContextManager(ABC):
    """上下文管理器接口"""
    
    @abstractmethod
    async def get_context(self, session_id: str) -> Dict[str, Any]:
        """获取会话上下文"""
        pass
    
    @abstractmethod
    async def update_context(self, session_id: str, message: Message) -> bool:
        """更新上下文"""
        pass
    
    @abstractmethod
    async def clear_context(self, session_id: str) -> bool:
        """清除上下文"""
        pass


class SessionManager(ABC):
    """会话管理器接口"""
    
    @abstractmethod
    async def create_session(self, config: SessionConfig) -> str:
        """创建新会话"""
        pass
    
    @abstractmethod  
    async def get_session(self, session_id: str) -> Optional[SessionConfig]:
        """获取会话信息"""
        pass
    
    @abstractmethod
    async def update_session_status(self, session_id: str, status: SessionStatus) -> bool:
        """更新会话状态"""
        pass
    
    @abstractmethod
    async def cleanup_expired_sessions(self) -> int:
        """清理过期会话"""
        pass


class AgentService(ABC):
    """智能体服务主接口"""
    
    @abstractmethod
    async def create_agent(self, config: AgentConfig) -> str:
        """创建智能体"""
        pass
    
    @abstractmethod
    async def get_agent(self, agent_id: str) -> Optional[AgentConfig]:
        """获取智能体配置"""
        pass
    
    @abstractmethod
    async def update_agent(self, agent_id: str, config: AgentConfig) -> bool:
        """更新智能体配置"""
        pass
    
    @abstractmethod
    async def delete_agent(self, agent_id: str) -> bool:
        """删除智能体"""
        pass
    
    @abstractmethod
    async def create_session(self, agent_id: str, user_id: str) -> str:
        """创建会话"""
        pass
    
    @abstractmethod
    async def process_message(
        self, 
        session_id: str, 
        message: str, 
        user_id: str,
        stream: bool = False
    ) -> ProcessResult:
        """处理消息"""
        pass
    
    @abstractmethod
    async def stream_response(
        self, 
        session_id: str, 
        message: str, 
        user_id: str
    ) -> AsyncGenerator[str, None]:
        """流式响应"""
        pass
    
    @abstractmethod
    async def get_session_history(self, session_id: str) -> List[Message]:
        """获取会话历史"""
        pass


# 具体实现
class DefaultAgentService(AgentService):
    """默认智能体服务实现"""
    
    def __init__(
        self,
        context_manager: ContextManager,
        session_manager: SessionManager,
        model_router_client,
        tool_orchestrator_client,
        vector_store_client
    ):
        self.context_manager = context_manager
        self.session_manager = session_manager  
        self.model_router = model_router_client
        self.tool_orchestrator = tool_orchestrator_client
        self.vector_store = vector_store_client
        
        self.agents: Dict[str, AgentConfig] = {}
        self.active_sessions: Dict[str, SessionConfig] = {}
    
    async def create_agent(self, config: AgentConfig) -> str:
        """创建智能体"""
        agent_id = config.agent_id or str(uuid.uuid4())
        config.agent_id = agent_id
        
        # 验证配置
        self._validate_agent_config(config)
        
        # 存储配置
        self.agents[agent_id] = config
        
        # 初始化智能体相关资源
        await self._initialize_agent_resources(config)
        
        return agent_id
    
    async def get_agent(self, agent_id: str) -> Optional[AgentConfig]:
        """获取智能体配置"""
        return self.agents.get(agent_id)
    
    async def create_session(self, agent_id: str, user_id: str) -> str:
        """创建会话"""
        if agent_id not in self.agents:
            raise ValueError(f"Agent {agent_id} not found")
        
        session_id = str(uuid.uuid4())
        session_config = SessionConfig(
            session_id=session_id,
            agent_id=agent_id,
            user_id=user_id
        )
        
        # 创建会话
        await self.session_manager.create_session(session_config)
        self.active_sessions[session_id] = session_config
        
        # 初始化会话上下文
        await self._initialize_session_context(session_id, agent_id)
        
        return session_id
    
    async def process_message(
        self, 
        session_id: str, 
        message: str, 
        user_id: str,
        stream: bool = False
    ) -> ProcessResult:
        """处理消息"""
        
        # 验证会话
        session = await self._get_active_session(session_id)
        agent_config = self.agents[session.agent_id]
        
        # 创建消息对象
        user_message = Message(
            message_id=str(uuid.uuid4()),
            session_id=session_id,
            role="user",
            content=message,
            timestamp=datetime.now()
        )
        
        # 更新上下文
        await self.context_manager.update_context(session_id, user_message)
        
        try:
            # 获取当前上下文
            context = await self.context_manager.get_context(session_id)
            
            # 构建AI请求
            ai_request = await self._build_ai_request(
                agent_config, context, message
            )
            
            # 路由到合适的模型
            model_response = await self.model_router.route_and_call(ai_request)
            
            # 处理工具调用
            if model_response.tool_calls:
                tool_results = await self._handle_tool_calls(
                    session_id, model_response.tool_calls
                )
                model_response = await self._merge_tool_results(
                    model_response, tool_results
                )
            
            # 安全过滤
            filtered_response = await self._apply_safety_filters(
                agent_config, model_response.content
            )
            
            # 创建助手消息
            assistant_message = Message(
                message_id=str(uuid.uuid4()),
                session_id=session_id,
                role="assistant", 
                content=filtered_response,
                timestamp=datetime.now(),
                tool_calls=model_response.tool_calls
            )
            
            # 更新上下文
            await self.context_manager.update_context(session_id, assistant_message)
            
            # 返回结果
            return ProcessResult(
                response=filtered_response,
                model_used=model_response.model,
                usage=model_response.usage,
                cost=model_response.cost,
                tools_called=[tc.get('function', {}).get('name') for tc in (model_response.tool_calls or [])],
                processing_time_ms=model_response.processing_time,
                session_updated=True,
                context_length=len(context.get('messages', []))
            )
            
        except Exception as e:
            # 错误处理和恢复
            await self._handle_processing_error(session_id, str(e))
            raise
    
    async def stream_response(
        self, 
        session_id: str, 
        message: str, 
        user_id: str
    ) -> AsyncGenerator[str, None]:
        """流式响应实现"""
        
        session = await self._get_active_session(session_id)
        agent_config = self.agents[session.agent_id]
        
        # 构建流式请求
        context = await self.context_manager.get_context(session_id)
        ai_request = await self._build_ai_request(agent_config, context, message)
        ai_request.stream = True
        
        # 流式调用模型
        async for chunk in self.model_router.stream_call(ai_request):
            if chunk.content:
                yield chunk.content
    
    async def get_session_history(self, session_id: str) -> List[Message]:
        """获取会话历史"""
        context = await self.context_manager.get_context(session_id)
        return context.get('messages', [])
    
    # 私有辅助方法
    def _validate_agent_config(self, config: AgentConfig) -> None:
        """验证智能体配置"""
        if not config.name:
            raise ValueError("Agent name is required")
        if not config.system_prompt:
            raise ValueError("System prompt is required")
        if config.max_context_length <= 0:
            raise ValueError("Max context length must be positive")
    
    async def _initialize_agent_resources(self, config: AgentConfig) -> None:
        """初始化智能体资源"""
        # 初始化向量存储
        if config.memory_enabled:
            await self.vector_store.create_collection(
                f"agent_{config.agent_id}_memory"
            )
    
    async def _initialize_session_context(self, session_id: str, agent_id: str) -> None:
        """初始化会话上下文"""
        agent_config = self.agents[agent_id]
        
        initial_context = {
            'messages': [{
                'role': 'system',
                'content': agent_config.system_prompt
            }],
            'agent_config': agent_config.__dict__,
            'session_metadata': {}
        }
        
        await self.context_manager.update_context(session_id, None)  # 初始化空上下文
        
    async def _get_active_session(self, session_id: str) -> SessionConfig:
        """获取活跃会话"""
        session = self.active_sessions.get(session_id)
        if not session:
            session = await self.session_manager.get_session(session_id)
            if session:
                self.active_sessions[session_id] = session
            else:
                raise ValueError(f"Session {session_id} not found")
        return session
    
    async def _build_ai_request(self, agent_config: AgentConfig, context: Dict, message: str) -> Dict:
        """构建AI请求"""
        messages = context.get('messages', [])
        messages.append({'role': 'user', 'content': message})
        
        return {
            'messages': messages,
            'temperature': agent_config.temperature,
            'max_tokens': 2000,
            'tools_enabled': agent_config.tools_enabled,
            'agent_type': agent_config.agent_type.value
        }
    
    async def _handle_tool_calls(self, session_id: str, tool_calls: List[Dict]) -> List[Dict]:
        """处理工具调用"""
        results = []
        for tool_call in tool_calls:
            result = await self.tool_orchestrator.execute_tool(
                session_id, tool_call
            )
            results.append(result)
        return results
    
    async def _merge_tool_results(self, model_response, tool_results: List[Dict]):
        """合并工具执行结果"""
        # 实现工具结果合并逻辑
        return model_response
    
    async def _apply_safety_filters(self, agent_config: AgentConfig, content: str) -> str:
        """应用安全过滤器"""
        if not agent_config.safety_filters:
            return content
            
        # 实现安全过滤逻辑
        filtered_content = content
        
        for filter_name in agent_config.safety_filters:
            # 应用特定过滤器
            pass
            
        return filtered_content
    
    async def _handle_processing_error(self, session_id: str, error_msg: str) -> None:
        """处理处理错误"""
        # 记录错误
        # 更新会话状态
        # 发送错误通知
        pass


# 工厂函数
def create_agent_service(
    context_manager: ContextManager,
    session_manager: SessionManager, 
    model_router_client,
    tool_orchestrator_client,
    vector_store_client
) -> AgentService:
    """创建智能体服务实例"""
    return DefaultAgentService(
        context_manager,
        session_manager,
        model_router_client, 
        tool_orchestrator_client,
        vector_store_client
    )