"""
Model Router - AI模型智能路由系统

职责:
- 智能模型选择和负载均衡
- 模型健康监控和故障转移  
- 成本优化和性能调优
- A/B测试和流量分流
"""

from abc import ABC, abstractmethod
from dataclasses import dataclass
from typing import Dict, List, Optional, Any, Tuple
from enum import Enum
import asyncio
import time
import random
from datetime import datetime, timedelta


class ModelProvider(Enum):
    """AI模型提供商"""
    OPENAI = "openai"
    ANTHROPIC = "anthropic" 
    OPENROUTER = "openrouter"
    GLM = "glm"
    HUGGINGFACE = "huggingface"
    LOCAL = "local"


class ModelCapability(Enum):
    """模型能力类型"""
    TEXT_GENERATION = "text_generation"
    CODE_GENERATION = "code_generation"
    REASONING = "reasoning"
    MULTIMODAL = "multimodal"
    FUNCTION_CALLING = "function_calling"
    LONG_CONTEXT = "long_context"


class RoutingStrategy(Enum):
    """路由策略"""
    COST_OPTIMIZED = "cost_optimized"
    PERFORMANCE_OPTIMIZED = "performance_optimized"
    BALANCED = "balanced"
    FAILOVER = "failover"
    A_B_TEST = "ab_test"
    LOAD_BALANCING = "load_balancing"


@dataclass
class ModelConfig:
    """模型配置"""
    model_id: str
    provider: ModelProvider
    name: str
    capabilities: List[ModelCapability]
    
    # 性能参数
    max_tokens: int
    context_window: int
    response_time_p50: float  # ms
    response_time_p95: float  # ms
    
    # 成本参数  
    cost_per_1k_input_tokens: float
    cost_per_1k_output_tokens: float
    free_tier_limit: int  # tokens per month
    
    # 可用性参数
    max_requests_per_minute: int
    max_concurrent_requests: int
    availability_sla: float  # 0.0 - 1.0
    
    # 质量参数
    quality_score: float  # 0.0 - 10.0
    safety_score: float   # 0.0 - 10.0
    
    # 连接配置
    api_endpoint: str
    timeout_seconds: int = 30
    retry_count: int = 3
    
    # 特性标志
    supports_streaming: bool = True
    supports_function_calling: bool = False
    supports_vision: bool = False
    
    # 权重和优先级
    routing_weight: float = 1.0
    priority: int = 1  # 1=highest, 10=lowest


@dataclass
class RouteRequest:
    """路由请求"""
    messages: List[Dict[str, str]]
    required_capabilities: List[ModelCapability]
    max_tokens: int
    temperature: float
    
    # 约束条件
    max_cost_per_request: Optional[float] = None
    max_response_time_ms: Optional[int] = None
    preferred_providers: Optional[List[ModelProvider]] = None
    excluded_models: Optional[List[str]] = None
    
    # 上下文信息
    user_id: str = ""
    session_id: str = ""
    agent_type: str = ""
    
    # 特殊需求
    require_streaming: bool = False
    require_function_calling: bool = False
    require_vision: bool = False
    
    # 路由策略
    routing_strategy: RoutingStrategy = RoutingStrategy.BALANCED


@dataclass  
class RouteResponse:
    """路由响应"""
    selected_model: ModelConfig
    backup_models: List[ModelConfig]
    estimated_cost: float
    estimated_response_time_ms: int
    route_reason: str
    routing_metadata: Dict[str, Any]


@dataclass
class ModelHealth:
    """模型健康状态"""
    model_id: str
    is_healthy: bool
    response_time_avg_ms: float
    error_rate_percent: float
    success_rate_percent: float
    requests_per_minute: int
    queue_depth: int
    last_check_timestamp: datetime
    consecutive_failures: int
    status_message: str


class ModelHealthChecker(ABC):
    """模型健康检查接口"""
    
    @abstractmethod
    async def check_model_health(self, model_config: ModelConfig) -> ModelHealth:
        """检查单个模型健康状态"""
        pass
    
    @abstractmethod
    async def check_all_models(self) -> Dict[str, ModelHealth]:
        """检查所有模型健康状态"""
        pass


