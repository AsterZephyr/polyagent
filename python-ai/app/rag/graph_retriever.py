"""
Graph RAG Retriever - 知识图谱检索器
基于实体关系图进行语义检索和关系推理
"""

from typing import List, Dict, Any, Optional, Set, Tuple
import asyncio
import json
import networkx as nx
from dataclasses import dataclass, field
from collections import defaultdict
import spacy
from spacy.matcher import Matcher
import re

from app.core.logging import LoggerMixin
from app.rag.core import BaseRetriever, DocumentChunk, RAGQuery, RetrievalResult


@dataclass
class Entity:
    """实体"""
    id: str
    name: str
    entity_type: str
    properties: Dict[str, Any] = field(default_factory=dict)
    mentions: List[str] = field(default_factory=list)
    frequency: int = 0
    importance_score: float = 0.0


@dataclass
class Relation:
    """关系"""
    id: str
    subject_id: str
    predicate: str
    object_id: str
    properties: Dict[str, Any] = field(default_factory=dict)
    confidence: float = 1.0
    source_chunk_ids: List[str] = field(default_factory=list)


class KnowledgeGraph(LoggerMixin):
    """知识图谱"""
    
    def __init__(self):
        self.graph = nx.MultiDiGraph()
        self.entities: Dict[str, Entity] = {}
        self.relations: Dict[str, Relation] = {}
        self.entity_index: Dict[str, Set[str]] = defaultdict(set)  # mention -> entity_ids
        
        # 加载 spaCy 模型
        try:
            self.nlp = spacy.load("zh_core_web_sm")
        except OSError:
            try:
                self.nlp = spacy.load("en_core_web_sm")
            except OSError:
                self.logger.warning("No spaCy model found, using blank model")
                self.nlp = spacy.blank("zh")
        
        # 设置实体匹配器
        self.matcher = Matcher(self.nlp.vocab)
        self._setup_patterns()
    
    def _setup_patterns(self):
        """设置实体识别模式"""
        
        # 技术实体模式
        tech_patterns = [
            [{"LOWER": {"IN": ["api", "sdk", "framework", "library", "algorithm"]}},
             {"IS_ALPHA": True, "OP": "?"}],
            [{"LOWER": "machine"}, {"LOWER": "learning"}],
            [{"LOWER": "deep"}, {"LOWER": "learning"}],
            [{"LOWER": "neural"}, {"LOWER": "network"}],
            [{"LOWER": "data"}, {"LOWER": {"IN": ["science", "mining", "analysis"]}}]
        ]
        
        # 组织机构模式
        org_patterns = [
            [{"IS_TITLE": True}, {"LOWER": {"IN": ["inc", "corp", "ltd", "company", "university"]}}],
            [{"IS_TITLE": True}, {"IS_TITLE": True, "OP": "?"}, {"LOWER": "团队"}],
            [{"IS_TITLE": True}, {"LOWER": "实验室"}]
        ]
        
        self.matcher.add("TECH_ENTITY", tech_patterns)
        self.matcher.add("ORG_ENTITY", org_patterns)
    
    def add_entity(self, entity: Entity):
        """添加实体"""
        self.entities[entity.id] = entity
        self.graph.add_node(entity.id, **entity.__dict__)
        
        # 建立提及索引
        for mention in entity.mentions:
            self.entity_index[mention.lower()].add(entity.id)
        
        self.logger.debug(f"Added entity: {entity.name} ({entity.entity_type})")
    
    def add_relation(self, relation: Relation):
        """添加关系"""
        self.relations[relation.id] = relation
        self.graph.add_edge(
            relation.subject_id,
            relation.object_id,
            key=relation.id,
            predicate=relation.predicate,
            **relation.properties
        )
        
        self.logger.debug(f"Added relation: {relation.predicate}")
    
    def find_entities_by_mention(self, text: str) -> List[Entity]:
        """通过提及文本查找实体"""
        text_lower = text.lower()
        entity_ids = set()
        
        # 精确匹配
        if text_lower in self.entity_index:
            entity_ids.update(self.entity_index[text_lower])
        
        # 模糊匹配
        for mention, ids in self.entity_index.items():
            if text_lower in mention or mention in text_lower:
                entity_ids.update(ids)
        
        return [self.entities[eid] for eid in entity_ids if eid in self.entities]
    
    def get_related_entities(
        self,
        entity_id: str,
        max_depth: int = 2,
        relation_types: Optional[List[str]] = None
    ) -> List[Tuple[str, int, str]]:
        """获取相关实体 (entity_id, distance, path_type)"""
        
        if entity_id not in self.graph:
            return []
        
        related = []
        visited = set()
        queue = [(entity_id, 0, "")]
        
        while queue:
            current_id, depth, path = queue.pop(0)
            
            if depth > max_depth or current_id in visited:
                continue
            
            visited.add(current_id)
            
            if depth > 0:  # 不包括起始实体
                related.append((current_id, depth, path))
            
            # 扩展邻居
            for neighbor in self.graph.neighbors(current_id):
                for edge_data in self.graph[current_id][neighbor].values():
                    predicate = edge_data.get('predicate', 'unknown')
                    
                    if relation_types is None or predicate in relation_types:
                        new_path = f"{path}-{predicate}" if path else predicate
                        queue.append((neighbor, depth + 1, new_path))
        
        return related
    
    def extract_entities_from_text(self, text: str) -> List[Entity]:
        """从文本中提取实体"""
        doc = self.nlp(text)
        entities = []
        
        # 使用 spaCy NER
        for ent in doc.ents:
            entity = Entity(
                id=f"entity_{len(self.entities)}_{hash(ent.text) % 10000}",
                name=ent.text,
                entity_type=ent.label_,
                mentions=[ent.text],
                properties={"start": ent.start_char, "end": ent.end_char}
            )
            entities.append(entity)
        
        # 使用模式匹配
        matches = self.matcher(doc)
        for match_id, start, end in matches:
            span = doc[start:end]
            label = self.nlp.vocab.strings[match_id]
            
            entity = Entity(
                id=f"entity_{len(self.entities)}_{hash(span.text) % 10000}",
                name=span.text,
                entity_type=label,
                mentions=[span.text],
                properties={"start": span.start_char, "end": span.end_char}
            )
            entities.append(entity)
        
        return entities
    
    def extract_relations_from_text(
        self,
        text: str,
        entities: List[Entity],
        chunk_id: str
    ) -> List[Relation]:
        """从文本中提取关系"""
        relations = []
        doc = self.nlp(text)
        
        # 简单的关系抽取基于依存句法
        for token in doc:
            if token.dep_ in ["nsubj", "dobj", "pobj"] and token.head.pos_ == "VERB":
                # 查找相关实体
                subj_entities = self._find_entities_in_span(entities, token.text)
                obj_entities = self._find_entities_in_span(entities, token.head.text)
                
                for subj_ent in subj_entities:
                    for obj_ent in obj_entities:
                        relation = Relation(
                            id=f"rel_{len(self.relations)}_{hash(f'{subj_ent.id}_{obj_ent.id}') % 10000}",
                            subject_id=subj_ent.id,
                            predicate=token.head.lemma_,
                            object_id=obj_ent.id,
                            confidence=0.7,
                            source_chunk_ids=[chunk_id]
                        )
                        relations.append(relation)
        
        return relations
    
    def _find_entities_in_span(self, entities: List[Entity], text: str) -> List[Entity]:
        """在文本片段中查找实体"""
        found = []
        for entity in entities:
            for mention in entity.mentions:
                if mention.lower() in text.lower() or text.lower() in mention.lower():
                    found.append(entity)
                    break
        return found
    
    def build_from_chunks(self, chunks: List[DocumentChunk]):
        """从文档块构建知识图谱"""
        
        self.logger.info(f"Building knowledge graph from {len(chunks)} chunks")
        
        all_entities = []
        
        # 第一轮：提取所有实体
        for chunk in chunks:
            entities = self.extract_entities_from_text(chunk.content)
            
            for entity in entities:
                # 检查是否已存在相似实体
                existing = self._find_similar_entity(entity)
                if existing:
                    # 合并实体
                    existing.mentions.extend(entity.mentions)
                    existing.frequency += 1
                else:
                    # 添加新实体
                    entity.frequency = 1
                    self.add_entity(entity)
                    all_entities.append(entity)
        
        # 第二轮：提取关系
        for chunk in chunks:
            chunk_entities = []
            for entity in all_entities:
                for mention in entity.mentions:
                    if mention.lower() in chunk.content.lower():
                        chunk_entities.append(entity)
            
            relations = self.extract_relations_from_text(
                chunk.content, chunk_entities, chunk.id
            )
            
            for relation in relations:
                self.add_relation(relation)
        
        # 计算实体重要性
        self._calculate_entity_importance()
        
        self.logger.info(f"Built knowledge graph: {len(self.entities)} entities, {len(self.relations)} relations")
    
    def _find_similar_entity(self, entity: Entity) -> Optional[Entity]:
        """查找相似实体"""
        for existing_entity in self.entities.values():
            if (existing_entity.name.lower() == entity.name.lower() or
                existing_entity.entity_type == entity.entity_type and
                self._similarity_score(existing_entity.name, entity.name) > 0.8):
                return existing_entity
        return None
    
    def _similarity_score(self, text1: str, text2: str) -> float:
        """计算文本相似度"""
        # 简单的字符级相似度
        if text1 == text2:
            return 1.0
        
        longer = text1 if len(text1) > len(text2) else text2
        shorter = text2 if len(text1) > len(text2) else text1
        
        if len(longer) == 0:
            return 0.0
        
        # 计算编辑距离
        edit_distance = self._levenshtein_distance(longer, shorter)
        return (len(longer) - edit_distance) / len(longer)
    
    def _levenshtein_distance(self, s1: str, s2: str) -> int:
        """计算编辑距离"""
        if len(s1) < len(s2):
            return self._levenshtein_distance(s2, s1)
        
        if len(s2) == 0:
            return len(s1)
        
        previous_row = list(range(len(s2) + 1))
        for i, c1 in enumerate(s1):
            current_row = [i + 1]
            for j, c2 in enumerate(s2):
                insertions = previous_row[j + 1] + 1
                deletions = current_row[j] + 1
                substitutions = previous_row[j] + (c1 != c2)
                current_row.append(min(insertions, deletions, substitutions))
            previous_row = current_row
        
        return previous_row[-1]
    
    def _calculate_entity_importance(self):
        """计算实体重要性分数"""
        
        # 基于图中心性计算重要性
        try:
            centrality = nx.degree_centrality(self.graph)
            betweenness = nx.betweenness_centrality(self.graph)
            
            for entity_id, entity in self.entities.items():
                degree_score = centrality.get(entity_id, 0)
                between_score = betweenness.get(entity_id, 0)
                frequency_score = entity.frequency / max(ent.frequency for ent in self.entities.values())
                
                entity.importance_score = (
                    0.4 * degree_score +
                    0.3 * between_score +
                    0.3 * frequency_score
                )
        
        except Exception as e:
            self.logger.error(f"Failed to calculate entity importance: {e}")
            # 回退到基于频率的计算
            max_freq = max(ent.frequency for ent in self.entities.values()) if self.entities else 1
            for entity in self.entities.values():
                entity.importance_score = entity.frequency / max_freq


