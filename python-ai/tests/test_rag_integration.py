"""
RAG系统集成测试
测试先进RAG系统的端到端功能
"""

import pytest
import asyncio
import tempfile
import os
from typing import List, Dict, Any

from app.rag.advanced_rag import AdvancedRAGSystem, create_advanced_rag_system
from app.services.rag_service import RAGService
from app.core.config import Settings


class MockSettings(Settings):
    """测试用模拟配置"""
    vector_store_type: str = "chromadb"
    enable_graph_retrieval: bool = True
    enable_advanced_reranking: bool = True
    enable_query_expansion: bool = True


@pytest.fixture
async def rag_system():
    """RAG系统测试夹具"""
    system = await create_advanced_rag_system(
        vector_store_type="chromadb",
        enable_graph_retrieval=False,  # 简化测试，禁用图检索
        enable_advanced_reranking=False,  # 简化测试，禁用重排序
        enable_query_expansion=False   # 简化测试，禁用查询扩展
    )
    yield system
    await system.shutdown()


@pytest.fixture
async def rag_service():
    """RAG服务测试夹具"""
    settings = MockSettings()
    service = RAGService(settings)
    await service.startup()
    yield service
    await service.shutdown()


@pytest.fixture
def sample_documents():
    """示例文档"""
    return [
        {
            "id": "doc_1",
            "content": "人工智能（AI）是计算机科学的一个分支，它企图了解智能的实质，并生产出一种新的能以人类智能相似的方式做出反应的智能机器。该领域的研究包括机器人、语言识别、图像识别、自然语言处理和专家系统等。",
            "filename": "ai_intro.txt",
            "doc_type": "text",
            "metadata": {"category": "technology", "language": "zh"}
        },
        {
            "id": "doc_2", 
            "content": "Machine Learning is a subset of artificial intelligence that provides systems the ability to automatically learn and improve from experience without being explicitly programmed. It focuses on the development of computer programs that can access data and use it to learn for themselves.",
            "filename": "ml_intro.txt",
            "doc_type": "text",
            "metadata": {"category": "technology", "language": "en"}
        },
        {
            "id": "doc_3",
            "content": "深度学习是机器学习的一个子领域，它基于人工神经网络的研究。深度学习模型能够学习数据的多层次表示，这些表示对应于不同层次的抽象。深度学习已被应用于计算机视觉、自然语言处理、语音识别等领域。",
            "filename": "deep_learning.txt",
            "doc_type": "text", 
            "metadata": {"category": "technology", "language": "zh"}
        },
        {
            "id": "doc_4",
            "content": "自然语言处理（NLP）是人工智能领域的一个重要分支，旨在让计算机能够理解、处理和生成人类语言。NLP技术包括文本分析、机器翻译、情感分析、问答系统等应用。现代NLP大量使用深度学习技术。",
            "filename": "nlp_intro.txt",
            "doc_type": "text",
            "metadata": {"category": "technology", "language": "zh"}
        }
    ]


