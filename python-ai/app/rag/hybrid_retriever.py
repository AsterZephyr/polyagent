"""
Hybrid Retrieval System
Combines semantic search with BM25 keyword search
"""

import asyncio
import math
from typing import List, Dict, Any, Optional, Tuple
from dataclasses import dataclass, field
from collections import Counter, defaultdict
import re
import numpy as np

from ..core.logging import LoggerMixin
from .core import DocumentChunk

@dataclass
class RetrievalResult:
    """Retrieval result with score breakdown"""
    chunk: DocumentChunk
    semantic_score: float
    bm25_score: float
    combined_score: float
    rank_fusion_score: float = 0.0
    metadata: Dict[str, Any] = field(default_factory=dict)

class BM25Retriever(LoggerMixin):
    """BM25 keyword-based retriever"""
    
    def __init__(self, k1: float = 1.2, b: float = 0.75):
        super().__init__()
        self.k1 = k1  # Term frequency saturation parameter
        self.b = b    # Length normalization parameter
        
        self.documents: List[DocumentChunk] = []
        self.doc_freqs: Dict[str, int] = {}
        self.idf_cache: Dict[str, float] = {}
        self.doc_lens: List[int] = []
        self.avgdl: float = 0.0
        
    def _tokenize(self, text: str) -> List[str]:
        """Simple tokenization - can be enhanced with proper NLP"""
        # Remove punctuation and split on whitespace
        text = re.sub(r'[^\w\s]', ' ', text.lower())
        return [token for token in text.split() if len(token) > 1]
    
    def add_documents(self, chunks: List[DocumentChunk]):
        """Add documents to BM25 index"""
        self.documents.extend(chunks)
        
        # Build term frequency dictionary
        all_doc_tokens = []
        for chunk in chunks:
            tokens = self._tokenize(chunk.content)
            all_doc_tokens.append(tokens)
            self.doc_lens.append(len(tokens))
            
            # Count document frequencies
            unique_tokens = set(tokens)
            for token in unique_tokens:
                self.doc_freqs[token] = self.doc_freqs.get(token, 0) + 1
        
        # Calculate average document length
        self.avgdl = sum(self.doc_lens) / len(self.doc_lens) if self.doc_lens else 0
        
        # Pre-calculate IDF scores
        total_docs = len(self.documents)
        for term, df in self.doc_freqs.items():
            self.idf_cache[term] = math.log((total_docs - df + 0.5) / (df + 0.5))
        
        self.logger.info(f"Built BM25 index for {total_docs} documents, "
                        f"{len(self.doc_freqs)} unique terms")
    
    def search(self, query: str, top_k: int = 10) -> List[Tuple[DocumentChunk, float]]:
        """BM25 search"""
        query_tokens = self._tokenize(query)
        query_token_counts = Counter(query_tokens)
        
        doc_scores = []
        
        for i, chunk in enumerate(self.documents):
            doc_tokens = self._tokenize(chunk.content)
            doc_token_counts = Counter(doc_tokens)
            doc_len = self.doc_lens[i]
            
            score = 0.0
            for token, query_tf in query_token_counts.items():
                if token in doc_token_counts:
                    doc_tf = doc_token_counts[token]
                    idf = self.idf_cache.get(token, 0)
                    
                    # BM25 formula
                    numerator = doc_tf * (self.k1 + 1)
                    denominator = doc_tf + self.k1 * (1 - self.b + self.b * doc_len / self.avgdl)
                    score += idf * numerator / denominator
            
            if score > 0:
                doc_scores.append((chunk, score))
        
        # Sort by score and return top_k
        doc_scores.sort(key=lambda x: x[1], reverse=True)
        return doc_scores[:top_k]

