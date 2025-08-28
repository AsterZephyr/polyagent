"""
Vector Database Integration
支持ChromaDB、Pinecone等向量数据库
"""

import asyncio
import json
from abc import ABC, abstractmethod
from typing import List, Dict, Any, Optional, Tuple
import numpy as np
from datetime import datetime
import hashlib

import chromadb
from chromadb.config import Settings as ChromaSettings
import pinecone
from sentence_transformers import SentenceTransformer

from app.core.logging import LoggerMixin
from app.core.exceptions import VectorDBException
from app.rag.core import DocumentChunk

class BaseVectorStore(ABC, LoggerMixin):
    """向量数据库基类"""
    
    @abstractmethod
    async def add_documents(self, chunks: List[DocumentChunk]) -> bool:
        """添加文档块"""
        pass
    
    @abstractmethod
    async def similarity_search(
        self,
        query_embedding: np.ndarray,
        top_k: int = 10,
        filters: Dict[str, Any] = None
    ) -> List[Tuple[DocumentChunk, float]]:
        """相似性搜索"""
        pass
    
    @abstractmethod
    async def delete_documents(self, doc_ids: List[str]) -> bool:
        """删除文档"""
        pass
    
    @abstractmethod
    async def update_document(self, chunk: DocumentChunk) -> bool:
        """更新文档块"""
        pass
    
    @abstractmethod
    async def get_collection_stats(self) -> Dict[str, Any]:
        """获取集合统计信息"""
        pass

class EmbeddingModel(LoggerMixin):
    """嵌入模型管理"""
    
    def __init__(self, model_name: str = "sentence-transformers/all-MiniLM-L6-v2"):
        self.model_name = model_name
        self.model = None
        self.dimension = None
        
    async def initialize(self):
        """初始化模型"""
        try:
            self.logger.info(f"Loading embedding model: {self.model_name}")
            self.model = SentenceTransformer(self.model_name)
            
            # 获取向量维度
            test_embedding = self.model.encode(["test"])
            self.dimension = test_embedding.shape[1]
            
            self.logger.info(f"Embedding model loaded, dimension: {self.dimension}")
            
        except Exception as e:
            raise VectorDBException("embedding_model_init", f"Failed to load embedding model: {str(e)}")
    
    async def encode_texts(self, texts: List[str]) -> np.ndarray:
        """编码文本"""
        if not self.model:
            await self.initialize()
        
        try:
            embeddings = self.model.encode(texts, show_progress_bar=False)
            return embeddings
        except Exception as e:
            raise VectorDBException("text_encoding", f"Failed to encode texts: {str(e)}")
    
    async def encode_text(self, text: str) -> np.ndarray:
        """编码单个文本"""
        embeddings = await self.encode_texts([text])
        return embeddings[0]