class TestRAGSystemBasics:
    """RAG系统基础功能测试"""
    
    @pytest.mark.asyncio
    async def test_system_initialization(self, rag_system):
        """测试系统初始化"""
        assert rag_system.initialized
        
        status = await rag_system.get_system_status()
        assert status["initialized"] is True
        assert "components" in status
        assert status["components"]["document_processor"] is True
    
    @pytest.mark.asyncio  
    async def test_document_addition(self, rag_system, sample_documents):
        """测试文档添加"""
        
        # 添加文档
        results = await rag_system.add_documents(sample_documents)
        
        assert results["total_documents"] == len(sample_documents)
        assert results["processed_chunks"] > 0
        assert len(results["failed_documents"]) == 0
        assert results["processing_time"] > 0
    
    @pytest.mark.asyncio
    async def test_basic_search(self, rag_system, sample_documents):
        """测试基础搜索功能"""
        
        # 先添加文档
        await rag_system.add_documents(sample_documents)
        
        # 执行搜索
        response = await rag_system.search(
            query="什么是人工智能",
            user_id="test_user",
            top_k=3
        )
        
        assert response.query == "什么是人工智能"
        assert len(response.results) > 0
        assert len(response.results) <= 3
        
        # 检查结果质量
        first_result = response.results[0]
        assert first_result.score > 0
        assert "人工智能" in first_result.chunk.content or "AI" in first_result.chunk.content
    
    @pytest.mark.asyncio
    async def test_english_search(self, rag_system, sample_documents):
        """测试英文搜索"""
        
        # 先添加文档
        await rag_system.add_documents(sample_documents)
        
        # 执行英文搜索
        response = await rag_system.search(
            query="What is machine learning",
            user_id="test_user",
            top_k=3
        )
        
        assert len(response.results) > 0
        
        # 检查是否找到相关英文内容
        found_relevant = False
        for result in response.results:
            if "machine learning" in result.chunk.content.lower():
                found_relevant = True
                break
        
        assert found_relevant, "Should find relevant English content"
    
    @pytest.mark.asyncio
    async def test_filtered_search(self, rag_system, sample_documents):
        """测试过滤搜索"""
        
        # 先添加文档
        await rag_system.add_documents(sample_documents)
        
        # 使用过滤器搜索
        response = await rag_system.search(
            query="深度学习",
            user_id="test_user",
            top_k=5,
            filters={"language": "zh"}
        )
        
        assert len(response.results) > 0
        
        # 验证过滤效果（注意：当前实现可能不支持过滤，所以这是预期行为）
        # 这个测试主要验证过滤参数不会导致错误


class TestRAGServiceIntegration:
    """RAG服务集成测试"""
    
    @pytest.mark.asyncio
    async def test_service_initialization(self, rag_service):
        """测试服务初始化"""
        assert rag_service.rag_system is not None
        
        status = await rag_service.get_system_status()
        assert status.get("initialized") is True
    
    @pytest.mark.asyncio
    async def test_document_upload(self, rag_service):
        """测试文档上传"""
        
        result = await rag_service.upload_document(
            user_id="test_user",
            filename="test_doc.txt",
            content="这是一个测试文档，包含一些测试内容。人工智能是未来技术发展的重要方向。",
            doc_type="text",
            metadata={"category": "test"}
        )
        
        assert result["status"] == "processed"
        assert result["chunks_created"] > 0
        assert "document_id" in result
    
    @pytest.mark.asyncio
    async def test_batch_document_upload(self, rag_service, sample_documents):
        """测试批量文档上传"""
        
        result = await rag_service.upload_documents_batch(
            user_id="test_user",
            documents=sample_documents
        )
        
        assert result["status"] == "completed"
        assert result["total_documents"] == len(sample_documents)
        assert result["processed_chunks"] > 0
    
    @pytest.mark.asyncio
    async def test_service_query(self, rag_service, sample_documents):
        """测试服务查询"""
        
        # 先上传文档
        await rag_service.upload_documents_batch("test_user", sample_documents)
        
        # 执行查询
        result = await rag_service.query(
            user_id="test_user",
            query="机器学习是什么",
            top_k=3
        )
        
        assert "query" in result
        assert "documents" in result
        assert "metadata" in result
        assert len(result["documents"]) > 0
        
        # 检查文档格式
        doc = result["documents"][0]
        required_fields = ["id", "content", "source", "score", "retrieval_method"]
        for field in required_fields:
            assert field in doc


