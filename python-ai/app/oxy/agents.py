"""
Oxy Agents - 智能体协作系统
基于OxyGent的设计理念实现多智能体协作
支持动态组织、实时协商、角色切换
"""

import asyncio
import time
import json
from typing import Dict, List, Any, Optional, Set, Callable
from dataclasses import dataclass, field
from enum import Enum
import uuid

from .core import (
    BaseOxy, OxyContext, OxyResult, OxyMessage, OxyStatus, OxyType,
    oxy_registry
)
from .workflow import WorkflowEngine, WorkflowDefinition, WorkflowBuilder
from ..core.logging import LoggerMixin
from ..core.exceptions import AgentException
from ..adapters.unified_adapter import UnifiedAIAdapter

class AgentRole(Enum):
    """智能体角色"""
    COORDINATOR = "coordinator"      # 协调者
    EXECUTOR = "executor"           # 执行者
    ANALYST = "analyst"            # 分析师
    SPECIALIST = "specialist"       # 专家
    REVIEWER = "reviewer"          # 审核者
    PLANNER = "planner"           # 规划者
    NEGOTIATOR = "negotiator"      # 协商者
    MONITOR = "monitor"           # 监控者

class CollaborationPattern(Enum):
    """协作模式"""
    HIERARCHICAL = "hierarchical"      # 层次化协作
    PEER_TO_PEER = "peer_to_peer"     # 点对点协作
    CONSENSUS = "consensus"            # 共识协作
    COMPETITIVE = "competitive"        # 竞争协作
    PIPELINE = "pipeline"              # 流水线协作
    SWARM = "swarm"                   # 群体协作

class TaskPriority(Enum):
    """任务优先级"""
    LOW = 1
    MEDIUM = 2
    HIGH = 3
    CRITICAL = 4
    URGENT = 5

@dataclass
class AgentCapability:
    """智能体能力定义"""
    capability_id: str
    name: str
    description: str
    proficiency_level: float = 0.8  # 0-1
    confidence_level: float = 0.8   # 0-1
    cost_factor: float = 1.0        # 相对成本
    execution_time_estimate: float = 1.0  # 相对执行时间

@dataclass
class CollaborationTask:
    """协作任务"""
    task_id: str = field(default_factory=lambda: str(uuid.uuid4()))
    name: str = ""
    description: str = ""
    priority: TaskPriority = TaskPriority.MEDIUM
    required_capabilities: List[str] = field(default_factory=list)
    max_agents: int = 5
    deadline: Optional[float] = None
    
    # 任务状态
    status: str = "pending"
    assigned_agents: List[str] = field(default_factory=list)
    start_time: Optional[float] = None
    end_time: Optional[float] = None
    
    # 任务数据
    input_data: Dict[str, Any] = field(default_factory=dict)
    output_data: Dict[str, Any] = field(default_factory=dict)
    context: Optional[OxyContext] = None

@dataclass
class NegotiationProposal:
    """协商提案"""
    proposal_id: str = field(default_factory=lambda: str(uuid.uuid4()))
    proposer_id: str = ""
    task_id: str = ""
    proposal_type: str = "task_assignment"  # task_assignment, resource_allocation, priority_change
    content: Dict[str, Any] = field(default_factory=dict)
    justification: str = ""
    confidence: float = 0.8
    created_at: float = field(default_factory=time.time)
    
    # 投票状态
    votes_for: Set[str] = field(default_factory=set)
    votes_against: Set[str] = field(default_factory=set)
    status: str = "pending"  # pending, accepted, rejected