class ChromaDBStore(BaseVectorStore):
    """ChromaDB向量数据库"""
    
    def __init__(
        self,
        collection_name: str = "polyagent_docs",
        host: str = "localhost",
        port: int = 8001,
        embedding_model: str = "sentence-transformers/all-MiniLM-L6-v2"
    ):
        self.collection_name = collection_name
        self.host = host
        self.port = port
        self.client = None
        self.collection = None
        self.embedding_model = EmbeddingModel(embedding_model)
    
    async def initialize(self):
        """初始化ChromaDB连接"""
        try:
            # 初始化客户端
            self.client = chromadb.HttpClient(
                host=self.host,
                port=self.port,
                settings=ChromaSettings(anonymized_telemetry=False)
            )
            
            # 初始化嵌入模型
            await self.embedding_model.initialize()
            
            # 获取或创建集合
            try:
                self.collection = self.client.get_collection(
                    name=self.collection_name,
                    embedding_function=None  # 我们手动管理嵌入
                )
                self.logger.info(f"Connected to existing ChromaDB collection: {self.collection_name}")
            except:
                self.collection = self.client.create_collection(
                    name=self.collection_name,
                    embedding_function=None
                )
                self.logger.info(f"Created new ChromaDB collection: {self.collection_name}")
            
        except Exception as e:
            raise VectorDBException("chromadb_init", f"Failed to initialize ChromaDB: {str(e)}")
    
    async def add_documents(self, chunks: List[DocumentChunk]) -> bool:
        """添加文档块到ChromaDB"""
        
        if not self.collection:
            await self.initialize()
        
        if not chunks:
            return True
        
        try:
            # 准备数据
            ids = []
            texts = []
            embeddings = []
            metadatas = []
            
            for chunk in chunks:
                ids.append(chunk.id)
                texts.append(chunk.content)
                
                # 准备元数据（ChromaDB只接受基本类型）
                metadata = {
                    "source_doc_id": chunk.source_doc_id,
                    "chunk_type": chunk.chunk_type.value,
                    "quality_score": chunk.quality_score,
                    "start_char": chunk.start_char,
                    "end_char": chunk.end_char,
                    "page_number": chunk.page_number or 0,
                    "created_at": chunk.created_at.isoformat(),
                }
                
                # 添加扁平化的自定义元数据
                for key, value in chunk.metadata.items():
                    if isinstance(value, (str, int, float, bool)):
                        metadata[f"meta_{key}"] = value
                    elif isinstance(value, (list, dict)):
                        metadata[f"meta_{key}"] = json.dumps(value)
                
                metadatas.append(metadata)
            
            # 生成嵌入向量
            self.logger.info(f"Generating embeddings for {len(texts)} chunks...")
            chunk_embeddings = await self.embedding_model.encode_texts(texts)
            embeddings = chunk_embeddings.tolist()
            
            # 存储到ChromaDB
            self.collection.add(
                ids=ids,
                embeddings=embeddings,
                documents=texts,
                metadatas=metadatas
            )
            
            self.logger.info(f"Successfully added {len(chunks)} chunks to ChromaDB")
            return True
            
        except Exception as e:
            self.logger.error(f"Failed to add documents to ChromaDB: {str(e)}")
            raise VectorDBException("chromadb_add", str(e))
    
    async def similarity_search(
        self,
        query_embedding: np.ndarray,
        top_k: int = 10,
        filters: Dict[str, Any] = None
    ) -> List[Tuple[DocumentChunk, float]]:
        """相似性搜索"""
        
        if not self.collection:
            await self.initialize()
        
        try:
            # 构建查询过滤器
            where_clause = {}
            if filters:
                for key, value in filters.items():
                    if key == "doc_types" and isinstance(value, list):
                        where_clause["chunk_type"] = {"$in": value}
                    elif key == "min_quality_score":
                        where_clause["quality_score"] = {"$gte": value}
                    elif key.startswith("meta_"):
                        where_clause[key] = value
            
            # 执行搜索
            results = self.collection.query(
                query_embeddings=[query_embedding.tolist()],
                n_results=top_k,
                where=where_clause if where_clause else None,
                include=["documents", "metadatas", "distances"]
            )
            
            # 解析结果
            search_results = []
            
            if results["ids"] and results["ids"][0]:
                ids = results["ids"][0]
                documents = results["documents"][0]
                metadatas = results["metadatas"][0]
                distances = results["distances"][0]
                
                for i, (chunk_id, content, metadata, distance) in enumerate(
                    zip(ids, documents, metadatas, distances)
                ):
                    # 重构DocumentChunk
                    chunk = DocumentChunk(
                        id=chunk_id,
                        content=content,
                        chunk_type=ChunkType(metadata["chunk_type"]),
                        source_doc_id=metadata["source_doc_id"],
                        start_char=metadata.get("start_char", 0),
                        end_char=metadata.get("end_char", 0),
                        page_number=metadata.get("page_number") if metadata.get("page_number", 0) > 0 else None,
                        quality_score=metadata.get("quality_score", 1.0),
                        created_at=datetime.fromisoformat(metadata["created_at"])
                    )
                    
                    # 恢复自定义元数据
                    chunk.metadata = {}
                    for key, value in metadata.items():
                        if key.startswith("meta_"):
                            original_key = key[5:]  # 移除"meta_"前缀
                            try:
                                # 尝试解析JSON
                                chunk.metadata[original_key] = json.loads(value)
                            except:
                                chunk.metadata[original_key] = value
                    
                    # 转换距离为相似度分数 (ChromaDB使用L2距离)
                    similarity_score = 1.0 / (1.0 + distance)
                    
                    search_results.append((chunk, similarity_score))
            
            self.logger.info(f"ChromaDB search returned {len(search_results)} results")
            return search_results
            
        except Exception as e:
            self.logger.error(f"ChromaDB similarity search failed: {str(e)}")
            raise VectorDBException("chromadb_search", str(e))
    
    async def delete_documents(self, doc_ids: List[str]) -> bool:
        """删除文档"""
        
        if not self.collection:
            await self.initialize()
        
        try:
            # 查找要删除的块
            all_chunks = self.collection.get(
                where={"source_doc_id": {"$in": doc_ids}},
                include=["metadatas"]
            )
            
            if all_chunks["ids"]:
                self.collection.delete(ids=all_chunks["ids"])
                self.logger.info(f"Deleted {len(all_chunks['ids'])} chunks from {len(doc_ids)} documents")
            
            return True
            
        except Exception as e:
            self.logger.error(f"Failed to delete documents from ChromaDB: {str(e)}")
            raise VectorDBException("chromadb_delete", str(e))
    
    async def update_document(self, chunk: DocumentChunk) -> bool:
        """更新文档块"""
        
        # ChromaDB中的更新操作：先删除再添加
        try:
            await self.delete_documents([chunk.id])
            await self.add_documents([chunk])
            return True
        except Exception as e:
            raise VectorDBException("chromadb_update", str(e))
    
    async def get_collection_stats(self) -> Dict[str, Any]:
        """获取集合统计信息"""
        
        if not self.collection:
            await self.initialize()
        
        try:
            # 获取集合信息
            collection_info = self.collection.count()
            
            return {
                "total_chunks": collection_info,
                "collection_name": self.collection_name,
                "embedding_dimension": self.embedding_model.dimension,
                "embedding_model": self.embedding_model.model_name
            }
            
        except Exception as e:
            self.logger.error(f"Failed to get ChromaDB stats: {str(e)}")
            return {}

