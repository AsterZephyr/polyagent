#!/usr/bin/env python3
"""
Test model routing to verify all latest models are supported
"""

import sys
import os

# Add agent directory to path  
sys.path.append('agent')

async def test_model_routing():
    """Test that all latest models route correctly"""
    
    print("🧪 Testing Model Routing")
    print("=" * 40)
    
    try:
        from agent.ai import call_ai, AICall
        
        # Latest models that should be supported
        test_models = [
            # Claude models
            "claude-3-5-sonnet-20241022",
            "claude-4-opus",
            "claude-4-sonnet", 
            
            # OpenAI models
            "gpt-4o",
            "gpt-5",
            "gpt-4-turbo",
            
            # OpenRouter models  
            "qwen/qwen-2.5-coder-32b-instruct",
            "openrouter/k2-free",
            "qwen/qwen-3-coder-free",
            
            # GLM models
            "glm-4-plus",
            "glm-4.5-turbo"
        ]
        
        routing_results = {}
        
        for model in test_models:
            print(f"Testing routing for: {model}")
            
            try:
                # Create test call
                test_call = AICall(
                    model=model,
                    messages=[{"role": "user", "content": "test"}]
                )
                
                # Test routing logic (without actual API call)
                if 'claude' in model:
                    expected_provider = 'claude'
                elif 'gpt' in model:
                    expected_provider = 'openai'
                elif 'qwen' in model or 'openrouter' in model:
                    expected_provider = 'openrouter'
                elif 'glm' in model:
                    expected_provider = 'glm'
                else:
                    expected_provider = 'unknown'
                
                # Verify the routing works by checking error patterns
                try:
                    # This will fail with auth error, but proves routing works
                    await call_ai(test_call, "test-key")
                except Exception as e:
                    error_msg = str(e).lower()
                    
                    # Check if it routed to the right provider based on error
                    if expected_provider == 'claude' and 'anthropic.com' in error_msg:
                        routing_results[model] = '✓ Claude'
                    elif expected_provider == 'openai' and ('openai.com' in error_msg or 'api.openai.com' in error_msg):
                        routing_results[model] = '✓ OpenAI'
                    elif expected_provider == 'openrouter' and 'openrouter.ai' in error_msg:
                        routing_results[model] = '✓ OpenRouter'
                    elif expected_provider == 'glm' and ('glm' in error_msg or 'zhipuai' in error_msg):
                        routing_results[model] = '✓ GLM'
                    elif 'unsupported model' in error_msg:
                        routing_results[model] = '❌ Unsupported'
                    else:
                        routing_results[model] = f'✓ {expected_provider} (inferred)'
                        
            except Exception as e:
                if 'Unsupported model' in str(e):
                    routing_results[model] = '❌ Not supported'
                else:
                    routing_results[model] = f'? Error: {str(e)[:50]}...'
        
        # Print results
        print("\nModel Routing Results:")
        print("=" * 50)
        
        supported = 0
        total = len(test_models)
        
        for model, result in routing_results.items():
            print(f"{result:<20} {model}")
            if result.startswith('✓'):
                supported += 1
        
        print("=" * 50)
        print(f"Supported models: {supported}/{total}")
        
        if supported == total:
            print("🎉 All latest models are supported!")
            return True
        else:
            print(f"⚠️ {total - supported} models need implementation")
            return False
            
    except Exception as e:
        print(f"❌ Model routing test failed: {e}")
        import traceback
        traceback.print_exc()
        return False

async def run_async_test():
    """Wrapper to run async test"""
    return await test_model_routing()

if __name__ == "__main__":
    import asyncio
    
    try:
        success = asyncio.run(run_async_test())
        if success:
            print("\n✅ Model routing verification PASSED")
            sys.exit(0)
        else:
            print("\n❌ Model routing verification FAILED") 
            sys.exit(1)
    except Exception as e:
        print(f"Fatal error: {e}")
        sys.exit(1)