class OxyAgent(BaseOxy):
    """Oxy智能体组件"""
    
    def __init__(
        self,
        agent_id: str = None,
        name: str = "",
        role: AgentRole = AgentRole.EXECUTOR,
        capabilities: List[AgentCapability] = None,
        ai_adapter: UnifiedAIAdapter = None,
        **kwargs
    ):
        super().__init__(
            oxy_id=agent_id or f"agent_{uuid.uuid4().hex[:8]}",
            name=name or f"Agent_{role.value.title()}",
            oxy_type=OxyType.AGENT,
            **kwargs
        )
        
        self.role = role
        self.capabilities = capabilities or []
        self.ai_adapter = ai_adapter
        
        # 协作状态
        self.current_tasks: Dict[str, CollaborationTask] = {}
        self.collaboration_history: List[Dict[str, Any]] = []
        self.trust_scores: Dict[str, float] = {}  # 对其他智能体的信任度
        self.reputation_score: float = 0.8
        
        # 性能指标
        self.completed_tasks: int = 0
        self.successful_collaborations: int = 0
        self.average_task_time: float = 0.0
        self.quality_score: float = 0.8
        
        # 协商能力
        self.negotiation_style: str = "cooperative"  # cooperative, competitive, adaptive
        self.decision_threshold: float = 0.7
        
    async def execute(self, context: OxyContext, **kwargs) -> OxyResult:
        """执行智能体任务"""
        start_time = time.time()
        
        try:
            await self.pre_execute(context, **kwargs)
            
            # 获取任务信息
            task = kwargs.get('task')
            if not task:
                return OxyResult(
                    success=False,
                    message="No task provided",
                    execution_time=time.time() - start_time
                )
            
            # 根据角色执行不同逻辑
            if self.role == AgentRole.COORDINATOR:
                result_data = await self._coordinate_task(task, context, **kwargs)
            elif self.role == AgentRole.ANALYST:
                result_data = await self._analyze_task(task, context, **kwargs)
            elif self.role == AgentRole.SPECIALIST:
                result_data = await self._execute_specialized_task(task, context, **kwargs)
            elif self.role == AgentRole.PLANNER:
                result_data = await self._plan_task(task, context, **kwargs)
            else:
                result_data = await self._execute_general_task(task, context, **kwargs)
            
            # 更新性能指标
            self.completed_tasks += 1
            execution_time = time.time() - start_time
            self.average_task_time = (
                (self.average_task_time * (self.completed_tasks - 1) + execution_time) 
                / self.completed_tasks
            )
            
            result = OxyResult(
                success=True,
                data=result_data,
                message=f"Task completed by {self.role.value}",
                execution_time=execution_time
            )
            
            await self.post_execute(context, result)
            return result
            
        except Exception as e:
            self.logger.error(f"Agent execution failed: {e}")
            result = OxyResult(
                success=False,
                message=f"Agent execution error: {str(e)}",
                error=e,
                execution_time=time.time() - start_time
            )
            await self.post_execute(context, result)
            return result
    
    async def _coordinate_task(
        self, 
        task: CollaborationTask, 
        context: OxyContext, 
        **kwargs
    ) -> Dict[str, Any]:
        """协调任务执行"""
        
        if not self.ai_adapter:
            raise AgentException("AI adapter not configured for coordinator")
        
        # 分析任务并制定计划
        analysis_prompt = f"""
        作为协调者，分析以下任务并制定执行计划：
        
        任务名称：{task.name}
        任务描述：{task.description}
        优先级：{task.priority.name}
        需要能力：{', '.join(task.required_capabilities)}
        
        请提供：
        1. 任务分解建议
        2. 所需智能体角色
        3. 执行顺序
        4. 风险评估
        5. 资源需求
        
        以JSON格式回复：
        {{
            "subtasks": [{{
                "name": "子任务名称",
                "description": "子任务描述",
                "required_role": "所需角色",
                "estimated_time": "预计时间(分钟)",
                "dependencies": ["依赖的子任务"]
            }}],
            "execution_plan": {{
                "total_estimated_time": "总预计时间",
                "critical_path": ["关键路径任务"],
                "resource_requirements": {{"agents": 数量, "estimated_cost": "预计成本"}}
            }},
            "risks": [{{
                "risk": "风险描述",
                "probability": "概率(0-1)",
                "impact": "影响程度",
                "mitigation": "缓解措施"
            }}]
        }}
        """
        
        messages = [{"role": "user", "content": analysis_prompt}]
        response = await self.ai_adapter.generate(messages, temperature=0.3)
        
        try:
            plan = json.loads(response.content)
            
            # 记录协调决策
            self.collaboration_history.append({
                "action": "task_coordination",
                "task_id": task.task_id,
                "plan": plan,
                "timestamp": time.time()
            })
            
            return {
                "coordination_plan": plan,
                "coordinator_id": self.oxy_id,
                "task_breakdown": plan.get("subtasks", []),
                "execution_strategy": plan.get("execution_plan", {}),
                "risk_assessment": plan.get("risks", [])
            }
            
        except json.JSONDecodeError:
            # 如果JSON解析失败，返回基础分析
            return {
                "coordination_plan": {"status": "basic_analysis"},
                "coordinator_id": self.oxy_id,
                "analysis": response.content,
                "task_breakdown": [],
                "execution_strategy": {},
                "risk_assessment": []
            }
    
    async def _analyze_task(
        self, 
        task: CollaborationTask, 
        context: OxyContext, 
        **kwargs
    ) -> Dict[str, Any]:
        """分析任务"""
        
        if not self.ai_adapter:
            raise AgentException("AI adapter not configured for analyst")
        
        analysis_prompt = f"""
        作为数据分析专家，深入分析以下任务和相关数据：
        
        任务：{task.name} - {task.description}
        输入数据：{json.dumps(task.input_data, ensure_ascii=False, indent=2)}
        上下文变量：{json.dumps(context.variables, ensure_ascii=False, indent=2)}
        
        请提供详细的分析报告，包括：
        1. 数据质量评估
        2. 关键指标识别
        3. 趋势分析
        4. 异常检测
        5. 结论和建议
        
        以结构化格式回复分析结果。
        """
        
        messages = [{"role": "user", "content": analysis_prompt}]
        response = await self.ai_adapter.generate(messages, temperature=0.2)
        
        return {
            "analysis_report": response.content,
            "analyst_id": self.oxy_id,
            "analysis_timestamp": time.time(),
            "confidence_level": 0.85,
            "data_quality_score": 0.8
        }
    
    async def _execute_specialized_task(
        self, 
        task: CollaborationTask, 
        context: OxyContext, 
        **kwargs
    ) -> Dict[str, Any]:
        """执行专业任务"""
        
        # 根据智能体的专业能力执行任务
        primary_capability = self.capabilities[0] if self.capabilities else None
        
        if not primary_capability:
            return {"result": "No specialized capabilities defined", "specialist_id": self.oxy_id}
        
        if self.ai_adapter:
            specialist_prompt = f"""
            作为{primary_capability.name}专家，处理以下专业任务：
            
            任务：{task.name}
            描述：{task.description}
            专业领域：{primary_capability.description}
            熟练程度：{primary_capability.proficiency_level:.2f}
            
            请运用你的专业知识提供高质量的解决方案。
            """
            
            messages = [{"role": "user", "content": specialist_prompt}]
            response = await self.ai_adapter.generate(messages, temperature=0.4)
            
            return {
                "specialist_result": response.content,
                "specialist_id": self.oxy_id,
                "capability_used": primary_capability.name,
                "proficiency_level": primary_capability.proficiency_level,
                "confidence_level": primary_capability.confidence_level
            }
        
        return {
            "specialist_result": f"Specialized task execution for {primary_capability.name}",
            "specialist_id": self.oxy_id,
            "capability_used": primary_capability.name
        }
    
    async def _plan_task(
        self, 
        task: CollaborationTask, 
        context: OxyContext, 
        **kwargs
    ) -> Dict[str, Any]:
        """规划任务"""
        
        if not self.ai_adapter:
            return {
                "plan": "Basic task planning completed",
                "planner_id": self.oxy_id,
                "planning_steps": []
            }
        
        planning_prompt = f"""
        作为任务规划专家，为以下任务制定详细的执行计划：
        
        任务：{task.name}
        描述：{task.description}
        优先级：{task.priority.name}
        截止时间：{task.deadline if task.deadline else "无"}
        可用资源：{task.max_agents} 个智能体
        
        请制定包含以下内容的详细计划：
        1. 阶段划分和里程碑
        2. 资源分配策略
        3. 风险管控措施
        4. 质量保证流程
        5. 进度监控机制
        
        以JSON格式返回计划结构。
        """
        
        messages = [{"role": "user", "content": planning_prompt}]
        response = await self.ai_adapter.generate(messages, temperature=0.3)
        
        return {
            "detailed_plan": response.content,
            "planner_id": self.oxy_id,
            "planning_timestamp": time.time(),
            "plan_confidence": 0.8
        }
    
    async def _execute_general_task(
        self, 
        task: CollaborationTask, 
        context: OxyContext, 
        **kwargs
    ) -> Dict[str, Any]:
        """执行一般任务"""
        
        if self.ai_adapter:
            task_prompt = f"""
            执行以下任务：
            
            任务名称：{task.name}
            任务描述：{task.description}
            输入数据：{json.dumps(task.input_data, ensure_ascii=False)}
            
            请完成任务并提供结果。
            """
            
            messages = [{"role": "user", "content": task_prompt}]
            response = await self.ai_adapter.generate(messages, temperature=0.6)
            
            return {
                "task_result": response.content,
                "executor_id": self.oxy_id,
                "execution_timestamp": time.time()
            }
        
        return {
            "task_result": f"Task {task.name} executed by {self.oxy_id}",
            "executor_id": self.oxy_id
        }
    
    def can_handle_task(self, task: CollaborationTask) -> float:
        """评估是否能处理任务（返回能力匹配度 0-1）"""
        
        if not task.required_capabilities:
            return 0.5  # 默认能力
        
        if not self.capabilities:
            return 0.2  # 低能力
        
        # 计算能力匹配度
        capability_names = {cap.capability_id for cap in self.capabilities}
        required_capabilities = set(task.required_capabilities)
        
        match_count = len(capability_names.intersection(required_capabilities))
        total_required = len(required_capabilities)
        
        if total_required == 0:
            return 0.5
        
        match_ratio = match_count / total_required
        
        # 考虑熟练程度
        if match_count > 0:
            avg_proficiency = sum(
                cap.proficiency_level for cap in self.capabilities
                if cap.capability_id in required_capabilities
            ) / match_count
            
            return match_ratio * avg_proficiency
        
        return match_ratio
    
    def estimate_task_cost(self, task: CollaborationTask) -> float:
        """估算任务成本"""
        
        base_cost = 1.0
        
        # 根据能力匹配度调整成本
        capability_match = self.can_handle_task(task)
        if capability_match < 0.5:
            base_cost *= 2.0  # 能力不匹配时成本增加
        
        # 根据任务优先级调整
        priority_multiplier = {
            TaskPriority.LOW: 0.8,
            TaskPriority.MEDIUM: 1.0,
            TaskPriority.HIGH: 1.2,
            TaskPriority.CRITICAL: 1.5,
            TaskPriority.URGENT: 2.0
        }
        
        base_cost *= priority_multiplier.get(task.priority, 1.0)
        
        # 考虑能力成本因子
        if self.capabilities:
            avg_cost_factor = sum(cap.cost_factor for cap in self.capabilities) / len(self.capabilities)
            base_cost *= avg_cost_factor
        
        return base_cost
    
    def update_trust_score(self, other_agent_id: str, success: bool, quality: float = 0.8):
        """更新对其他智能体的信任度"""
        
        current_trust = self.trust_scores.get(other_agent_id, 0.5)
        
        # 简单的信任度更新算法
        if success:
            # 成功时增加信任度
            new_trust = current_trust + (quality * 0.1 * (1 - current_trust))
        else:
            # 失败时降低信任度
            new_trust = current_trust * 0.9
        
        self.trust_scores[other_agent_id] = max(0.0, min(1.0, new_trust))
    
    def get_agent_info(self) -> Dict[str, Any]:
        """获取智能体详细信息"""
        base_info = self.get_info()
        
        agent_info = {
            **base_info,
            "role": self.role.value,
            "capabilities": [
                {
                    "id": cap.capability_id,
                    "name": cap.name,
                    "proficiency": cap.proficiency_level,
                    "confidence": cap.confidence_level
                }
                for cap in self.capabilities
            ],
            "performance_metrics": {
                "completed_tasks": self.completed_tasks,
                "successful_collaborations": self.successful_collaborations,
                "average_task_time": self.average_task_time,
                "quality_score": self.quality_score,
                "reputation_score": self.reputation_score
            },
            "trust_scores": self.trust_scores,
            "current_tasks": len(self.current_tasks),
            "collaboration_history_count": len(self.collaboration_history)
        }
        
        return agent_info