class PineconeStore(BaseVectorStore):
    """Pinecone向量数据库（云服务）"""
    
    def __init__(
        self,
        api_key: str,
        environment: str,
        index_name: str = "polyagent-docs",
        embedding_model: str = "sentence-transformers/all-MiniLM-L6-v2"
    ):
        self.api_key = api_key
        self.environment = environment
        self.index_name = index_name
        self.index = None
        self.embedding_model = EmbeddingModel(embedding_model)
    
    async def initialize(self):
        """初始化Pinecone连接"""
        try:
            # 初始化Pinecone
            pinecone.init(api_key=self.api_key, environment=self.environment)
            
            # 初始化嵌入模型
            await self.embedding_model.initialize()
            
            # 检查索引是否存在
            if self.index_name not in pinecone.list_indexes():
                # 创建索引
                pinecone.create_index(
                    name=self.index_name,
                    dimension=self.embedding_model.dimension,
                    metric="cosine",
                    pod_type="p1.x1"
                )
                self.logger.info(f"Created new Pinecone index: {self.index_name}")
            
            # 连接到索引
            self.index = pinecone.Index(self.index_name)
            self.logger.info(f"Connected to Pinecone index: {self.index_name}")
            
        except Exception as e:
            raise VectorDBException("pinecone_init", f"Failed to initialize Pinecone: {str(e)}")
    
    async def add_documents(self, chunks: List[DocumentChunk]) -> bool:
        """添加文档块到Pinecone"""
        
        if not self.index:
            await self.initialize()
        
        if not chunks:
            return True
        
        try:
            # 准备数据
            texts = [chunk.content for chunk in chunks]
            
            # 生成嵌入向量
            self.logger.info(f"Generating embeddings for {len(texts)} chunks...")
            embeddings = await self.embedding_model.encode_texts(texts)
            
            # 准备upsert数据
            vectors_to_upsert = []
            
            for chunk, embedding in zip(chunks, embeddings):
                # 准备元数据
                metadata = {
                    "content": chunk.content,
                    "source_doc_id": chunk.source_doc_id,
                    "chunk_type": chunk.chunk_type.value,
                    "quality_score": chunk.quality_score,
                    "start_char": chunk.start_char,
                    "end_char": chunk.end_char,
                    "created_at": chunk.created_at.isoformat(),
                }
                
                # 添加可选字段
                if chunk.page_number:
                    metadata["page_number"] = chunk.page_number
                
                # 添加扁平化的自定义元数据
                for key, value in chunk.metadata.items():
                    if isinstance(value, (str, int, float, bool)):
                        metadata[f"meta_{key}"] = value
                
                vectors_to_upsert.append((
                    chunk.id,
                    embedding.tolist(),
                    metadata
                ))
            
            # 批量插入
            self.index.upsert(vectors=vectors_to_upsert)
            
            self.logger.info(f"Successfully added {len(chunks)} chunks to Pinecone")
            return True
            
        except Exception as e:
            self.logger.error(f"Failed to add documents to Pinecone: {str(e)}")
            raise VectorDBException("pinecone_add", str(e))
    
    async def similarity_search(
        self,
        query_embedding: np.ndarray,
        top_k: int = 10,
        filters: Dict[str, Any] = None
    ) -> List[Tuple[DocumentChunk, float]]:
        """相似性搜索"""
        
        if not self.index:
            await self.initialize()
        
        try:
            # 构建查询过滤器
            filter_dict = {}
            if filters:
                for key, value in filters.items():
                    if key == "doc_types" and isinstance(value, list):
                        filter_dict["chunk_type"] = {"$in": value}
                    elif key == "min_quality_score":
                        filter_dict["quality_score"] = {"$gte": value}
                    elif key.startswith("meta_"):
                        filter_dict[key] = value
            
            # 执行搜索
            search_results = self.index.query(
                vector=query_embedding.tolist(),
                top_k=top_k,
                include_metadata=True,
                filter=filter_dict if filter_dict else None
            )
            
            # 解析结果
            results = []
            
            for match in search_results["matches"]:
                metadata = match["metadata"]
                
                # 重构DocumentChunk
                chunk = DocumentChunk(
                    id=match["id"],
                    content=metadata["content"],
                    chunk_type=ChunkType(metadata["chunk_type"]),
                    source_doc_id=metadata["source_doc_id"],
                    start_char=metadata.get("start_char", 0),
                    end_char=metadata.get("end_char", 0),
                    page_number=metadata.get("page_number"),
                    quality_score=metadata.get("quality_score", 1.0),
                    created_at=datetime.fromisoformat(metadata["created_at"])
                )
                
                # 恢复自定义元数据
                chunk.metadata = {}
                for key, value in metadata.items():
                    if key.startswith("meta_"):
                        original_key = key[5:]
                        chunk.metadata[original_key] = value
                
                results.append((chunk, match["score"]))
            
            self.logger.info(f"Pinecone search returned {len(results)} results")
            return results
            
        except Exception as e:
            self.logger.error(f"Pinecone similarity search failed: {str(e)}")
            raise VectorDBException("pinecone_search", str(e))
    
    async def delete_documents(self, doc_ids: List[str]) -> bool:
        """删除文档"""
        
        if not self.index:
            await self.initialize()
        
        try:
            # Pinecone需要通过filter删除
            for doc_id in doc_ids:
                self.index.delete(filter={"source_doc_id": doc_id})
            
            self.logger.info(f"Deleted documents: {doc_ids}")
            return True
            
        except Exception as e:
            self.logger.error(f"Failed to delete documents from Pinecone: {str(e)}")
            raise VectorDBException("pinecone_delete", str(e))
    
    async def update_document(self, chunk: DocumentChunk) -> bool:
        """更新文档块"""
        
        # Pinecone的upsert操作会自动更新
        return await self.add_documents([chunk])
    
    async def get_collection_stats(self) -> Dict[str, Any]:
        """获取集合统计信息"""
        
        if not self.index:
            await self.initialize()
        
        try:
            stats = self.index.describe_index_stats()
            
            return {
                "total_chunks": stats["total_vector_count"],
                "index_name": self.index_name,
                "dimension": stats["dimension"],
                "embedding_model": self.embedding_model.model_name
            }
            
        except Exception as e:
            self.logger.error(f"Failed to get Pinecone stats: {str(e)}")
            return {}

class VectorStoreFactory:
    """向量数据库工厂"""
    
    @staticmethod
    def create_vector_store(
        store_type: str,
        config: Dict[str, Any]
    ) -> BaseVectorStore:
        """创建向量数据库实例"""
        
        if store_type.lower() == "chromadb":
            return ChromaDBStore(
                collection_name=config.get("collection_name", "polyagent_docs"),
                host=config.get("host", "localhost"),
                port=config.get("port", 8001),
                embedding_model=config.get("embedding_model", "sentence-transformers/all-MiniLM-L6-v2")
            )
        
        elif store_type.lower() == "pinecone":
            if not config.get("api_key"):
                raise VectorDBException("config", "Pinecone API key is required")
            
            return PineconeStore(
                api_key=config["api_key"],
                environment=config.get("environment", "us-west1-gcp"),
                index_name=config.get("index_name", "polyagent-docs"),
                embedding_model=config.get("embedding_model", "sentence-transformers/all-MiniLM-L6-v2")
            )
        
        else:
            raise VectorDBException("config", f"Unsupported vector store type: {store_type}")