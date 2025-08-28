"""
Advanced RAG Core Module
实现混合RAG架构：Vector RAG + Graph RAG + Keyword Search + Reranking
"""

from abc import ABC, abstractmethod
from typing import List, Dict, Any, Optional, Union, Tuple
from dataclasses import dataclass, field
from enum import Enum
import asyncio
import numpy as np
from datetime import datetime

from app.core.logging import LoggerMixin
from app.core.exceptions import RAGException

class RetrievalMode(Enum):
    """检索模式"""
    VECTOR_ONLY = "vector_only"
    KEYWORD_ONLY = "keyword_only"
    GRAPH_ONLY = "graph_only" 
    HYBRID_VECTOR_KEYWORD = "hybrid_vector_keyword"
    HYBRID_ALL = "hybrid_all"
    ADAPTIVE = "adaptive"

class ChunkType(Enum):
    """文档块类型"""
    PARAGRAPH = "paragraph"
    SENTENCE = "sentence"
    SECTION = "section"
    TABLE = "table" 
    CODE = "code"
    IMAGE_CAPTION = "image_caption"
    METADATA = "metadata"

@dataclass
class DocumentChunk:
    """文档块"""
    id: str
    content: str
    chunk_type: ChunkType
    source_doc_id: str
    metadata: Dict[str, Any] = field(default_factory=dict)
    
    # 向量相关
    embedding: Optional[np.ndarray] = None
    embedding_model: Optional[str] = None
    
    # 位置信息
    start_char: int = 0
    end_char: int = 0
    page_number: Optional[int] = None
    
    # 层次结构
    parent_chunk_id: Optional[str] = None
    child_chunk_ids: List[str] = field(default_factory=list)
    
    # 图结构
    entities: List[Dict[str, Any]] = field(default_factory=list)
    relations: List[Dict[str, Any]] = field(default_factory=list)
    
    # 质量评分
    quality_score: float = 1.0
    relevance_keywords: List[str] = field(default_factory=list)
    
    created_at: datetime = field(default_factory=datetime.now)

@dataclass 
class RetrievalResult:
    """检索结果"""
    chunk: DocumentChunk
    score: float
    retrieval_method: str
    rank: int = 0
    explanation: Optional[str] = None

@dataclass
class RAGQuery:
    """RAG查询"""
    query: str
    user_id: str
    session_id: Optional[str] = None
    
    # 检索参数
    top_k: int = 10
    retrieval_mode: RetrievalMode = RetrievalMode.ADAPTIVE
    
    # 过滤条件
    filters: Dict[str, Any] = field(default_factory=dict)
    date_range: Optional[Tuple[datetime, datetime]] = None
    
    # 查询增强
    query_expansion: bool = True
    semantic_search: bool = True
    
    # 重排序参数
    enable_reranking: bool = True
    rerank_top_k: int = 5
    
    # 上下文
    conversation_history: List[str] = field(default_factory=list)
    user_preferences: Dict[str, Any] = field(default_factory=dict)

@dataclass
class RAGResponse:
    """RAG响应"""
    query: str
    results: List[RetrievalResult]
    context: str
    
    # 元数据
    total_docs_searched: int = 0
    retrieval_time_ms: float = 0
    rerank_time_ms: float = 0
    
    # 置信度评分
    confidence_score: float = 0.0
    coverage_score: float = 0.0
    
    # 调试信息
    debug_info: Dict[str, Any] = field(default_factory=dict)

class BaseRetriever(ABC, LoggerMixin):
    """检索器基类"""
    
    @abstractmethod
    async def retrieve(
        self,
        query: RAGQuery,
        chunk_pool: List[DocumentChunk]
    ) -> List[RetrievalResult]:
        """执行检索"""
        pass
    
    @abstractmethod
    def get_retriever_name(self) -> str:
        """获取检索器名称"""
        pass

class BaseReranker(ABC, LoggerMixin):
    """重排序器基类"""
    
    @abstractmethod
    async def rerank(
        self,
        query: str,
        results: List[RetrievalResult]
    ) -> List[RetrievalResult]:
        """重排序结果"""
        pass

