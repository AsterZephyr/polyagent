"""
Advanced Rerankers - 先进的重排序器
支持多种重排序策略：语义重排序、交叉编码器、多因子融合
"""

from typing import List, Dict, Any, Optional, Union
import asyncio
import numpy as np
from dataclasses import dataclass
import torch
import torch.nn.functional as F
from transformers import AutoTokenizer, AutoModel, AutoModelForSequenceClassification
from sentence_transformers import CrossEncoder
import math
from datetime import datetime

from app.core.logging import LoggerMixin  
from app.rag.core import BaseReranker, RetrievalResult


@dataclass
class RerankingFeatures:
    """重排序特征"""
    semantic_score: float = 0.0
    lexical_score: float = 0.0
    position_score: float = 0.0
    length_score: float = 0.0
    quality_score: float = 0.0
    recency_score: float = 0.0
    diversity_score: float = 0.0


class SemanticReranker(BaseReranker):
    """语义重排序器 - 使用交叉编码器进行深度语义匹配"""
    
    def __init__(
        self,
        model_name: str = "cross-encoder/ms-marco-MiniLM-L-2-v2",
        device: Optional[str] = None,
        batch_size: int = 32
    ):
        super().__init__()
        self.model_name = model_name
        self.device = device or ("cuda" if torch.cuda.is_available() else "cpu")
        self.batch_size = batch_size
        
        # 加载模型
        self._load_model()
    
    def _load_model(self):
        """加载交叉编码器模型"""
        try:
            self.cross_encoder = CrossEncoder(self.model_name, device=self.device)
            self.logger.info(f"Loaded cross-encoder model: {self.model_name}")
        except Exception as e:
            self.logger.error(f"Failed to load cross-encoder model: {e}")
            # 回退到基础模型
            self._load_fallback_model()
    
    def _load_fallback_model(self):
        """加载回退模型"""
        try:
            model_name = "sentence-transformers/all-MiniLM-L6-v2"
            self.tokenizer = AutoTokenizer.from_pretrained(model_name)
            self.model = AutoModel.from_pretrained(model_name)
            self.model.to(self.device)
            self.model.eval()
            self.cross_encoder = None
            self.logger.info(f"Loaded fallback model: {model_name}")
        except Exception as e:
            self.logger.error(f"Failed to load fallback model: {e}")
            self.tokenizer = None
            self.model = None
            self.cross_encoder = None
    
    async def rerank(
        self,
        query: str,
        results: List[RetrievalResult]
    ) -> List[RetrievalResult]:
        """语义重排序"""
        
        if not results:
            return results
        
        # 使用交叉编码器
        if self.cross_encoder:
            scores = await self._rerank_with_cross_encoder(query, results)
        elif self.model:
            scores = await self._rerank_with_similarity(query, results)
        else:
            self.logger.warning("No model available for semantic reranking")
            return results
        
        # 更新分数并重排序
        for i, result in enumerate(results):
            if i < len(scores):
                # 结合原始分数和语义分数
                original_score = result.score
                semantic_score = scores[i]
                result.score = 0.3 * original_score + 0.7 * semantic_score
                result.retrieval_method = f"{result.retrieval_method}+semantic_rerank"
        
        # 排序
        results.sort(key=lambda x: x.score, reverse=True)
        
        self.logger.info(f"Semantic reranking completed for {len(results)} results")
        return results
    
    async def _rerank_with_cross_encoder(
        self,
        query: str,
        results: List[RetrievalResult]
    ) -> List[float]:
        """使用交叉编码器重排序"""
        
        # 构造查询-文档对
        pairs = [(query, result.chunk.content) for result in results]
        
        # 分批处理
        scores = []
        for i in range(0, len(pairs), self.batch_size):
            batch_pairs = pairs[i:i + self.batch_size]
            
            # 异步执行
            batch_scores = await asyncio.get_event_loop().run_in_executor(
                None, self.cross_encoder.predict, batch_pairs
            )
            
            scores.extend(batch_scores.tolist() if hasattr(batch_scores, 'tolist') else batch_scores)
        
        # 归一化到 [0, 1]
        if scores:
            min_score, max_score = min(scores), max(scores)
            if max_score > min_score:
                scores = [(s - min_score) / (max_score - min_score) for s in scores]
        
        return scores
    
    async def _rerank_with_similarity(
        self,
        query: str,
        results: List[RetrievalResult]
    ) -> List[float]:
        """使用相似度重排序"""
        
        # 编码查询
        query_embedding = await self._encode_text(query)
        
        # 编码文档
        doc_embeddings = []
        for result in results:
            doc_embedding = await self._encode_text(result.chunk.content)
            doc_embeddings.append(doc_embedding)
        
        # 计算相似度
        scores = []
        for doc_embedding in doc_embeddings:
            similarity = self._cosine_similarity(query_embedding, doc_embedding)
            scores.append(similarity)
        
        return scores
    
    async def _encode_text(self, text: str) -> torch.Tensor:
        """编码文本"""
        
        def _encode():
            inputs = self.tokenizer(
                text,
                return_tensors="pt",
                max_length=512,
                truncation=True,
                padding=True
            )
            inputs = {k: v.to(self.device) for k, v in inputs.items()}
            
            with torch.no_grad():
                outputs = self.model(**inputs)
                # 使用 [CLS] token 的表示
                embeddings = outputs.last_hidden_state[:, 0, :]
                return embeddings.cpu()
        
        return await asyncio.get_event_loop().run_in_executor(None, _encode)
    
    def _cosine_similarity(self, tensor1: torch.Tensor, tensor2: torch.Tensor) -> float:
        """计算余弦相似度"""
        similarity = F.cosine_similarity(tensor1, tensor2, dim=1)
        return similarity.item()


