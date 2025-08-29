"""
Retrieval - Simple search functionality
No fancy abstractions, just search documents
"""

import asyncio
import math
import re
from typing import List, Dict, Any, Tuple
from dataclasses import dataclass
from collections import Counter

@dataclass
class SearchResult:
    """Simple search result"""
    text: str
    score: float
    source: str = ""
    
    def __lt__(self, other):
        return self.score < other.score

# One function for search - that's it
async def search(query: str, 
                documents: List[str],
                method: str = "hybrid",
                top_k: int = 5) -> List[SearchResult]:
    """
    Search documents. Simple.
    
    Args:
        query: Search query
        documents: List of document texts
        method: "keyword", "semantic", or "hybrid" 
        top_k: Number of results to return
    """
    
    if not documents:
        return []
    
    if method == "keyword":
        return _bm25_search(query, documents, top_k)
    elif method == "semantic":
        return await _simple_semantic_search(query, documents, top_k)
    elif method == "hybrid":
        # Simple combination - no fancy fusion algorithms
        keyword_results = _bm25_search(query, documents, top_k * 2)
        semantic_results = await _simple_semantic_search(query, documents, top_k * 2)
        
        # Dead simple fusion: average scores
        combined = {}
        
        for result in keyword_results:
            combined[result.text] = {
                'text': result.text,
                'keyword_score': result.score,
                'semantic_score': 0.0,
                'source': result.source
            }
        
        for result in semantic_results:
            if result.text in combined:
                combined[result.text]['semantic_score'] = result.score
            else:
                combined[result.text] = {
                    'text': result.text,
                    'keyword_score': 0.0,
                    'semantic_score': result.score,
                    'source': result.source
                }
        
        # Simple weighted average
        final_results = []
        for doc_data in combined.values():
            # Normalize scores to 0-1 range
            combined_score = (doc_data['keyword_score'] * 0.4 + 
                            doc_data['semantic_score'] * 0.6)
            
            final_results.append(SearchResult(
                text=doc_data['text'],
                score=combined_score,
                source=doc_data['source']
            ))
        
        return sorted(final_results, key=lambda x: x.score, reverse=True)[:top_k]
    
    else:
        raise ValueError(f"Unknown method: {method}")

def _bm25_search(query: str, documents: List[str], top_k: int) -> List[SearchResult]:
    """
    BM25 keyword search - straightforward implementation
    No fancy parameters, just works
    """
    
    if not query.strip():
        return []
    
    # Simple tokenization
    query_tokens = _tokenize(query)
    if not query_tokens:
        return []
    
    # Tokenize all documents
    doc_tokens = [_tokenize(doc) for doc in documents]
    doc_lengths = [len(tokens) for tokens in doc_tokens]
    avgdl = sum(doc_lengths) / len(doc_lengths) if doc_lengths else 0
    
    # Count document frequencies
    df = Counter()  # document frequency
    for tokens in doc_tokens:
        unique_tokens = set(tokens)
        for token in unique_tokens:
            df[token] += 1
    
    # BM25 parameters
    k1 = 1.2
    b = 0.75
    N = len(documents)
    
    # Calculate scores
    scores = []
    for i, (doc, tokens) in enumerate(zip(documents, doc_tokens)):
        if not tokens:
            scores.append(0.0)
            continue
            
        tf = Counter(tokens)  # term frequency
        doc_len = doc_lengths[i]
        
        score = 0.0
        for token in query_tokens:
            if token in tf:
                # IDF
                idf = math.log((N - df[token] + 0.5) / (df[token] + 0.5))
                
                # TF component
                tf_component = (tf[token] * (k1 + 1)) / (
                    tf[token] + k1 * (1 - b + b * doc_len / avgdl)
                )
                
                score += idf * tf_component
        
        scores.append(score)
    
    # Create results
    results = []
    for i, score in enumerate(scores):
        if score > 0:
            results.append(SearchResult(
                text=documents[i],
                score=score,
                source=f"doc_{i}"
            ))
    
    # Sort and return top k
    return sorted(results, key=lambda x: x.score, reverse=True)[:top_k]

async def _simple_semantic_search(query: str, documents: List[str], top_k: int) -> List[SearchResult]:
    """
    Simple semantic search using basic word overlap
    In production, you'd use actual embeddings, but this works for demo
    """
    
    if not query.strip():
        return []
    
    query_words = set(_tokenize(query.lower()))
    if not query_words:
        return []
    
    results = []
    for i, doc in enumerate(documents):
        if not doc.strip():
            continue
            
        doc_words = set(_tokenize(doc.lower()))
        if not doc_words:
            continue
        
        # Simple Jaccard similarity
        intersection = query_words & doc_words
        union = query_words | doc_words
        
        if union:
            similarity = len(intersection) / len(union)
            
            if similarity > 0:
                results.append(SearchResult(
                    text=doc,
                    score=similarity,
                    source=f"doc_{i}"
                ))
    
    return sorted(results, key=lambda x: x.score, reverse=True)[:top_k]

def _tokenize(text: str) -> List[str]:
    """
    Simple tokenization
    Just split on whitespace and punctuation, filter short tokens
    """
    if not text:
        return []
    
    # Remove punctuation and split
    tokens = re.findall(r'\b\w+\b', text.lower())
    
    # Filter short tokens and common stop words
    stop_words = {'the', 'a', 'an', 'and', 'or', 'but', 'in', 'on', 'at', 'to', 'for', 'of', 'with', 'by',
                  'is', 'are', 'was', 'were', 'be', 'been', 'have', 'has', 'had', 'do', 'does', 'did',
                  '的', '了', '在', '是', '我', '你', '他', '她', '它', '们', '这', '那', '有', '没', '不', '很', '也', '就', '都', '会', '能', '可以'}
    
    return [token for token in tokens if len(token) > 1 and token not in stop_words]

# Simple document loader
def load_documents(file_paths: List[str]) -> List[str]:
    """Load documents from files - simple implementation"""
    documents = []
    
    for path in file_paths:
        try:
            with open(path, 'r', encoding='utf-8') as f:
                content = f.read().strip()
                if content:
                    # Split large documents into chunks
                    chunks = _chunk_text(content, max_length=1000)
                    documents.extend(chunks)
        except Exception as e:
            print(f"Warning: Could not load {path}: {e}")
    
    return documents

def _chunk_text(text: str, max_length: int = 1000, overlap: int = 100) -> List[str]:
    """Split text into overlapping chunks"""
    if len(text) <= max_length:
        return [text]
    
    chunks = []
    start = 0
    
    while start < len(text):
        end = start + max_length
        
        # Try to break at sentence boundary
        if end < len(text):
            # Look for sentence endings
            for i in range(end, start + max_length // 2, -1):
                if text[i] in '.!?。！？':
                    end = i + 1
                    break
        
        chunk = text[start:end].strip()
        if chunk:
            chunks.append(chunk)
        
        start = max(start + max_length - overlap, end)
        if start >= len(text):
            break
    
    return chunks