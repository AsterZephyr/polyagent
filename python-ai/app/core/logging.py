"""
日志配置模块
"""

import logging
import sys
from typing import Dict, Any
import json
from datetime import datetime

class JSONFormatter(logging.Formatter):
    """JSON格式日志格式化器"""
    
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
    """设置日志配置"""
    
    # 设置根日志级别
    logging.getLogger().setLevel(getattr(logging, log_level.upper()))
    
    # 创建控制台处理器
    console_handler = logging.StreamHandler(sys.stdout)
    console_handler.setLevel(getattr(logging, log_level.upper()))
    
    # 设置格式化器
    if log_level.upper() == "DEBUG":
        # 开发环境使用简单格式
        formatter = logging.Formatter(
            '%(asctime)s - %(name)s - %(levelname)s - %(message)s'
        )
    else:
        # 生产环境使用JSON格式
        formatter = JSONFormatter()
    
    console_handler.setFormatter(formatter)
    
    # 配置根日志器
    root_logger = logging.getLogger()
    root_logger.handlers.clear()
    root_logger.addHandler(console_handler)
    
    # 配置第三方库日志级别
    logging.getLogger("uvicorn").setLevel(logging.INFO)
    logging.getLogger("fastapi").setLevel(logging.INFO)
    logging.getLogger("httpx").setLevel(logging.WARNING)
    logging.getLogger("openai").setLevel(logging.WARNING)
    logging.getLogger("anthropic").setLevel(logging.WARNING)
    
    # 创建应用专用日志器
    app_logger = logging.getLogger("polyagent")
    app_logger.setLevel(getattr(logging, log_level.upper()))
    
    return app_logger

def get_logger(name: str) -> logging.Logger:
    """获取命名日志器"""
    return logging.getLogger(f"polyagent.{name}")

class LoggerMixin:
    """日志混入类"""
    
    @property
    def logger(self) -> logging.Logger:
        return get_logger(self.__class__.__name__)
    
    def log_with_context(self, level: str, message: str, **context):
        """带上下文的日志记录"""
        logger = self.logger
        extra_data = {"context": context} if context else {}
        
        getattr(logger, level.lower())(
            message, 
            extra={"extra_data": extra_data}
        )