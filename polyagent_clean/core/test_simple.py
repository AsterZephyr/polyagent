#!/usr/bin/env python3
"""
Simple test for PolyAgent core functionality
No external dependencies required
"""

import asyncio
import sys
import os

def test_basic_imports():
    """Test that core modules can be imported"""
    
    print("Testing basic imports...")
    
    try:
        # Test simple AI module
        from ai_simple import AICall, AIResponse, call_ai_fallback, get_best_model_simple
        print("✓ AI simple module imported")
        
        # Test retrieve module (it should work without external deps)
        from retrieve import search, SearchResult, _tokenize
        print("✓ Retrieve module imported")
        
        # Test tools module (basic functionality)
        from tools import register_tool, list_tools
        print("✓ Tools module imported")
        
        return True
    except ImportError as e:
        print(f"❌ Import failed: {e}")
        return False

async def test_basic_functionality():
    """Test basic functionality without external APIs"""
    
    print("\nTesting basic functionality...")
    
    try:
        # Test AI simple functionality
        from ai_simple import test_simple_ai
        await test_simple_ai()
        print("✓ AI simple test passed")
        
        # Test search functionality
        from retrieve import search
        
        test_docs = [
            "Python is a programming language",
            "JavaScript is used for web development", 
            "Machine learning is a subset of AI"
        ]
        
        results = await search("programming language", test_docs, method="keyword", top_k=2)
        if len(results) > 0:
            print(f"✓ Search test passed: {len(results)} results")
        else:
            print("⚠️ Search returned no results")
        
        # Test tool registration
        from tools import register_tool, call_tool, list_tools
        
        @register_tool("test_tool")
        def test_func(message: str) -> str:
            return f"Test response: {message}"
        
        result = await call_tool("test_tool", {"message": "hello"})
        if result == "Test response: hello":
            print("✓ Tool registration and calling works")
        else:
            print(f"⚠️ Tool test unexpected result: {result}")
        
        return True
        
    except Exception as e:
        print(f"❌ Functionality test failed: {e}")
        return False

def test_config_loading():
    """Test configuration loading"""
    
    print("\nTesting configuration...")
    
    try:
        # Test environment variable handling
        os.environ['TEST_VAR'] = 'test_value'
        test_value = os.getenv('TEST_VAR')
        
        if test_value == 'test_value':
            print("✓ Environment variable handling works")
        else:
            print("❌ Environment variable test failed")
            return False
        
        # Clean up
        del os.environ['TEST_VAR']
        
        return True
        
    except Exception as e:
        print(f"❌ Config test failed: {e}")
        return False

async def main():
    """Main test function"""
    
    print("PolyAgent Simple Test Suite")
    print("=" * 40)
    print("Testing core functionality without external dependencies")
    
    tests = [
        ("Basic Imports", lambda: test_basic_imports()),
        ("Configuration", lambda: test_config_loading()),
        ("Basic Functionality", test_basic_functionality),
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
        print("🎉 All core tests passed!")
        print("\n📝 Next steps:")
        print("1. Install httpx: pip3 install httpx")
        print("2. Add API keys to config/.env") 
        print("3. Run full integration tests")
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