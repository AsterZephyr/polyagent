"""
Request Tracing and Chain Monitoring System
"""

import uuid
import time
import asyncio
from typing import Dict, List, Any, Optional, Set, Callable
from dataclasses import dataclass, field, asdict
from enum import Enum
from abc import ABC, abstractmethod
import json
from contextlib import asynccontextmanager

from .logging import LoggerMixin

class TraceLevel(Enum):
    """Trace level enum"""
    DEBUG = "debug"
    INFO = "info"
    WARNING = "warning"
    ERROR = "error"

class SpanStatus(Enum):
    """Span status enum"""
    RUNNING = "running"
    SUCCESS = "success"
    ERROR = "error"
    TIMEOUT = "timeout"

@dataclass
class SpanContext:
    """Span context data structure"""
    trace_id: str
    span_id: str
    parent_span_id: Optional[str] = None
    baggage: Dict[str, Any] = field(default_factory=dict)
    
    def with_span_id(self, span_id: str) -> 'SpanContext':
        """Create new context with different span ID"""
        return SpanContext(
            trace_id=self.trace_id,
            span_id=span_id,
            parent_span_id=self.span_id,
            baggage=self.baggage.copy()
        )

@dataclass
class Span:
    """Distributed tracing span"""
    context: SpanContext
    operation_name: str
    start_time: float = field(default_factory=time.time)
    end_time: Optional[float] = None
    duration: Optional[float] = None
    status: SpanStatus = SpanStatus.RUNNING
    tags: Dict[str, Any] = field(default_factory=dict)
    logs: List[Dict[str, Any]] = field(default_factory=list)
    component: str = ""
    error: Optional[str] = None
    
    def set_tag(self, key: str, value: Any):
        """Set span tag"""
        self.tags[key] = value
    
    def log(self, level: TraceLevel, message: str, **kwargs):
        """Add log entry to span"""
        self.logs.append({
            "timestamp": time.time(),
            "level": level.value,
            "message": message,
            "fields": kwargs
        })
    
    def finish(self, status: SpanStatus = SpanStatus.SUCCESS, error: str = None):
        """Finish the span"""
        self.end_time = time.time()
        self.duration = self.end_time - self.start_time
        self.status = status
        if error:
            self.error = error
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert span to dictionary"""
        return asdict(self)

class TraceCollector(ABC):
    """Abstract trace collector interface"""
    
    @abstractmethod
    async def collect_span(self, span: Span):
        """Collect completed span"""
        pass
    
    @abstractmethod
    async def get_trace(self, trace_id: str) -> List[Span]:
        """Get trace by ID"""
        pass

class MemoryTraceCollector(TraceCollector, LoggerMixin):
    """In-memory trace collector for development/testing"""
    
    def __init__(self, max_traces: int = 1000):
        super().__init__()
        self.max_traces = max_traces
        self.traces: Dict[str, List[Span]] = {}
        self._lock = asyncio.Lock()
    
    async def collect_span(self, span: Span):
        """Collect span in memory"""
        async with self._lock:
            trace_id = span.context.trace_id
            
            if trace_id not in self.traces:
                self.traces[trace_id] = []
            
            self.traces[trace_id].append(span)
            
            # Cleanup old traces if needed
            if len(self.traces) > self.max_traces:
                oldest_trace = min(self.traces.keys(), 
                                 key=lambda tid: min(s.start_time for s in self.traces[tid]))
                del self.traces[oldest_trace]
            
            self.logger.debug(f"Collected span: {span.operation_name} in trace {trace_id}")
    
    async def get_trace(self, trace_id: str) -> List[Span]:
        """Get trace spans by trace ID"""
        async with self._lock:
            return self.traces.get(trace_id, [])
    
    async def get_trace_summary(self, trace_id: str) -> Optional[Dict[str, Any]]:
        """Get trace summary statistics"""
        spans = await self.get_trace(trace_id)
        if not spans:
            return None
        
        root_spans = [s for s in spans if s.context.parent_span_id is None]
        
        return {
            "trace_id": trace_id,
            "total_spans": len(spans),
            "root_spans": len(root_spans),
            "total_duration": sum(s.duration or 0 for s in spans),
            "start_time": min(s.start_time for s in spans),
            "end_time": max(s.end_time or s.start_time for s in spans),
            "success_count": sum(1 for s in spans if s.status == SpanStatus.SUCCESS),
            "error_count": sum(1 for s in spans if s.status == SpanStatus.ERROR),
            "components": list(set(s.component for s in spans if s.component))
        }

class Tracer(LoggerMixin):
    """Distributed tracer implementation"""
    
    def __init__(self, service_name: str, collector: TraceCollector = None):
        super().__init__()
        self.service_name = service_name
        self.collector = collector or MemoryTraceCollector()
        self._active_spans: Dict[str, Span] = {}
    
    def start_trace(self, operation_name: str, **tags) -> SpanContext:
        """Start a new trace with root span"""
        trace_id = str(uuid.uuid4())
        span_id = str(uuid.uuid4())
        
        context = SpanContext(trace_id=trace_id, span_id=span_id)
        span = Span(
            context=context,
            operation_name=operation_name,
            component=self.service_name
        )
        
        for key, value in tags.items():
            span.set_tag(key, value)
        
        self._active_spans[span_id] = span
        self.logger.debug(f"Started trace {trace_id} with operation {operation_name}")
        
        return context
    
    def start_span(self, context: SpanContext, operation_name: str, **tags) -> SpanContext:
        """Start a child span"""
        span_id = str(uuid.uuid4())
        child_context = context.with_span_id(span_id)
        
        span = Span(
            context=child_context,
            operation_name=operation_name,
            component=self.service_name
        )
        
        for key, value in tags.items():
            span.set_tag(key, value)
        
        self._active_spans[span_id] = span
        self.logger.debug(f"Started span {operation_name} in trace {context.trace_id}")
        
        return child_context
    
    async def finish_span(self, context: SpanContext, status: SpanStatus = SpanStatus.SUCCESS, 
                         error: str = None, **tags):
        """Finish a span"""
        span_id = context.span_id
        
        if span_id not in self._active_spans:
            self.logger.warning(f"Span {span_id} not found in active spans")
            return
        
        span = self._active_spans.pop(span_id)
        
        # Add final tags
        for key, value in tags.items():
            span.set_tag(key, value)
        
        span.finish(status, error)
        
        # Collect the completed span
        await self.collector.collect_span(span)
        
        self.logger.debug(f"Finished span {span.operation_name} with status {status.value}")
    
    @asynccontextmanager
    async def span(self, context: SpanContext, operation_name: str, **tags):
        """Context manager for automatic span lifecycle"""
        child_context = self.start_span(context, operation_name, **tags)
        
        try:
            yield child_context
            await self.finish_span(child_context, SpanStatus.SUCCESS)
        except Exception as e:
            await self.finish_span(child_context, SpanStatus.ERROR, str(e))
            raise
    
    def inject_context(self, context: SpanContext) -> Dict[str, str]:
        """Inject tracing context into headers/metadata"""
        return {
            "x-trace-id": context.trace_id,
            "x-span-id": context.span_id,
            "x-parent-span-id": context.parent_span_id or "",
        }
    
    def extract_context(self, headers: Dict[str, str]) -> Optional[SpanContext]:
        """Extract tracing context from headers/metadata"""
        trace_id = headers.get("x-trace-id")
        span_id = headers.get("x-span-id")
        
        if not trace_id or not span_id:
            return None
        
        parent_span_id = headers.get("x-parent-span-id") or None
        if parent_span_id == "":
            parent_span_id = None
        
        return SpanContext(
            trace_id=trace_id,
            span_id=span_id,
            parent_span_id=parent_span_id
        )
    
    async def get_trace_info(self, trace_id: str) -> Optional[Dict[str, Any]]:
        """Get comprehensive trace information"""
        return await self.collector.get_trace_summary(trace_id)

class ChainMonitor(LoggerMixin):
    """Chain connectivity and health monitor"""
    
    def __init__(self, tracer: Tracer):
        super().__init__()
        self.tracer = tracer
        self.health_checks: Dict[str, Callable] = {}
        self.component_dependencies: Dict[str, Set[str]] = {}
    
    def register_component(self, name: str, health_check: Callable, dependencies: List[str] = None):
        """Register a component for monitoring"""
        self.health_checks[name] = health_check
        self.component_dependencies[name] = set(dependencies or [])
        self.logger.info(f"Registered component {name} with dependencies: {dependencies}")
    
    async def check_component_health(self, component_name: str, context: SpanContext) -> bool:
        """Check health of a specific component"""
        if component_name not in self.health_checks:
            return False
        
        async with self.tracer.span(context, f"health_check_{component_name}") as span_ctx:
            try:
                health_check = self.health_checks[component_name]
                if asyncio.iscoroutinefunction(health_check):
                    result = await health_check()
                else:
                    result = health_check()
                
                span = self.tracer._active_spans.get(span_ctx.span_id)
                if span:
                    span.set_tag("health_status", "healthy" if result else "unhealthy")
                
                return result
            except Exception as e:
                self.logger.error(f"Health check failed for {component_name}: {e}")
                return False
    
    async def check_chain_connectivity(self, trace_context: SpanContext = None) -> Dict[str, Any]:
        """Check end-to-end chain connectivity"""
        if not trace_context:
            trace_context = self.tracer.start_trace("chain_connectivity_check")
        
        async with self.tracer.span(trace_context, "full_chain_check") as span_ctx:
            results = {}
            failed_components = []
            
            # Check all components
            for component_name in self.health_checks:
                is_healthy = await self.check_component_health(component_name, span_ctx)
                results[component_name] = is_healthy
                
                if not is_healthy:
                    failed_components.append(component_name)
            
            # Check dependency chains
            chain_status = "healthy"
            broken_chains = []
            
            for component, deps in self.component_dependencies.items():
                for dep in deps:
                    if not results.get(dep, False):
                        broken_chains.append(f"{component} -> {dep}")
                        chain_status = "broken"
            
            span = self.tracer._active_spans.get(span_ctx.span_id)
            if span:
                span.set_tag("chain_status", chain_status)
                span.set_tag("failed_components", len(failed_components))
                span.set_tag("broken_chains", len(broken_chains))
            
            return {
                "overall_status": chain_status,
                "component_results": results,
                "failed_components": failed_components,
                "broken_chains": broken_chains,
                "trace_id": trace_context.trace_id
            }

# Global tracer instance
_global_tracer: Optional[Tracer] = None

def init_tracing(service_name: str, collector: TraceCollector = None) -> Tracer:
    """Initialize global tracer"""
    global _global_tracer
    _global_tracer = Tracer(service_name, collector)
    return _global_tracer

def get_tracer() -> Tracer:
    """Get global tracer instance"""
    if _global_tracer is None:
        raise RuntimeError("Tracer not initialized. Call init_tracing() first.")
    return _global_tracer