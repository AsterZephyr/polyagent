#!/usr/bin/env python3
"""
Simple test for PolyAgent
Like Unix test utilities - simple, fast, reliable
"""

import asyncio
import os
import sys
from main import PolyAgent, load_config

async def test_basic_functionality():
    """Test basic functionality without API calls"""
    
    print("Testing basic functionality...")
    
    # Test configuration loading
    config = load_config()
    print(f"✓ Configuration loaded")
    
    # Test imports
    from ai import AICall, AIResponse
    from retrieve import search, SearchResult  
    from tools import list_tools, register_tool
    
    print(f"✓ All modules imported successfully")
    
    # Test search without documents
    results = await search("test query", ["This is a test document", "Another test document"])
    print(f"✓ Search works: {len(results)} results")
    
    # Test tools
    tools = list_tools()
    print(f"✓ Tools available: {len(tools)} tools")
    
    # Test tool registration
    @register_tool("test_tool")
    def test_func(message: str) -> str:
        return f"Test: {message}"
    
    from tools import call_tool
    result = await call_tool("test_tool", {"message": "hello"})
    assert result == "Test: hello"
    print(f"✓ Tool registration and execution works")
    
    print("All basic tests passed!")
    return True

async def test_with_api_keys():
    """Test with actual API keys if available"""
    
    config = load_config()
    api_keys = {k: v for k, v in config['api_keys'].items() if v}
    
    if not api_keys:
        print("No API keys found - skipping API tests")
        return True
    
    print(f"Testing with API keys: {list(api_keys.keys())}")
    
    agent = PolyAgent(api_keys)
    
    # Health check
    health = await agent.health_check()
    print(f"Health check: {health['status']}")
    
    working_models = [k for k, v in health['models'].items() if v == 'working']
    if working_models:
        print(f"✓ Working models: {working_models}")
        
        # Simple chat test
        response = await agent.chat("Hello, this is a test. Please respond briefly.")
        print(f"✓ Chat test response: {response[:100]}...")
        
    else:
        print("No working models found")
        return False
    
    return True

def test_cli_parsing():
    """Test command line argument parsing"""
    
    # Test environment variables
    test_env = {
        'POLYAGENT_VERBOSE': '1',
        'POLYAGENT_TOOLS': 'true', 
        'POLYAGENT_DOCS': 'doc1.txt,doc2.txt'
    }
    
    # Temporarily set environment
    original_env = {}
    for key, value in test_env.items():
        original_env[key] = os.getenv(key)
        os.environ[key] = value
    
    try:
        config = load_config()
        assert config['verbose'] == True
        assert config['tools_enabled'] == True
        assert config['document_paths'] == ['doc1.txt', 'doc2.txt']
        print("✓ Environment variable parsing works")
    finally:
        # Restore environment
        for key, original_value in original_env.items():
            if original_value is None:
                os.environ.pop(key, None)
            else:
                os.environ[key] = original_value
    
    return True

async def main():
    """Main test function"""
    
    print("PolyAgent Simple Test Suite")
    print("=" * 40)
    
    tests = [
        ("Basic Functionality", test_basic_functionality),
        ("CLI Parsing", lambda: test_cli_parsing()),
        ("API Integration", test_with_api_keys),
    ]
    
    passed = 0
    total = len(tests)
    
    for test_name, test_func in tests:
        print(f"\n🧪 {test_name}:")
        try:
            if asyncio.iscoroutinefunction(test_func):
                result = await test_func()
            else:
                result = test_func()
            
            if result:
                print(f"✅ {test_name} PASSED")
                passed += 1
            else:
                print(f"❌ {test_name} FAILED")
        except Exception as e:
            print(f"💥 {test_name} CRASHED: {e}")
            import traceback
            traceback.print_exc()
    
    print("\n" + "=" * 40)
    print(f"Test Results: {passed}/{total} passed")
    
    if passed == total:
        print("🎉 All tests passed!")
        return 0
    else:
        print("❌ Some tests failed")
        return 1

if __name__ == "__main__":
    sys.exit(asyncio.run(main()))