"""
AI Models Configuration - 最新AI模型支持配置
支持 Claude-4, GPT-5, OpenRouter, GLM-4.5 等最新模型
"""

from enum import Enum
from dataclasses import dataclass
from typing import Dict, List, Optional, Any
import os

class ModelProvider(Enum):
    """AI模型提供商"""
    OPENAI = "openai"
    ANTHROPIC = "anthropic"
    OPENROUTER = "openrouter"
    GLM = "glm"
    CLAUDE = "claude"
    LOCAL = "local"

class ModelCapability(Enum):
    """模型能力"""
    TEXT_GENERATION = "text_generation"
    CODE_GENERATION = "code_generation"
    VISION = "vision"
    FUNCTION_CALLING = "function_calling"
    REASONING = "reasoning"
    CREATIVE = "creative"
    ANALYSIS = "analysis"
    EMBEDDING = "embedding"
    MULTIMODAL = "multimodal"

@dataclass
class ModelConfig:
    """模型配置"""
    provider: ModelProvider
    model_id: str
    display_name: str
    description: str
    capabilities: List[ModelCapability]
    max_tokens: int
    context_window: int
    cost_per_1k_input: float  # USD
    cost_per_1k_output: float  # USD
    supports_streaming: bool = True
    supports_system_message: bool = True
    supports_tools: bool = False
    api_key_env: str = ""
    base_url: Optional[str] = None
    free_tier: bool = False
    free_quota: Optional[str] = None
    quality_score: int = 85  # 1-100