class AgentCollaborationEngine(LoggerMixin):
    """智能体协作引擎"""
    
    def __init__(self, workflow_engine: WorkflowEngine = None):
        super().__init__()
        
        self.workflow_engine = workflow_engine or WorkflowEngine()
        self.agents: Dict[str, OxyAgent] = {}
        self.active_tasks: Dict[str, CollaborationTask] = {}
        self.negotiation_proposals: Dict[str, NegotiationProposal] = {}
        
        # 协作统计
        self.total_collaborations = 0
        self.successful_collaborations = 0
        self.total_negotiation_rounds = 0
        self.average_collaboration_time = 0.0
    
    def register_agent(self, agent: OxyAgent) -> str:
        """注册智能体"""
        self.agents[agent.oxy_id] = agent
        oxy_registry.register(agent)
        self.logger.info(f"Registered agent: {agent.name} ({agent.oxy_id})")
        return agent.oxy_id
    
    def unregister_agent(self, agent_id: str) -> bool:
        """注销智能体"""
        if agent_id in self.agents:
            del self.agents[agent_id]
            oxy_registry.unregister(agent_id)
            self.logger.info(f"Unregistered agent: {agent_id}")
            return True
        return False
    
    async def execute_collaborative_task(
        self,
        task: CollaborationTask,
        collaboration_pattern: CollaborationPattern = CollaborationPattern.HIERARCHICAL
    ) -> Dict[str, Any]:
        """执行协作任务"""
        
        self.logger.info(f"Starting collaborative task: {task.name} ({task.task_id})")
        
        start_time = time.time()
        task.start_time = start_time
        task.status = "running"
        self.active_tasks[task.task_id] = task
        
        try:
            # 选择合适的智能体
            selected_agents = await self._select_agents_for_task(task)
            
            if not selected_agents:
                raise AgentException("No suitable agents found for the task")
            
            task.assigned_agents = [agent.oxy_id for agent in selected_agents]
            
            # 根据协作模式执行任务
            if collaboration_pattern == CollaborationPattern.HIERARCHICAL:
                result = await self._execute_hierarchical_collaboration(task, selected_agents)
            elif collaboration_pattern == CollaborationPattern.PEER_TO_PEER:
                result = await self._execute_peer_to_peer_collaboration(task, selected_agents)
            elif collaboration_pattern == CollaborationPattern.CONSENSUS:
                result = await self._execute_consensus_collaboration(task, selected_agents)
            else:
                result = await self._execute_default_collaboration(task, selected_agents)
            
            # 完成任务
            task.status = "completed"
            task.end_time = time.time()
            task.output_data = result
            
            self.total_collaborations += 1
            self.successful_collaborations += 1
            
            # 更新平均协作时间
            collaboration_time = task.end_time - task.start_time
            self.average_collaboration_time = (
                (self.average_collaboration_time * (self.total_collaborations - 1) + collaboration_time)
                / self.total_collaborations
            )
            
            self.logger.info(f"Collaborative task completed: {task.task_id}")
            
            return {
                "success": True,
                "task_id": task.task_id,
                "result": result,
                "execution_time": collaboration_time,
                "agents_involved": task.assigned_agents,
                "collaboration_pattern": collaboration_pattern.value
            }
            
        except Exception as e:
            task.status = "failed"
            task.end_time = time.time()
            
            self.total_collaborations += 1
            self.logger.error(f"Collaborative task failed: {task.task_id}, error: {e}")
            
            return {
                "success": False,
                "task_id": task.task_id,
                "error": str(e),
                "execution_time": time.time() - start_time,
                "agents_involved": task.assigned_agents,
                "collaboration_pattern": collaboration_pattern.value
            }
        
        finally:
            self.active_tasks.pop(task.task_id, None)
    
    async def _select_agents_for_task(self, task: CollaborationTask) -> List[OxyAgent]:
        """为任务选择合适的智能体"""
        
        if not self.agents:
            return []
        
        # 计算每个智能体的适合度分数
        agent_scores = []
        
        for agent in self.agents.values():
            capability_match = agent.can_handle_task(task)
            cost_estimate = agent.estimate_task_cost(task)
            
            # 综合评分（能力匹配度高，成本低，信誉好）
            score = (
                capability_match * 0.5 +           # 能力匹配 50%
                (1.0 / cost_estimate) * 0.3 +      # 成本效益 30%
                agent.reputation_score * 0.2       # 信誉度 20%
            )
            
            agent_scores.append((agent, score, capability_match))
        
        # 按评分排序
        agent_scores.sort(key=lambda x: x[1], reverse=True)
        
        # 选择最合适的智能体（不超过max_agents）
        selected_agents = []
        selected_capabilities = set()
        
        for agent, score, capability_match in agent_scores:
            if len(selected_agents) >= task.max_agents:
                break
            
            # 确保能力覆盖
            agent_capabilities = {cap.capability_id for cap in agent.capabilities}
            
            if (capability_match > 0.3 and  # 最低能力要求
                (not selected_capabilities or  # 第一个智能体
                 agent_capabilities - selected_capabilities)):  # 提供新能力
                
                selected_agents.append(agent)
                selected_capabilities.update(agent_capabilities)
        
        # 确保至少有一个智能体
        if not selected_agents and agent_scores:
            selected_agents.append(agent_scores[0][0])
        
        return selected_agents
    
    async def _execute_hierarchical_collaboration(
        self,
        task: CollaborationTask,
        agents: List[OxyAgent]
    ) -> Dict[str, Any]:
        """执行层次化协作"""
        
        # 选择协调者（优先选择COORDINATOR角色）
        coordinator = None
        executors = []
        
        for agent in agents:
            if agent.role == AgentRole.COORDINATOR and not coordinator:
                coordinator = agent
            else:
                executors.append(agent)
        
        # 如果没有专门的协调者，选择信誉最高的作为协调者
        if not coordinator and agents:
            coordinator = max(agents, key=lambda a: a.reputation_score)
            executors = [a for a in agents if a != coordinator]
        
        # 协调者制定计划
        coordination_result = await coordinator.execute(task.context, task=task)
        
        if not coordination_result.success:
            raise AgentException(f"Coordination failed: {coordination_result.message}")
        
        coordination_plan = coordination_result.data
        
        # 执行者按计划执行任务
        execution_results = []
        
        for executor in executors:
            # 为每个执行者分配子任务
            subtask = CollaborationTask(
                name=f"{task.name}_subtask_{executor.oxy_id}",
                description=f"Subtask for {executor.name}",
                input_data=task.input_data,
                context=task.context
            )
            
            result = await executor.execute(task.context, task=subtask)
            execution_results.append({
                "agent_id": executor.oxy_id,
                "agent_name": executor.name,
                "result": result.data if result.success else None,
                "success": result.success,
                "error": str(result.error) if result.error else None
            })
        
        return {
            "collaboration_type": "hierarchical",
            "coordinator": {
                "agent_id": coordinator.oxy_id,
                "agent_name": coordinator.name,
                "coordination_plan": coordination_plan
            },
            "execution_results": execution_results,
            "overall_success": all(r["success"] for r in execution_results)
        }
    
    async def _execute_peer_to_peer_collaboration(
        self,
        task: CollaborationTask,
        agents: List[OxyAgent]
    ) -> Dict[str, Any]:
        """执行点对点协作"""
        
        # 所有智能体并行执行任务
        tasks_coroutines = []
        
        for agent in agents:
            agent_task = CollaborationTask(
                name=f"{task.name}_peer_{agent.oxy_id}",
                description=f"Peer task for {agent.name}",
                input_data=task.input_data,
                context=task.context
            )
            
            coroutine = agent.execute(task.context, task=agent_task)
            tasks_coroutines.append((agent, coroutine))
        
        # 等待所有任务完成
        results = []
        
        for agent, coroutine in tasks_coroutines:
            try:
                result = await coroutine
                results.append({
                    "agent_id": agent.oxy_id,
                    "agent_name": agent.name,
                    "result": result.data if result.success else None,
                    "success": result.success,
                    "execution_time": result.execution_time,
                    "error": str(result.error) if result.error else None
                })
            except Exception as e:
                results.append({
                    "agent_id": agent.oxy_id,
                    "agent_name": agent.name,
                    "result": None,
                    "success": False,
                    "execution_time": 0.0,
                    "error": str(e)
                })
        
        # 选择最佳结果
        successful_results = [r for r in results if r["success"]]
        
        if successful_results:
            # 选择质量最高的结果（简化实现：选择第一个成功的）
            best_result = successful_results[0]
        else:
            best_result = None
        
        return {
            "collaboration_type": "peer_to_peer",
            "all_results": results,
            "best_result": best_result,
            "success_rate": len(successful_results) / len(results) if results else 0
        }
    
    async def _execute_consensus_collaboration(
        self,
        task: CollaborationTask,
        agents: List[OxyAgent]
    ) -> Dict[str, Any]:
        """执行共识协作"""
        
        # 第一轮：所有智能体提出解决方案
        proposals = []
        
        for agent in agents:
            agent_task = CollaborationTask(
                name=f"{task.name}_proposal_{agent.oxy_id}",
                description=f"Proposal from {agent.name}",
                input_data=task.input_data,
                context=task.context
            )
            
            result = await agent.execute(task.context, task=agent_task)
            
            if result.success:
                proposals.append({
                    "agent_id": agent.oxy_id,
                    "agent_name": agent.name,
                    "proposal": result.data,
                    "confidence": getattr(result, 'confidence', 0.8)
                })
        
        if not proposals:
            raise AgentException("No valid proposals generated")
        
        # 第二轮：协商和投票
        negotiation_results = await self._conduct_negotiation(agents, proposals, task)
        
        return {
            "collaboration_type": "consensus",
            "initial_proposals": proposals,
            "negotiation_results": negotiation_results,
            "final_consensus": negotiation_results.get("consensus")
        }
    
    async def _execute_default_collaboration(
        self,
        task: CollaborationTask,
        agents: List[OxyAgent]
    ) -> Dict[str, Any]:
        """执行默认协作"""
        
        # 简单的顺序执行
        results = []
        
        for agent in agents:
            agent_task = CollaborationTask(
                name=f"{task.name}_sequential_{agent.oxy_id}",
                description=f"Sequential task for {agent.name}",
                input_data=task.input_data,
                context=task.context
            )
            
            result = await agent.execute(task.context, task=agent_task)
            
            results.append({
                "agent_id": agent.oxy_id,
                "agent_name": agent.name,
                "result": result.data if result.success else None,
                "success": result.success,
                "error": str(result.error) if result.error else None
            })
            
            # 如果当前任务成功，将结果传递给下一个智能体
            if result.success:
                task.input_data.update({"previous_result": result.data})
        
        return {
            "collaboration_type": "sequential",
            "execution_chain": results,
            "final_result": results[-1] if results else None
        }
    
    async def _conduct_negotiation(
        self,
        agents: List[OxyAgent],
        proposals: List[Dict[str, Any]],
        task: CollaborationTask
    ) -> Dict[str, Any]:
        """进行协商"""
        
        # 简化的协商实现
        # 实际实现会更复杂，包括多轮投票、提案修改等
        
        vote_results = {}
        
        for proposal in proposals:
            votes_for = 0
            votes_against = 0
            
            for agent in agents:
                # 简化的投票逻辑：高信任度的智能体更容易获得支持
                proposer_trust = agent.trust_scores.get(proposal["agent_id"], 0.5)
                proposal_confidence = proposal["confidence"]
                
                vote_score = (proposer_trust + proposal_confidence) / 2
                
                if vote_score > agent.decision_threshold:
                    votes_for += 1
                else:
                    votes_against += 1
            
            vote_results[proposal["agent_id"]] = {
                "proposal": proposal,
                "votes_for": votes_for,
                "votes_against": votes_against,
                "support_ratio": votes_for / len(agents)
            }
        
        # 选择支持度最高的提案
        if vote_results:
            best_proposal = max(vote_results.items(), key=lambda x: x[1]["support_ratio"])
            consensus = best_proposal[1]["proposal"]
        else:
            consensus = None
        
        self.total_negotiation_rounds += 1
        
        return {
            "vote_results": vote_results,
            "consensus": consensus,
            "negotiation_rounds": 1
        }
    
    def get_collaboration_stats(self) -> Dict[str, Any]:
        """获取协作统计信息"""
        
        success_rate = (
            self.successful_collaborations / self.total_collaborations
            if self.total_collaborations > 0 else 0
        )
        
        return {
            "total_agents": len(self.agents),
            "active_tasks": len(self.active_tasks),
            "total_collaborations": self.total_collaborations,
            "successful_collaborations": self.successful_collaborations,
            "success_rate": success_rate,
            "average_collaboration_time": self.average_collaboration_time,
            "total_negotiation_rounds": self.total_negotiation_rounds,
            "agent_summary": [
                {
                    "agent_id": agent.oxy_id,
                    "name": agent.name,
                    "role": agent.role.value,
                    "completed_tasks": agent.completed_tasks,
                    "reputation_score": agent.reputation_score
                }
                for agent in self.agents.values()
            ]
        }

# 全局协作引擎实例
collaboration_engine = AgentCollaborationEngine()