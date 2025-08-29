"""
Advanced Reasoning Engine - 先进推理引擎
实现Chain-of-Thought、Plan-Execute、Self-Reflection等推理模式
"""

from typing import List, Dict, Any, Optional, Union, Tuple
import asyncio
import json
import time
from dataclasses import dataclass, field
from enum import Enum
from datetime import datetime

from app.core.logging import LoggerMixin
from app.core.exceptions import AgentException
from app.adapters.base import BaseAIAdapter
from app.services.tool_service import ToolService


class ReasoningMode(Enum):
    """推理模式"""
    CHAIN_OF_THOUGHT = "cot"           # 思维链推理
    PLAN_AND_EXECUTE = "plan_execute"  # 计划执行模式
    REACT = "react"                    # ReAct模式 (Reasoning + Acting)
    SELF_REFLECTION = "reflection"      # 自我反思模式
    TREE_OF_THOUGHTS = "tot"           # 思维树探索
    AUTO = "auto"                      # 自动选择模式


class TaskComplexity(Enum):
    """任务复杂度"""
    SIMPLE = "simple"      # 简单任务，直接回答
    MODERATE = "moderate"  # 中等任务，需要推理
    COMPLEX = "complex"    # 复杂任务，需要规划
    EXPERT = "expert"      # 专家级任务，需要深度分析


@dataclass
class ReasoningStep:
    """推理步骤"""
    step_id: str
    step_type: str  # thought, action, observation, reflection
    content: str
    timestamp: datetime = field(default_factory=datetime.now)
    confidence: float = 1.0
    metadata: Dict[str, Any] = field(default_factory=dict)


@dataclass
class ReasoningChain:
    """推理链"""
    chain_id: str
    mode: ReasoningMode
    steps: List[ReasoningStep] = field(default_factory=list)
    final_answer: Optional[str] = None
    success: bool = False
    total_time: float = 0.0
    metadata: Dict[str, Any] = field(default_factory=dict)


@dataclass
class ExecutionPlan:
    """执行计划"""
    plan_id: str
    goal: str
    tasks: List[Dict[str, Any]] = field(default_factory=list)
    current_task_index: int = 0
    status: str = "pending"  # pending, executing, completed, failed
    results: List[Any] = field(default_factory=list)