class HybridRAGEngine(LoggerMixin):
    """混合RAG引擎核心"""
    
    def __init__(self):
        self.retrievers: Dict[str, BaseRetriever] = {}
        self.rerankers: List[BaseReranker] = []
        
        # 检索权重配置
        self.retrieval_weights = {
            "vector": 0.6,
            "keyword": 0.3, 
            "graph": 0.1
        }
        
        # 自适应检索配置
        self.adaptive_config = {
            "query_length_threshold": 50,
            "entity_count_threshold": 3,
            "technical_keywords": ["algorithm", "implementation", "code", "API"]
        }
    
    def register_retriever(self, name: str, retriever: BaseRetriever):
        """注册检索器"""
        self.retrievers[name] = retriever
        self.logger.info(f"Registered retriever: {name}")
    
    def register_reranker(self, reranker: BaseReranker):
        """注册重排序器"""
        self.rerankers.append(reranker)
        self.logger.info(f"Registered reranker: {reranker.__class__.__name__}")
    
    async def retrieve(
        self,
        query: RAGQuery,
        document_chunks: List[DocumentChunk]
    ) -> RAGResponse:
        """执行混合检索"""
        
        start_time = datetime.now()
        
        try:
            # 1. 查询分析和模式选择
            selected_mode = self._select_retrieval_mode(query)
            self.logger.info(f"Selected retrieval mode: {selected_mode}")
            
            # 2. 查询增强
            enhanced_query = await self._enhance_query(query)
            
            # 3. 并行检索
            retrieval_results = await self._parallel_retrieve(
                enhanced_query, document_chunks, selected_mode
            )
            
            # 4. 结果融合
            fused_results = await self._fuse_results(retrieval_results, query.top_k)
            
            # 5. 重排序
            if query.enable_reranking and self.rerankers:
                fused_results = await self._rerank_results(
                    query.query, fused_results[:query.rerank_top_k * 2]
                )
            
            # 6. 后处理
            final_results = self._post_process_results(fused_results, query)
            
            # 7. 生成上下文
            context = self._generate_context(final_results)
            
            # 8. 计算质量指标
            response = RAGResponse(
                query=query.query,
                results=final_results[:query.top_k],
                context=context,
                total_docs_searched=len(document_chunks),
                retrieval_time_ms=(datetime.now() - start_time).total_seconds() * 1000
            )
            
            response.confidence_score = self._calculate_confidence(response)
            response.coverage_score = self._calculate_coverage(response, query)
            
            return response
            
        except Exception as e:
            self.logger.error(f"RAG retrieval failed: {str(e)}")
            raise RAGException(f"Retrieval failed: {str(e)}")
    
    def _select_retrieval_mode(self, query: RAGQuery) -> RetrievalMode:
        """智能选择检索模式"""
        
        if query.retrieval_mode != RetrievalMode.ADAPTIVE:
            return query.retrieval_mode
        
        query_text = query.query.lower()
        query_length = len(query_text.split())
        
        # 基于查询特征选择模式
        if query_length > self.adaptive_config["query_length_threshold"]:
            return RetrievalMode.HYBRID_ALL
        
        # 检查是否包含技术关键词
        technical_score = sum(
            1 for kw in self.adaptive_config["technical_keywords"]
            if kw in query_text
        )
        
        if technical_score > 0:
            return RetrievalMode.HYBRID_VECTOR_KEYWORD
        
        # 检查是否需要图检索（实体关系查询）
        if any(word in query_text for word in ["relationship", "connected", "related", "between"]):
            return RetrievalMode.HYBRID_ALL
        
        # 默认混合向量+关键词
        return RetrievalMode.HYBRID_VECTOR_KEYWORD
    
    async def _enhance_query(self, query: RAGQuery) -> RAGQuery:
        """查询增强"""
        
        if not query.query_expansion:
            return query
        
        enhanced_query = query.query
        
        # TODO: 实现查询扩展
        # 1. 同义词扩展
        # 2. 实体识别
        # 3. 意图分类
        # 4. 历史查询上下文
        
        # 创建增强后的查询对象
        enhanced = query
        enhanced.query = enhanced_query
        
        return enhanced
    
    async def _parallel_retrieve(
        self,
        query: RAGQuery,
        chunks: List[DocumentChunk],
        mode: RetrievalMode
    ) -> Dict[str, List[RetrievalResult]]:
        """并行检索"""
        
        tasks = []
        active_retrievers = []
        
        # 根据模式选择激活的检索器
        if mode in [RetrievalMode.VECTOR_ONLY, RetrievalMode.HYBRID_VECTOR_KEYWORD, RetrievalMode.HYBRID_ALL]:
            if "vector" in self.retrievers:
                tasks.append(self.retrievers["vector"].retrieve(query, chunks))
                active_retrievers.append("vector")
        
        if mode in [RetrievalMode.KEYWORD_ONLY, RetrievalMode.HYBRID_VECTOR_KEYWORD, RetrievalMode.HYBRID_ALL]:
            if "keyword" in self.retrievers:
                tasks.append(self.retrievers["keyword"].retrieve(query, chunks))
                active_retrievers.append("keyword")
        
        if mode in [RetrievalMode.GRAPH_ONLY, RetrievalMode.HYBRID_ALL]:
            if "graph" in self.retrievers:
                tasks.append(self.retrievers["graph"].retrieve(query, chunks))
                active_retrievers.append("graph")
        
        # 并行执行检索
        results = await asyncio.gather(*tasks, return_exceptions=True)
        
        # 组织结果
        retrieval_results = {}
        for i, (retriever_name, result) in enumerate(zip(active_retrievers, results)):
            if isinstance(result, Exception):
                self.logger.error(f"Retriever {retriever_name} failed: {str(result)}")
                retrieval_results[retriever_name] = []
            else:
                retrieval_results[retriever_name] = result
        
        return retrieval_results
    
    async def _fuse_results(
        self,
        retrieval_results: Dict[str, List[RetrievalResult]],
        top_k: int
    ) -> List[RetrievalResult]:
        """结果融合"""
        
        # 收集所有结果
        all_results = []
        
        for retriever_name, results in retrieval_results.items():
            weight = self.retrieval_weights.get(retriever_name, 1.0)
            
            for result in results:
                # 应用权重
                result.score *= weight
                result.retrieval_method = f"{result.retrieval_method}({retriever_name})"
                all_results.append(result)
        
        # 去重（基于chunk_id）
        unique_results = {}
        for result in all_results:
            chunk_id = result.chunk.id
            if chunk_id not in unique_results or result.score > unique_results[chunk_id].score:
                unique_results[chunk_id] = result
        
        # 排序
        fused_results = sorted(
            unique_results.values(),
            key=lambda x: x.score,
            reverse=True
        )
        
        return fused_results[:top_k * 3]  # 保留更多结果供重排序
    
    async def _rerank_results(
        self,
        query: str,
        results: List[RetrievalResult]
    ) -> List[RetrievalResult]:
        """重排序结果"""
        
        if not self.rerankers:
            return results
        
        reranked_results = results
        
        for reranker in self.rerankers:
            try:
                reranked_results = await reranker.rerank(query, reranked_results)
            except Exception as e:
                self.logger.error(f"Reranking failed: {str(e)}")
                continue
        
        return reranked_results
    
    def _post_process_results(
        self,
        results: List[RetrievalResult],
        query: RAGQuery
    ) -> List[RetrievalResult]:
        """结果后处理"""
        
        processed_results = []
        
        for i, result in enumerate(results):
            # 设置排名
            result.rank = i + 1
            
            # 应用过滤器
            if self._apply_filters(result, query.filters):
                processed_results.append(result)
        
        return processed_results
    
    def _apply_filters(
        self,
        result: RetrievalResult,
        filters: Dict[str, Any]
    ) -> bool:
        """应用过滤器"""
        
        if not filters:
            return True
        
        chunk = result.chunk
        
        # 文档类型过滤
        if "doc_types" in filters:
            doc_type = chunk.metadata.get("doc_type", "unknown")
            if doc_type not in filters["doc_types"]:
                return False
        
        # 日期范围过滤
        if "date_range" in filters:
            doc_date = chunk.metadata.get("created_date")
            if doc_date:
                start_date, end_date = filters["date_range"]
                if not (start_date <= doc_date <= end_date):
                    return False
        
        # 质量阈值过滤
        if "min_quality_score" in filters:
            if chunk.quality_score < filters["min_quality_score"]:
                return False
        
        return True
    
    def _generate_context(self, results: List[RetrievalResult]) -> str:
        """生成上下文"""
        
        context_parts = []
        
        for result in results:
            chunk = result.chunk
            
            # 添加来源信息
            source_info = f"[Source: {chunk.source_doc_id}"
            if chunk.page_number:
                source_info += f", Page: {chunk.page_number}"
            source_info += f", Score: {result.score:.3f}]"
            
            # 组装内容
            content = f"{source_info}\n{chunk.content}\n"
            context_parts.append(content)
        
        return "\n".join(context_parts)
    
    def _calculate_confidence(self, response: RAGResponse) -> float:
        """计算置信度"""
        
        if not response.results:
            return 0.0
        
        # 基于分数分布计算置信度
        scores = [r.score for r in response.results]
        
        if len(scores) == 1:
            return min(scores[0], 1.0)
        
        # 考虑分数差异和平均分
        avg_score = np.mean(scores)
        score_std = np.std(scores)
        
        # 分数越高，差异越小，置信度越高
        confidence = avg_score * (1.0 - min(score_std, 0.5))
        
        return min(confidence, 1.0)
    
    def _calculate_coverage(self, response: RAGResponse, query: RAGQuery) -> float:
        """计算覆盖率"""
        
        # 简化实现：基于返回结果数量
        expected_results = min(query.top_k, 10)
        actual_results = len(response.results)
        
        coverage = actual_results / expected_results
        
        return min(coverage, 1.0)