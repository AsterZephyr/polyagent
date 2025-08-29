"""
Oxy Workflow Engine - 工作流引擎
实现智能体协作和任务编排
基于OxyGent的动态规划范式
"""

import asyncio
import time
from typing import Dict, List, Any, Optional, Union, Set
from dataclasses import dataclass, field
from enum import Enum
import json
import uuid

from .core import (
    BaseOxy, OxyContext, OxyResult, OxyMessage, OxyStatus, OxyType,
    oxy_registry
)
from ..core.logging import LoggerMixin
from ..core.exceptions import AgentException

class WorkflowStatus(Enum):
    """工作流状态"""
    PENDING = "pending"
    RUNNING = "running"
    COMPLETED = "completed"
    FAILED = "failed"
    PAUSED = "paused"
    CANCELLED = "cancelled"

class ExecutionMode(Enum):
    """执行模式"""
    SEQUENTIAL = "sequential"      # 顺序执行
    PARALLEL = "parallel"          # 并行执行
    CONDITIONAL = "conditional"    # 条件执行
    LOOP = "loop"                 # 循环执行
    DYNAMIC = "dynamic"           # 动态执行

@dataclass
class WorkflowStep:
    """工作流步骤"""
    step_id: str
    oxy_id: str
    name: str = ""
    description: str = ""
    execution_mode: ExecutionMode = ExecutionMode.SEQUENTIAL
    condition: Optional[str] = None
    loop_condition: Optional[str] = None
    max_iterations: int = 100
    timeout: float = 300.0  # 5分钟超时
    retry_count: int = 3
    inputs: Dict[str, Any] = field(default_factory=dict)
    outputs: Dict[str, str] = field(default_factory=dict)  # output_key -> variable_name
    
    # 执行状态
    status: WorkflowStatus = WorkflowStatus.PENDING
    start_time: Optional[float] = None
    end_time: Optional[float] = None
    execution_time: float = 0.0
    iterations: int = 0
    current_retry: int = 0
    result: Optional[OxyResult] = None
    error: Optional[Exception] = None

@dataclass
class WorkflowDefinition:
    """工作流定义"""
    workflow_id: str
    name: str
    description: str = ""
    steps: List[WorkflowStep] = field(default_factory=list)
    global_timeout: float = 3600.0  # 1小时
    max_concurrent_steps: int = 10
    
    # 元数据
    created_by: Optional[str] = None
    created_at: float = field(default_factory=time.time)
    version: str = "1.0"
    tags: List[str] = field(default_factory=list)

@dataclass
class WorkflowExecution:
    """工作流执行实例"""
    execution_id: str
    workflow_id: str
    context: OxyContext
    status: WorkflowStatus = WorkflowStatus.PENDING
    start_time: Optional[float] = None
    end_time: Optional[float] = None
    execution_time: float = 0.0
    
    # 执行状态
    completed_steps: Set[str] = field(default_factory=set)
    failed_steps: Set[str] = field(default_factory=set)
    running_steps: Set[str] = field(default_factory=set)
    
    # 结果
    results: Dict[str, OxyResult] = field(default_factory=dict)
    final_result: Optional[Any] = None
    error: Optional[Exception] = None

class ConditionalEvaluator:
    """条件评估器"""
    
    @staticmethod
    def evaluate(condition: str, context: OxyContext, step_results: Dict[str, OxyResult]) -> bool:
        """评估条件表达式"""
        if not condition:
            return True
        
        # 构建评估环境
        eval_env = {
            "context": context,
            "variables": context.variables,
            "results": step_results,
        }
        
        # 添加常用函数
        eval_env.update({
            "len": len,
            "str": str,
            "int": int,
            "float": float,
            "bool": bool,
            "isinstance": isinstance,
        })
        
        try:
            # 安全的条件评估
            return bool(eval(condition, {"__builtins__": {}}, eval_env))
        except Exception as e:
            # 条件评估失败时默认返回False
            return False