class CostCalculator(ABC):
    """成本计算接口"""
    
    @abstractmethod
    def calculate_request_cost(
        self, 
        model_config: ModelConfig, 
        input_tokens: int, 
        output_tokens: int
    ) -> float:
        """计算单次请求成本"""
        pass
    
    @abstractmethod
    def estimate_request_cost(
        self, 
        model_config: ModelConfig, 
        request: RouteRequest
    ) -> float:
        """估算请求成本"""
        pass


class ModelRouter(ABC):
    """模型路由器主接口"""
    
    @abstractmethod
    async def route_request(self, request: RouteRequest) -> RouteResponse:
        """路由请求到最优模型"""
        pass
    
    @abstractmethod
    async def get_available_models(
        self, 
        capabilities: List[ModelCapability] = None
    ) -> List[ModelConfig]:
        """获取可用模型列表"""
        pass
    
    @abstractmethod
    async def get_model_health(self, model_id: str = None) -> Dict[str, ModelHealth]:
        """获取模型健康状态"""
        pass
    
    @abstractmethod
    async def update_model_weights(self, performance_data: Dict[str, float]) -> bool:
        """基于性能数据更新模型权重"""
        pass
    
    @abstractmethod
    async def enable_ab_test(
        self, 
        model_a: str, 
        model_b: str, 
        traffic_split: float
    ) -> str:
        """启用A/B测试"""
        pass


