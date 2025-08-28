"""
配置管理模块
"""

import os
from typing import List, Optional
from pydantic_settings import BaseSettings
from pydantic import Field
from functools import lru_cache

class Settings(BaseSettings):
    """应用配置"""
    
    # 基础配置
    APP_NAME: str = "PolyAgent AI Service"
    DEBUG: bool = Field(default=False, env="DEBUG")
    HOST: str = Field(default="0.0.0.0", env="HOST")
    PORT: int = Field(default=8000, env="PORT")
    LOG_LEVEL: str = Field(default="INFO", env="LOG_LEVEL")
    
    # 数据库配置
    REDIS_URL: str = Field(default="redis://localhost:6379/0", env="REDIS_URL")
    POSTGRES_URL: str = Field(default="postgresql://user:pass@localhost:5432/polyagent", env="POSTGRES_URL")
    
    # AI 模型配置
    OPENAI_API_KEY: Optional[str] = Field(default=None, env="OPENAI_API_KEY")
    OPENAI_BASE_URL: str = Field(default="https://api.openai.com/v1", env="OPENAI_BASE_URL")
    OPENAI_MODEL: str = Field(default="gpt-4", env="OPENAI_MODEL")
    
    ANTHROPIC_API_KEY: Optional[str] = Field(default=None, env="ANTHROPIC_API_KEY")
    ANTHROPIC_MODEL: str = Field(default="claude-3-sonnet-20240229", env="ANTHROPIC_MODEL")
    
    # 其他AI模型配置
    AVAILABLE_MODELS: List[str] = Field(default=[
        "gpt-4", "gpt-3.5-turbo", 
        "claude-3-sonnet", "claude-3-haiku",
        "gemini-pro"
    ])
    
    # RAG配置
    VECTOR_DB_TYPE: str = Field(default="chromadb", env="VECTOR_DB_TYPE")  # chromadb, pinecone
    CHROMADB_HOST: str = Field(default="localhost", env="CHROMADB_HOST")
    CHROMADB_PORT: int = Field(default=8001, env="CHROMADB_PORT")
    
    PINECONE_API_KEY: Optional[str] = Field(default=None, env="PINECONE_API_KEY")
    PINECONE_ENVIRONMENT: str = Field(default="us-west1-gcp", env="PINECONE_ENVIRONMENT")
    PINECONE_INDEX_NAME: str = Field(default="polyagent", env="PINECONE_INDEX_NAME")
    
    # Embedding模型配置
    EMBEDDING_MODEL: str = Field(default="sentence-transformers/all-MiniLM-L6-v2", env="EMBEDDING_MODEL")
    EMBEDDING_DIMENSION: int = Field(default=384, env="EMBEDDING_DIMENSION")
    
    # 文档处理配置
    CHUNK_SIZE: int = Field(default=1000, env="CHUNK_SIZE")
    CHUNK_OVERLAP: int = Field(default=200, env="CHUNK_OVERLAP")
    MAX_DOCUMENT_SIZE: int = Field(default=50*1024*1024, env="MAX_DOCUMENT_SIZE")  # 50MB
    
    # 工具配置
    ENABLE_WEB_SEARCH: bool = Field(default=True, env="ENABLE_WEB_SEARCH")
    ENABLE_CODE_EXECUTION: bool = Field(default=False, env="ENABLE_CODE_EXECUTION")  # 安全考虑
    
    # 搜索引擎配置
    GOOGLE_API_KEY: Optional[str] = Field(default=None, env="GOOGLE_API_KEY")
    GOOGLE_CSE_ID: Optional[str] = Field(default=None, env="GOOGLE_CSE_ID")
    SERPAPI_API_KEY: Optional[str] = Field(default=None, env="SERPAPI_API_KEY")
    
    # 性能配置
    MAX_CONCURRENT_REQUESTS: int = Field(default=10, env="MAX_CONCURRENT_REQUESTS")
    REQUEST_TIMEOUT: int = Field(default=30, env="REQUEST_TIMEOUT")
    MAX_TOKENS: int = Field(default=4000, env="MAX_TOKENS")
    TEMPERATURE: float = Field(default=0.7, env="TEMPERATURE")
    
    # 缓存配置
    ENABLE_CACHE: bool = Field(default=True, env="ENABLE_CACHE")
    CACHE_TTL: int = Field(default=3600, env="CACHE_TTL")  # 1小时
    
    # 安全配置
    API_KEY_HEADER: str = Field(default="X-API-Key", env="API_KEY_HEADER")
    ALLOWED_HOSTS: List[str] = Field(default=["*"], env="ALLOWED_HOSTS")
    
    class Config:
        env_file = ".env"
        case_sensitive = True

@lru_cache()
def get_settings() -> Settings:
    """获取设置实例（单例）"""
    return Settings()