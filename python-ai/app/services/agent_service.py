"""
Agent Service - 智能体服务
集成先进的Agent核心系统，提供统一的智能体服务接口
"""

from typing import Dict, List, Any, Optional, AsyncIterator
import asyncio
import uuid
from datetime import datetime

from app.core.logging import LoggerMixin
from app.core.config import Settings
from app.core.exceptions import AgentException
from app.adapters.base import BaseAIAdapter
from app.adapters.openai_adapter import OpenAIAdapter
from app.adapters.claude_adapter import ClaudeAdapter
from app.services.tool_service import ToolService
from app.services.rag_service import RAGService
from app.agents.core import IntelligentAgent, AgentRequest, AgentResponse, AgentMode
from app.agents.memory import AdvancedMemorySystem


class AgentService(LoggerMixin):
    """智能体服务管理器"""
    
    def __init__(self, settings: Settings):
        super().__init__()
        self.settings = settings
        
        # 核心组件
        self.ai_adapters: Dict[str, BaseAIAdapter] = {}
        self.tool_service: Optional[ToolService] = None
        self.rag_service: Optional[RAGService] = None
        self.memory_system: Optional[AdvancedMemorySystem] = None
        
        # 智能体实例
        self.agents: Dict[str, IntelligentAgent] = {}
        self.default_agent: Optional[IntelligentAgent] = None
        
        # 服务状态
        self.initialized = False
    
    async def startup(self):
        """启动Agent服务"""
        
        if self.initialized:
            return
        
        self.logger.info("Initializing Agent Service...")
        
        try:
            # 初始化AI适配器
            await self._initialize_ai_adapters()
            
            # 初始化工具服务
            await self._initialize_tool_service()
            
            # 初始化RAG服务
            await self._initialize_rag_service()
            
            # 初始化记忆系统
            await self._initialize_memory_system()
            
            # 创建默认智能体
            await self._create_default_agent()
            
            self.initialized = True
            self.logger.info("Agent Service initialized successfully")
            
        except Exception as e:
            self.logger.error(f"Failed to initialize Agent Service: {e}")
            raise AgentException(f"Agent service initialization failed: {str(e)}")
    
    async def shutdown(self):
        """关闭Agent服务"""
        
        self.logger.info("Shutting down Agent Service...")
        
        try:
            # 清理智能体
            for agent in self.agents.values():
                if hasattr(agent, 'cleanup_inactive_sessions'):
                    await agent.cleanup_inactive_sessions(max_idle_hours=0)  # 清理所有会话
            
            # 清理记忆系统
            if self.memory_system:
                await self.memory_system.cleanup_expired_memories(max_age_days=0)
            
            # 关闭RAG服务
            if self.rag_service:
                await self.rag_service.shutdown()
            
            self.agents.clear()
            self.default_agent = None
            self.initialized = False
            
            self.logger.info("Agent Service shutdown completed")
            
        except Exception as e:
            self.logger.error(f"Error during Agent Service shutdown: {e}")
    
    async def _initialize_ai_adapters(self):
        """初始化AI适配器"""
        
        # OpenAI适配器
        if hasattr(self.settings, 'openai_api_key') and self.settings.openai_api_key:
            openai_adapter = OpenAIAdapter(
                api_key=self.settings.openai_api_key,
                base_url=getattr(self.settings, 'openai_base_url', None),
                model=getattr(self.settings, 'openai_model', 'gpt-3.5-turbo')
            )
            self.ai_adapters['openai'] = openai_adapter
            self.logger.info("OpenAI adapter initialized")
        
        # Claude适配器
        if hasattr(self.settings, 'anthropic_api_key') and self.settings.anthropic_api_key:
            claude_adapter = ClaudeAdapter(
                api_key=self.settings.anthropic_api_key,
                model=getattr(self.settings, 'anthropic_model', 'claude-3-sonnet-20240229')
            )
            self.ai_adapters['claude'] = claude_adapter
            self.logger.info("Claude adapter initialized")
        
        if not self.ai_adapters:
            raise AgentException("No AI adapters initialized. Please check your API keys.")
    
    async def _initialize_tool_service(self):
        """初始化工具服务"""
        
        try:
            self.tool_service = ToolService(self.settings)
            await self.tool_service.startup()
            self.logger.info("Tool service initialized")
        except Exception as e:
            self.logger.warning(f"Tool service initialization failed: {e}")
            self.tool_service = None
    
    async def _initialize_rag_service(self):
        """初始化RAG服务"""
        
        try:
            self.rag_service = RAGService(self.settings)
            await self.rag_service.startup()
            self.logger.info("RAG service initialized")
        except Exception as e:
            self.logger.warning(f"RAG service initialization failed: {e}")
            self.rag_service = None
    
    async def _initialize_memory_system(self):
        """初始化记忆系统"""
        
        try:
            self.memory_system = AdvancedMemorySystem(
                max_working_memory=getattr(self.settings, 'max_working_memory', 20),
                max_episodic_memory=getattr(self.settings, 'max_episodic_memory', 1000),
                max_semantic_memory=getattr(self.settings, 'max_semantic_memory', 5000)
            )
            self.logger.info("Memory system initialized")
        except Exception as e:
            self.logger.warning(f"Memory system initialization failed: {e}")
            self.memory_system = None
    
    async def _create_default_agent(self):
        """创建默认智能体"""
        
        # 选择主要的AI适配器
        primary_adapter = None
        if 'openai' in self.ai_adapters:
            primary_adapter = self.ai_adapters['openai']
        elif 'claude' in self.ai_adapters:
            primary_adapter = self.ai_adapters['claude']
        else:
            raise AgentException("No AI adapter available for default agent")
        
        # 创建默认智能体
        self.default_agent = IntelligentAgent(
            ai_adapter=primary_adapter,
            tool_service=self.tool_service,
            rag_service=self.rag_service,
            memory_system=self.memory_system,
            agent_id="default_agent"
        )
        
        self.agents["default"] = self.default_agent
        self.logger.info("Default agent created successfully")
    
    async def chat(
        self,
        message: str,
        user_id: str,
        session_id: Optional[str] = None,
        agent_id: Optional[str] = None,
        mode: Optional[str] = None,
        stream: bool = False,
        temperature: float = 0.7,
        max_tokens: int = 2000,
        tools_enabled: bool = True,
        rag_enabled: bool = True,
        **kwargs
    ) -> Dict[str, Any]:
        """对话接口"""
        
        if not self.initialized:
            raise AgentException("Agent service not initialized")
        
        # 选择智能体
        agent = self.agents.get(agent_id or "default", self.default_agent)
        if not agent:
            raise AgentException(f"Agent {agent_id} not found")
        
        # 生成会话ID
        if not session_id:
            session_id = str(uuid.uuid4())
        
        # 解析模式
        agent_mode = AgentMode.AUTO
        if mode:
            try:
                agent_mode = AgentMode(mode.lower())
            except ValueError:
                self.logger.warning(f"Unknown agent mode: {mode}, using AUTO")
        
        # 构建请求
        request = AgentRequest(
            message=message,
            user_id=user_id,
            session_id=session_id,
            mode=agent_mode,
            tools_enabled=tools_enabled,
            rag_enabled=rag_enabled,
            temperature=temperature,
            max_tokens=max_tokens,
            stream=stream,
            context=kwargs
        )
        
        try:
            # 处理请求
            response = await agent.process_request(request)
            
            # 转换为API响应格式
            return self._format_chat_response(response, session_id)
            
        except Exception as e:
            self.logger.error(f"Chat processing failed: {e}")
            raise AgentException(f"Chat failed: {str(e)}")
    
    async def chat_stream(
        self,
        message: str,
        user_id: str,
        session_id: Optional[str] = None,
        agent_id: Optional[str] = None,
        **kwargs
    ) -> AsyncIterator[Dict[str, Any]]:
        """流式对话接口"""
        
        # 设置流式模式
        kwargs['stream'] = True
        
        try:
            response = await self.chat(
                message=message,
                user_id=user_id,
                session_id=session_id,
                agent_id=agent_id,
                **kwargs
            )
            
            # 模拟流式输出
            content = response.get('content', '')
            words = content.split()
            
            for i, word in enumerate(words):
                chunk = {
                    "delta": {"content": word + " "},
                    "session_id": session_id,
                    "finish_reason": "length" if i == len(words) - 1 else None
                }
                yield chunk
                
                # 添加小延迟模拟实时效果
                await asyncio.sleep(0.05)
            
        except Exception as e:
            error_chunk = {
                "error": str(e),
                "session_id": session_id,
                "finish_reason": "error"
            }
            yield error_chunk
    
    def _format_chat_response(
        self,
        response: AgentResponse,
        session_id: str
    ) -> Dict[str, Any]:
        """格式化对话响应"""
        
        formatted_response = {
            "content": response.content,
            "session_id": session_id,
            "agent_mode": response.mode_used.value,
            "response_type": response.response_type.value,
            "processing_time": response.processing_time,
            "confidence_score": response.confidence_score,
            "metadata": response.metadata
        }
        
        # 添加推理链信息
        if response.reasoning_chain:
            formatted_response["reasoning"] = {
                "chain_id": response.reasoning_chain.chain_id,
                "mode": response.reasoning_chain.mode.value,
                "steps_count": len(response.reasoning_chain.steps),
                "success": response.reasoning_chain.success,
                "reasoning_time": response.reasoning_chain.total_time
            }
        
        # 添加工具使用信息
        if response.tools_used:
            formatted_response["tools_used"] = response.tools_used
        
        # 添加RAG结果信息
        if response.rag_results:
            formatted_response["rag_sources"] = [
                {
                    "source": doc.get("source", "unknown"),
                    "score": doc.get("score", 0.0),
                    "chunk_type": doc.get("chunk_type", "text")
                }
                for doc in response.rag_results[:3]  # 只返回前3个源
            ]
        
        return formatted_response
    
    async def create_agent(
        self,
        agent_id: str,
        ai_provider: str = "openai",
        configuration: Optional[Dict[str, Any]] = None
    ) -> Dict[str, Any]:
        """创建新的智能体"""
        
        if agent_id in self.agents:
            raise AgentException(f"Agent {agent_id} already exists")
        
        # 选择AI适配器
        if ai_provider not in self.ai_adapters:
            raise AgentException(f"AI provider {ai_provider} not available")
        
        ai_adapter = self.ai_adapters[ai_provider]
        
        # 创建智能体
        agent = IntelligentAgent(
            ai_adapter=ai_adapter,
            tool_service=self.tool_service,
            rag_service=self.rag_service,
            memory_system=self.memory_system,
            agent_id=agent_id
        )
        
        self.agents[agent_id] = agent
        
        self.logger.info(f"Created agent {agent_id} with provider {ai_provider}")
        
        return {
            "agent_id": agent_id,
            "ai_provider": ai_provider,
            "status": "created",
            "configuration": configuration or {}
        }
    
    async def delete_agent(self, agent_id: str) -> Dict[str, Any]:
        """删除智能体"""
        
        if agent_id == "default":
            raise AgentException("Cannot delete default agent")
        
        if agent_id not in self.agents:
            raise AgentException(f"Agent {agent_id} not found")
        
        # 清理会话
        agent = self.agents[agent_id]
        await agent.cleanup_inactive_sessions(max_idle_hours=0)
        
        # 删除智能体
        del self.agents[agent_id]
        
        self.logger.info(f"Deleted agent {agent_id}")
        
        return {
            "agent_id": agent_id,
            "status": "deleted"
        }
    
    async def list_agents(self) -> Dict[str, Any]:
        """列出所有智能体"""
        
        agents_info = {}
        
        for agent_id, agent in self.agents.items():
            status = await agent.get_agent_status()
            agents_info[agent_id] = {
                "agent_id": agent_id,
                "active_sessions": status["active_sessions"],
                "total_requests": status["performance"]["total_requests"],
                "success_rate": (
                    status["performance"]["successful_requests"] / 
                    max(status["performance"]["total_requests"], 1) * 100
                ),
                "average_response_time": status["performance"]["average_response_time"],
                "components": status["components"]
            }
        
        return {
            "agents": agents_info,
            "total_agents": len(self.agents),
            "default_agent": "default"
        }
    
    async def get_agent_status(self, agent_id: Optional[str] = None) -> Dict[str, Any]:
        """获取智能体状态"""
        
        target_agent_id = agent_id or "default"
        
        if target_agent_id not in self.agents:
            raise AgentException(f"Agent {target_agent_id} not found")
        
        agent = self.agents[target_agent_id]
        return await agent.get_agent_status()
    
    async def get_service_status(self) -> Dict[str, Any]:
        """获取服务状态"""
        
        return {
            "initialized": self.initialized,
            "ai_adapters": list(self.ai_adapters.keys()),
            "components": {
                "tool_service": self.tool_service is not None,
                "rag_service": self.rag_service is not None,
                "memory_system": self.memory_system is not None
            },
            "agents": {
                "total": len(self.agents),
                "active": len([a for a in self.agents.values() if a.active_sessions])
            },
            "memory_stats": (
                self.memory_system.get_memory_statistics() 
                if self.memory_system else {}
            )
        }
    
    async def upload_knowledge(
        self,
        user_id: str,
        file_name: str,
        content: str,
        content_type: str = "text",
        metadata: Optional[Dict[str, Any]] = None
    ) -> Dict[str, Any]:
        """上传知识到RAG系统"""
        
        if not self.rag_service:
            raise AgentException("RAG service not available")
        
        try:
            result = await self.rag_service.upload_document(
                user_id=user_id,
                filename=file_name,
                content=content,
                doc_type=content_type,
                metadata=metadata
            )
            
            return result
            
        except Exception as e:
            self.logger.error(f"Knowledge upload failed: {e}")
            raise AgentException(f"Knowledge upload failed: {str(e)}")
    
    async def search_knowledge(
        self,
        user_id: str,
        query: str,
        top_k: int = 5,
        session_id: Optional[str] = None
    ) -> Dict[str, Any]:
        """搜索知识库"""
        
        if not self.rag_service:
            raise AgentException("RAG service not available")
        
        try:
            result = await self.rag_service.query(
                user_id=user_id,
                query=query,
                top_k=top_k,
                session_id=session_id
            )
            
            return result
            
        except Exception as e:
            self.logger.error(f"Knowledge search failed: {e}")
            raise AgentException(f"Knowledge search failed: {str(e)}")