# 具体实现
class DefaultModelRouter(ModelRouter):
    """默认模型路由器实现"""
    
    def __init__(
        self,
        model_configs: List[ModelConfig],
        health_checker: ModelHealthChecker,
        cost_calculator: CostCalculator,
        config: Dict[str, Any] = None
    ):
        self.models = {model.model_id: model for model in model_configs}
        self.health_checker = health_checker
        self.cost_calculator = cost_calculator
        self.config = config or {}
        
        # 运行时状态
        self.model_health: Dict[str, ModelHealth] = {}
        self.performance_metrics: Dict[str, Dict] = {}
        self.ab_tests: Dict[str, Dict] = {}
        
        # 启动后台任务
        asyncio.create_task(self._health_check_loop())
        asyncio.create_task(self._performance_monitoring_loop())
    
    async def route_request(self, request: RouteRequest) -> RouteResponse:
        """智能路由请求"""
        
        # 1. 获取候选模型
        candidates = await self._get_candidate_models(request)
        
        if not candidates:
            raise ValueError("No suitable models available for this request")
        
        # 2. 应用路由策略
        selected_model, backup_models = await self._apply_routing_strategy(
            request, candidates
        )
        
        # 3. 计算预估成本和响应时间
        estimated_cost = self.cost_calculator.estimate_request_cost(
            selected_model, request
        )
        
        estimated_time = self._estimate_response_time(selected_model, request)
        
        # 4. 生成路由原因
        route_reason = self._generate_route_reason(
            request, selected_model, candidates
        )
        
        return RouteResponse(
            selected_model=selected_model,
            backup_models=backup_models,
            estimated_cost=estimated_cost,
            estimated_response_time_ms=estimated_time,
            route_reason=route_reason,
            routing_metadata={
                'candidates_count': len(candidates),
                'strategy': request.routing_strategy.value,
                'timestamp': datetime.now().isoformat()
            }
        )
    
    async def get_available_models(
        self, 
        capabilities: List[ModelCapability] = None
    ) -> List[ModelConfig]:
        """获取可用模型"""
        
        available_models = []
        health_status = await self.get_model_health()
        
        for model_id, model_config in self.models.items():
            # 检查健康状态
            health = health_status.get(model_id)
            if not health or not health.is_healthy:
                continue
            
            # 检查能力匹配
            if capabilities:
                if not all(cap in model_config.capabilities for cap in capabilities):
                    continue
            
            available_models.append(model_config)
        
        return available_models
    
    async def get_model_health(self, model_id: str = None) -> Dict[str, ModelHealth]:
        """获取模型健康状态"""
        if model_id:
            return {model_id: self.model_health.get(model_id)}
        return self.model_health.copy()
    
    async def update_model_weights(self, performance_data: Dict[str, float]) -> bool:
        """更新模型权重"""
        try:
            for model_id, performance_score in performance_data.items():
                if model_id in self.models:
                    # 基于性能调整权重
                    current_weight = self.models[model_id].routing_weight
                    new_weight = current_weight * (0.9 + performance_score * 0.2)
                    new_weight = max(0.1, min(2.0, new_weight))  # 限制权重范围
                    
                    self.models[model_id].routing_weight = new_weight
            
            return True
        except Exception:
            return False
    
    # 私有方法
    async def _get_candidate_models(self, request: RouteRequest) -> List[ModelConfig]:
        """获取候选模型"""
        candidates = []
        
        for model in self.models.values():
            # 检查健康状态
            health = self.model_health.get(model.model_id)
            if not health or not health.is_healthy:
                continue
            
            # 检查能力要求
            if not all(cap in model.capabilities for cap in request.required_capabilities):
                continue
            
            # 检查特殊需求
            if request.require_streaming and not model.supports_streaming:
                continue
            
            if request.require_function_calling and not model.supports_function_calling:
                continue
            
            if request.require_vision and not model.supports_vision:
                continue
            
            # 检查提供商偏好
            if request.preferred_providers and model.provider not in request.preferred_providers:
                continue
            
            # 检查排除列表
            if request.excluded_models and model.model_id in request.excluded_models:
                continue
            
            # 检查成本约束
            if request.max_cost_per_request:
                estimated_cost = self.cost_calculator.estimate_request_cost(model, request)
                if estimated_cost > request.max_cost_per_request:
                    continue
            
            # 检查响应时间约束
            if request.max_response_time_ms:
                if model.response_time_p95 > request.max_response_time_ms:
                    continue
            
            candidates.append(model)
        
        return candidates
    
    async def _apply_routing_strategy(
        self, 
        request: RouteRequest, 
        candidates: List[ModelConfig]
    ) -> Tuple[ModelConfig, List[ModelConfig]]:
        """应用路由策略"""
        
        strategy = request.routing_strategy
        
        if strategy == RoutingStrategy.COST_OPTIMIZED:
            return self._route_by_cost(request, candidates)
        
        elif strategy == RoutingStrategy.PERFORMANCE_OPTIMIZED:
            return self._route_by_performance(candidates)
        
        elif strategy == RoutingStrategy.BALANCED:
            return self._route_balanced(request, candidates)
        
        elif strategy == RoutingStrategy.FAILOVER:
            return self._route_failover(candidates)
        
        elif strategy == RoutingStrategy.A_B_TEST:
            return self._route_ab_test(request, candidates)
        
        elif strategy == RoutingStrategy.LOAD_BALANCING:
            return self._route_load_balanced(candidates)
        
        else:
            # 默认策略：综合评分
            return self._route_balanced(request, candidates)
    
    def _route_by_cost(
        self, 
        request: RouteRequest, 
        candidates: List[ModelConfig]
    ) -> Tuple[ModelConfig, List[ModelConfig]]:
        """成本优化路由"""
        # 按成本排序
        sorted_candidates = sorted(
            candidates,
            key=lambda m: self.cost_calculator.estimate_request_cost(m, request)
        )
        
        return sorted_candidates[0], sorted_candidates[1:3]
    
    def _route_by_performance(
        self, 
        candidates: List[ModelConfig]
    ) -> Tuple[ModelConfig, List[ModelConfig]]:
        """性能优化路由"""
        # 按质量分数和响应时间排序
        sorted_candidates = sorted(
            candidates,
            key=lambda m: (-m.quality_score, m.response_time_p50)
        )
        
        return sorted_candidates[0], sorted_candidates[1:3]
    
    def _route_balanced(
        self, 
        request: RouteRequest, 
        candidates: List[ModelConfig]
    ) -> Tuple[ModelConfig, List[ModelConfig]]:
        """平衡路由：综合考虑成本、性能、健康状态"""
        
        def calculate_score(model: ModelConfig) -> float:
            # 性能分数 (40%)
            performance_score = model.quality_score / 10.0 * 0.4
            
            # 成本分数 (30%) - 成本越低分数越高
            cost = self.cost_calculator.estimate_request_cost(model, request)
            max_cost = max(self.cost_calculator.estimate_request_cost(m, request) for m in candidates)
            cost_score = (1.0 - cost / max_cost) * 0.3 if max_cost > 0 else 0.3
            
            # 健康分数 (20%)
            health = self.model_health.get(model.model_id)
            health_score = (health.success_rate_percent / 100.0) * 0.2 if health else 0.0
            
            # 权重分数 (10%)
            weight_score = model.routing_weight / 2.0 * 0.1
            
            return performance_score + cost_score + health_score + weight_score
        
        # 按综合分数排序
        sorted_candidates = sorted(candidates, key=calculate_score, reverse=True)
        
        return sorted_candidates[0], sorted_candidates[1:3]
    
    def _route_load_balanced(
        self, 
        candidates: List[ModelConfig]
    ) -> Tuple[ModelConfig, List[ModelConfig]]:
        """负载均衡路由：基于权重随机选择"""
        
        # 计算权重总和
        total_weight = sum(model.routing_weight for model in candidates)
        
        # 按权重随机选择
        random_value = random.random() * total_weight
        current_weight = 0.0
        
        for model in candidates:
            current_weight += model.routing_weight
            if current_value <= current_weight:
                selected = model
                break
        else:
            selected = candidates[0]  # 后备选择
        
        # 移除选中的模型，剩余作为备选
        backup_candidates = [m for m in candidates if m.model_id != selected.model_id]
        
        return selected, backup_candidates[:2]
    
    def _estimate_response_time(
        self, 
        model: ModelConfig, 
        request: RouteRequest
    ) -> int:
        """估算响应时间"""
        # 基础响应时间
        base_time = model.response_time_p50
        
        # 根据请求长度调整
        input_length = sum(len(msg.get('content', '')) for msg in request.messages)
        length_factor = 1.0 + (input_length / 1000) * 0.1  # 每1k字符增加10%
        
        # 根据输出token数调整
        output_factor = 1.0 + (request.max_tokens / 1000) * 0.2  # 每1k token增加20%
        
        estimated_time = base_time * length_factor * output_factor
        
        return int(estimated_time)
    
    def _generate_route_reason(
        self, 
        request: RouteRequest, 
        selected_model: ModelConfig, 
        candidates: List[ModelConfig]
    ) -> str:
        """生成路由原因说明"""
        
        reasons = []
        
        # 策略原因
        strategy_reasons = {
            RoutingStrategy.COST_OPTIMIZED: f"最低成本选择 (${self.cost_calculator.estimate_request_cost(selected_model, request):.4f})",
            RoutingStrategy.PERFORMANCE_OPTIMIZED: f"最高性能选择 (质量分数: {selected_model.quality_score}/10)",
            RoutingStrategy.BALANCED: "综合最优选择",
            RoutingStrategy.FAILOVER: "故障转移选择",
            RoutingStrategy.LOAD_BALANCING: "负载均衡选择"
        }
        
        reasons.append(strategy_reasons.get(
            request.routing_strategy, 
            "默认路由策略"
        ))
        
        # 能力匹配
        if request.required_capabilities:
            cap_names = [cap.value for cap in request.required_capabilities]
            reasons.append(f"满足能力需求: {', '.join(cap_names)}")
        
        # 健康状态
        health = self.model_health.get(selected_model.model_id)
        if health:
            reasons.append(f"健康状态良好 (成功率: {health.success_rate_percent:.1f}%)")
        
        return "; ".join(reasons)
    
    # 后台监控任务
    async def _health_check_loop(self):
        """健康检查循环"""
        while True:
            try:
                health_results = await self.health_checker.check_all_models()
                self.model_health.update(health_results)
                
                await asyncio.sleep(30)  # 30秒检查一次
            except Exception:
                await asyncio.sleep(10)  # 出错时10秒后重试
    
    async def _performance_monitoring_loop(self):
        """性能监控循环"""  
        while True:
            try:
                # 收集性能指标
                # 更新模型权重
                # 记录统计信息
                
                await asyncio.sleep(300)  # 5分钟更新一次
            except Exception:
                await asyncio.sleep(60)  # 出错时1分钟后重试


# 工厂函数
def create_model_router(
    model_configs: List[ModelConfig],
    health_checker: ModelHealthChecker,
    cost_calculator: CostCalculator,
    config: Dict[str, Any] = None
) -> ModelRouter:
    """创建模型路由器实例"""
    return DefaultModelRouter(
        model_configs, 
        health_checker, 
        cost_calculator, 
        config
    )