class MultiFactorReranker(BaseReranker):
    """多因子重排序器 - 综合考虑多种因素"""
    
    def __init__(
        self,
        semantic_weight: float = 0.4,
        lexical_weight: float = 0.2,
        quality_weight: float = 0.15,
        recency_weight: float = 0.1,
        diversity_weight: float = 0.1,
        position_weight: float = 0.05
    ):
        super().__init__()
        self.weights = {
            "semantic": semantic_weight,
            "lexical": lexical_weight,
            "quality": quality_weight,
            "recency": recency_weight,
            "diversity": diversity_weight,
            "position": position_weight
        }
        
        # 内部语义重排序器
        self.semantic_reranker = SemanticReranker()
    
    async def rerank(
        self,
        query: str,
        results: List[RetrievalResult]
    ) -> List[RetrievalResult]:
        """多因子重排序"""
        
        if not results:
            return results
        
        # 计算各种特征分数
        features = await self._extract_features(query, results)
        
        # 计算综合分数
        for i, result in enumerate(results):
            feature = features[i]
            
            combined_score = (
                self.weights["semantic"] * feature.semantic_score +
                self.weights["lexical"] * feature.lexical_score +
                self.weights["quality"] * feature.quality_score +
                self.weights["recency"] * feature.recency_score +
                self.weights["diversity"] * feature.diversity_score +
                self.weights["position"] * feature.position_score
            )
            
            # 结合原始分数
            result.score = 0.3 * result.score + 0.7 * combined_score
            result.retrieval_method = f"{result.retrieval_method}+multifactor_rerank"
        
        # 排序
        results.sort(key=lambda x: x.score, reverse=True)
        
        self.logger.info(f"Multi-factor reranking completed for {len(results)} results")
        return results
    
    async def _extract_features(
        self,
        query: str,
        results: List[RetrievalResult]
    ) -> List[RerankingFeatures]:
        """提取重排序特征"""
        
        features = []
        
        # 语义分数（使用语义重排序器）
        semantic_results = await self.semantic_reranker.rerank(query, results.copy())
        semantic_scores = [r.score for r in semantic_results]
        
        # 词汇分数
        lexical_scores = self._calculate_lexical_scores(query, results)
        
        # 质量分数
        quality_scores = [r.chunk.quality_score for r in results]
        
        # 时效性分数
        recency_scores = self._calculate_recency_scores(results)
        
        # 多样性分数
        diversity_scores = self._calculate_diversity_scores(results)
        
        # 位置分数（原始排名）
        position_scores = self._calculate_position_scores(len(results))
        
        # 归一化所有分数
        semantic_scores = self._normalize_scores(semantic_scores)
        lexical_scores = self._normalize_scores(lexical_scores)
        quality_scores = self._normalize_scores(quality_scores)
        recency_scores = self._normalize_scores(recency_scores)
        diversity_scores = self._normalize_scores(diversity_scores)
        position_scores = self._normalize_scores(position_scores)
        
        # 构建特征对象
        for i in range(len(results)):
            feature = RerankingFeatures(
                semantic_score=semantic_scores[i] if i < len(semantic_scores) else 0.0,
                lexical_score=lexical_scores[i] if i < len(lexical_scores) else 0.0,
                quality_score=quality_scores[i] if i < len(quality_scores) else 0.0,
                recency_score=recency_scores[i] if i < len(recency_scores) else 0.0,
                diversity_score=diversity_scores[i] if i < len(diversity_scores) else 0.0,
                position_score=position_scores[i] if i < len(position_scores) else 0.0
            )
            features.append(feature)
        
        return features
    
    def _calculate_lexical_scores(
        self,
        query: str,
        results: List[RetrievalResult]
    ) -> List[float]:
        """计算词汇匹配分数"""
        
        query_words = set(query.lower().split())
        scores = []
        
        for result in results:
            content_words = set(result.chunk.content.lower().split())
            
            # 计算 Jaccard 相似度
            intersection = len(query_words & content_words)
            union = len(query_words | content_words)
            
            if union > 0:
                jaccard_score = intersection / union
            else:
                jaccard_score = 0.0
            
            # 计算词频权重
            tf_score = sum(
                result.chunk.content.lower().count(word) 
                for word in query_words
            ) / max(len(result.chunk.content.split()), 1)
            
            # 综合词汇分数
            lexical_score = 0.6 * jaccard_score + 0.4 * tf_score
            scores.append(lexical_score)
        
        return scores
    
    def _calculate_recency_scores(self, results: List[RetrievalResult]) -> List[float]:
        """计算时效性分数"""
        
        scores = []
        current_time = datetime.now()
        
        for result in results:
            # 使用文档创建时间
            doc_time = result.chunk.created_at
            time_diff = (current_time - doc_time).total_seconds()
            
            # 时间衰减函数（指数衰减）
            decay_factor = 0.1  # 控制衰减速度
            recency_score = math.exp(-decay_factor * time_diff / (24 * 3600))  # 按天衰减
            
            scores.append(recency_score)
        
        return scores
    
    def _calculate_diversity_scores(self, results: List[RetrievalResult]) -> List[float]:
        """计算多样性分数"""
        
        scores = []
        seen_sources = set()
        seen_types = set()
        
        for result in results:
            diversity_score = 1.0
            
            # 来源多样性
            source = result.chunk.source_doc_id
            if source in seen_sources:
                diversity_score *= 0.7
            else:
                seen_sources.add(source)
            
            # 类型多样性
            chunk_type = result.chunk.chunk_type.value
            if chunk_type in seen_types:
                diversity_score *= 0.8
            else:
                seen_types.add(chunk_type)
            
            scores.append(diversity_score)
        
        return scores
    
    def _calculate_position_scores(self, num_results: int) -> List[float]:
        """计算位置分数（原始排名权重）"""
        
        scores = []
        for i in range(num_results):
            # 位置分数随排名递减
            position_score = 1.0 / (1.0 + i)
            scores.append(position_score)
        
        return scores
    
    def _normalize_scores(self, scores: List[float]) -> List[float]:
        """归一化分数到 [0, 1]"""
        
        if not scores:
            return scores
        
        min_score = min(scores)
        max_score = max(scores)
        
        if max_score <= min_score:
            return [0.5] * len(scores)
        
        return [(s - min_score) / (max_score - min_score) for s in scores]


