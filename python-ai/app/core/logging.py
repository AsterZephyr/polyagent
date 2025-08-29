"""
Logging configuration module
"""

import logging
import sys
from typing import Dict, Any
import json
from datetime import datetime

class JSONFormatter(logging.Formatter):
    """JSON log formatter"""
    
    def format(self, record: logging.LogRecord) -> str:
        log_entry = {
            "timestamp": datetime.utcnow().isoformat(),
            "level": record.levelname,
            "logger": record.name,
            "message": record.getMessage(),
            "module": record.module,
            "function": record.funcName,
            "line": record.lineno,
        }
        
        if record.exc_info:
            log_entry["exception"] = self.formatException(record.exc_info)
            
        if hasattr(record, 'extra_data'):
            log_entry.update(record.extra_data)
            
        return json.dumps(log_entry, ensure_ascii=False)

def setup_logging(log_level: str = "INFO") -> logging.Logger:
    """Setup logging configuration"""
    
    # Set root log level
    logging.getLogger().setLevel(getattr(logging, log_level.upper()))
    
    # Create console handler
    console_handler = logging.StreamHandler(sys.stdout)
    console_handler.setLevel(getattr(logging, log_level.upper()))
    
    # Setup formatter
    if log_level.upper() == "DEBUG":
        # Development environment - simple format
        formatter = logging.Formatter(
            '%(asctime)s - %(name)s - %(levelname)s - %(message)s'
        )
    else:
        # Production environment - JSON format
        formatter = JSONFormatter()
    
    console_handler.setFormatter(formatter)
    
    # Configure root logger
    root_logger = logging.getLogger()
    root_logger.handlers.clear()
    root_logger.addHandler(console_handler)
    
    # Configure third-party library log levels
    logging.getLogger("uvicorn").setLevel(logging.INFO)
    logging.getLogger("fastapi").setLevel(logging.INFO)
    logging.getLogger("httpx").setLevel(logging.WARNING)
    logging.getLogger("openai").setLevel(logging.WARNING)
    logging.getLogger("anthropic").setLevel(logging.WARNING)
    
    # Create application logger
    app_logger = logging.getLogger("polyagent")
    app_logger.setLevel(getattr(logging, log_level.upper()))
    
    return app_logger

def get_logger(name: str) -> logging.Logger:
    """Get named logger"""
    return logging.getLogger(f"polyagent.{name}")

class LoggerMixin:
    """Logger mixin class"""
    
    @property
    def logger(self) -> logging.Logger:
        return get_logger(self.__class__.__name__)
    
    def log_with_context(self, level: str, message: str, **context):
        """Log with context data"""
        logger = self.logger
        extra_data = {"context": context} if context else {}
        
        getattr(logger, level.lower())(
            message, 
            extra={"extra_data": extra_data}
        )