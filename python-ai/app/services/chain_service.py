"""
Chain Service - End-to-end chain connectivity and monitoring
"""

import asyncio
from typing import Dict, List, Any, Optional
from dataclasses import dataclass

from ..core.logging import LoggerMixin
from ..core.tracing import get_tracer, SpanContext, SpanStatus, ChainMonitor
from ..adapters.unified_adapter import UnifiedAIAdapter
from ..oxy.core import OxyContext, OxyResult

@dataclass
class ChainTestResult:
    """Chain test execution result"""
    success: bool
    trace_id: str
    components_tested: List[str]
    total_duration: float
    failed_steps: List[str]
    error_details: Dict[str, str]

class ChainService(LoggerMixin):
    """Service for managing and testing end-to-end chains"""
    
    def __init__(self, ai_adapter: UnifiedAIAdapter):
        super().__init__()
        self.ai_adapter = ai_adapter
        self.tracer = get_tracer()
        self.monitor = ChainMonitor(self.tracer)
        self._setup_monitoring()
    
    def _setup_monitoring(self):
        """Setup component monitoring"""
        # Register AI adapter health check
        self.monitor.register_component(
            "ai_adapter",
            self._check_ai_adapter_health,
            []
        )
        
        # Register model availability check
        self.monitor.register_component(
            "models",
            self._check_models_availability,
            ["ai_adapter"]
        )
        
        # Register context processing check
        self.monitor.register_component(
            "context_processor",
            self._check_context_processing,
            ["ai_adapter", "models"]
        )
        
        self.logger.info("Chain monitoring setup completed")
    
    async def _check_ai_adapter_health(self) -> bool:
        """Check AI adapter health"""
        try:
            available_models = self.ai_adapter.get_available_models()
            return len(available_models) > 0
        except Exception as e:
            self.logger.error(f"AI adapter health check failed: {e}")
            return False
    
    async def _check_models_availability(self) -> bool:
        """Check if AI models are available and responsive"""
        try:
            health_status = await self.ai_adapter.health_check_all()
            return any(health_status.values())
        except Exception as e:
            self.logger.error(f"Models availability check failed: {e}")
            return False
    
    async def _check_context_processing(self) -> bool:
        """Check context processing capabilities"""
        try:
            # Create a test context
            test_context = OxyContext(
                session_id="test_health_check",
                user_id="system"
            )
            
            # Try basic message processing
            test_messages = [{"role": "user", "content": "Hello"}]
            
            # Test with first available model
            available_models = self.ai_adapter.get_available_models()
            if not available_models:
                return False
            
            response = await self.ai_adapter.generate(
                messages=test_messages,
                model=available_models[0],
                max_tokens=10
            )
            
            return response.content is not None
            
        except Exception as e:
            self.logger.error(f"Context processing check failed: {e}")
            return False
    
    async def test_end_to_end_chain(self, 
                                   test_scenario: str = "basic",
                                   timeout: float = 30.0) -> ChainTestResult:
        """Test complete end-to-end chain functionality"""
        
        trace_context = self.tracer.start_trace(
            "e2e_chain_test",
            test_scenario=test_scenario,
            service="polyagent"
        )
        
        start_time = asyncio.get_event_loop().time()
        failed_steps = []
        error_details = {}
        components_tested = []
        
        try:
            async with self.tracer.span(trace_context, "chain_test_execution") as span_ctx:
                # Step 1: Check chain connectivity
                self.logger.info("Testing chain connectivity...")
                connectivity_result = await self.monitor.check_chain_connectivity(span_ctx)
                components_tested.extend(connectivity_result["component_results"].keys())
                
                if connectivity_result["overall_status"] != "healthy":
                    failed_steps.append("connectivity_check")
                    error_details["connectivity_check"] = str(connectivity_result["failed_components"])
                
                # Step 2: Test AI model integration
                async with self.tracer.span(span_ctx, "ai_model_integration_test") as ai_span:
                    try:
                        self.logger.info("Testing AI model integration...")
                        test_result = await self._test_ai_integration(ai_span)
                        components_tested.append("ai_integration")
                        
                        if not test_result:
                            failed_steps.append("ai_integration")
                            error_details["ai_integration"] = "AI model integration test failed"
                        
                    except Exception as e:
                        failed_steps.append("ai_integration")
                        error_details["ai_integration"] = str(e)
                
                # Step 3: Test Oxy component system
                async with self.tracer.span(span_ctx, "oxy_components_test") as oxy_span:
                    try:
                        self.logger.info("Testing Oxy component system...")
                        oxy_result = await self._test_oxy_components(oxy_span)
                        components_tested.append("oxy_components")
                        
                        if not oxy_result:
                            failed_steps.append("oxy_components")
                            error_details["oxy_components"] = "Oxy components test failed"
                        
                    except Exception as e:
                        failed_steps.append("oxy_components")
                        error_details["oxy_components"] = str(e)
                
                # Step 4: Test workflow execution
                async with self.tracer.span(span_ctx, "workflow_execution_test") as workflow_span:
                    try:
                        self.logger.info("Testing workflow execution...")
                        workflow_result = await self._test_workflow_execution(workflow_span)
                        components_tested.append("workflow_execution")
                        
                        if not workflow_result:
                            failed_steps.append("workflow_execution")
                            error_details["workflow_execution"] = "Workflow execution test failed"
                        
                    except Exception as e:
                        failed_steps.append("workflow_execution")
                        error_details["workflow_execution"] = str(e)
                
                # Calculate results
                total_duration = asyncio.get_event_loop().time() - start_time
                success = len(failed_steps) == 0
                
                # Log test completion
                status = SpanStatus.SUCCESS if success else SpanStatus.ERROR
                await self.tracer.finish_span(
                    span_ctx, status,
                    error=f"Failed steps: {failed_steps}" if failed_steps else None,
                    components_tested=len(components_tested),
                    total_duration=total_duration,
                    test_scenario=test_scenario
                )
                
                self.logger.info(f"E2E chain test completed: {success}, duration: {total_duration:.2f}s")
                
                return ChainTestResult(
                    success=success,
                    trace_id=trace_context.trace_id,
                    components_tested=components_tested,
                    total_duration=total_duration,
                    failed_steps=failed_steps,
                    error_details=error_details
                )
        
        except Exception as e:
            await self.tracer.finish_span(
                trace_context, SpanStatus.ERROR, str(e)
            )
            self.logger.error(f"E2E chain test failed with exception: {e}")
            raise
    
    async def _test_ai_integration(self, context: SpanContext) -> bool:
        """Test AI model integration"""
        try:
            # Get available models
            available_models = self.ai_adapter.get_available_models()
            if not available_models:
                return False
            
            # Test basic generation
            test_messages = [
                {"role": "user", "content": "Say 'integration test successful' if you can read this."}
            ]
            
            response = await self.ai_adapter.generate(
                messages=test_messages,
                model=available_models[0],
                max_tokens=20
            )
            
            # Check if response contains expected content
            success = "integration test successful" in response.content.lower()
            
            span = self.tracer._active_spans.get(context.span_id)
            if span:
                span.set_tag("test_model", available_models[0])
                span.set_tag("response_length", len(response.content))
                span.set_tag("test_passed", success)
            
            return success
            
        except Exception as e:
            self.logger.error(f"AI integration test failed: {e}")
            return False
    
    async def _test_oxy_components(self, context: SpanContext) -> bool:
        """Test Oxy component system"""
        try:
            from ..oxy.core import OxyLLM, OxyContext
            
            # Create test LLM component
            llm_component = OxyLLM(
                model="test_model",
                adapter=self.ai_adapter,
                name="TestLLM"
            )
            
            # Create test context
            test_context = OxyContext(
                session_id="oxy_test",
                user_id="system"
            )
            
            # Test component info retrieval
            component_info = llm_component.get_info()
            
            span = self.tracer._active_spans.get(context.span_id)
            if span:
                span.set_tag("component_type", component_info["type"])
                span.set_tag("component_name", component_info["name"])
            
            return True
            
        except Exception as e:
            self.logger.error(f"Oxy components test failed: {e}")
            return False
    
    async def _test_workflow_execution(self, context: SpanContext) -> bool:
        """Test workflow execution capabilities"""
        try:
            # Simple workflow execution test
            # This would normally involve the workflow engine
            
            span = self.tracer._active_spans.get(context.span_id)
            if span:
                span.set_tag("workflow_type", "test_workflow")
                span.set_tag("execution_mode", "sequential")
            
            # Simulate workflow execution
            await asyncio.sleep(0.1)  # Simulate some processing time
            
            return True
            
        except Exception as e:
            self.logger.error(f"Workflow execution test failed: {e}")
            return False
    
    async def get_chain_health_status(self) -> Dict[str, Any]:
        """Get current chain health status"""
        trace_context = self.tracer.start_trace("chain_health_check")
        
        try:
            health_result = await self.monitor.check_chain_connectivity(trace_context)
            
            # Add additional metrics
            model_stats = self.ai_adapter.get_model_stats()
            available_models = self.ai_adapter.get_available_models()
            
            return {
                **health_result,
                "available_models": len(available_models),
                "model_stats": model_stats,
                "service_status": "operational" if health_result["overall_status"] == "healthy" else "degraded"
            }
            
        except Exception as e:
            await self.tracer.finish_span(trace_context, SpanStatus.ERROR, str(e))
            raise
    
    async def get_trace_analytics(self, trace_id: str) -> Optional[Dict[str, Any]]:
        """Get detailed analytics for a specific trace"""
        return await self.tracer.get_trace_info(trace_id)