class DiversityReranker(BaseReranker):
    """多样性重排序器 - 确保结果的多样性"""
    
    def __init__(
        self,
        diversity_threshold: float = 0.8,
        max_same_source: int = 3
    ):
        super().__init__()
        self.diversity_threshold = diversity_threshold
        self.max_same_source = max_same_source
    
    async def rerank(
        self,
        query: str,
        results: List[RetrievalResult]
    ) -> List[RetrievalResult]:
        """多样性重排序"""
        
        if not results:
            return results
        
        # 使用贪心算法确保多样性
        diversified_results = []
        remaining_results = results.copy()
        source_counts = {}
        
        # 先添加最高分的结果
        if remaining_results:
            best_result = remaining_results.pop(0)
            diversified_results.append(best_result)
            source_counts[best_result.chunk.source_doc_id] = 1
        
        # 逐个添加剩余结果，考虑多样性
        for _ in range(len(remaining_results)):
            if not remaining_results:
                break
            
            best_candidate = None
            best_score = -1
            best_index = -1
            
            for i, candidate in enumerate(remaining_results):
                # 计算多样性奖励分数
                diversity_bonus = self._calculate_diversity_bonus(
                    candidate, diversified_results, source_counts
                )
                
                # 综合分数：原始分数 + 多样性奖励
                combined_score = candidate.score + diversity_bonus
                
                if combined_score > best_score:
                    best_score = combined_score
                    best_candidate = candidate
                    best_index = i
            
            # 添加最佳候选
            if best_candidate:
                diversified_results.append(best_candidate)
                remaining_results.pop(best_index)
                
                source_id = best_candidate.chunk.source_doc_id
                source_counts[source_id] = source_counts.get(source_id, 0) + 1
        
        # 更新排名
        for i, result in enumerate(diversified_results):
            result.rank = i + 1
            result.retrieval_method = f"{result.retrieval_method}+diversity_rerank"
        
        self.logger.info(f"Diversity reranking completed for {len(diversified_results)} results")
        return diversified_results
    
    def _calculate_diversity_bonus(
        self,
        candidate: RetrievalResult,
        selected_results: List[RetrievalResult],
        source_counts: Dict[str, int]
    ) -> float:
        """计算多样性奖励分数"""
        
        bonus = 0.0
        
        # 来源多样性奖励
        source_id = candidate.chunk.source_doc_id
        source_count = source_counts.get(source_id, 0)
        
        if source_count == 0:
            bonus += 0.2  # 新来源奖励
        elif source_count >= self.max_same_source:
            bonus -= 0.3  # 相同来源过多的惩罚
        
        # 内容相似度惩罚
        max_similarity = 0.0
        for selected_result in selected_results:
            similarity = self._content_similarity(
                candidate.chunk.content,
                selected_result.chunk.content
            )
            max_similarity = max(max_similarity, similarity)
        
        if max_similarity > self.diversity_threshold:
            bonus -= 0.2 * (max_similarity - self.diversity_threshold)
        
        # 类型多样性奖励
        candidate_type = candidate.chunk.chunk_type
        selected_types = {r.chunk.chunk_type for r in selected_results}
        
        if candidate_type not in selected_types:
            bonus += 0.1  # 新类型奖励
        
        return bonus
    
    def _content_similarity(self, content1: str, content2: str) -> float:
        """计算内容相似度（简化版）"""
        
        # 使用词袋模型计算 Jaccard 相似度
        words1 = set(content1.lower().split())
        words2 = set(content2.lower().split())
        
        intersection = len(words1 & words2)
        union = len(words1 | words2)
        
        return intersection / union if union > 0 else 0.0