class ReasoningEngine(LoggerMixin):
    """推理引擎"""
    
    def __init__(
        self,
        ai_adapter: BaseAIAdapter,
        tool_service: Optional[ToolService] = None,
        default_mode: ReasoningMode = ReasoningMode.AUTO
    ):
        super().__init__()
        self.ai_adapter = ai_adapter
        self.tool_service = tool_service
        self.default_mode = default_mode
        
        # 推理链存储
        self.reasoning_chains: Dict[str, ReasoningChain] = {}
        self.execution_plans: Dict[str, ExecutionPlan] = {}
        
        # 推理模式配置
        self.mode_configs = {
            ReasoningMode.CHAIN_OF_THOUGHT: {
                "max_steps": 10,
                "require_explanation": True,
                "step_verification": True
            },
            ReasoningMode.PLAN_AND_EXECUTE: {
                "max_tasks": 20,
                "allow_task_modification": True,
                "require_task_validation": True
            },
            ReasoningMode.REACT: {
                "max_iterations": 15,
                "action_timeout": 30,
                "observation_required": True
            },
            ReasoningMode.SELF_REFLECTION: {
                "reflection_frequency": 3,
                "max_reflections": 5,
                "improvement_threshold": 0.1
            }
        }
    
    async def reason(
        self,
        query: str,
        context: Optional[Dict[str, Any]] = None,
        mode: Optional[ReasoningMode] = None,
        user_id: Optional[str] = None
    ) -> ReasoningChain:
        """执行推理"""
        
        start_time = time.time()
        
        # 选择推理模式
        selected_mode = mode or await self._select_reasoning_mode(query, context)
        
        # 创建推理链
        chain_id = f"reasoning_{int(time.time())}_{hash(query) % 10000}"
        reasoning_chain = ReasoningChain(
            chain_id=chain_id,
            mode=selected_mode,
            metadata={
                "query": query,
                "user_id": user_id,
                "start_time": start_time,
                "context": context or {}
            }
        )
        
        self.reasoning_chains[chain_id] = reasoning_chain
        
        try:
            self.logger.info(f"Starting reasoning with mode: {selected_mode.value}")
            
            # 根据模式执行推理
            if selected_mode == ReasoningMode.CHAIN_OF_THOUGHT:
                await self._chain_of_thought_reasoning(reasoning_chain, query, context)
            elif selected_mode == ReasoningMode.PLAN_AND_EXECUTE:
                await self._plan_and_execute_reasoning(reasoning_chain, query, context)
            elif selected_mode == ReasoningMode.REACT:
                await self._react_reasoning(reasoning_chain, query, context)
            elif selected_mode == ReasoningMode.SELF_REFLECTION:
                await self._self_reflection_reasoning(reasoning_chain, query, context)
            elif selected_mode == ReasoningMode.TREE_OF_THOUGHTS:
                await self._tree_of_thoughts_reasoning(reasoning_chain, query, context)
            
            reasoning_chain.success = True
            
        except Exception as e:
            self.logger.error(f"Reasoning failed: {e}")
            reasoning_chain.success = False
            
            # 添加错误步骤
            error_step = ReasoningStep(
                step_id=f"error_{len(reasoning_chain.steps)}",
                step_type="error",
                content=f"Reasoning failed: {str(e)}",
                confidence=0.0
            )
            reasoning_chain.steps.append(error_step)
        
        finally:
            reasoning_chain.total_time = time.time() - start_time
            self.logger.info(f"Reasoning completed in {reasoning_chain.total_time:.2f}s")
        
        return reasoning_chain
    
    async def _select_reasoning_mode(
        self,
        query: str,
        context: Optional[Dict[str, Any]] = None
    ) -> ReasoningMode:
        """智能选择推理模式"""
        
        # 分析任务复杂度
        complexity = await self._analyze_task_complexity(query, context)
        
        # 检查是否需要工具调用
        needs_tools = await self._needs_tool_usage(query, context)
        
        # 检查是否需要多步规划
        needs_planning = await self._needs_multi_step_planning(query, context)
        
        # 根据分析结果选择模式
        if needs_planning and complexity in [TaskComplexity.COMPLEX, TaskComplexity.EXPERT]:
            return ReasoningMode.PLAN_AND_EXECUTE
        elif needs_tools:
            return ReasoningMode.REACT
        elif complexity == TaskComplexity.EXPERT:
            return ReasoningMode.TREE_OF_THOUGHTS
        elif complexity in [TaskComplexity.MODERATE, TaskComplexity.COMPLEX]:
            return ReasoningMode.CHAIN_OF_THOUGHT
        else:
            # 简单任务，直接使用CoT
            return ReasoningMode.CHAIN_OF_THOUGHT
    
    async def _analyze_task_complexity(
        self,
        query: str,
        context: Optional[Dict[str, Any]] = None
    ) -> TaskComplexity:
        """分析任务复杂度"""
        
        # 复杂度指标
        complexity_indicators = {
            "multi_step_keywords": ["步骤", "首先", "然后", "最后", "计划", "分析", "step", "first", "then", "finally"],
            "expert_keywords": ["算法", "设计", "架构", "优化", "研究", "algorithm", "design", "architecture", "optimize"],
            "tool_keywords": ["搜索", "计算", "查询", "获取", "处理", "search", "calculate", "query", "fetch", "process"],
            "complex_patterns": ["如何", "为什么", "比较", "评估", "分析", "how", "why", "compare", "evaluate", "analyze"]
        }
        
        query_lower = query.lower()
        
        # 计算复杂度分数
        score = 0
        
        # 多步骤任务
        multi_step_matches = sum(1 for keyword in complexity_indicators["multi_step_keywords"] 
                               if keyword in query_lower)
        if multi_step_matches >= 2:
            score += 2
        elif multi_step_matches == 1:
            score += 1
        
        # 专家级任务
        expert_matches = sum(1 for keyword in complexity_indicators["expert_keywords"] 
                           if keyword in query_lower)
        score += expert_matches * 2
        
        # 工具使用需求
        tool_matches = sum(1 for keyword in complexity_indicators["tool_keywords"] 
                         if keyword in query_lower)
        score += tool_matches
        
        # 复杂模式
        pattern_matches = sum(1 for pattern in complexity_indicators["complex_patterns"] 
                            if pattern in query_lower)
        score += pattern_matches
        
        # 查询长度影响
        if len(query) > 200:
            score += 2
        elif len(query) > 100:
            score += 1
        
        # 根据分数确定复杂度
        if score >= 8:
            return TaskComplexity.EXPERT
        elif score >= 5:
            return TaskComplexity.COMPLEX
        elif score >= 2:
            return TaskComplexity.MODERATE
        else:
            return TaskComplexity.SIMPLE
    
    async def _needs_tool_usage(
        self,
        query: str,
        context: Optional[Dict[str, Any]] = None
    ) -> bool:
        """判断是否需要工具调用"""
        
        if not self.tool_service:
            return False
        
        tool_indicators = [
            "搜索", "查询", "获取", "计算", "处理", "下载", "上传",
            "search", "query", "fetch", "calculate", "process", "download", "upload",
            "当前", "最新", "实时", "current", "latest", "real-time"
        ]
        
        query_lower = query.lower()
        return any(indicator in query_lower for indicator in tool_indicators)
    
    async def _needs_multi_step_planning(
        self,
        query: str,
        context: Optional[Dict[str, Any]] = None
    ) -> bool:
        """判断是否需要多步骤规划"""
        
        planning_indicators = [
            "计划", "规划", "步骤", "流程", "方案", "策略",
            "plan", "planning", "steps", "process", "strategy", "approach",
            "如何实现", "怎样做", "分步骤", "how to implement", "step by step"
        ]
        
        query_lower = query.lower()
        return any(indicator in query_lower for indicator in planning_indicators)
    
    async def _chain_of_thought_reasoning(
        self,
        chain: ReasoningChain,
        query: str,
        context: Optional[Dict[str, Any]] = None
    ):
        """链式思维推理"""
        
        config = self.mode_configs[ReasoningMode.CHAIN_OF_THOUGHT]
        
        # 构建CoT提示
        cot_prompt = f"""请使用链式思维方法回答以下问题。请按照以下格式思考：

问题: {query}

思考过程:
1. 首先，我需要理解问题的核心...
2. 然后，我需要分析相关因素...
3. 接下来，我需要考虑可能的解决方案...
4. 最后，我需要得出结论...

请详细展示你的思考步骤，每一步都要有清晰的推理过程。"""
        
        try:
            # 获取AI响应
            response = await self.ai_adapter.generate_response(
                messages=[{"role": "user", "content": cot_prompt}],
                temperature=0.7,
                max_tokens=2000
            )
            
            # 解析思维链
            await self._parse_chain_of_thought(chain, response)
            
        except Exception as e:
            raise AgentException(f"Chain-of-Thought reasoning failed: {str(e)}")
    
    async def _parse_chain_of_thought(self, chain: ReasoningChain, response: str):
        """解析思维链响应"""
        
        lines = response.split('\n')
        current_step = 1
        current_content = []
        
        for line in lines:
            line = line.strip()
            if not line:
                continue
            
            # 检测步骤标记
            if any(marker in line.lower() for marker in ['思考', '分析', '考虑', '结论', 'think', 'analyze', 'consider', 'conclude']):
                # 保存前一步
                if current_content:
                    step = ReasoningStep(
                        step_id=f"cot_step_{current_step}",
                        step_type="thought",
                        content='\n'.join(current_content).strip(),
                        confidence=0.8
                    )
                    chain.steps.append(step)
                    current_step += 1
                    current_content = []
            
            current_content.append(line)
        
        # 保存最后一步
        if current_content:
            final_content = '\n'.join(current_content).strip()
            if final_content:
                step = ReasoningStep(
                    step_id=f"cot_final",
                    step_type="conclusion",
                    content=final_content,
                    confidence=0.9
                )
                chain.steps.append(step)
                chain.final_answer = final_content
    
    async def _plan_and_execute_reasoning(
        self,
        chain: ReasoningChain,
        query: str,
        context: Optional[Dict[str, Any]] = None
    ):
        """计划执行推理"""
        
        # 第一步：制定计划
        plan_prompt = f"""请为以下任务制定详细的执行计划：

任务: {query}

请按照以下格式输出计划：
```json
{{
  "goal": "任务目标描述",
  "tasks": [
    {{"id": 1, "description": "任务描述", "dependencies": [], "estimated_time": "预估时间"}},
    {{"id": 2, "description": "任务描述", "dependencies": [1], "estimated_time": "预估时间"}}
  ]
}}
```

请确保计划完整、可执行，每个子任务都有明确的目标。"""
        
        try:
            # 获取计划
            plan_response = await self.ai_adapter.generate_response(
                messages=[{"role": "user", "content": plan_prompt}],
                temperature=0.3,
                max_tokens=1000
            )
            
            # 解析计划
            plan = await self._parse_execution_plan(plan_response, query)
            
            # 添加计划步骤
            plan_step = ReasoningStep(
                step_id="plan_creation",
                step_type="planning",
                content=f"Created execution plan with {len(plan.tasks)} tasks",
                metadata={"plan_id": plan.plan_id}
            )
            chain.steps.append(plan_step)
            
            # 执行计划
            await self._execute_plan(chain, plan)
            
        except Exception as e:
            raise AgentException(f"Plan-and-Execute reasoning failed: {str(e)}")
    
    async def _parse_execution_plan(self, response: str, goal: str) -> ExecutionPlan:
        """解析执行计划"""
        
        plan_id = f"plan_{int(time.time())}"
        
        try:
            # 提取JSON部分
            json_start = response.find('{')
            json_end = response.rfind('}') + 1
            
            if json_start >= 0 and json_end > json_start:
                json_str = response[json_start:json_end]
                plan_data = json.loads(json_str)
                
                plan = ExecutionPlan(
                    plan_id=plan_id,
                    goal=plan_data.get("goal", goal),
                    tasks=plan_data.get("tasks", [])
                )
                
                self.execution_plans[plan_id] = plan
                return plan
        
        except (json.JSONDecodeError, KeyError) as e:
            self.logger.warning(f"Failed to parse execution plan JSON: {e}")
        
        # 回退：创建简单计划
        simple_plan = ExecutionPlan(
            plan_id=plan_id,
            goal=goal,
            tasks=[{
                "id": 1,
                "description": f"Complete task: {goal}",
                "dependencies": [],
                "estimated_time": "unknown"
            }]
        )
        
        self.execution_plans[plan_id] = simple_plan
        return simple_plan
    
    async def _execute_plan(self, chain: ReasoningChain, plan: ExecutionPlan):
        """执行计划"""
        
        plan.status = "executing"
        
        for i, task in enumerate(plan.tasks):
            plan.current_task_index = i
            
            # 检查依赖
            if not await self._check_task_dependencies(task, plan.results):
                continue
            
            # 执行任务
            task_result = await self._execute_task(task, plan)
            plan.results.append(task_result)
            
            # 添加执行步骤
            exec_step = ReasoningStep(
                step_id=f"execute_task_{task['id']}",
                step_type="execution",
                content=f"Executed task: {task['description']}",
                metadata={
                    "task_id": task['id'],
                    "result": task_result
                }
            )
            chain.steps.append(exec_step)
        
        plan.status = "completed"
        
        # 生成最终答案
        final_answer = await self._generate_plan_summary(plan)
        chain.final_answer = final_answer
    
    async def _check_task_dependencies(self, task: Dict[str, Any], completed_results: List[Any]) -> bool:
        """检查任务依赖"""
        
        dependencies = task.get("dependencies", [])
        return len(completed_results) >= max(dependencies) if dependencies else True
    
    async def _execute_task(self, task: Dict[str, Any], plan: ExecutionPlan) -> str:
        """执行单个任务"""
        
        task_prompt = f"""请执行以下任务：

任务描述: {task['description']}
总体目标: {plan.goal}

请提供执行结果和关键信息。"""
        
        try:
            result = await self.ai_adapter.generate_response(
                messages=[{"role": "user", "content": task_prompt}],
                temperature=0.5,
                max_tokens=500
            )
            return result
        
        except Exception as e:
            return f"Task execution failed: {str(e)}"
    
    async def _generate_plan_summary(self, plan: ExecutionPlan) -> str:
        """生成计划执行总结"""
        
        summary_prompt = f"""请总结以下计划的执行结果：

目标: {plan.goal}
执行的任务数: {len(plan.tasks)}
所有任务结果:
{chr(10).join([f"{i+1}. {result}" for i, result in enumerate(plan.results)])}

请提供一个综合性的总结回答。"""
        
        try:
            summary = await self.ai_adapter.generate_response(
                messages=[{"role": "user", "content": summary_prompt}],
                temperature=0.3,
                max_tokens=800
            )
            return summary
        
        except Exception as e:
            return f"Plan summary generation failed: {str(e)}"
    
    async def _react_reasoning(
        self,
        chain: ReasoningChain,
        query: str,
        context: Optional[Dict[str, Any]] = None
    ):
        """ReAct推理模式"""
        
        if not self.tool_service:
            raise AgentException("ReAct mode requires tool service")
        
        config = self.mode_configs[ReasoningMode.REACT]
        max_iterations = config["max_iterations"]
        
        react_prompt = f"""你是一个能够进行推理并采取行动的AI助手。
请使用以下格式回答问题：

Thought: 我需要思考这个问题...
Action: [工具名称]
Action Input: [工具输入参数]
Observation: [工具执行结果]
... (这个Thought/Action/Observation可以重复多次)
Thought: 现在我知道最终答案了
Final Answer: [最终答案]

可用工具: {await self._get_available_tools()}

问题: {query}"""
        
        messages = [{"role": "user", "content": react_prompt}]
        iteration = 0
        
        while iteration < max_iterations:
            iteration += 1
            
            # 获取AI响应
            response = await self.ai_adapter.generate_response(
                messages=messages,
                temperature=0.7,
                max_tokens=1000
            )
            
            messages.append({"role": "assistant", "content": response})
            
            # 解析响应
            action_result = await self._parse_react_response(chain, response, iteration)
            
            if action_result["is_final"]:
                chain.final_answer = action_result["final_answer"]
                break
            elif action_result["action_needed"]:
                # 执行工具
                observation = await self._execute_react_action(
                    action_result["action_name"],
                    action_result["action_input"]
                )
                
                # 添加观察结果
                obs_step = ReasoningStep(
                    step_id=f"react_obs_{iteration}",
                    step_type="observation",
                    content=observation,
                    metadata={
                        "action_name": action_result["action_name"],
                        "action_input": action_result["action_input"]
                    }
                )
                chain.steps.append(obs_step)
                
                # 添加观察到消息
                messages.append({"role": "user", "content": f"Observation: {observation}"})
        
        if not chain.final_answer:
            chain.final_answer = "Unable to complete ReAct reasoning within maximum iterations."
    
    async def _parse_react_response(
        self,
        chain: ReasoningChain,
        response: str,
        iteration: int
    ) -> Dict[str, Any]:
        """解析ReAct响应"""
        
        lines = response.strip().split('\n')
        current_step = None
        action_name = None
        action_input = None
        is_final = False
        final_answer = None
        
        for line in lines:
            line = line.strip()
            
            if line.startswith("Thought:"):
                thought_content = line[8:].strip()
                thought_step = ReasoningStep(
                    step_id=f"react_thought_{iteration}",
                    step_type="thought",
                    content=thought_content
                )
                chain.steps.append(thought_step)
            
            elif line.startswith("Action:"):
                action_name = line[7:].strip()
            
            elif line.startswith("Action Input:"):
                action_input = line[13:].strip()
            
            elif line.startswith("Final Answer:"):
                final_answer = line[13:].strip()
                is_final = True
        
        return {
            "is_final": is_final,
            "final_answer": final_answer,
            "action_needed": action_name is not None and action_input is not None,
            "action_name": action_name,
            "action_input": action_input
        }
    
    async def _execute_react_action(self, action_name: str, action_input: str) -> str:
        """执行ReAct动作"""
        
        try:
            if self.tool_service:
                result = await self.tool_service.execute_tool(action_name, action_input)
                return str(result)
            else:
                return f"Tool execution not available: {action_name}"
        
        except Exception as e:
            return f"Tool execution error: {str(e)}"
    
    async def _get_available_tools(self) -> str:
        """获取可用工具列表"""
        
        if self.tool_service:
            tools = self.tool_service.list_available_tools()
            return ", ".join([f"{name}: {desc}" for name, desc in tools.items()])
        else:
            return "No tools available"
    
    async def _self_reflection_reasoning(
        self,
        chain: ReasoningChain,
        query: str,
        context: Optional[Dict[str, Any]] = None
    ):
        """自我反思推理"""
        
        # 初始回答
        initial_response = await self.ai_adapter.generate_response(
            messages=[{"role": "user", "content": query}],
            temperature=0.7
        )
        
        # 添加初始步骤
        initial_step = ReasoningStep(
            step_id="initial_response",
            step_type="initial_answer",
            content=initial_response
        )
        chain.steps.append(initial_step)
        
        current_answer = initial_response
        
        # 反思循环
        for reflection_round in range(self.mode_configs[ReasoningMode.SELF_REFLECTION]["max_reflections"]):
            
            # 生成反思提示
            reflection_prompt = f"""请仔细审视以下回答，并进行自我反思：

原问题: {query}
当前回答: {current_answer}

请按照以下方面进行反思：
1. 回答是否完整和准确？
2. 是否有遗漏的重要信息？
3. 逻辑是否清晰连贯？
4. 是否需要改进或补充？

如果需要改进，请提供改进后的回答。如果当前回答已经足够好，请确认。"""
            
            reflection_response = await self.ai_adapter.generate_response(
                messages=[{"role": "user", "content": reflection_prompt}],
                temperature=0.5
            )
            
            # 添加反思步骤
            reflection_step = ReasoningStep(
                step_id=f"reflection_{reflection_round + 1}",
                step_type="reflection",
                content=reflection_response
            )
            chain.steps.append(reflection_step)
            
            # 检查是否需要改进
            if "改进" in reflection_response or "improve" in reflection_response.lower():
                # 提取改进的回答
                improved_answer = await self._extract_improved_answer(reflection_response, current_answer)
                if improved_answer != current_answer:
                    current_answer = improved_answer
                    
                    # 添加改进步骤
                    improved_step = ReasoningStep(
                        step_id=f"improved_answer_{reflection_round + 1}",
                        step_type="improvement",
                        content=improved_answer
                    )
                    chain.steps.append(improved_step)
            else:
                # 反思确认当前答案足够好
                break
        
        chain.final_answer = current_answer
    
    async def _extract_improved_answer(self, reflection_response: str, current_answer: str) -> str:
        """提取改进后的答案"""
        
        # 简单的改进答案提取逻辑
        lines = reflection_response.split('\n')
        
        # 寻找改进部分
        for i, line in enumerate(lines):
            if any(keyword in line.lower() for keyword in ['改进', 'improve', '更好', 'better', '修正', 'correct']):
                # 返回后续内容作为改进答案
                improved_content = '\n'.join(lines[i+1:]).strip()
                if improved_content and len(improved_content) > 50:
                    return improved_content
        
        return current_answer
    
    async def _tree_of_thoughts_reasoning(
        self,
        chain: ReasoningChain,
        query: str,
        context: Optional[Dict[str, Any]] = None
    ):
        """思维树探索推理"""
        
        # 生成多个思路分支
        branches_prompt = f"""对于以下复杂问题，请生成3个不同的解决思路：

问题: {query}

请为每个思路提供：
1. 思路描述
2. 解决步骤
3. 优缺点分析

格式：
思路1: [描述]
步骤: [具体步骤]
优缺点: [分析]

思路2: [描述]
...

思路3: [描述]
..."""
        
        branches_response = await self.ai_adapter.generate_response(
            messages=[{"role": "user", "content": branches_prompt}],
            temperature=0.8,
            max_tokens=2000
        )
        
        # 解析分支
        branches = await self._parse_thought_branches(branches_response)
        
        # 为每个分支添加步骤
        for i, branch in enumerate(branches):
            branch_step = ReasoningStep(
                step_id=f"branch_{i+1}",
                step_type="branch_exploration",
                content=branch,
                metadata={"branch_index": i+1}
            )
            chain.steps.append(branch_step)
        
        # 评估和选择最佳分支
        evaluation_prompt = f"""请评估以下思路分支，并选择最佳方案：

{chr(10).join([f"分支{i+1}: {branch}" for i, branch in enumerate(branches)])}

请选择最佳分支并说明理由，然后提供最终的详细解决方案。"""
        
        final_response = await self.ai_adapter.generate_response(
            messages=[{"role": "user", "content": evaluation_prompt}],
            temperature=0.3,
            max_tokens=1500
        )
        
        # 添加评估步骤
        evaluation_step = ReasoningStep(
            step_id="branch_evaluation",
            step_type="evaluation",
            content=final_response
        )
        chain.steps.append(evaluation_step)
        
        chain.final_answer = final_response
    
    async def _parse_thought_branches(self, response: str) -> List[str]:
        """解析思维分支"""
        
        branches = []
        current_branch = []
        
        lines = response.split('\n')
        
        for line in lines:
            line = line.strip()
            if line.startswith('思路') and current_branch:
                # 保存前一个分支
                branches.append('\n'.join(current_branch))
                current_branch = []
            
            if line:
                current_branch.append(line)
        
        # 保存最后一个分支
        if current_branch:
            branches.append('\n'.join(current_branch))
        
        return branches[:3]  # 限制最多3个分支
    
    def get_reasoning_history(self, chain_id: str) -> Optional[ReasoningChain]:
        """获取推理历史"""
        return self.reasoning_chains.get(chain_id)
    
    def list_reasoning_chains(self) -> List[str]:
        """列出所有推理链ID"""
        return list(self.reasoning_chains.keys())
    
    async def explain_reasoning(self, chain_id: str) -> str:
        """解释推理过程"""
        
        chain = self.reasoning_chains.get(chain_id)
        if not chain:
            return "Reasoning chain not found."
        
        explanation = f"""推理过程解释 (模式: {chain.mode.value})

总体信息:
- 推理时间: {chain.total_time:.2f}秒
- 步骤数量: {len(chain.steps)}
- 成功状态: {"成功" if chain.success else "失败"}

详细步骤:
"""
        
        for i, step in enumerate(chain.steps, 1):
            explanation += f"\n{i}. [{step.step_type}] {step.content[:100]}..."
            if step.confidence < 1.0:
                explanation += f" (置信度: {step.confidence:.2f})"
        
        if chain.final_answer:
            explanation += f"\n\n最终答案:\n{chain.final_answer}"
        
        return explanation