# 最新AI模型配置
AVAILABLE_MODELS: Dict[str, ModelConfig] = {
    
    # OpenAI GPT系列 (最新)
    "gpt-5": ModelConfig(
        provider=ModelProvider.OPENAI,
        model_id="gpt-5",
        display_name="GPT-5",
        description="OpenAI最新旗舰模型，具备极强推理和多模态能力",
        capabilities=[ModelCapability.TEXT_GENERATION, ModelCapability.CODE_GENERATION, 
                     ModelCapability.REASONING, ModelCapability.FUNCTION_CALLING,
                     ModelCapability.VISION, ModelCapability.MULTIMODAL],
        max_tokens=8192,
        context_window=200000,
        cost_per_1k_input=0.015,
        cost_per_1k_output=0.060,
        supports_tools=True,
        api_key_env="OPENAI_API_KEY",
        quality_score=98
    ),
    
    "gpt-4o": ModelConfig(
        provider=ModelProvider.OPENAI,
        model_id="gpt-4o",
        display_name="GPT-4o",
        description="OpenAI多模态旗舰模型，支持文本、视觉、音频处理",
        capabilities=[ModelCapability.TEXT_GENERATION, ModelCapability.CODE_GENERATION,
                     ModelCapability.VISION, ModelCapability.MULTIMODAL, ModelCapability.FUNCTION_CALLING],
        max_tokens=4096,
        context_window=128000,
        cost_per_1k_input=0.005,
        cost_per_1k_output=0.015,
        supports_tools=True,
        api_key_env="OPENAI_API_KEY",
        quality_score=95
    ),
    
    # Anthropic Claude系列 (最新)
    "claude-4": ModelConfig(
        provider=ModelProvider.ANTHROPIC,
        model_id="claude-3-5-sonnet-20241022",  # Claude-4对应的实际API名称
        display_name="Claude-4",
        description="Anthropic最新Claude模型，卓越的推理和分析能力",
        capabilities=[ModelCapability.TEXT_GENERATION, ModelCapability.CODE_GENERATION,
                     ModelCapability.REASONING, ModelCapability.ANALYSIS, 
                     ModelCapability.FUNCTION_CALLING],
        max_tokens=8192,
        context_window=200000,
        cost_per_1k_input=0.003,
        cost_per_1k_output=0.015,
        supports_tools=True,
        api_key_env="ANTHROPIC_API_KEY",
        base_url="https://api.anthropic.com",
        quality_score=97
    ),
    
    "claude-3-5-sonnet": ModelConfig(
        provider=ModelProvider.ANTHROPIC,
        model_id="claude-3-5-sonnet-20241022",
        display_name="Claude 3.5 Sonnet",
        description="Claude 3.5 Sonnet - 平衡性能和成本的最佳选择",
        capabilities=[ModelCapability.TEXT_GENERATION, ModelCapability.CODE_GENERATION,
                     ModelCapability.REASONING, ModelCapability.ANALYSIS],
        max_tokens=8192,
        context_window=200000,
        cost_per_1k_input=0.003,
        cost_per_1k_output=0.015,
        supports_tools=True,
        api_key_env="ANTHROPIC_API_KEY",
        quality_score=94
    ),
    
    # OpenRouter 模型 (免费和高级)
    "openrouter-k2-free": ModelConfig(
        provider=ModelProvider.OPENROUTER,
        model_id="microsoft/wizardlm-2-8x22b",  # K2对应模型
        display_name="OpenRouter K2 (Free)",
        description="OpenRouter免费K2模型，适合轻量级任务",
        capabilities=[ModelCapability.TEXT_GENERATION, ModelCapability.CODE_GENERATION],
        max_tokens=4096,
        context_window=65536,
        cost_per_1k_input=0.0,
        cost_per_1k_output=0.0,
        api_key_env="OPENROUTER_API_KEY",
        base_url="https://openrouter.ai/api/v1",
        free_tier=True,
        free_quota="Unlimited (with rate limits)",
        quality_score=78
    ),
    
    "openrouter-qwen3-coder-free": ModelConfig(
        provider=ModelProvider.OPENROUTER,
        model_id="qwen/qwen-2.5-coder-32b-instruct",
        display_name="Qwen 3 Coder (Free)",
        description="通义千问3代码专家模型，免费使用",
        capabilities=[ModelCapability.TEXT_GENERATION, ModelCapability.CODE_GENERATION],
        max_tokens=8192,
        context_window=32768,
        cost_per_1k_input=0.0,
        cost_per_1k_output=0.0,
        api_key_env="OPENROUTER_API_KEY",
        base_url="https://openrouter.ai/api/v1",
        free_tier=True,
        free_quota="Unlimited (with rate limits)",
        quality_score=85
    ),
    
    "openrouter-claude-3-haiku": ModelConfig(
        provider=ModelProvider.OPENROUTER,
        model_id="anthropic/claude-3-haiku-20240307",
        display_name="Claude 3 Haiku (OpenRouter)",
        description="通过OpenRouter访问的Claude 3 Haiku",
        capabilities=[ModelCapability.TEXT_GENERATION, ModelCapability.REASONING],
        max_tokens=4096,
        context_window=200000,
        cost_per_1k_input=0.00025,
        cost_per_1k_output=0.00125,
        api_key_env="OPENROUTER_API_KEY",
        base_url="https://openrouter.ai/api/v1",
        quality_score=88
    ),
    
    # GLM-4.5 (智谱AI)
    "glm-4.5": ModelConfig(
        provider=ModelProvider.GLM,
        model_id="glm-4-plus",
        display_name="GLM-4.5",
        description="智谱AI GLM-4.5，赠送200万tokens",
        capabilities=[ModelCapability.TEXT_GENERATION, ModelCapability.CODE_GENERATION,
                     ModelCapability.REASONING, ModelCapability.ANALYSIS,
                     ModelCapability.FUNCTION_CALLING],
        max_tokens=8192,
        context_window=128000,
        cost_per_1k_input=0.0,  # 免费额度内
        cost_per_1k_output=0.0,
        supports_tools=True,
        api_key_env="GLM_API_KEY",
        base_url="https://open.bigmodel.cn/api/paas/v4",
        free_tier=True,
        free_quota="200万 tokens",
        quality_score=86
    ),
    
    "glm-4-vision": ModelConfig(
        provider=ModelProvider.GLM,
        model_id="glm-4v",
        display_name="GLM-4 Vision",
        description="GLM-4视觉版本，支持图像理解",
        capabilities=[ModelCapability.TEXT_GENERATION, ModelCapability.VISION,
                     ModelCapability.MULTIMODAL],
        max_tokens=4096,
        context_window=8192,
        cost_per_1k_input=0.01,
        cost_per_1k_output=0.01,
        api_key_env="GLM_API_KEY",
        base_url="https://open.bigmodel.cn/api/paas/v4",
        quality_score=82
    ),
    
    # Embedding 模型
    "text-embedding-3-large": ModelConfig(
        provider=ModelProvider.OPENAI,
        model_id="text-embedding-3-large",
        display_name="OpenAI Embedding v3 Large",
        description="OpenAI最新大型嵌入模型",
        capabilities=[ModelCapability.EMBEDDING],
        max_tokens=8191,
        context_window=8191,
        cost_per_1k_input=0.00013,
        cost_per_1k_output=0.0,
        supports_streaming=False,
        api_key_env="OPENAI_API_KEY",
        quality_score=95
    ),
    
    "text-embedding-3-small": ModelConfig(
        provider=ModelProvider.OPENAI,
        model_id="text-embedding-3-small",
        display_name="OpenAI Embedding v3 Small",
        description="OpenAI高性价比嵌入模型",
        capabilities=[ModelCapability.EMBEDDING],
        max_tokens=8191,
        context_window=8191,
        cost_per_1k_input=0.00002,
        cost_per_1k_output=0.0,
        supports_streaming=False,
        api_key_env="OPENAI_API_KEY",
        quality_score=90
    ),
}