class HybridRetriever(LoggerMixin):
    """Hybrid retriever combining semantic and keyword search"""
    
    def __init__(self, 
                 vector_store,
                 embedding_model,
                 semantic_weight: float = 0.6,
                 bm25_weight: float = 0.4,
                 rrf_k: int = 60):  # Reciprocal Rank Fusion parameter
        super().__init__()
        self.vector_store = vector_store
        self.embedding_model = embedding_model
        self.bm25_retriever = BM25Retriever()
        
        # Fusion weights
        self.semantic_weight = semantic_weight
        self.bm25_weight = bm25_weight
        self.rrf_k = rrf_k
        
        # Ensure weights sum to 1
        total_weight = semantic_weight + bm25_weight
        self.semantic_weight /= total_weight
        self.bm25_weight /= total_weight
    
    async def add_documents(self, chunks: List[DocumentChunk]):
        """Add documents to both semantic and keyword indices"""
        # Add to vector store
        await self.vector_store.add_documents(chunks)
        
        # Add to BM25 index
        self.bm25_retriever.add_documents(chunks)
        
        self.logger.info(f"Added {len(chunks)} documents to hybrid index")
    
    async def search(self, 
                    query: str, 
                    top_k: int = 10,
                    fusion_method: str = "weighted") -> List[RetrievalResult]:
        """
        Hybrid search using multiple fusion methods
        
        Args:
            query: Search query
            top_k: Number of results to return
            fusion_method: "weighted", "rrf" (Reciprocal Rank Fusion), or "max"
        """
        # Generate query embedding
        query_embedding = await self._get_query_embedding(query)
        
        # Semantic search
        semantic_results = await self.vector_store.similarity_search(
            query_embedding, top_k=top_k*2  # Get more for fusion
        )
        
        # BM25 search
        bm25_results = self.bm25_retriever.search(query, top_k=top_k*2)
        
        # Combine results
        if fusion_method == "weighted":
            return await self._weighted_fusion(semantic_results, bm25_results, top_k)
        elif fusion_method == "rrf":
            return await self._reciprocal_rank_fusion(semantic_results, bm25_results, top_k)
        elif fusion_method == "max":
            return await self._max_fusion(semantic_results, bm25_results, top_k)
        else:
            raise ValueError(f"Unknown fusion method: {fusion_method}")
    
    async def _weighted_fusion(self, 
                             semantic_results: List[Tuple[DocumentChunk, float]], 
                             bm25_results: List[Tuple[DocumentChunk, float]], 
                             top_k: int) -> List[RetrievalResult]:
        """Weighted score fusion"""
        # Normalize scores to [0, 1]
        semantic_scores = self._normalize_scores([s[1] for s in semantic_results])
        bm25_scores = self._normalize_scores([s[1] for s in bm25_results])
        
        # Create score mapping
        semantic_map = {chunk.chunk_id: score for (chunk, _), score in 
                       zip(semantic_results, semantic_scores)}
        bm25_map = {chunk.chunk_id: score for (chunk, _), score in 
                   zip(bm25_results, bm25_scores)}
        
        # Combine scores
        all_chunks = {}
        for chunk, _ in semantic_results + bm25_results:
            if chunk.chunk_id not in all_chunks:
                all_chunks[chunk.chunk_id] = chunk
        
        combined_results = []
        for chunk_id, chunk in all_chunks.items():
            semantic_score = semantic_map.get(chunk_id, 0.0)
            bm25_score = bm25_map.get(chunk_id, 0.0)
            
            combined_score = (self.semantic_weight * semantic_score + 
                            self.bm25_weight * bm25_score)
            
            combined_results.append(RetrievalResult(
                chunk=chunk,
                semantic_score=semantic_score,
                bm25_score=bm25_score,
                combined_score=combined_score
            ))
        
        # Sort by combined score
        combined_results.sort(key=lambda x: x.combined_score, reverse=True)
        return combined_results[:top_k]
    
    async def _reciprocal_rank_fusion(self, 
                                    semantic_results: List[Tuple[DocumentChunk, float]], 
                                    bm25_results: List[Tuple[DocumentChunk, float]], 
                                    top_k: int) -> List[RetrievalResult]:
        """Reciprocal Rank Fusion (RRF)"""
        # Create rank mappings
        semantic_ranks = {chunk.chunk_id: rank for rank, (chunk, _) in enumerate(semantic_results)}
        bm25_ranks = {chunk.chunk_id: rank for rank, (chunk, _) in enumerate(bm25_results)}
        
        # Get all unique chunks
        all_chunks = {}
        for chunk, score in semantic_results:
            all_chunks[chunk.chunk_id] = (chunk, score, 0.0)
        for chunk, score in bm25_results:
            if chunk.chunk_id in all_chunks:
                all_chunks[chunk.chunk_id] = (all_chunks[chunk.chunk_id][0], 
                                            all_chunks[chunk.chunk_id][1], score)
            else:
                all_chunks[chunk.chunk_id] = (chunk, 0.0, score)
        
        # Calculate RRF scores
        rrf_results = []
        for chunk_id, (chunk, semantic_score, bm25_score) in all_chunks.items():
            rrf_score = 0.0
            
            if chunk_id in semantic_ranks:
                rrf_score += 1.0 / (self.rrf_k + semantic_ranks[chunk_id] + 1)
            
            if chunk_id in bm25_ranks:
                rrf_score += 1.0 / (self.rrf_k + bm25_ranks[chunk_id] + 1)
            
            rrf_results.append(RetrievalResult(
                chunk=chunk,
                semantic_score=semantic_score,
                bm25_score=bm25_score,
                combined_score=rrf_score,
                rank_fusion_score=rrf_score
            ))
        
        # Sort by RRF score
        rrf_results.sort(key=lambda x: x.combined_score, reverse=True)
        return rrf_results[:top_k]
    
    async def _max_fusion(self, 
                         semantic_results: List[Tuple[DocumentChunk, float]], 
                         bm25_results: List[Tuple[DocumentChunk, float]], 
                         top_k: int) -> List[RetrievalResult]:
        """Max score fusion"""
        # Normalize scores
        semantic_scores = self._normalize_scores([s[1] for s in semantic_results])
        bm25_scores = self._normalize_scores([s[1] for s in bm25_results])
        
        semantic_map = {chunk.chunk_id: score for (chunk, _), score in 
                       zip(semantic_results, semantic_scores)}
        bm25_map = {chunk.chunk_id: score for (chunk, _), score in 
                   zip(bm25_results, bm25_scores)}
        
        # Take max of available scores
        all_chunks = {}
        for chunk, _ in semantic_results + bm25_results:
            if chunk.chunk_id not in all_chunks:
                all_chunks[chunk.chunk_id] = chunk
        
        max_results = []
        for chunk_id, chunk in all_chunks.items():
            semantic_score = semantic_map.get(chunk_id, 0.0)
            bm25_score = bm25_map.get(chunk_id, 0.0)
            max_score = max(semantic_score, bm25_score)
            
            max_results.append(RetrievalResult(
                chunk=chunk,
                semantic_score=semantic_score,
                bm25_score=bm25_score,
                combined_score=max_score
            ))
        
        max_results.sort(key=lambda x: x.combined_score, reverse=True)
        return max_results[:top_k]
    
    def _normalize_scores(self, scores: List[float]) -> List[float]:
        """Min-max normalization"""
        if not scores:
            return []
        
        min_score = min(scores)
        max_score = max(scores)
        
        if max_score == min_score:
            return [1.0] * len(scores)
        
        return [(score - min_score) / (max_score - min_score) for score in scores]
    
    async def _get_query_embedding(self, query: str) -> np.ndarray:
        """Generate query embedding"""
        if hasattr(self.embedding_model, 'encode'):
            return self.embedding_model.encode([query])[0]
        else:
            # Async embedding model
            return await self.embedding_model.embed_query(query)

