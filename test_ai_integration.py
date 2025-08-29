#!/usr/bin/env python3
"""
Test AI integration with the refactored PolyAgent system
"""

import asyncio
import sys
import os

# Add agent directory to path
sys.path.append('agent')

async def test_ai_integration():
    """Test the AI integration"""
    
    print("🧪 Testing AI Integration")
    print("=" * 40)
    
    try:
        from agent.ai import AICall, call_ai
        print("✓ AI module imported successfully")
        
        # Test with supported model pattern
        test_call = AICall(
            model="claude-3-sonnet",  # Uses claude pattern
            messages=[{"role": "user", "content": "Hello, this is a test"}]
        )
        
        response = await call_ai(test_call, "test-key")
        print(f"✓ Mock AI call successful")
        print(f"  Response: {response.content[:100]}...")
        print(f"  Usage: {response.usage}")
        print(f"  Cost: ${response.cost}")
        
        return True
        
    except Exception as e:
        print(f"❌ AI integration test failed: {e}")
        import traceback
        traceback.print_exc()
        return False

async def test_full_system():
    """Test the full PolyAgent system"""
    
    print("\n🧪 Testing Full System")
    print("=" * 40)
    
    try:
        from agent.main import PolyAgent
        
        # Initialize agent with mock API keys
        api_keys = {
            'openai': 'test-key',
            'anthropic': 'test-key',
            'openrouter': 'test-key',
            'glm': 'test-key'
        }
        agent = PolyAgent(api_keys)
        print("✓ PolyAgent initialized")
        
        # Test simple chat
        response = await agent.chat("Hello, can you help me?", use_tools=False)
        print(f"✓ Chat test successful")
        print(f"  Response: {response[:100]}...")
        
        # Test with tools enabled
        response = await agent.chat("What time is it?", use_tools=True)
        print(f"✓ Tools test successful")
        print(f"  Response: {response[:100]}...")
        
        return True
        
    except Exception as e:
        print(f"❌ Full system test failed: {e}")
        import traceback
        traceback.print_exc()
        return False

async def main():
    """Main test function"""
    
    print("PolyAgent Integration Test Suite")
    print("=" * 50)
    
    # Change to project directory
    os.chdir('/Users/hxz/code/polyagent')
    
    tests = [
        ("AI Integration", test_ai_integration),
        ("Full System", test_full_system),
    ]
    
    passed = 0
    total = len(tests)
    
    for test_name, test_func in tests:
        try:
            result = await test_func()
            if result:
                print(f"\n✅ {test_name} PASSED")
                passed += 1
            else:
                print(f"\n❌ {test_name} FAILED")
        except Exception as e:
            print(f"\n💥 {test_name} CRASHED: {e}")
    
    print("\n" + "=" * 50)
    print(f"Integration Test Results: {passed}/{total} passed")
    
    if passed == total:
        print("🎉 All integration tests passed!")
        print("\nSystem is ready for production use!")
        return 0
    else:
        print("❌ Some tests failed")
        return 1

if __name__ == "__main__":
    try:
        exit_code = asyncio.run(main())
        sys.exit(exit_code)
    except KeyboardInterrupt:
        print("\nTest interrupted")
        sys.exit(130)
    except Exception as e:
        print(f"Fatal error: {e}")
        sys.exit(1)