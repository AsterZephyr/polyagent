"""
Core Exception Classes (Independent of FastAPI)
"""

from typing import Any, Dict

class AgentException(Exception):
    """Base exception for agent system"""
    
    def __init__(self, message: str, code: str = "AGENT_ERROR", details: Dict[str, Any] = None):
        self.message = message
        self.code = code
        self.details = details or {}
        super().__init__(self.message)

class ModelNotAvailableException(AgentException):
    """Model not available exception"""
    
    def __init__(self, model_name: str):
        super().__init__(
            message=f"Model '{model_name}' is not available",
            code="MODEL_NOT_AVAILABLE",
            details={"model_name": model_name}
        )

class APIKeyMissingException(AgentException):
    """API key missing exception"""
    
    def __init__(self, provider: str):
        super().__init__(
            message=f"API key for '{provider}' is not configured",
            code="API_KEY_MISSING",
            details={"provider": provider}
        )

class ComponentExecutionException(AgentException):
    """Component execution exception"""
    
    def __init__(self, component_name: str, error: str):
        super().__init__(
            message=f"Component '{component_name}' execution failed: {error}",
            code="COMPONENT_EXECUTION_ERROR",
            details={"component_name": component_name, "error": error}
        )

class TracingException(AgentException):
    """Tracing system exception"""
    
    def __init__(self, operation: str, error: str):
        super().__init__(
            message=f"Tracing operation '{operation}' failed: {error}",
            code="TRACING_ERROR",
            details={"operation": operation, "error": error}
        )