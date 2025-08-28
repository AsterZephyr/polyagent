"""
Advanced RAG System Integration - 先进RAG系统集成
整合所有RAG组件：文档处理、向量存储、图检索、重排序、查询处理
"""

from typing import List, Dict, Any, Optional, Union
import asyncio
from datetime import datetime
import numpy as np

from app.core.logging import LoggerMixin
from app.core.exceptions import RAGException
from app.rag.core import (
    HybridRAGEngine, DocumentChunk, RAGQuery, RAGResponse,
    RetrievalMode, ChunkType
)
from app.rag.document_processor import DocumentProcessor, ProcessingConfig
from app.rag.vector_stores import ChromaDBStore, PineconeStore
from app.rag.graph_retriever import GraphRetriever
from app.rag.rerankers import (
    SemanticReranker, MultiFactorReranker, DiversityReranker, HybridReranker
)
from app.rag.query_processor import QueryProcessor


class AdvancedRAGSystem(LoggerMixin):
    """先进RAG系统"""
    
    def __init__(
        self,
        vector_store_type: str = "chromadb",
        enable_graph_retrieval: bool = True,
        enable_advanced_reranking: bool = True,
        enable_query_expansion: bool = True
    ):
        super().__init__()
        
        # 配置选项
        self.vector_store_type = vector_store_type
        self.enable_graph_retrieval = enable_graph_retrieval
        self.enable_advanced_reranking = enable_advanced_reranking
        self.enable_query_expansion = enable_query_expansion
        
        # 核心组件
        self.document_processor = None
        self.vector_store = None
        self.rag_engine = None
        self.graph_retriever = None
        self.query_processor = None
        
        # 重排序器
        self.rerankers = []
        
        # 初始化标志
        self.initialized = False
    
    async def initialize(self):
        """初始化RAG系统"""
        
        if self.initialized:
            return
        
        self.logger.info("Initializing Advanced RAG System...")
        
        # 1. 初始化文档处理器
        processing_config = ProcessingConfig(
            chunk_size=1000,
            chunk_overlap=200,
            enable_semantic_chunking=True,
            enable_quality_scoring=True,
            enable_keyword_extraction=True,
            languages=["zh", "en"]
        )
        self.document_processor = DocumentProcessor(processing_config)
        
        # 2. 初始化向量存储
        await self._initialize_vector_store()
        
        # 3. 初始化RAG引擎
        self.rag_engine = HybridRAGEngine()
        
        # 4. 注册向量检索器
        if self.vector_store:
            self.rag_engine.register_retriever("vector", self.vector_store)
        
        # 5. 注册图检索器
        if self.enable_graph_retrieval:
            self.graph_retriever = GraphRetriever()
            self.rag_engine.register_retriever("graph", self.graph_retriever)
        
        # 6. 初始化重排序器
        await self._initialize_rerankers()
        
        # 7. 初始化查询处理器
        if self.enable_query_expansion:
            self.query_processor = QueryProcessor()
        
        self.initialized = True
        self.logger.info("Advanced RAG System initialized successfully")
    
    async def _initialize_vector_store(self):
        """初始化向量存储"""
        
        try:
            if self.vector_store_type.lower() == "chromadb":
                self.vector_store = ChromaDBStore(
                    collection_name="polyagent_documents",
                    embedding_model="sentence-transformers/all-MiniLM-L6-v2"
                )
            elif self.vector_store_type.lower() == "pinecone":
                self.vector_store = PineconeStore(
                    index_name="polyagent-documents",
                    embedding_model="sentence-transformers/all-MiniLM-L6-v2"
                )
            else:
                raise RAGException(f"Unsupported vector store type: {self.vector_store_type}")
            
            await self.vector_store.initialize()
            self.logger.info(f"Vector store initialized: {self.vector_store_type}")
            
        except Exception as e:
            self.logger.error(f"Failed to initialize vector store: {e}")
            self.vector_store = None
    
    async def _initialize_rerankers(self):
        """初始化重排序器"""
        
        if not self.enable_advanced_reranking:
            return
        
        try:
            # 根据配置添加不同的重排序器
            self.rerankers = [
                HybridReranker()  # 使用混合重排序器作为默认
            ]
            
            # 注册到RAG引擎
            for reranker in self.rerankers:
                self.rag_engine.register_reranker(reranker)
            
            self.logger.info(f"Initialized {len(self.rerankers)} rerankers")
            
        except Exception as e:
            self.logger.error(f"Failed to initialize rerankers: {e}")
    
    async def add_documents(
        self,
        documents: List[Dict[str, Any]]
    ) -> Dict[str, Any]:
        """添加文档到RAG系统"""
        
        if not self.initialized:
            await self.initialize()
        
        self.logger.info(f"Adding {len(documents)} documents to RAG system")
        
        results = {
            "total_documents": len(documents),
            "processed_chunks": 0,
            "failed_documents": [],
            "processing_time": 0
        }
        
        start_time = datetime.now()
        
        try:
            all_chunks = []
            
            # 处理每个文档
            for doc_idx, document in enumerate(documents):
                try:
                    # 文档处理
                    chunks = await self.document_processor.process_document(document)
                    all_chunks.extend(chunks)
                    
                    self.logger.debug(f"Processed document {doc_idx}: {len(chunks)} chunks")
                    
                except Exception as e:
                    self.logger.error(f"Failed to process document {doc_idx}: {e}")
                    results["failed_documents"].append({
                        "index": doc_idx,
                        "error": str(e)
                    })
            
            # 存储到向量数据库
            if all_chunks and self.vector_store:
                await self.vector_store.add_chunks(all_chunks)
                self.logger.info(f"Added {len(all_chunks)} chunks to vector store")
            
            # 构建知识图谱
            if all_chunks and self.enable_graph_retrieval and self.graph_retriever:
                await self._build_knowledge_graph(all_chunks)
            
            results["processed_chunks"] = len(all_chunks)
            results["processing_time"] = (datetime.now() - start_time).total_seconds()
            
            self.logger.info(f"Document addition completed: {results}")
            return results
            
        except Exception as e:
            self.logger.error(f"Document addition failed: {e}")
            raise RAGException(f"Failed to add documents: {str(e)}")
    
    async def _build_knowledge_graph(self, chunks: List[DocumentChunk]):
        """构建知识图谱"""
        
        try:
            if self.graph_retriever:
                # 在后台构建图谱
                await asyncio.get_event_loop().run_in_executor(
                    None,
                    self.graph_retriever.knowledge_graph.build_from_chunks,
                    chunks
                )
                
                self.graph_retriever.graph_built = True
                self.logger.info("Knowledge graph construction completed")
        
        except Exception as e:
            self.logger.error(f"Knowledge graph construction failed: {e}")
    
    async def search(
        self,
        query: str,
        user_id: str,
        top_k: int = 10,
        retrieval_mode: Optional[str] = None,
        filters: Optional[Dict[str, Any]] = None,
        enable_reranking: bool = True,
        session_id: Optional[str] = None
    ) -> RAGResponse:
        """执行RAG搜索"""
        
        if not self.initialized:
            await self.initialize()
        
        self.logger.info(f"RAG search query: {query}")
        
        try:
            # 1. 查询处理和扩展
            processed_query = await self._process_query(query)
            
            # 2. 构建RAG查询
            rag_query = RAGQuery(
                query=processed_query.expanded_query if processed_query else query,
                user_id=user_id,
                session_id=session_id,
                top_k=top_k,
                retrieval_mode=self._parse_retrieval_mode(retrieval_mode),
                filters=filters or {},
                enable_reranking=enable_reranking
            )
            
            # 3. 获取文档块池
            document_chunks = await self._get_document_chunks()
            
            # 4. 执行混合检索
            response = await self.rag_engine.retrieve(rag_query, document_chunks)
            
            # 5. 增强响应信息
            if processed_query:
                response.debug_info["original_query"] = query
                response.debug_info["expanded_query"] = processed_query.expanded_query
                response.debug_info["query_entities"] = [
                    {"text": e.text, "type": e.entity_type}
                    for e in processed_query.entities
                ]
                response.debug_info["query_intent"] = processed_query.intent.value
            
            self.logger.info(f"RAG search completed: {len(response.results)} results")
            return response
            
        except Exception as e:
            self.logger.error(f"RAG search failed: {e}")
            raise RAGException(f"Search failed: {str(e)}")
    
    async def _process_query(self, query: str) -> Optional[Any]:
        """处理查询"""
        
        if not self.enable_query_expansion or not self.query_processor:
            return None
        
        try:
            processed_query = await self.query_processor.process_query(query)
            self.logger.debug(f"Query processed: {processed_query.expanded_query}")
            return processed_query
        
        except Exception as e:
            self.logger.warning(f"Query processing failed: {e}")
            return None
    
    def _parse_retrieval_mode(self, mode_str: Optional[str]) -> RetrievalMode:
        """解析检索模式"""
        
        if not mode_str:
            return RetrievalMode.ADAPTIVE
        
        mode_mapping = {
            "vector": RetrievalMode.VECTOR_ONLY,
            "keyword": RetrievalMode.KEYWORD_ONLY,
            "graph": RetrievalMode.GRAPH_ONLY,
            "hybrid": RetrievalMode.HYBRID_VECTOR_KEYWORD,
            "all": RetrievalMode.HYBRID_ALL,
            "adaptive": RetrievalMode.ADAPTIVE
        }
        
        return mode_mapping.get(mode_str.lower(), RetrievalMode.ADAPTIVE)
    
    async def _get_document_chunks(self) -> List[DocumentChunk]:
        """获取文档块池"""
        
        chunks = []
        
        # 从向量存储获取所有块
        if self.vector_store:
            try:
                # 这里需要vector store提供获取所有chunks的方法
                # 暂时返回空列表，实际实现中需要完善
                pass
            except Exception as e:
                self.logger.warning(f"Failed to get chunks from vector store: {e}")
        
        return chunks
    
    async def get_system_status(self) -> Dict[str, Any]:
        """获取系统状态"""
        
        status = {
            "initialized": self.initialized,
            "components": {
                "document_processor": self.document_processor is not None,
                "vector_store": self.vector_store is not None,
                "rag_engine": self.rag_engine is not None,
                "graph_retriever": self.graph_retriever is not None,
                "query_processor": self.query_processor is not None
            },
            "configuration": {
                "vector_store_type": self.vector_store_type,
                "enable_graph_retrieval": self.enable_graph_retrieval,
                "enable_advanced_reranking": self.enable_advanced_reranking,
                "enable_query_expansion": self.enable_query_expansion
            },
            "statistics": {}
        }
        
        # 获取组件统计信息
        if self.vector_store:
            try:
                vector_stats = await self.vector_store.get_statistics()
                status["statistics"]["vector_store"] = vector_stats
            except Exception as e:
                self.logger.warning(f"Failed to get vector store statistics: {e}")
        
        if self.graph_retriever and self.graph_retriever.graph_built:
            try:
                graph_stats = {
                    "entities_count": len(self.graph_retriever.knowledge_graph.entities),
                    "relations_count": len(self.graph_retriever.knowledge_graph.relations)
                }
                status["statistics"]["knowledge_graph"] = graph_stats
            except Exception as e:
                self.logger.warning(f"Failed to get knowledge graph statistics: {e}")
        
        return status
    
    async def update_configuration(self, config: Dict[str, Any]) -> bool:
        """更新系统配置"""
        
        try:
            # 更新检索权重
            if "retrieval_weights" in config and self.rag_engine:
                self.rag_engine.retrieval_weights.update(config["retrieval_weights"])
                self.logger.info("Updated retrieval weights")
            
            # 更新自适应配置
            if "adaptive_config" in config and self.rag_engine:
                self.rag_engine.adaptive_config.update(config["adaptive_config"])
                self.logger.info("Updated adaptive configuration")
            
            # 重新初始化重排序器（如果权重改变）
            if "reranker_weights" in config:
                await self._initialize_rerankers()
                self.logger.info("Reinitialized rerankers with new weights")
            
            return True
            
        except Exception as e:
            self.logger.error(f"Failed to update configuration: {e}")
            return False
    
    async def clear_data(self, data_type: str = "all") -> bool:
        """清除数据"""
        
        try:
            if data_type in ["all", "vector"]:
                if self.vector_store:
                    await self.vector_store.clear()
                    self.logger.info("Vector store data cleared")
            
            if data_type in ["all", "graph"]:
                if self.graph_retriever:
                    self.graph_retriever.knowledge_graph = type(self.graph_retriever.knowledge_graph)()
                    self.graph_retriever.graph_built = False
                    self.logger.info("Knowledge graph data cleared")
            
            return True
            
        except Exception as e:
            self.logger.error(f"Failed to clear data: {e}")
            return False
    
    async def shutdown(self):
        """关闭RAG系统"""
        
        self.logger.info("Shutting down Advanced RAG System...")
        
        try:
            # 关闭向量存储连接
            if self.vector_store and hasattr(self.vector_store, 'close'):
                await self.vector_store.close()
            
            # 清理其他资源
            self.document_processor = None
            self.vector_store = None
            self.rag_engine = None
            self.graph_retriever = None
            self.query_processor = None
            self.rerankers.clear()
            
            self.initialized = False
            self.logger.info("Advanced RAG System shutdown completed")
            
        except Exception as e:
            self.logger.error(f"Error during shutdown: {e}")


# 工厂函数
async def create_advanced_rag_system(
    vector_store_type: str = "chromadb",
    **kwargs
) -> AdvancedRAGSystem:
    """创建先进RAG系统实例"""
    
    system = AdvancedRAGSystem(
        vector_store_type=vector_store_type,
        **kwargs
    )
    
    await system.initialize()
    return system