# Models grouped by provider
MODELS_BY_PROVIDER = {
    provider: [model for model in AVAILABLE_MODELS.values() if model.provider == provider]
    for provider in ModelProvider
}

# Free models list
FREE_MODELS = [
    model_id for model_id, model in AVAILABLE_MODELS.items()
    if model.free_tier or model.cost_per_1k_input == 0.0
]

# Recommended models configuration
RECOMMENDED_MODELS = {
    "best_quality": "claude-4",
    "best_value": "glm-4.5", 
    "best_free": "openrouter-qwen3-coder-free",
    "best_coding": "openrouter-qwen3-coder-free",
    "best_reasoning": "claude-4",
    "best_multimodal": "gpt-5",
    "best_embedding": "text-embedding-3-large",
    "budget_friendly": "claude-3-haiku"
}

class ModelSelector:
    """Intelligent model selector"""
    
    @staticmethod
    def get_model_for_task(
        task_type: str,
        budget_limit: float = None,
        free_only: bool = False,
        capability_requirements: List[ModelCapability] = None
    ) -> str:
        """Select appropriate model based on task type"""
        
        if free_only:
            candidates = [
                model_id for model_id, model in AVAILABLE_MODELS.items()
                if model.free_tier or model.cost_per_1k_input == 0.0
            ]
        else:
            candidates = list(AVAILABLE_MODELS.keys())
        
        # Filter capability requirements
        if capability_requirements:
            candidates = [
                model_id for model_id in candidates
                if all(cap in AVAILABLE_MODELS[model_id].capabilities 
                      for cap in capability_requirements)
            ]
        
        # Sort by quality score
        candidates.sort(
            key=lambda x: AVAILABLE_MODELS[x].quality_score,
            reverse=True
        )
        
        if not candidates:
            return "gpt-4o"  # Default fallback
        
        return candidates[0]
    
    @staticmethod
    def get_available_models(
        provider: ModelProvider = None,
        capability: ModelCapability = None,
        free_only: bool = False
    ) -> List[str]:
        """Get available models list"""
        
        models = AVAILABLE_MODELS.items()
        
        if provider:
            models = [(k, v) for k, v in models if v.provider == provider]
        
        if capability:
            models = [(k, v) for k, v in models if capability in v.capabilities]
        
        if free_only:
            models = [(k, v) for k, v in models if v.free_tier or v.cost_per_1k_input == 0.0]
        
        return [model_id for model_id, _ in models]
    
    @staticmethod
    def estimate_cost(model_id: str, input_tokens: int, output_tokens: int = 0) -> float:
        """Estimate usage cost"""
        
        if model_id not in AVAILABLE_MODELS:
            return 0.0
        
        model = AVAILABLE_MODELS[model_id]
        
        input_cost = (input_tokens / 1000) * model.cost_per_1k_input
        output_cost = (output_tokens / 1000) * model.cost_per_1k_output
        
        return input_cost + output_cost
    
    @staticmethod
    def get_model_info(model_id: str) -> Optional[ModelConfig]:
        """Get model information"""
        return AVAILABLE_MODELS.get(model_id)

# Environment variables check
def check_api_keys() -> Dict[str, bool]:
    """Check API keys configuration"""
    key_status = {}
    
    required_keys = {
        "OPENAI_API_KEY": "OpenAI API key",
        "ANTHROPIC_API_KEY": "Anthropic API key", 
        "OPENROUTER_API_KEY": "OpenRouter API key",
        "GLM_API_KEY": "Zhipu AI API key"
    }
    
    for key, desc in required_keys.items():
        key_status[key] = bool(os.getenv(key))
    
    return key_status

# Get default configuration
def get_default_model_config() -> Dict[str, Any]:
    """Get default model configuration"""
    
    # Check available free models
    available_free = [
        model_id for model_id in FREE_MODELS
        if os.getenv(AVAILABLE_MODELS[model_id].api_key_env)
    ]
    
    # Select default model
    if "glm-4.5" in available_free:
        default_model = "glm-4.5"
    elif "openrouter-qwen3-coder-free" in available_free:
        default_model = "openrouter-qwen3-coder-free"
    elif os.getenv("ANTHROPIC_API_KEY"):
        default_model = "claude-3-5-sonnet"
    else:
        default_model = "gpt-4o"
    
    return {
        "default_model": default_model,
        "embedding_model": "text-embedding-3-small",
        "vision_model": "gpt-4o",
        "code_model": "openrouter-qwen3-coder-free",
        "reasoning_model": "claude-4" if os.getenv("ANTHROPIC_API_KEY") else default_model,
        "available_free_models": available_free,
        "recommended_models": RECOMMENDED_MODELS
    }