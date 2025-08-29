"""
Agent Core System - 智能体核心系统
整合推理引擎、记忆管理、工具调用和任务规划
"""

from typing import List, Dict, Any, Optional, Union, AsyncIterator
import asyncio
import time
import json
from dataclasses import dataclass, field
from datetime import datetime
from enum import Enum

from app.core.logging import LoggerMixin
from app.core.exceptions import AgentException
from app.core.performance import monitor_performance, cache_result, performance_monitor
from app.adapters.base import BaseAIAdapter
from app.services.tool_service import ToolService
from app.services.rag_service import RAGService
from app.agents.reasoning import ReasoningEngine, ReasoningMode, ReasoningChain
from app.agents.memory import AdvancedMemorySystem, MemoryType, MemoryImportance


class AgentMode(Enum):
    """智能体模式"""
    CHAT = "chat"                    # 对话模式
    RAG = "rag"                      # RAG检索模式
    REASONING = "reasoning"          # 深度推理模式
    TOOL_USING = "tool_using"        # 工具使用模式
    PLANNING = "planning"            # 任务规划模式
    CREATIVE = "creative"            # 创意生成模式
    ANALYTICAL = "analytical"        # 分析模式
    AUTO = "auto"                    # 自动选择模式


class ResponseType(Enum):
    """响应类型"""
    TEXT = "text"
    STREAMING = "streaming"
    STRUCTURED = "structured"
    TOOL_CALL = "tool_call"
    REASONING_CHAIN = "reasoning_chain"


@dataclass
class AgentRequest:
    """智能体请求"""
    message: str
    user_id: str
    session_id: Optional[str] = None
    mode: AgentMode = AgentMode.AUTO
    context: Dict[str, Any] = field(default_factory=dict)
    tools_enabled: bool = True
    rag_enabled: bool = True
    reasoning_enabled: bool = True
    memory_enabled: bool = True
    temperature: float = 0.7
    max_tokens: int = 2000
    stream: bool = False


@dataclass
class AgentResponse:
    """智能体响应"""
    content: str
    response_type: ResponseType
    mode_used: AgentMode
    reasoning_chain: Optional[ReasoningChain] = None
    tools_used: List[str] = field(default_factory=list)
    rag_results: Optional[List[Dict[str, Any]]] = None
    memories_accessed: List[str] = field(default_factory=list)
    processing_time: float = 0.0
    confidence_score: float = 1.0
    metadata: Dict[str, Any] = field(default_factory=dict)


