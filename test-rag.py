#!/usr/bin/env python3
"""
RAG系统快速集成测试脚本
"""

import asyncio
import sys
import os

# 添加项目路径
sys.path.insert(0, os.path.join(os.path.dirname(__file__), 'python-ai'))

try:
    from python_ai.tests.test_rag_integration import main
    
    if __name__ == "__main__":
        print("🧪 PolyAgent RAG系统集成测试")
        print("=" * 50)
        
        asyncio.run(main())
        
except ImportError as e:
    print(f"❌ 导入错误: {e}")
    print("请确保在python-ai目录下运行此脚本")
    
    # 备用简单测试
    print("\n🔄 运行备用简单测试...")
    
    async def simple_test():
        """简单RAG测试"""
        try:
            # 直接导入和测试核心组件
            from app.rag.advanced_rag import create_advanced_rag_system
            
            print("✅ 成功导入RAG模块")
            
            # 创建最小化RAG系统
            rag_system = await create_advanced_rag_system(
                vector_store_type="chromadb",
                enable_graph_retrieval=False,
                enable_advanced_reranking=False,
                enable_query_expansion=False
            )
            
            print("✅ RAG系统创建成功")
            
            # 获取系统状态
            status = await rag_system.get_system_status()
            print(f"✅ 系统状态检查: {status.get('initialized', False)}")
            
            # 清理
            await rag_system.shutdown()
            print("✅ 简单测试完成")
            
        except Exception as e:
            print(f"❌ 简单测试失败: {e}")
            import traceback
            traceback.print_exc()
    
    asyncio.run(simple_test())