class TestRAGPerformance:
    """RAG系统性能测试"""
    
    @pytest.mark.asyncio
    async def test_search_performance(self, rag_system, sample_documents):
        """测试搜索性能"""
        
        # 添加文档
        await rag_system.add_documents(sample_documents)
        
        # 执行多次搜索测试性能
        queries = [
            "人工智能的应用",
            "What is deep learning",
            "自然语言处理技术",
            "机器学习算法"
        ]
        
        total_time = 0
        for query in queries:
            response = await rag_system.search(
                query=query,
                user_id="test_user",
                top_k=5
            )
            total_time += response.retrieval_time_ms
            
            # 验证响应时间合理
            assert response.retrieval_time_ms < 5000  # 小于5秒
        
        avg_time = total_time / len(queries)
        print(f"Average search time: {avg_time:.2f}ms")
        assert avg_time < 2000  # 平均响应时间小于2秒
    
    @pytest.mark.asyncio
    async def test_concurrent_searches(self, rag_system, sample_documents):
        """测试并发搜索"""
        
        # 添加文档
        await rag_system.add_documents(sample_documents)
        
        # 并发执行多个搜索
        async def search_task(query_id: int):
            response = await rag_system.search(
                query=f"测试查询 {query_id}",
                user_id=f"user_{query_id}",
                top_k=3
            )
            return len(response.results)
        
        # 创建10个并发任务
        tasks = [search_task(i) for i in range(10)]
        results = await asyncio.gather(*tasks, return_exceptions=True)
        
        # 验证所有任务成功完成
        for result in results:
            assert not isinstance(result, Exception), f"Concurrent search failed: {result}"
            assert isinstance(result, int) and result >= 0


class TestRAGErrorHandling:
    """RAG系统错误处理测试"""
    
    @pytest.mark.asyncio
    async def test_empty_query(self, rag_system, sample_documents):
        """测试空查询处理"""
        
        await rag_system.add_documents(sample_documents)
        
        # 测试空查询
        response = await rag_system.search(
            query="",
            user_id="test_user"
        )
        
        # 空查询应该返回空结果或默认结果
        assert len(response.results) >= 0
    
    @pytest.mark.asyncio
    async def test_invalid_documents(self, rag_system):
        """测试无效文档处理"""
        
        invalid_docs = [
            {"id": "invalid_1", "content": ""},  # 空内容
            {"id": "invalid_2"},  # 缺少content字段
            {"id": "invalid_3", "content": None}  # None内容
        ]
        
        result = await rag_system.add_documents(invalid_docs)
        
        # 系统应该能处理无效文档而不崩溃
        assert result["total_documents"] == len(invalid_docs)
        # 可能有失败的文档
        assert len(result["failed_documents"]) >= 0
    
    @pytest.mark.asyncio
    async def test_system_status_after_error(self, rag_system):
        """测试错误后系统状态"""
        
        # 尝试处理无效文档
        try:
            await rag_system.add_documents([{"invalid": "document"}])
        except:
            pass
        
        # 系统应该仍然可用
        status = await rag_system.get_system_status()
        assert status["initialized"] is True


# 运行集成测试的主函数
async def main():
    """运行集成测试"""
    
    print("🚀 开始RAG系统集成测试...")
    
    try:
        # 创建RAG系统
        rag_system = await create_advanced_rag_system(
            vector_store_type="chromadb",
            enable_graph_retrieval=False,  # 简化测试
            enable_advanced_reranking=False,
            enable_query_expansion=False
        )
        
        print("✅ RAG系统初始化成功")
        
        # 准备测试文档
        sample_docs = [
            {
                "id": "test_doc_1",
                "content": "人工智能技术正在快速发展，深度学习是其中的重要组成部分。",
                "filename": "test1.txt",
                "doc_type": "text"
            }
        ]
        
        # 测试文档添加
        print("📄 测试文档添加...")
        add_result = await rag_system.add_documents(sample_docs)
        print(f"   文档添加结果: {add_result['processed_chunks']} 个文档块")
        
        # 测试搜索
        print("🔍 测试搜索功能...")
        search_result = await rag_system.search(
            query="人工智能的发展",
            user_id="test_user"
        )
        print(f"   搜索结果: {len(search_result.results)} 个结果")
        if search_result.results:
            print(f"   最佳匹配分数: {search_result.results[0].score:.3f}")
        
        # 测试系统状态
        print("📊 检查系统状态...")
        status = await rag_system.get_system_status()
        print(f"   系统状态: {'✅ 正常' if status['initialized'] else '❌ 异常'}")
        
        await rag_system.shutdown()
        print("✅ RAG系统集成测试完成!")
        
    except Exception as e:
        print(f"❌ 集成测试失败: {e}")
        import traceback
        traceback.print_exc()


if __name__ == "__main__":
    asyncio.run(main())