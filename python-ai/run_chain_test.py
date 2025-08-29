#!/usr/bin/env python3
"""
PolyAgent Chain Connectivity Test Runner
"""

import asyncio
import os
import sys
from pathlib import Path

# Add the app directory to the Python path
sys.path.insert(0, str(Path(__file__).parent))

from app.core.tracing import init_tracing, MemoryTraceCollector
from app.adapters.unified_adapter import UnifiedAIAdapter
from app.services.chain_service import ChainService

async def main():
    """Run chain connectivity tests"""
    
    print("🚀 Starting PolyAgent Chain Connectivity Test")
    print("=" * 50)
    
    # Initialize tracing system
    print("📊 Initializing tracing system...")
    collector = MemoryTraceCollector(max_traces=100)
    tracer = init_tracing("polyagent", collector)
    print(f"✅ Tracing initialized with service name: polyagent")
    
    # Setup API keys
    print("\n🔑 Checking API keys...")
    api_keys = {
        "OPENAI_API_KEY": os.getenv("OPENAI_API_KEY"),
        "ANTHROPIC_API_KEY": os.getenv("ANTHROPIC_API_KEY"), 
        "OPENROUTER_API_KEY": os.getenv("OPENROUTER_API_KEY"),
        "GLM_API_KEY": os.getenv("GLM_API_KEY")
    }
    
    available_keys = {k: v for k, v in api_keys.items() if v is not None}
    print(f"✅ Found {len(available_keys)} API keys: {list(available_keys.keys())}")
    
    if not available_keys:
        print("❌ No API keys found! Please set at least one of:")
        print("   - OPENAI_API_KEY")
        print("   - ANTHROPIC_API_KEY") 
        print("   - OPENROUTER_API_KEY")
        print("   - GLM_API_KEY")
        return
    
    # Initialize AI adapter with proxy support
    print("\n🤖 Initializing AI adapter...")
    proxy_config = {}
    if os.getenv("OPENAI_PROXY_URL"):
        proxy_config["openai"] = os.getenv("OPENAI_PROXY_URL")
        print(f"🔗 Using OpenAI proxy: {proxy_config['openai']}")
    
    ai_adapter = UnifiedAIAdapter(api_keys=available_keys, proxy_config=proxy_config)
    available_models = ai_adapter.get_available_models()
    print(f"✅ AI adapter initialized with {len(available_models)} models")
    
    if available_models:
        print("📋 Available models:")
        for model in available_models[:5]:  # Show first 5 models
            print(f"   - {model}")
        if len(available_models) > 5:
            print(f"   ... and {len(available_models) - 5} more")
    
    # Initialize chain service
    print("\n⛓️ Initializing chain service...")
    chain_service = ChainService(ai_adapter)
    print("✅ Chain service initialized")
    
    # Test 1: Chain Health Check
    print("\n🩺 Running chain health check...")
    try:
        health_status = await chain_service.get_chain_health_status()
        print(f"📊 Overall status: {health_status['overall_status']}")
        print(f"📦 Components checked: {len(health_status['component_results'])}")
        print(f"🤖 Available models: {health_status['available_models']}")
        
        for component, status in health_status['component_results'].items():
            status_icon = "✅" if status else "❌"
            print(f"   {status_icon} {component}: {'healthy' if status else 'unhealthy'}")
        
        if health_status['failed_components']:
            print(f"⚠️  Failed components: {health_status['failed_components']}")
        
        if health_status['broken_chains']:
            print(f"🔗💥 Broken chains: {health_status['broken_chains']}")
            
    except Exception as e:
        print(f"❌ Health check failed: {e}")
        return
    
    # Test 2: End-to-End Chain Test
    print("\n🔄 Running end-to-end chain test...")
    try:
        e2e_result = await chain_service.test_end_to_end_chain("basic", timeout=45.0)
        
        success_icon = "✅" if e2e_result.success else "❌"
        print(f"{success_icon} E2E test result: {'PASSED' if e2e_result.success else 'FAILED'}")
        print(f"⏱️  Total duration: {e2e_result.total_duration:.2f}s")
        print(f"🔧 Components tested: {len(e2e_result.components_tested)}")
        print(f"📋 Test components: {e2e_result.components_tested}")
        
        if e2e_result.failed_steps:
            print(f"❌ Failed steps: {e2e_result.failed_steps}")
            for step, error in e2e_result.error_details.items():
                print(f"   💥 {step}: {error}")
        
        # Show trace information
        if e2e_result.trace_id:
            trace_info = await chain_service.get_trace_analytics(e2e_result.trace_id)
            if trace_info:
                print(f"🔍 Trace ID: {e2e_result.trace_id}")
                print(f"📊 Total spans: {trace_info['total_spans']}")
                print(f"⚡ Components involved: {trace_info['components']}")
        
    except Exception as e:
        print(f"❌ E2E test failed with exception: {e}")
        import traceback
        traceback.print_exc()
        return
    
    # Test 3: AI Model Integration Test  
    print("\n🧠 Testing AI model integration...")
    try:
        if available_models:
            # Test basic generation with first available model
            test_messages = [{"role": "user", "content": "Hello! Please respond with 'AI integration test successful' if you can read this."}]
            
            response = await ai_adapter.generate(
                messages=test_messages,
                model=available_models[0],
                max_tokens=50
            )
            
            print(f"🤖 Model: {available_models[0]}")
            print(f"💬 Response: {response.content[:100]}...")
            print(f"📊 Tokens used: {response.usage}")
            print(f"💰 Estimated cost: ${response.cost_estimate:.6f}")
            
            if "integration test successful" in response.content.lower():
                print("✅ AI integration test PASSED")
            else:
                print("⚠️  AI responded but didn't include expected phrase")
        else:
            print("⏭️  No models available for integration test")
            
    except Exception as e:
        print(f"❌ AI integration test failed: {e}")
    
    # Summary
    print("\n" + "=" * 50)
    print("📊 CHAIN CONNECTIVITY TEST SUMMARY")
    print("=" * 50)
    
    if e2e_result.success:
        print("🎉 OVERALL RESULT: ✅ ALL TESTS PASSED")
        print("🔗 Chain connectivity: ESTABLISHED")
        print("📡 System status: OPERATIONAL")
    else:
        print("⚠️  OVERALL RESULT: ❌ SOME TESTS FAILED")
        print("🔗 Chain connectivity: PARTIAL")
        print("📡 System status: DEGRADED")
    
    print(f"\n📈 Performance metrics:")
    print(f"   ⏱️  E2E test duration: {e2e_result.total_duration:.2f}s")
    print(f"   🧪 Components tested: {len(e2e_result.components_tested)}")
    print(f"   🤖 Available AI models: {len(available_models)}")
    
    print(f"\n🔍 Trace information:")
    print(f"   📋 Trace ID: {e2e_result.trace_id}")
    print(f"   💾 Traces stored in memory for debugging")
    
    if not e2e_result.success:
        print(f"\n❌ Issues found:")
        for step in e2e_result.failed_steps:
            print(f"   🔧 {step}: {e2e_result.error_details.get(step, 'Unknown error')}")
    
    print("\n🚀 Chain connectivity test completed!")

if __name__ == "__main__":
    # Check if we're in the right directory
    if not Path("app").exists():
        print("❌ Error: Please run this script from the python-ai directory")
        print("Current directory should contain the 'app' folder")
        sys.exit(1)
    
    try:
        asyncio.run(main())
    except KeyboardInterrupt:
        print("\n⏹️  Test interrupted by user")
    except Exception as e:
        print(f"\n💥 Unexpected error: {e}")
        import traceback
        traceback.print_exc()
        sys.exit(1)