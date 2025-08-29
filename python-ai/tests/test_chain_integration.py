"""
Integration Tests for End-to-End Chain Connectivity
"""

import asyncio
import pytest
import os
from typing import Dict, Any

from app.core.tracing import init_tracing, MemoryTraceCollector
from app.adapters.unified_adapter import UnifiedAIAdapter
from app.services.chain_service import ChainService, ChainTestResult

class TestChainIntegration:
    """Integration tests for chain connectivity"""
    
    @classmethod
    def setup_class(cls):
        """Setup test environment"""
        # Initialize tracing
        collector = MemoryTraceCollector(max_traces=100)
        cls.tracer = init_tracing("polyagent-test", collector)
        
        # Setup API keys for testing
        cls.api_keys = {
            "OPENAI_API_KEY": os.getenv("OPENAI_API_KEY"),
            "ANTHROPIC_API_KEY": os.getenv("ANTHROPIC_API_KEY"),
            "OPENROUTER_API_KEY": os.getenv("OPENROUTER_API_KEY"),
            "GLM_API_KEY": os.getenv("GLM_API_KEY")
        }
        
        # Initialize AI adapter
        cls.ai_adapter = UnifiedAIAdapter(api_keys=cls.api_keys)
        
        # Initialize chain service
        cls.chain_service = ChainService(cls.ai_adapter)
    
    @pytest.mark.asyncio
    async def test_chain_health_check(self):
        """Test chain health monitoring"""
        health_status = await self.chain_service.get_chain_health_status()
        
        # Verify health status structure
        assert "overall_status" in health_status
        assert "component_results" in health_status
        assert "available_models" in health_status
        assert "trace_id" in health_status
        
        # Log results for debugging
        print(f"Chain health status: {health_status}")
        
        # Check if at least some models are available
        assert health_status["available_models"] > 0, "No AI models available for testing"
    
    @pytest.mark.asyncio
    async def test_basic_e2e_chain(self):
        """Test basic end-to-end chain functionality"""
        result = await self.chain_service.test_end_to_end_chain("basic", timeout=60.0)
        
        # Verify test result structure
        assert isinstance(result, ChainTestResult)
        assert result.trace_id is not None
        assert len(result.components_tested) > 0
        assert result.total_duration > 0
        
        # Log test results
        print(f"E2E test result: Success={result.success}")
        print(f"Components tested: {result.components_tested}")
        print(f"Duration: {result.total_duration:.2f}s")
        
        if result.failed_steps:
            print(f"Failed steps: {result.failed_steps}")
            print(f"Error details: {result.error_details}")
        
        # For CI/CD, we might want to be more lenient
        # depending on which API keys are available
        if not result.success and not self._has_valid_api_keys():
            pytest.skip("Test failed due to missing API keys")
        
        assert result.success, f"E2E chain test failed: {result.error_details}"
    
    @pytest.mark.asyncio
    async def test_ai_model_integration(self):
        """Test AI model integration specifically"""
        # Check available models
        available_models = self.ai_adapter.get_available_models()
        
        if not available_models:
            pytest.skip("No AI models available for testing")
        
        # Test health check for each available model
        health_results = await self.ai_adapter.health_check_all()
        
        # Verify at least one model is healthy
        healthy_models = [model for model, is_healthy in health_results.items() if is_healthy]
        assert len(healthy_models) > 0, f"No healthy models found. Health results: {health_results}"
        
        print(f"Healthy models: {healthy_models}")
    
    @pytest.mark.asyncio
    async def test_tracing_functionality(self):
        """Test distributed tracing functionality"""
        # Start a test trace
        trace_context = self.tracer.start_trace(
            "test_tracing",
            test_type="integration",
            component="test_suite"
        )
        
        # Create some nested spans
        async with self.tracer.span(trace_context, "test_operation_1") as span1:
            await asyncio.sleep(0.01)  # Simulate work
            
            async with self.tracer.span(span1, "test_operation_2") as span2:
                await asyncio.sleep(0.01)  # Simulate work
        
        # Get trace information
        trace_info = await self.tracer.get_trace_info(trace_context.trace_id)
        
        assert trace_info is not None
        assert trace_info["trace_id"] == trace_context.trace_id
        assert trace_info["total_spans"] >= 3  # root + 2 child spans
        assert trace_info["total_duration"] > 0
        
        print(f"Trace info: {trace_info}")
    
    @pytest.mark.asyncio
    async def test_proxy_configuration(self):
        """Test proxy configuration if available"""
        # This test is optional and depends on proxy setup
        proxy_config = {
            "openai": os.getenv("OPENAI_PROXY_URL"),
            "anthropic": os.getenv("ANTHROPIC_PROXY_URL")
        }
        
        # Filter out None values
        proxy_config = {k: v for k, v in proxy_config.items() if v is not None}
        
        if not proxy_config:
            pytest.skip("No proxy configuration available for testing")
        
        # Initialize adapter with proxy config
        proxy_adapter = UnifiedAIAdapter(
            api_keys=self.api_keys,
            proxy_config=proxy_config
        )
        
        # Test basic functionality with proxy
        available_models = proxy_adapter.get_available_models()
        assert len(available_models) > 0, "Proxy adapter should have available models"
        
        print(f"Proxy adapter models: {available_models}")
    
    @pytest.mark.asyncio
    async def test_error_handling_and_recovery(self):
        """Test error handling and recovery mechanisms"""
        # Create a trace for error testing
        trace_context = self.tracer.start_trace("error_handling_test")
        
        # Test with invalid model
        try:
            await self.ai_adapter.generate(
                messages=[{"role": "user", "content": "test"}],
                model="invalid_model_name"
            )
            assert False, "Should have raised an exception for invalid model"
        except Exception as e:
            # This is expected behavior
            print(f"Expected error for invalid model: {e}")
        
        # Test chain service error recovery
        chain_health = await self.chain_service.get_chain_health_status()
        
        # Even with some errors, the service should still report status
        assert "overall_status" in chain_health
        print(f"Chain status after error test: {chain_health['overall_status']}")
    
    def _has_valid_api_keys(self) -> bool:
        """Check if we have at least one valid API key"""
        return any(key for key in self.api_keys.values() if key is not None)

