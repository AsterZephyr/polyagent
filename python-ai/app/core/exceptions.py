"""
异常处理模块
"""

from typing import Any, Dict
from fastapi import FastAPI, Request, HTTPException
from fastapi.responses import JSONResponse
from fastapi.exceptions import RequestValidationError
from starlette.exceptions import HTTPException as StarletteHTTPException
import traceback
import logging

logger = logging.getLogger("polyagent.exceptions")

class AIServiceException(Exception):
    """AI服务基础异常"""
    
    def __init__(self, message: str, code: str = "AI_ERROR", details: Dict[str, Any] = None):
        self.message = message
        self.code = code
        self.details = details or {}
        super().__init__(self.message)

class ModelNotAvailableException(AIServiceException):
    """模型不可用异常"""
    
    def __init__(self, model_name: str):
        super().__init__(
            message=f"Model '{model_name}' is not available",
            code="MODEL_NOT_AVAILABLE",
            details={"model_name": model_name}
        )

class APIKeyMissingException(AIServiceException):
    """API密钥缺失异常"""
    
    def __init__(self, provider: str):
        super().__init__(
            message=f"API key for '{provider}' is not configured",
            code="API_KEY_MISSING",
            details={"provider": provider}
        )

class ToolExecutionException(AIServiceException):
    """工具执行异常"""
    
    def __init__(self, tool_name: str, error: str):
        super().__init__(
            message=f"Tool '{tool_name}' execution failed: {error}",
            code="TOOL_EXECUTION_ERROR",
            details={"tool_name": tool_name, "error": error}
        )

class DocumentProcessingException(AIServiceException):
    """文档处理异常"""
    
    def __init__(self, filename: str, error: str):
        super().__init__(
            message=f"Document processing failed for '{filename}': {error}",
            code="DOCUMENT_PROCESSING_ERROR", 
            details={"filename": filename, "error": error}
        )

class VectorDBException(AIServiceException):
    """向量数据库异常"""
    
    def __init__(self, operation: str, error: str):
        super().__init__(
            message=f"Vector DB operation '{operation}' failed: {error}",
            code="VECTOR_DB_ERROR",
            details={"operation": operation, "error": error}
        )

async def ai_service_exception_handler(request: Request, exc: AIServiceException):
    """AI服务异常处理器"""
    logger.error(f"AI Service Exception: {exc.message}", extra={
        "code": exc.code,
        "details": exc.details,
        "path": request.url.path
    })
    
    return JSONResponse(
        status_code=500,
        content={
            "error": exc.message,
            "code": exc.code,
            "details": exc.details
        }
    )

async def http_exception_handler(request: Request, exc: HTTPException):
    """HTTP异常处理器"""
    logger.warning(f"HTTP Exception: {exc.detail}", extra={
        "status_code": exc.status_code,
        "path": request.url.path
    })
    
    return JSONResponse(
        status_code=exc.status_code,
        content={
            "error": exc.detail,
            "code": f"HTTP_{exc.status_code}",
            "path": request.url.path
        }
    )

async def validation_exception_handler(request: Request, exc: RequestValidationError):
    """请求验证异常处理器"""
    logger.warning(f"Validation Error: {exc.errors()}", extra={
        "path": request.url.path
    })
    
    return JSONResponse(
        status_code=422,
        content={
            "error": "Request validation failed",
            "code": "VALIDATION_ERROR",
            "details": exc.errors()
        }
    )

async def generic_exception_handler(request: Request, exc: Exception):
    """通用异常处理器"""
    error_traceback = traceback.format_exc()
    logger.error(f"Unhandled Exception: {str(exc)}", extra={
        "traceback": error_traceback,
        "path": request.url.path
    })
    
    return JSONResponse(
        status_code=500,
        content={
            "error": "Internal server error",
            "code": "INTERNAL_ERROR",
            "details": {
                "message": str(exc),
                "type": exc.__class__.__name__
            }
        }
    )

def setup_exception_handlers(app: FastAPI):
    """设置异常处理器"""
    app.add_exception_handler(AIServiceException, ai_service_exception_handler)
    app.add_exception_handler(HTTPException, http_exception_handler)
    app.add_exception_handler(StarletteHTTPException, http_exception_handler)
    app.add_exception_handler(RequestValidationError, validation_exception_handler)
    app.add_exception_handler(Exception, generic_exception_handler)