class GraphRetriever(BaseRetriever):
    """图检索器"""
    
    def __init__(self):
        super().__init__()
        self.knowledge_graph = KnowledgeGraph()
        self.graph_built = False
    
    def get_retriever_name(self) -> str:
        return "graph"
    
    async def retrieve(
        self,
        query: RAGQuery,
        chunk_pool: List[DocumentChunk]
    ) -> List[RetrievalResult]:
        """执行图检索"""
        
        # 构建知识图谱（如果尚未构建）
        if not self.graph_built:
            await self._build_graph_if_needed(chunk_pool)
        
        # 从查询中识别实体
        query_entities = self.knowledge_graph.extract_entities_from_text(query.query)
        
        if not query_entities:
            self.logger.info("No entities found in query, falling back to entity mention matching")
            query_entities = self._find_query_entities_by_keywords(query.query)
        
        if not query_entities:
            self.logger.warning("No entities found for graph retrieval")
            return []
        
        # 扩展相关实体
        related_entities = set()
        for entity in query_entities:
            if entity.id in self.knowledge_graph.entities:
                related = self.knowledge_graph.get_related_entities(
                    entity.id,
                    max_depth=2
                )
                related_entities.update([rel[0] for rel in related])
                related_entities.add(entity.id)
        
        # 检索包含相关实体的文档块
        results = []
        for chunk in chunk_pool:
            score = self._calculate_graph_relevance_score(
                chunk, related_entities, query_entities
            )
            
            if score > 0:
                result = RetrievalResult(
                    chunk=chunk,
                    score=score,
                    retrieval_method="graph",
                    explanation=f"Contains entities: {self._get_chunk_entities(chunk, related_entities)}"
                )
                results.append(result)
        
        # 按分数排序
        results.sort(key=lambda x: x.score, reverse=True)
        
        self.logger.info(f"Graph retrieval found {len(results)} relevant chunks")
        return results[:query.top_k * 2]  # 返回更多结果供融合
    
    async def _build_graph_if_needed(self, chunks: List[DocumentChunk]):
        """按需构建知识图谱"""
        
        if self.graph_built:
            return
        
        # 在后台构建图谱
        await asyncio.get_event_loop().run_in_executor(
            None, self.knowledge_graph.build_from_chunks, chunks
        )
        
        self.graph_built = True
        self.logger.info("Knowledge graph construction completed")
    
    def _find_query_entities_by_keywords(self, query: str) -> List[Entity]:
        """通过关键词匹配查找查询实体"""
        
        query_words = query.lower().split()
        found_entities = []
        
        for word in query_words:
            entities = self.knowledge_graph.find_entities_by_mention(word)
            found_entities.extend(entities)
        
        # 去重并按重要性排序
        unique_entities = list({ent.id: ent for ent in found_entities}.values())
        unique_entities.sort(key=lambda x: x.importance_score, reverse=True)
        
        return unique_entities[:5]  # 限制实体数量
    
    def _calculate_graph_relevance_score(
        self,
        chunk: DocumentChunk,
        related_entity_ids: Set[str],
        query_entities: List[Entity]
    ) -> float:
        """计算图相关性分数"""
        
        # 检查块中包含的实体
        chunk_entities = self._get_chunk_entity_ids(chunk, related_entity_ids)
        
        if not chunk_entities:
            return 0.0
        
        # 基础分数：包含实体数量比例
        base_score = len(chunk_entities) / max(len(related_entity_ids), 1)
        
        # 查询实体权重
        query_entity_ids = {ent.id for ent in query_entities if ent.id in self.knowledge_graph.entities}
        query_match_bonus = len(chunk_entities & query_entity_ids) / max(len(query_entity_ids), 1)
        
        # 实体重要性权重
        importance_bonus = sum(
            self.knowledge_graph.entities[eid].importance_score
            for eid in chunk_entities
            if eid in self.knowledge_graph.entities
        ) / max(len(chunk_entities), 1)
        
        # 综合分数
        score = (
            0.4 * base_score +
            0.4 * query_match_bonus +
            0.2 * importance_bonus
        )
        
        return min(score, 1.0)
    
    def _get_chunk_entity_ids(
        self,
        chunk: DocumentChunk,
        candidate_entity_ids: Set[str]
    ) -> Set[str]:
        """获取文档块中包含的实体ID"""
        
        chunk_entities = set()
        content_lower = chunk.content.lower()
        
        for entity_id in candidate_entity_ids:
            if entity_id in self.knowledge_graph.entities:
                entity = self.knowledge_graph.entities[entity_id]
                
                # 检查实体提及是否在块中出现
                for mention in entity.mentions:
                    if mention.lower() in content_lower:
                        chunk_entities.add(entity_id)
                        break
        
        return chunk_entities
    
    def _get_chunk_entities(
        self,
        chunk: DocumentChunk,
        entity_ids: Set[str]
    ) -> List[str]:
        """获取文档块中包含的实体名称列表"""
        
        chunk_entity_ids = self._get_chunk_entity_ids(chunk, entity_ids)
        
        return [
            self.knowledge_graph.entities[eid].name
            for eid in chunk_entity_ids
            if eid in self.knowledge_graph.entities
        ][:5]  # 限制显示数量