class IntelligentAgent(LoggerMixin):
    """智能Agent核心系统"""
    
    def __init__(
        self,
        ai_adapter: BaseAIAdapter,
        tool_service: Optional[ToolService] = None,
        rag_service: Optional[RAGService] = None,
        memory_system: Optional[AdvancedMemorySystem] = None,
        agent_id: Optional[str] = None
    ):
        super().__init__()
        
        self.agent_id = agent_id or f"agent_{int(time.time())}"
        
        # 核心组件
        self.ai_adapter = ai_adapter
        self.tool_service = tool_service
        self.rag_service = rag_service
        
        # 记忆系统
        self.memory_system = memory_system or AdvancedMemorySystem()
        
        # 推理引擎
        self.reasoning_engine = ReasoningEngine(
            ai_adapter=ai_adapter,
            tool_service=tool_service
        )
        
        # 智能体状态
        self.active_sessions: Dict[str, Dict[str, Any]] = {}
        self.performance_stats: Dict[str, Any] = {
            "total_requests": 0,
            "successful_requests": 0,
            "average_response_time": 0.0,
            "mode_usage": {mode.value: 0 for mode in AgentMode},
            "tool_usage": {},
            "error_count": 0
        }
        
        # 性能优化：缓存配置
        self.context_cache_enabled = True
        self.response_cache_enabled = True
        
        # 模式选择配置
        self.mode_selection_config = {
            "keywords": {
                AgentMode.RAG: ["搜索", "查询", "找到", "search", "find", "lookup", "知识库", "文档"],
                AgentMode.REASONING: ["分析", "推理", "思考", "解释", "analyze", "reason", "think", "explain", "为什么", "如何", "why", "how"],
                AgentMode.TOOL_USING: ["计算", "获取", "处理", "执行", "calculate", "fetch", "process", "execute", "工具", "tool"],
                AgentMode.PLANNING: ["计划", "规划", "步骤", "方案", "plan", "planning", "steps", "strategy"],
                AgentMode.CREATIVE: ["创作", "设计", "想象", "创意", "creative", "design", "imagine", "brainstorm"],
                AgentMode.ANALYTICAL: ["统计", "对比", "评估", "分析", "statistics", "compare", "evaluate", "analyze"]
            }
        }
    
    @monitor_performance("agent_request_processing")
    async def process_request(self, request: AgentRequest) -> AgentResponse:
        """处理智能体请求"""
        
        start_time = time.time()
        self.performance_stats["total_requests"] += 1
        
        try:
            # 选择处理模式
            selected_mode = await self._select_agent_mode(request)
            self.performance_stats["mode_usage"][selected_mode.value] += 1
            
            self.logger.info(f"Processing request with mode: {selected_mode.value}")
            
            # 准备上下文
            context = await self._prepare_context(request)
            
            # 根据模式处理请求
            response = await self._process_by_mode(request, selected_mode, context)
            
            # 后处理
            await self._post_process_response(request, response, context)
            
            # 更新统计
            response.processing_time = time.time() - start_time
            self.performance_stats["successful_requests"] += 1
            
            # 更新平均响应时间
            self._update_average_response_time(response.processing_time)
            
            # 记录性能指标
            performance_monitor.record_time("total_request_time", response.processing_time)
            performance_monitor.increment_counter("successful_requests")
            
            return response
            
        except Exception as e:
            self.logger.error(f"Agent request processing failed: {e}")
            self.performance_stats["error_count"] += 1
            
            # 返回错误响应
            error_response = AgentResponse(
                content=f"I apologize, but I encountered an error processing your request: {str(e)}",
                response_type=ResponseType.TEXT,
                mode_used=AgentMode.CHAT,
                processing_time=time.time() - start_time,
                confidence_score=0.0,
                metadata={"error": str(e)}
            )
            
            return error_response
    
    async def _select_agent_mode(self, request: AgentRequest) -> AgentMode:
        """智能选择处理模式"""
        
        if request.mode != AgentMode.AUTO:
            return request.mode
        
        message_lower = request.message.lower()
        
        # 计算每种模式的匹配分数
        mode_scores = {}
        
        for mode, keywords in self.mode_selection_config["keywords"].items():
            score = sum(1 for keyword in keywords if keyword in message_lower)
            if score > 0:
                mode_scores[mode] = score
        
        # 特殊规则
        # RAG模式：包含知识查询需求
        if any(word in message_lower for word in ["什么是", "定义", "解释", "介绍", "what is", "define", "explain"]):
            mode_scores[AgentMode.RAG] = mode_scores.get(AgentMode.RAG, 0) + 2
        
        # 推理模式：复杂问题
        if any(pattern in message_lower for pattern in ["为什么", "怎么", "如何", "分析", "why", "how", "analyze"]):
            mode_scores[AgentMode.REASONING] = mode_scores.get(AgentMode.REASONING, 0) + 2
        
        # 工具使用模式：需要实时数据或计算
        if any(word in message_lower for word in ["当前", "最新", "实时", "计算", "current", "latest", "calculate"]):
            mode_scores[AgentMode.TOOL_USING] = mode_scores.get(AgentMode.TOOL_USING, 0) + 2
        
        # 选择得分最高的模式
        if mode_scores:
            selected_mode = max(mode_scores.items(), key=lambda x: x[1])[0]
            self.logger.debug(f"Mode selection scores: {mode_scores}, selected: {selected_mode.value}")
            return selected_mode
        
        # 默认模式：对话
        return AgentMode.CHAT
    
    @monitor_performance("context_preparation")
    async def _prepare_context(self, request: AgentRequest) -> Dict[str, Any]:
        """准备上下文信息"""
        
        context = {
            "user_context": None,
            "rag_context": None,
            "memory_context": None,
            "session_context": None
        }
        
        # 获取用户上下文
        if request.memory_enabled and self.memory_system:
            try:
                user_context = await self.memory_system.get_context_for_user(
                    user_id=request.user_id,
                    query=request.message,
                    limit=5
                )
                context["user_context"] = user_context
            except Exception as e:
                self.logger.warning(f"Failed to get user context: {e}")
        
        # 获取RAG上下文
        if request.rag_enabled and self.rag_service:
            try:
                rag_results = await self.rag_service.query(
                    user_id=request.user_id,
                    query=request.message,
                    top_k=5,
                    session_id=request.session_id
                )
                context["rag_context"] = rag_results
            except Exception as e:
                self.logger.warning(f"Failed to get RAG context: {e}")
        
        # 获取会话上下文
        if request.session_id and request.session_id in self.active_sessions:
            context["session_context"] = self.active_sessions[request.session_id]
        
        return context
    
    async def _process_by_mode(
        self,
        request: AgentRequest,
        mode: AgentMode,
        context: Dict[str, Any]
    ) -> AgentResponse:
        """根据模式处理请求"""
        
        if mode == AgentMode.CHAT:
            return await self._process_chat_mode(request, context)
        elif mode == AgentMode.RAG:
            return await self._process_rag_mode(request, context)
        elif mode == AgentMode.REASONING:
            return await self._process_reasoning_mode(request, context)
        elif mode == AgentMode.TOOL_USING:
            return await self._process_tool_mode(request, context)
        elif mode == AgentMode.PLANNING:
            return await self._process_planning_mode(request, context)
        elif mode == AgentMode.CREATIVE:
            return await self._process_creative_mode(request, context)
        elif mode == AgentMode.ANALYTICAL:
            return await self._process_analytical_mode(request, context)
        else:
            return await self._process_chat_mode(request, context)
    
    @monitor_performance("chat_mode_processing")
    @cache_result(ttl=300)  # 缓存5分钟
    async def _process_chat_mode(
        self,
        request: AgentRequest,
        context: Dict[str, Any]
    ) -> AgentResponse:
        """处理对话模式"""
        
        # 构建对话提示
        messages = await self._build_chat_messages(request, context)
        
        # 生成响应
        if request.stream:
            # 流式响应
            response_content = ""
            async for chunk in self.ai_adapter.generate_response_stream(
                messages=messages,
                temperature=request.temperature,
                max_tokens=request.max_tokens
            ):
                response_content += chunk
            
            return AgentResponse(
                content=response_content,
                response_type=ResponseType.STREAMING,
                mode_used=AgentMode.CHAT
            )
        else:
            # 标准响应
            response_content = await self.ai_adapter.generate_response(
                messages=messages,
                temperature=request.temperature,
                max_tokens=request.max_tokens
            )
            
            return AgentResponse(
                content=response_content,
                response_type=ResponseType.TEXT,
                mode_used=AgentMode.CHAT
            )
    
    @monitor_performance("rag_mode_processing")
    async def _process_rag_mode(
        self,
        request: AgentRequest,
        context: Dict[str, Any]
    ) -> AgentResponse:
        """处理RAG模式"""
        
        rag_results = context.get("rag_context")
        
        if not rag_results or not rag_results.get("documents"):
            # 回退到对话模式
            return await self._process_chat_mode(request, context)
        
        # 构建RAG增强的提示
        documents = rag_results["documents"][:3]  # 使用前3个最相关的文档
        
        rag_context = "\n\n".join([
            f"文档{i+1}: {doc['content'][:500]}..."
            for i, doc in enumerate(documents)
        ])
        
        rag_prompt = f"""基于以下相关信息回答用户问题：

相关信息：
{rag_context}

用户问题：{request.message}

请基于提供的信息给出准确、有帮助的回答。如果信息不足，请说明并提供你能确定的部分。"""
        
        messages = [{"role": "user", "content": rag_prompt}]
        
        response_content = await self.ai_adapter.generate_response(
            messages=messages,
            temperature=request.temperature * 0.8,  # RAG模式使用较低温度
            max_tokens=request.max_tokens
        )
        
        return AgentResponse(
            content=response_content,
            response_type=ResponseType.TEXT,
            mode_used=AgentMode.RAG,
            rag_results=rag_results.get("documents", [])
        )
    
    @monitor_performance("reasoning_mode_processing")
    async def _process_reasoning_mode(
        self,
        request: AgentRequest,
        context: Dict[str, Any]
    ) -> AgentResponse:
        """处理推理模式"""
        
        # 执行推理链
        reasoning_chain = await self.reasoning_engine.reason(
            query=request.message,
            context=request.context,
            user_id=request.user_id
        )
        
        return AgentResponse(
            content=reasoning_chain.final_answer or "Unable to complete reasoning process.",
            response_type=ResponseType.REASONING_CHAIN,
            mode_used=AgentMode.REASONING,
            reasoning_chain=reasoning_chain,
            confidence_score=0.9 if reasoning_chain.success else 0.3
        )
    
    async def _process_tool_mode(
        self,
        request: AgentRequest,
        context: Dict[str, Any]
    ) -> AgentResponse:
        """处理工具使用模式"""
        
        if not self.tool_service:
            return await self._process_chat_mode(request, context)
        
        # 使用ReAct推理模式进行工具调用
        reasoning_chain = await self.reasoning_engine.reason(
            query=request.message,
            context=request.context,
            mode=ReasoningMode.REACT,
            user_id=request.user_id
        )
        
        # 提取使用的工具
        tools_used = []
        for step in reasoning_chain.steps:
            if step.step_type == "observation" and "action_name" in step.metadata:
                tools_used.append(step.metadata["action_name"])
        
        return AgentResponse(
            content=reasoning_chain.final_answer or "Tool execution completed.",
            response_type=ResponseType.TOOL_CALL,
            mode_used=AgentMode.TOOL_USING,
            reasoning_chain=reasoning_chain,
            tools_used=tools_used,
            confidence_score=0.8 if reasoning_chain.success else 0.4
        )
    
    async def _process_planning_mode(
        self,
        request: AgentRequest,
        context: Dict[str, Any]
    ) -> AgentResponse:
        """处理规划模式"""
        
        # 使用Plan-and-Execute推理模式
        reasoning_chain = await self.reasoning_engine.reason(
            query=request.message,
            context=request.context,
            mode=ReasoningMode.PLAN_AND_EXECUTE,
            user_id=request.user_id
        )
        
        return AgentResponse(
            content=reasoning_chain.final_answer or "Planning completed.",
            response_type=ResponseType.REASONING_CHAIN,
            mode_used=AgentMode.PLANNING,
            reasoning_chain=reasoning_chain,
            confidence_score=0.85 if reasoning_chain.success else 0.4
        )
    
    async def _process_creative_mode(
        self,
        request: AgentRequest,
        context: Dict[str, Any]
    ) -> AgentResponse:
        """处理创意模式"""
        
        creative_prompt = f"""你是一个富有创意的AI助手。请以创新、有想象力的方式回应以下请求：

用户请求：{request.message}

请发挥创意，提供原创、有趣、有价值的回答。可以包含：
- 创新的想法和概念
- 多角度的思考
- 富有想象力的方案
- 创意性的解决方案

回答要求：创新性强、实用性高、表达生动。"""
        
        messages = [{"role": "user", "content": creative_prompt}]
        
        response_content = await self.ai_adapter.generate_response(
            messages=messages,
            temperature=min(request.temperature * 1.3, 1.0),  # 创意模式使用更高温度
            max_tokens=request.max_tokens
        )
        
        return AgentResponse(
            content=response_content,
            response_type=ResponseType.TEXT,
            mode_used=AgentMode.CREATIVE,
            confidence_score=0.8
        )
    
    async def _process_analytical_mode(
        self,
        request: AgentRequest,
        context: Dict[str, Any]
    ) -> AgentResponse:
        """处理分析模式"""
        
        analytical_prompt = f"""你是一个专业的分析师。请对以下问题进行深入、系统的分析：

分析对象：{request.message}

请按照以下结构进行分析：
1. 问题理解和定义
2. 关键要素识别
3. 数据和事实分析
4. 多角度评估
5. 结论和建议

要求：
- 逻辑清晰，层次分明
- 基于事实和数据
- 客观中立的分析
- 提供可行的建议"""
        
        messages = [{"role": "user", "content": analytical_prompt}]
        
        response_content = await self.ai_adapter.generate_response(
            messages=messages,
            temperature=request.temperature * 0.7,  # 分析模式使用较低温度
            max_tokens=request.max_tokens
        )
        
        return AgentResponse(
            content=response_content,
            response_type=ResponseType.STRUCTURED,
            mode_used=AgentMode.ANALYTICAL,
            confidence_score=0.9
        )
    
    async def _build_chat_messages(
        self,
        request: AgentRequest,
        context: Dict[str, Any]
    ) -> List[Dict[str, str]]:
        """构建对话消息"""
        
        messages = []
        
        # 系统提示
        system_prompt = self._build_system_prompt(request, context)
        if system_prompt:
            messages.append({"role": "system", "content": system_prompt})
        
        # 添加上下文信息
        if context.get("user_context") and context["user_context"].get("user_profile"):
            profile = context["user_context"]["user_profile"]
            if profile.preferences or profile.interests:
                context_info = f"用户偏好：{profile.preferences}，兴趣领域：{list(profile.interests)}"
                messages.append({"role": "system", "content": f"用户背景信息：{context_info}"})
        
        # 添加会话历史
        if context.get("session_context") and "messages" in context["session_context"]:
            recent_messages = context["session_context"]["messages"][-6:]  # 最近6条消息
            messages.extend(recent_messages)
        
        # 用户消息
        messages.append({"role": "user", "content": request.message})
        
        return messages
    
    def _build_system_prompt(
        self,
        request: AgentRequest,
        context: Dict[str, Any]
    ) -> Optional[str]:
        """构建系统提示"""
        
        base_prompt = """你是PolyAgent，一个高级智能助手。你具备以下能力：

1. 知识检索：可以从知识库中查找相关信息
2. 逻辑推理：可以进行复杂的逻辑思考和推理
3. 工具使用：可以调用各种工具完成任务
4. 记忆管理：可以记住和调用历史对话信息
5. 任务规划：可以制定和执行复杂任务计划

请根据用户的需求选择合适的方式回应，提供准确、有用、友好的帮助。"""
        
        # 根据上下文调整提示
        if context.get("rag_context") and context["rag_context"].get("documents"):
            base_prompt += "\n\n你可以使用提供的知识库信息来增强你的回答。"
        
        if context.get("user_context") and context["user_context"].get("recent_memories"):
            base_prompt += "\n\n请考虑用户的历史对话和偏好。"
        
        return base_prompt
    
    async def _post_process_response(
        self,
        request: AgentRequest,
        response: AgentResponse,
        context: Dict[str, Any]
    ):
        """后处理响应"""
        
        # 存储对话记忆
        if request.memory_enabled and self.memory_system:
            try:
                # 存储用户消息
                await self.memory_system.add_conversation_memory(
                    conversation_id=request.session_id or "default",
                    user_id=request.user_id,
                    message={"role": "user", "content": request.message}
                )
                
                # 存储助手响应
                await self.memory_system.add_conversation_memory(
                    conversation_id=request.session_id or "default",
                    user_id=request.user_id,
                    message={"role": "assistant", "content": response.content}
                )
                
                # 存储重要信息为长期记忆
                if response.confidence_score >= 0.8:
                    await self.memory_system.store_memory(
                        content=f"用户问题: {request.message}\n助手回答: {response.content}",
                        memory_type=MemoryType.EPISODIC,
                        importance=MemoryImportance.MEDIUM,
                        metadata={"user_id": request.user_id, "mode": response.mode_used.value}
                    )
            except Exception as e:
                self.logger.warning(f"Failed to store memory: {e}")
        
        # 更新会话状态
        if request.session_id:
            if request.session_id not in self.active_sessions:
                self.active_sessions[request.session_id] = {
                    "user_id": request.user_id,
                    "created_at": datetime.now(),
                    "messages": [],
                    "context": {}
                }
            
            session = self.active_sessions[request.session_id]
            session["messages"].append({"role": "user", "content": request.message})
            session["messages"].append({"role": "assistant", "content": response.content})
            session["last_activity"] = datetime.now()
            
            # 限制会话历史长度
            if len(session["messages"]) > 20:
                session["messages"] = session["messages"][-20:]
        
        # 更新工具使用统计
        for tool in response.tools_used:
            if tool in self.performance_stats["tool_usage"]:
                self.performance_stats["tool_usage"][tool] += 1
            else:
                self.performance_stats["tool_usage"][tool] = 1
    
    def _update_average_response_time(self, response_time: float):
        """更新平均响应时间"""
        current_avg = self.performance_stats["average_response_time"]
        total_requests = self.performance_stats["successful_requests"]
        
        if total_requests == 1:
            self.performance_stats["average_response_time"] = response_time
        else:
            # 使用滑动平均
            self.performance_stats["average_response_time"] = (
                (current_avg * (total_requests - 1) + response_time) / total_requests
            )
    
    async def get_agent_status(self) -> Dict[str, Any]:
        """获取智能体状态"""
        
        memory_stats = {}
        if self.memory_system:
            memory_stats = self.memory_system.get_memory_statistics()
        
        reasoning_stats = {
            "active_chains": len(self.reasoning_engine.reasoning_chains),
            "execution_plans": len(self.reasoning_engine.execution_plans)
        }
        
        return {
            "agent_id": self.agent_id,
            "active_sessions": len(self.active_sessions),
            "performance": self.performance_stats,
            "memory": memory_stats,
            "reasoning": reasoning_stats,
            "components": {
                "ai_adapter": self.ai_adapter is not None,
                "tool_service": self.tool_service is not None,
                "rag_service": self.rag_service is not None,
                "memory_system": self.memory_system is not None,
                "reasoning_engine": self.reasoning_engine is not None
            }
        }
    
    async def cleanup_inactive_sessions(self, max_idle_hours: int = 24):
        """清理不活跃的会话"""
        
        current_time = datetime.now()
        inactive_sessions = []
        
        for session_id, session_data in self.active_sessions.items():
            last_activity = session_data.get("last_activity", session_data.get("created_at"))
            if last_activity:
                idle_time = (current_time - last_activity).total_seconds() / 3600
                if idle_time > max_idle_hours:
                    inactive_sessions.append(session_id)
        
        for session_id in inactive_sessions:
            del self.active_sessions[session_id]
        
        self.logger.info(f"Cleaned up {len(inactive_sessions)} inactive sessions")
        performance_monitor.increment_counter("sessions_cleaned", len(inactive_sessions))
        return len(inactive_sessions)
    
    async def get_performance_metrics(self) -> Dict[str, Any]:
        """获取性能指标"""
        
        # 获取基础性能统计
        perf_stats = performance_monitor.get_stats()
        
        # 添加Agent特定指标
        agent_metrics = {
            "agent_stats": self.performance_stats,
            "active_sessions": len(self.active_sessions),
            "memory_usage": {},
            "reasoning_metrics": {}
        }
        
        # 内存系统指标
        if self.memory_system:
            agent_metrics["memory_usage"] = self.memory_system.get_memory_statistics()
        
        # 推理引擎指标
        if hasattr(self.reasoning_engine, 'get_performance_stats'):
            agent_metrics["reasoning_metrics"] = self.reasoning_engine.get_performance_stats()
        
        return {
            "performance": perf_stats,
            "agent_metrics": agent_metrics,
            "timestamp": datetime.now().isoformat()
        }
    
    async def optimize_performance(self):
        """性能优化操作"""
        
        # 清理过期缓存
        if hasattr(self, '_context_cache'):
            self._context_cache.cleanup_expired()
        
        # 清理不活跃会话
        cleaned_sessions = await self.cleanup_inactive_sessions()
        
        # 优化内存系统
        if self.memory_system and hasattr(self.memory_system, 'optimize'):
            await self.memory_system.optimize()
        
        # 优化推理引擎
        if hasattr(self.reasoning_engine, 'optimize'):
            await self.reasoning_engine.optimize()
        
        self.logger.info(f"Performance optimization completed. Cleaned {cleaned_sessions} sessions.")
        
        return {
            "sessions_cleaned": cleaned_sessions,
            "timestamp": datetime.now().isoformat()
        }