class WorkflowEngine(LoggerMixin):
    """工作流引擎"""
    
    def __init__(self):
        super().__init__()
        self.workflows: Dict[str, WorkflowDefinition] = {}
        self.executions: Dict[str, WorkflowExecution] = {}
        self.running_executions: Set[str] = set()
        
        self.condition_evaluator = ConditionalEvaluator()
        
        # 执行统计
        self.total_executions = 0
        self.successful_executions = 0
        self.failed_executions = 0
    
    def register_workflow(self, workflow: WorkflowDefinition) -> str:
        """注册工作流"""
        self.workflows[workflow.workflow_id] = workflow
        self.logger.info(f"Registered workflow: {workflow.name} ({workflow.workflow_id})")
        return workflow.workflow_id
    
    def unregister_workflow(self, workflow_id: str) -> bool:
        """注销工作流"""
        if workflow_id in self.workflows:
            del self.workflows[workflow_id]
            self.logger.info(f"Unregistered workflow: {workflow_id}")
            return True
        return False
    
    def get_workflow(self, workflow_id: str) -> Optional[WorkflowDefinition]:
        """获取工作流定义"""
        return self.workflows.get(workflow_id)
    
    async def execute_workflow(
        self,
        workflow_id: str,
        context: OxyContext,
        inputs: Dict[str, Any] = None
    ) -> WorkflowExecution:
        """执行工作流"""
        
        workflow = self.get_workflow(workflow_id)
        if not workflow:
            raise AgentException(f"Workflow {workflow_id} not found")
        
        # 创建执行实例
        execution = WorkflowExecution(
            execution_id=str(uuid.uuid4()),
            workflow_id=workflow_id,
            context=context,
            start_time=time.time()
        )
        
        # 添加输入到上下文
        if inputs:
            context.variables.update(inputs)
        
        self.executions[execution.execution_id] = execution
        self.running_executions.add(execution.execution_id)
        self.total_executions += 1
        
        self.logger.info(f"Starting workflow execution: {execution.execution_id}")
        
        try:
            execution.status = WorkflowStatus.RUNNING
            
            # 执行工作流
            await self._execute_workflow_steps(workflow, execution)
            
            # 完成执行
            execution.status = WorkflowStatus.COMPLETED
            execution.end_time = time.time()
            execution.execution_time = execution.end_time - execution.start_time
            
            self.successful_executions += 1
            self.logger.info(f"Workflow execution completed: {execution.execution_id}")
            
        except Exception as e:
            execution.status = WorkflowStatus.FAILED
            execution.error = e
            execution.end_time = time.time()
            execution.execution_time = execution.end_time - execution.start_time
            
            self.failed_executions += 1
            self.logger.error(f"Workflow execution failed: {execution.execution_id}, error: {e}")
            
        finally:
            self.running_executions.discard(execution.execution_id)
        
        return execution
    
    async def _execute_workflow_steps(
        self,
        workflow: WorkflowDefinition,
        execution: WorkflowExecution
    ):
        """执行工作流步骤"""
        
        step_results = {}
        
        # 按依赖关系排序步骤
        execution_order = self._get_execution_order(workflow.steps)
        
        # 执行步骤
        for step in execution_order:
            await self._execute_step(step, execution, step_results)
    
    def _get_execution_order(self, steps: List[WorkflowStep]) -> List[WorkflowStep]:
        """获取步骤执行顺序"""
        # 简化实现：按定义顺序执行
        # 实际应该根据依赖关系进行拓扑排序
        return steps
    
    async def _execute_step(
        self,
        step: WorkflowStep,
        execution: WorkflowExecution,
        step_results: Dict[str, OxyResult]
    ):
        """执行单个步骤"""
        
        self.logger.debug(f"Executing step: {step.name} ({step.step_id})")
        
        # 检查条件
        if step.condition and not self.condition_evaluator.evaluate(
            step.condition, execution.context, step_results
        ):
            self.logger.debug(f"Step condition not met: {step.step_id}")
            return
        
        step.status = WorkflowStatus.RUNNING
        step.start_time = time.time()
        execution.running_steps.add(step.step_id)
        
        try:
            # 获取Oxy组件
            oxy_component = oxy_registry.get(step.oxy_id)
            if not oxy_component:
                raise AgentException(f"Oxy component {step.oxy_id} not found")
            
            # 准备输入参数
            kwargs = step.inputs.copy()
            
            # 处理循环执行
            if step.execution_mode == ExecutionMode.LOOP:
                await self._execute_loop_step(step, oxy_component, execution, kwargs)
            else:
                # 执行组件
                result = await asyncio.wait_for(
                    oxy_component.execute(execution.context, **kwargs),
                    timeout=step.timeout
                )
                
                step.result = result
                step_results[step.step_id] = result
                execution.results[step.step_id] = result
                
                # 处理输出映射
                if result.success and step.outputs:
                    self._map_outputs(result, step.outputs, execution.context)
            
            step.status = WorkflowStatus.COMPLETED
            execution.completed_steps.add(step.step_id)
            
        except asyncio.TimeoutError:
            step.status = WorkflowStatus.FAILED
            step.error = AgentException(f"Step {step.step_id} timeout")
            execution.failed_steps.add(step.step_id)
            self.logger.error(f"Step timeout: {step.step_id}")
            
        except Exception as e:
            step.status = WorkflowStatus.FAILED
            step.error = e
            execution.failed_steps.add(step.step_id)
            self.logger.error(f"Step execution failed: {step.step_id}, error: {e}")
            
            # 重试逻辑
            if step.current_retry < step.retry_count:
                step.current_retry += 1
                self.logger.info(f"Retrying step {step.step_id}, attempt {step.current_retry}")
                await asyncio.sleep(1)  # 重试延迟
                await self._execute_step(step, execution, step_results)
                return
        
        finally:
            step.end_time = time.time()
            step.execution_time = step.end_time - step.start_time
            execution.running_steps.discard(step.step_id)
    
    async def _execute_loop_step(
        self,
        step: WorkflowStep,
        oxy_component: BaseOxy,
        execution: WorkflowExecution,
        kwargs: Dict[str, Any]
    ):
        """执行循环步骤"""
        
        step_results = {}
        
        while step.iterations < step.max_iterations:
            step.iterations += 1
            
            # 检查循环条件
            if step.loop_condition and not self.condition_evaluator.evaluate(
                step.loop_condition, execution.context, step_results
            ):
                break
            
            # 执行组件
            result = await oxy_component.execute(execution.context, **kwargs)
            step_results[f"iteration_{step.iterations}"] = result
            
            if not result.success:
                step.error = result.error
                break
            
            # 处理输出映射
            if step.outputs:
                self._map_outputs(result, step.outputs, execution.context)
        
        # 设置最终结果
        step.result = OxyResult(
            success=step.error is None,
            data=step_results,
            message=f"Loop completed with {step.iterations} iterations",
            metadata={"iterations": step.iterations}
        )
    
    def _map_outputs(
        self,
        result: OxyResult,
        output_mapping: Dict[str, str],
        context: OxyContext
    ):
        """映射输出到上下文变量"""
        
        for output_key, variable_name in output_mapping.items():
            if hasattr(result.data, output_key):
                value = getattr(result.data, output_key)
            elif isinstance(result.data, dict) and output_key in result.data:
                value = result.data[output_key]
            else:
                value = result.data
            
            context.set_variable(variable_name, value)
    
    async def pause_execution(self, execution_id: str) -> bool:
        """暂停执行"""
        execution = self.executions.get(execution_id)
        if execution and execution.status == WorkflowStatus.RUNNING:
            execution.status = WorkflowStatus.PAUSED
            self.logger.info(f"Paused execution: {execution_id}")
            return True
        return False
    
    async def resume_execution(self, execution_id: str) -> bool:
        """恢复执行"""
        execution = self.executions.get(execution_id)
        if execution and execution.status == WorkflowStatus.PAUSED:
            execution.status = WorkflowStatus.RUNNING
            self.logger.info(f"Resumed execution: {execution_id}")
            return True
        return False
    
    async def cancel_execution(self, execution_id: str) -> bool:
        """取消执行"""
        execution = self.executions.get(execution_id)
        if execution and execution.status in [WorkflowStatus.RUNNING, WorkflowStatus.PAUSED]:
            execution.status = WorkflowStatus.CANCELLED
            execution.end_time = time.time()
            execution.execution_time = execution.end_time - (execution.start_time or time.time())
            
            self.running_executions.discard(execution_id)
            self.logger.info(f"Cancelled execution: {execution_id}")
            return True
        return False
    
    def get_execution(self, execution_id: str) -> Optional[WorkflowExecution]:
        """获取执行实例"""
        return self.executions.get(execution_id)
    
    def get_execution_status(self, execution_id: str) -> Optional[WorkflowStatus]:
        """获取执行状态"""
        execution = self.executions.get(execution_id)
        return execution.status if execution else None
    
    def list_executions(
        self,
        workflow_id: Optional[str] = None,
        status: Optional[WorkflowStatus] = None
    ) -> List[WorkflowExecution]:
        """列出执行实例"""
        executions = list(self.executions.values())
        
        if workflow_id:
            executions = [e for e in executions if e.workflow_id == workflow_id]
        
        if status:
            executions = [e for e in executions if e.status == status]
        
        return executions
    
    def get_engine_stats(self) -> Dict[str, Any]:
        """获取引擎统计信息"""
        return {
            "total_workflows": len(self.workflows),
            "total_executions": self.total_executions,
            "successful_executions": self.successful_executions,
            "failed_executions": self.failed_executions,
            "running_executions": len(self.running_executions),
            "success_rate": (
                self.successful_executions / self.total_executions
                if self.total_executions > 0 else 0
            ),
        }