class HybridReranker(BaseReranker):
    """混合重排序器 - 组合多个重排序策略"""
    
    def __init__(self):
        super().__init__()
        
        # 初始化各种重排序器
        self.semantic_reranker = SemanticReranker()
        self.multifactor_reranker = MultiFactorReranker()
        self.diversity_reranker = DiversityReranker()
        
        # 重排序器权重
        self.reranker_weights = {
            "semantic": 0.4,
            "multifactor": 0.4,
            "diversity": 0.2
        }
    
    async def rerank(
        self,
        query: str,
        results: List[RetrievalResult]
    ) -> List[RetrievalResult]:
        """混合重排序"""
        
        if not results:
            return results
        
        # 保存原始分数
        original_scores = [r.score for r in results]
        
        # 应用语义重排序
        semantic_results = await self.semantic_reranker.rerank(query, results.copy())
        semantic_scores = [r.score for r in semantic_results]
        
        # 应用多因子重排序
        multifactor_results = await self.multifactor_reranker.rerank(query, results.copy())
        multifactor_scores = [r.score for r in multifactor_results]
        
        # 应用多样性重排序
        diversity_results = await self.diversity_reranker.rerank(query, results.copy())
        diversity_scores = [r.score for r in diversity_results]
        
        # 融合分数
        final_scores = []
        for i in range(len(results)):
            weighted_score = (
                self.reranker_weights["semantic"] * semantic_scores[i] +
                self.reranker_weights["multifactor"] * multifactor_scores[i] +
                self.reranker_weights["diversity"] * diversity_scores[i]
            )
            final_scores.append(weighted_score)
        
        # 更新结果分数并排序
        for i, result in enumerate(results):
            result.score = final_scores[i]
            result.retrieval_method = f"{result.retrieval_method}+hybrid_rerank"
        
        results.sort(key=lambda x: x.score, reverse=True)
        
        # 更新排名
        for i, result in enumerate(results):
            result.rank = i + 1
        
        self.logger.info(f"Hybrid reranking completed for {len(results)} results")
        return results