class RetrievalEvaluator(LoggerMixin):
    """Evaluation system for retrieval methods"""
    
    def __init__(self):
        super().__init__()
        self.metrics = {}
    
    def evaluate_retrieval(self, 
                         results: List[RetrievalResult], 
                         relevant_doc_ids: List[str],
                         at_k: List[int] = [1, 3, 5, 10]) -> Dict[str, float]:
        """
        Evaluate retrieval results using standard IR metrics
        
        Returns:
            - precision@k
            - recall@k  
            - MAP (Mean Average Precision)
            - MRR (Mean Reciprocal Rank)
        """
        retrieved_ids = [r.chunk.chunk_id for r in results]
        relevant_set = set(relevant_doc_ids)
        
        metrics = {}
        
        # Precision@k and Recall@k
        for k in at_k:
            retrieved_at_k = set(retrieved_ids[:k])
            relevant_retrieved = retrieved_at_k & relevant_set
            
            precision_k = len(relevant_retrieved) / k if k > 0 else 0
            recall_k = len(relevant_retrieved) / len(relevant_set) if relevant_set else 0
            
            metrics[f'precision@{k}'] = precision_k
            metrics[f'recall@{k}'] = recall_k
            
            # F1@k
            if precision_k + recall_k > 0:
                metrics[f'f1@{k}'] = 2 * precision_k * recall_k / (precision_k + recall_k)
            else:
                metrics[f'f1@{k}'] = 0.0
        
        # Mean Average Precision (MAP)
        ap = 0.0
        relevant_found = 0
        for i, doc_id in enumerate(retrieved_ids):
            if doc_id in relevant_set:
                relevant_found += 1
                precision_at_i = relevant_found / (i + 1)
                ap += precision_at_i
        
        metrics['map'] = ap / len(relevant_set) if relevant_set else 0.0
        
        # Mean Reciprocal Rank (MRR)
        mrr = 0.0
        for i, doc_id in enumerate(retrieved_ids):
            if doc_id in relevant_set:
                mrr = 1.0 / (i + 1)
                break
        metrics['mrr'] = mrr
        
        return metrics
    
    async def run_evaluation_suite(self, 
                                 retriever: HybridRetriever,
                                 test_queries: List[Dict[str, Any]]) -> Dict[str, Any]:
        """
        Run comprehensive evaluation
        
        test_queries format:
        [
            {
                "query": "search query",
                "relevant_docs": ["doc_id1", "doc_id2"],
                "category": "factual" | "conceptual" | "procedural"
            }
        ]
        """
        all_metrics = defaultdict(list)
        category_metrics = defaultdict(lambda: defaultdict(list))
        
        fusion_methods = ["weighted", "rrf", "max"]
        
        for query_data in test_queries:
            query = query_data["query"]
            relevant_docs = query_data["relevant_docs"]
            category = query_data.get("category", "general")
            
            for method in fusion_methods:
                results = await retriever.search(query, top_k=20, fusion_method=method)
                metrics = self.evaluate_retrieval(results, relevant_docs)
                
                for metric, value in metrics.items():
                    all_metrics[f"{method}_{metric}"].append(value)
                    category_metrics[category][f"{method}_{metric}"].append(value)
        
        # Calculate averages
        final_metrics = {}
        for metric, values in all_metrics.items():
            final_metrics[metric] = sum(values) / len(values)
        
        # Category breakdowns
        category_results = {}
        for category, metrics in category_metrics.items():
            category_results[category] = {}
            for metric, values in metrics.items():
                category_results[category][metric] = sum(values) / len(values)
        
        return {
            "overall_metrics": final_metrics,
            "category_metrics": category_results,
            "num_queries": len(test_queries)
        }