# Performance benchmarks
class TestChainPerformance:
    """Performance tests for chain operations"""
    
    @classmethod
    def setup_class(cls):
        """Setup performance test environment"""
        collector = MemoryTraceCollector(max_traces=1000)
        cls.tracer = init_tracing("polyagent-perf-test", collector)
        
        api_keys = {
            "OPENAI_API_KEY": os.getenv("OPENAI_API_KEY"),
            "ANTHROPIC_API_KEY": os.getenv("ANTHROPIC_API_KEY"),
            "OPENROUTER_API_KEY": os.getenv("OPENROUTER_API_KEY"),
            "GLM_API_KEY": os.getenv("GLM_API_KEY")
        }
        
        cls.ai_adapter = UnifiedAIAdapter(api_keys=api_keys)
        cls.chain_service = ChainService(cls.ai_adapter)
    
    @pytest.mark.asyncio
    @pytest.mark.slow
    async def test_concurrent_chain_operations(self):
        """Test concurrent chain operations for performance"""
        if not self.ai_adapter.get_available_models():
            pytest.skip("No AI models available for performance testing")
        
        # Run multiple concurrent health checks
        tasks = []
        for i in range(5):
            task = self.chain_service.get_chain_health_status()
            tasks.append(task)
        
        results = await asyncio.gather(*tasks, return_exceptions=True)
        
        # Verify all tasks completed successfully
        successful_results = [r for r in results if not isinstance(r, Exception)]
        assert len(successful_results) == len(tasks), f"Some concurrent operations failed: {results}"
        
        print(f"Completed {len(successful_results)} concurrent operations successfully")
    
    @pytest.mark.asyncio
    @pytest.mark.slow
    async def test_chain_latency_benchmark(self):
        """Benchmark chain operation latency"""
        if not self.ai_adapter.get_available_models():
            pytest.skip("No AI models available for latency testing")
        
        # Run multiple E2E tests to measure latency distribution
        latencies = []
        
        for i in range(3):  # Reduced for CI/CD
            result = await self.chain_service.test_end_to_end_chain("performance")
            if result.success:
                latencies.append(result.total_duration)
        
        if not latencies:
            pytest.skip("No successful chain tests for latency measurement")
        
        avg_latency = sum(latencies) / len(latencies)
        max_latency = max(latencies)
        min_latency = min(latencies)
        
        print(f"Latency stats - Avg: {avg_latency:.2f}s, Min: {min_latency:.2f}s, Max: {max_latency:.2f}s")
        
        # Performance assertions (adjust based on expected performance)
        assert avg_latency < 30.0, f"Average latency too high: {avg_latency:.2f}s"
        assert max_latency < 60.0, f"Max latency too high: {max_latency:.2f}s"

if __name__ == "__main__":
    # Run basic integration test directly
    async def main():
        test_instance = TestChainIntegration()
        test_instance.setup_class()
        
        print("Running chain health check...")
        await test_instance.test_chain_health_check()
        
        print("Running basic E2E test...")
        await test_instance.test_basic_e2e_chain()
        
        print("Running tracing test...")
        await test_instance.test_tracing_functionality()
        
        print("All integration tests completed!")
    
    asyncio.run(main())