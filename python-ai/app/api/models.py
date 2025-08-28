"""
API请求和响应模型
"""

from typing import Dict, List, Any, Optional
from pydantic import BaseModel, Field

# 任务相关模型

class TaskRequest(BaseModel):
    """任务执行请求"""
    task_id: str
    user_id: str
    session_id: str
    agent_type: str
    input: str
    context: Optional[Dict[str, Any]] = None
    tools: Optional[List[str]] = None
    memory: Optional[Dict[str, Any]] = None

class TaskResponse(BaseModel):
    """任务执行响应"""
    task_id: str
    status: str  # success, error, streaming
    output: str
    tool_calls: Optional[List[Dict[str, Any]]] = None
    memory: Optional[Dict[str, Any]] = None
    metadata: Optional[Dict[str, Any]] = None
    error: Optional[str] = None
    timestamp: str

# 聊天相关模型

class ChatMessage(BaseModel):
    """聊天消息"""
    role: str = Field(..., description="消息角色: system, user, assistant, tool")
    content: str = Field(..., description="消息内容")
    tool_calls: Optional[List[Dict[str, Any]]] = None
    tool_call_id: Optional[str] = None

class ChatRequest(BaseModel):
    """聊天请求"""
    messages: List[ChatMessage] = Field(..., description="对话消息列表")
    model: str = Field(default="gpt-3.5-turbo", description="使用的模型")
    temperature: float = Field(default=0.7, ge=0.0, le=2.0, description="温度参数")
    max_tokens: int = Field(default=1000, gt=0, le=4000, description="最大token数")
    tools: Optional[List[Dict[str, Any]]] = Field(None, description="可用工具列表")
    stream: bool = Field(default=False, description="是否流式响应")

class ChatResponse(BaseModel):
    """聊天响应"""
    content: str
    tool_calls: Optional[List[Dict[str, Any]]] = None
    usage: Optional[Dict[str, int]] = None
    model: str
    finish_reason: Optional[str] = None

# RAG相关模型

class RAGQuery(BaseModel):
    """RAG查询请求"""
    user_id: str
    query: str = Field(..., description="查询内容")
    top_k: int = Field(default=5, ge=1, le=20, description="返回结果数量")
    filters: Optional[Dict[str, Any]] = Field(None, description="过滤条件")

class RAGDocument(BaseModel):
    """RAG文档结果"""
    id: str
    content: str
    metadata: Dict[str, Any]
    score: float

class RAGResult(BaseModel):
    """RAG查询结果"""
    documents: List[RAGDocument]
    scores: List[float]
    context: str

# 文档相关模型

class DocumentUploadRequest(BaseModel):
    """文档上传请求"""
    user_id: str
    filename: str
    content: str  # Base64编码或直接文本

class DocumentUploadResponse(BaseModel):
    """文档上传响应"""
    document_id: str
    status: str
    message: str

# 工具相关模型

class ToolInfo(BaseModel):
    """工具信息"""
    name: str
    description: str
    parameters: Dict[str, Any]
    category: str

class ToolExecutionRequest(BaseModel):
    """工具执行请求"""
    name: str = Field(..., description="工具名称")
    parameters: Dict[str, Any] = Field(default_factory=dict, description="工具参数")

class ToolExecutionResult(BaseModel):
    """工具执行结果"""
    success: bool
    result: Any
    error: Optional[str] = None

# 通用响应模型

class ErrorResponse(BaseModel):
    """错误响应"""
    error: str
    code: str
    details: Optional[Dict[str, Any]] = None

class HealthResponse(BaseModel):
    """健康检查响应"""
    status: str
    adapters: Dict[str, Dict[str, Any]]
    models: List[str]
    total_adapters: int

class ModelListResponse(BaseModel):
    """模型列表响应"""
    models: List[str]
    total: int

class ToolListResponse(BaseModel):
    """工具列表响应"""
    tools: List[ToolInfo]
    total: int