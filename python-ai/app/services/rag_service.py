"""
RAG (检索增强生成) 服务 - 集成先进RAG系统
"""

import asyncio
from typing import List, Dict, Any, Optional
from app.core.config import Settings
from app.core.logging import LoggerMixin
from app.core.exceptions import VectorDBException, DocumentProcessingException, RAGException
from app.rag.advanced_rag import AdvancedRAGSystem, create_advanced_rag_system

class RAGService(LoggerMixin):
    """RAG服务管理器 - 集成先进RAG系统"""
    
    def __init__(self, settings: Settings):
        self.settings = settings
        self.rag_system: Optional[AdvancedRAGSystem] = None
        
    async def startup(self):
        """启动RAG服务"""
        self.logger.info("Initializing Advanced RAG Service...")
        
        try:
            # 创建先进RAG系统
            self.rag_system = await create_advanced_rag_system(
                vector_store_type=getattr(self.settings, 'vector_store_type', 'chromadb'),
                enable_graph_retrieval=getattr(self.settings, 'enable_graph_retrieval', True),
                enable_advanced_reranking=getattr(self.settings, 'enable_advanced_reranking', True),
                enable_query_expansion=getattr(self.settings, 'enable_query_expansion', True)
            )
            
            self.logger.info("Advanced RAG Service initialized successfully")
            
        except Exception as e:
            self.logger.error(f"Failed to initialize RAG service: {e}")
            raise RAGException(f"RAG service initialization failed: {str(e)}")
    
    async def shutdown(self):
        """关闭RAG服务"""
        self.logger.info("Shutting down Advanced RAG Service...")
        
        try:
            if self.rag_system:
                await self.rag_system.shutdown()
            self.logger.info("Advanced RAG Service shutdown completed")
            
        except Exception as e:
            self.logger.error(f"Error during RAG service shutdown: {e}")
        
    async def query(
        self,
        user_id: str,
        query: str,
        top_k: int = 5,
        filters: Optional[Dict[str, Any]] = None,
        retrieval_mode: Optional[str] = None,
        session_id: Optional[str] = None
    ) -> Dict[str, Any]:
        """执行RAG查询"""
        
        if not self.rag_system:
            raise RAGException("RAG system not initialized")
        
        try:
            self.logger.info(f"Advanced RAG query for user {user_id}: {query}")
            
            # 执行RAG搜索
            response = await self.rag_system.search(
                query=query,
                user_id=user_id,
                top_k=top_k,
                retrieval_mode=retrieval_mode,
                filters=filters,
                session_id=session_id
            )
            
            # 转换响应格式
            result = {
                "query": response.query,
                "context": response.context,
                "documents": [
                    {
                        "id": result.chunk.id,
                        "content": result.chunk.content,
                        "source": result.chunk.source_doc_id,
                        "score": result.score,
                        "retrieval_method": result.retrieval_method,
                        "rank": result.rank,
                        "chunk_type": result.chunk.chunk_type.value,
                        "metadata": result.chunk.metadata
                    }
                    for result in response.results
                ],
                "metadata": {
                    "total_docs_searched": response.total_docs_searched,
                    "retrieval_time_ms": response.retrieval_time_ms,
                    "rerank_time_ms": response.rerank_time_ms,
                    "confidence_score": response.confidence_score,
                    "coverage_score": response.coverage_score,
                    "debug_info": response.debug_info
                }
            }
            
            self.logger.info(f"RAG query completed: {len(response.results)} results found")
            return result
            
        except Exception as e:
            self.logger.error(f"RAG query failed: {e}")
            raise RAGException(f"Query failed: {str(e)}")
    
    async def upload_document(
        self,
        user_id: str,
        filename: str,
        content: str,
        doc_type: str = "text",
        metadata: Optional[Dict[str, Any]] = None
    ) -> Dict[str, Any]:
        """上传和处理文档"""
        
        if not self.rag_system:
            raise RAGException("RAG system not initialized")
        
        try:
            self.logger.info(f"Document upload for user {user_id}: {filename}")
            
            # 构建文档对象
            document = {
                "id": f"doc_{user_id}_{filename}_{hash(content) % 100000}",
                "content": content,
                "filename": filename,
                "doc_type": doc_type,
                "user_id": user_id,
                "metadata": metadata or {}
            }
            
            # 添加到RAG系统
            results = await self.rag_system.add_documents([document])
            
            return {
                "document_id": document["id"],
                "status": "processed" if results["processed_chunks"] > 0 else "failed",
                "chunks_created": results["processed_chunks"],
                "processing_time": results["processing_time"],
                "message": f"Document processed successfully, created {results['processed_chunks']} chunks"
            }
            
        except Exception as e:
            self.logger.error(f"Document upload failed: {e}")
            raise DocumentProcessingException(f"Document upload failed: {str(e)}")
    
    async def upload_documents_batch(
        self,
        user_id: str,
        documents: List[Dict[str, Any]]
    ) -> Dict[str, Any]:
        """批量上传文档"""
        
        if not self.rag_system:
            raise RAGException("RAG system not initialized")
        
        try:
            self.logger.info(f"Batch document upload for user {user_id}: {len(documents)} documents")
            
            # 为每个文档添加用户ID和文档ID
            for i, doc in enumerate(documents):
                if "id" not in doc:
                    doc["id"] = f"doc_{user_id}_{i}_{hash(str(doc)) % 100000}"
                doc["user_id"] = user_id
            
            # 批量添加到RAG系统
            results = await self.rag_system.add_documents(documents)
            
            return {
                "total_documents": results["total_documents"],
                "processed_chunks": results["processed_chunks"],
                "failed_documents": results["failed_documents"],
                "processing_time": results["processing_time"],
                "status": "completed"
            }
            
        except Exception as e:
            self.logger.error(f"Batch document upload failed: {e}")
            raise DocumentProcessingException(f"Batch upload failed: {str(e)}")
    
    async def get_system_status(self) -> Dict[str, Any]:
        """获取RAG系统状态"""
        
        if not self.rag_system:
            return {"initialized": False, "error": "RAG system not initialized"}
        
        try:
            status = await self.rag_system.get_system_status()
            return status
            
        except Exception as e:
            self.logger.error(f"Failed to get system status: {e}")
            return {"error": str(e)}
    
    async def update_system_configuration(self, config: Dict[str, Any]) -> Dict[str, Any]:
        """更新系统配置"""
        
        if not self.rag_system:
            raise RAGException("RAG system not initialized")
        
        try:
            success = await self.rag_system.update_configuration(config)
            
            return {
                "status": "success" if success else "failed",
                "message": "Configuration updated successfully" if success else "Failed to update configuration"
            }
            
        except Exception as e:
            self.logger.error(f"Failed to update configuration: {e}")
            raise RAGException(f"Configuration update failed: {str(e)}")
    
    async def clear_user_data(self, user_id: str, data_type: str = "all") -> Dict[str, Any]:
        """清除用户数据"""
        
        if not self.rag_system:
            raise RAGException("RAG system not initialized")
        
        try:
            # 目前清除所有数据，后续可以实现按用户过滤
            success = await self.rag_system.clear_data(data_type)
            
            return {
                "status": "success" if success else "failed",
                "message": f"User data cleared successfully" if success else "Failed to clear user data"
            }
            
        except Exception as e:
            self.logger.error(f"Failed to clear user data: {e}")
            raise RAGException(f"Data clearing failed: {str(e)}")