# 工作流构建器
class WorkflowBuilder:
    """工作流构建器"""
    
    def __init__(self, workflow_id: str = None, name: str = ""):
        self.workflow = WorkflowDefinition(
            workflow_id=workflow_id or str(uuid.uuid4()),
            name=name
        )
    
    def set_description(self, description: str) -> 'WorkflowBuilder':
        """设置描述"""
        self.workflow.description = description
        return self
    
    def set_timeout(self, timeout: float) -> 'WorkflowBuilder':
        """设置全局超时"""
        self.workflow.global_timeout = timeout
        return self
    
    def add_step(
        self,
        oxy_id: str,
        name: str = "",
        **kwargs
    ) -> 'WorkflowBuilder':
        """添加步骤"""
        step = WorkflowStep(
            step_id=str(uuid.uuid4()),
            oxy_id=oxy_id,
            name=name or f"Step_{len(self.workflow.steps) + 1}",
            **kwargs
        )
        self.workflow.steps.append(step)
        return self
    
    def add_sequential_steps(self, *oxy_ids: str) -> 'WorkflowBuilder':
        """添加顺序执行步骤"""
        for i, oxy_id in enumerate(oxy_ids):
            self.add_step(
                oxy_id=oxy_id,
                name=f"Sequential_Step_{i+1}",
                execution_mode=ExecutionMode.SEQUENTIAL
            )
        return self
    
    def add_conditional_step(
        self,
        oxy_id: str,
        condition: str,
        name: str = ""
    ) -> 'WorkflowBuilder':
        """添加条件步骤"""
        return self.add_step(
            oxy_id=oxy_id,
            name=name or f"Conditional_Step_{len(self.workflow.steps) + 1}",
            execution_mode=ExecutionMode.CONDITIONAL,
            condition=condition
        )
    
    def add_loop_step(
        self,
        oxy_id: str,
        loop_condition: str,
        max_iterations: int = 100,
        name: str = ""
    ) -> 'WorkflowBuilder':
        """添加循环步骤"""
        return self.add_step(
            oxy_id=oxy_id,
            name=name or f"Loop_Step_{len(self.workflow.steps) + 1}",
            execution_mode=ExecutionMode.LOOP,
            loop_condition=loop_condition,
            max_iterations=max_iterations
        )
    
    def build(self) -> WorkflowDefinition:
        """构建工作流"""
        return self.workflow

# 全局工作流引擎实例
workflow_engine = WorkflowEngine()