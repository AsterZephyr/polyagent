"""
Advanced Query Processor - 先进的查询处理器
支持查询扩展、实体识别、意图分类和上下文增强
"""

from typing import List, Dict, Any, Optional, Set, Tuple
import re
import asyncio
import json
from dataclasses import dataclass, field
from enum import Enum
import spacy
from spacy.matcher import Matcher, PhraseMatcher
import nltk
from nltk.corpus import wordnet
from nltk.tokenize import word_tokenize
from nltk.tag import pos_tag
from nltk.chunk import ne_chunk
import jieba
import jieba.posseg as pseg

from app.core.logging import LoggerMixin
from app.core.config import settings


class QueryIntent(Enum):
    """查询意图"""
    FACTUAL = "factual"  # 事实性查询
    PROCEDURAL = "procedural"  # 流程性查询
    COMPARATIVE = "comparative"  # 比较性查询
    ANALYTICAL = "analytical"  # 分析性查询
    CREATIVE = "creative"  # 创造性查询
    TECHNICAL = "technical"  # 技术性查询
    UNKNOWN = "unknown"  # 未知意图


@dataclass
class QueryEntity:
    """查询实体"""
    text: str
    entity_type: str
    start: int
    end: int
    confidence: float = 1.0
    synonyms: List[str] = field(default_factory=list)
    related_terms: List[str] = field(default_factory=list)


@dataclass
class ExpandedQuery:
    """扩展查询"""
    original_query: str
    expanded_query: str
    entities: List[QueryEntity]
    intent: QueryIntent
    keywords: List[str]
    synonyms: List[str]
    related_terms: List[str]
    semantic_variations: List[str]
    confidence: float = 1.0


class QueryProcessor(LoggerMixin):
    """查询处理器"""
    
    def __init__(self):
        super().__init__()
        self._initialize_nlp_models()
        self._initialize_keyword_patterns()
        self._initialize_synonym_dict()
    
    def _initialize_nlp_models(self):
        """初始化NLP模型"""
        
        # 加载中文模型
        try:
            self.zh_nlp = spacy.load("zh_core_web_sm")
        except OSError:
            self.logger.warning("Chinese spaCy model not found, using blank model")
            self.zh_nlp = spacy.blank("zh")
        
        # 加载英文模型
        try:
            self.en_nlp = spacy.load("en_core_web_sm")
        except OSError:
            self.logger.warning("English spaCy model not found, using blank model")
            self.en_nlp = spacy.blank("en")
        
        # 初始化jieba
        jieba.initialize()
        
        # 下载NLTK数据
        try:
            nltk.data.find('corpora/wordnet')
        except LookupError:
            nltk.download('wordnet', quiet=True)
        
        try:
            nltk.data.find('tokenizers/punkt')
        except LookupError:
            nltk.download('punkt', quiet=True)
        
        try:
            nltk.data.find('taggers/averaged_perceptron_tagger')
        except LookupError:
            nltk.download('averaged_perceptron_tagger', quiet=True)
        
        try:
            nltk.data.find('chunkers/maxent_ne_chunker')
        except LookupError:
            nltk.download('maxent_ne_chunker', quiet=True)
    
    def _initialize_keyword_patterns(self):
        """初始化关键词模式"""
        
        # 意图识别模式
        self.intent_patterns = {
            QueryIntent.FACTUAL: [
                "什么是", "what is", "define", "definition", "解释",
                "介绍", "describe", "explain"
            ],
            QueryIntent.PROCEDURAL: [
                "如何", "怎么", "how to", "steps", "步骤", "流程",
                "process", "procedure", "方法"
            ],
            QueryIntent.COMPARATIVE: [
                "比较", "对比", "compare", "difference", "区别",
                "vs", "versus", "哪个更好", "which is better"
            ],
            QueryIntent.ANALYTICAL: [
                "分析", "analyze", "why", "为什么", "原因",
                "cause", "reason", "impact", "影响"
            ],
            QueryIntent.TECHNICAL: [
                "api", "sdk", "代码", "code", "算法", "algorithm",
                "implementation", "实现", "技术", "technical"
            ]
        }
        
        # 技术实体模式
        self.tech_patterns = [
            "API", "SDK", "框架", "framework", "库", "library",
            "算法", "algorithm", "数据结构", "data structure",
            "机器学习", "machine learning", "深度学习", "deep learning",
            "人工智能", "artificial intelligence", "AI"
        ]
        
        # 初始化匹配器
        self.zh_matcher = Matcher(self.zh_nlp.vocab)
        self.en_matcher = Matcher(self.en_nlp.vocab)
        self._setup_matchers()
    
    def _setup_matchers(self):
        """设置模式匹配器"""
        
        # 技术术语模式
        tech_patterns_zh = [[{"LOWER": {"IN": ["api", "sdk", "框架", "算法", "机器学习"]}}]]
        tech_patterns_en = [[{"LOWER": {"IN": ["api", "sdk", "framework", "algorithm", "machine"]}, 
                             {"LOWER": "learning", "OP": "?"}}]]
        
        self.zh_matcher.add("TECH_TERM", tech_patterns_zh)
        self.en_matcher.add("TECH_TERM", tech_patterns_en)
    
    def _initialize_synonym_dict(self):
        """初始化同义词词典"""
        
        self.synonym_dict = {
            # 技术术语同义词
            "api": ["接口", "应用程序接口", "application programming interface"],
            "sdk": ["软件开发包", "software development kit", "开发工具包"],
            "framework": ["框架", "软件框架"],
            "algorithm": ["算法", "演算法"],
            "machine learning": ["机器学习", "ML", "机器学习算法"],
            "deep learning": ["深度学习", "DL", "神经网络"],
            "artificial intelligence": ["人工智能", "AI", "智能系统"],
            
            # 中文同义词
            "算法": ["演算法", "algorithm", "计算方法"],
            "框架": ["framework", "软件框架", "开发框架"],
            "机器学习": ["machine learning", "ML", "智能学习"],
            "深度学习": ["deep learning", "DL", "神经网络学习"],
            "人工智能": ["artificial intelligence", "AI", "智能系统"]
        }
    
    async def process_query(self, query: str, context: Optional[Dict[str, Any]] = None) -> ExpandedQuery:
        """处理和扩展查询"""
        
        self.logger.info(f"Processing query: {query}")
        
        # 1. 语言检测和预处理
        language = self._detect_language(query)
        cleaned_query = self._preprocess_query(query)
        
        # 2. 实体识别
        entities = await self._extract_entities(cleaned_query, language)
        
        # 3. 意图识别
        intent = self._classify_intent(cleaned_query)
        
        # 4. 关键词提取
        keywords = await self._extract_keywords(cleaned_query, language)
        
        # 5. 同义词扩展
        synonyms = await self._expand_synonyms(keywords, language)
        
        # 6. 相关词扩展
        related_terms = await self._find_related_terms(keywords, entities, language)
        
        # 7. 语义变化生成
        semantic_variations = await self._generate_semantic_variations(
            cleaned_query, entities, intent, language
        )
        
        # 8. 构建扩展查询
        expanded_query_text = self._build_expanded_query(
            cleaned_query, keywords, synonyms, related_terms
        )
        
        # 9. 计算置信度
        confidence = self._calculate_expansion_confidence(
            entities, keywords, synonyms, intent
        )
        
        expanded_query = ExpandedQuery(
            original_query=query,
            expanded_query=expanded_query_text,
            entities=entities,
            intent=intent,
            keywords=keywords,
            synonyms=synonyms,
            related_terms=related_terms,
            semantic_variations=semantic_variations,
            confidence=confidence
        )
        
        self.logger.info(f"Query processing completed, expanded query: {expanded_query_text}")
        return expanded_query
    
    def _detect_language(self, text: str) -> str:
        """检测文本语言"""
        
        # 简单的语言检测基于字符集
        zh_chars = len(re.findall(r'[\u4e00-\u9fff]', text))
        en_chars = len(re.findall(r'[a-zA-Z]', text))
        
        if zh_chars > en_chars:
            return "zh"
        else:
            return "en"
    
    def _preprocess_query(self, query: str) -> str:
        """预处理查询"""
        
        # 清理特殊字符但保留重要标点
        cleaned = re.sub(r'[^\w\s\u4e00-\u9fff\-\.]', ' ', query)
        
        # 规范化空白字符
        cleaned = re.sub(r'\s+', ' ', cleaned).strip()
        
        return cleaned
    
    async def _extract_entities(self, query: str, language: str) -> List[QueryEntity]:
        """提取查询实体"""
        
        entities = []
        
        if language == "zh":
            entities.extend(await self._extract_chinese_entities(query))
        else:
            entities.extend(await self._extract_english_entities(query))
        
        # 去重和合并
        entities = self._merge_overlapping_entities(entities)
        
        return entities
    
    async def _extract_chinese_entities(self, query: str) -> List[QueryEntity]:
        """提取中文实体"""
        
        entities = []
        
        # 使用jieba进行词性标注
        words = pseg.cut(query)
        
        offset = 0
        for word, flag in words:
            # 根据词性判断实体类型
            entity_type = self._pos_to_entity_type(flag, "zh")
            
            if entity_type:
                entity = QueryEntity(
                    text=word,
                    entity_type=entity_type,
                    start=offset,
                    end=offset + len(word)
                )
                entities.append(entity)
            
            offset += len(word)
        
        # 使用spaCy NER
        if hasattr(self.zh_nlp, 'pipe'):
            doc = self.zh_nlp(query)
            for ent in doc.ents:
                entity = QueryEntity(
                    text=ent.text,
                    entity_type=ent.label_,
                    start=ent.start_char,
                    end=ent.end_char,
                    confidence=0.8
                )
                entities.append(entity)
        
        # 模式匹配
        matches = self.zh_matcher(self.zh_nlp(query))
        for match_id, start, end in matches:
            span = self.zh_nlp(query)[start:end]
            label = self.zh_nlp.vocab.strings[match_id]
            
            entity = QueryEntity(
                text=span.text,
                entity_type=label,
                start=span.start_char,
                end=span.end_char,
                confidence=0.9
            )
            entities.append(entity)
        
        return entities
    
    async def _extract_english_entities(self, query: str) -> List[QueryEntity]:
        """提取英文实体"""
        
        entities = []
        
        # 使用spaCy NER
        doc = self.en_nlp(query)
        for ent in doc.ents:
            entity = QueryEntity(
                text=ent.text,
                entity_type=ent.label_,
                start=ent.start_char,
                end=ent.end_char,
                confidence=0.8
            )
            entities.append(entity)
        
        # 使用NLTK NER
        try:
            tokens = word_tokenize(query)
            pos_tags = pos_tag(tokens)
            named_entities = ne_chunk(pos_tags)
            
            offset = 0
            for chunk in named_entities:
                if hasattr(chunk, 'label'):
                    # 这是一个命名实体
                    text = ' '.join([token for token, pos in chunk])
                    start_pos = query.find(text, offset)
                    
                    if start_pos != -1:
                        entity = QueryEntity(
                            text=text,
                            entity_type=chunk.label(),
                            start=start_pos,
                            end=start_pos + len(text),
                            confidence=0.7
                        )
                        entities.append(entity)
                        offset = start_pos + len(text)
        
        except Exception as e:
            self.logger.warning(f"NLTK NER failed: {e}")
        
        # 模式匹配
        matches = self.en_matcher(doc)
        for match_id, start, end in matches:
            span = doc[start:end]
            label = self.en_nlp.vocab.strings[match_id]
            
            entity = QueryEntity(
                text=span.text,
                entity_type=label,
                start=span.start_char,
                end=span.end_char,
                confidence=0.9
            )
            entities.append(entity)
        
        return entities
    
    def _pos_to_entity_type(self, pos_flag: str, language: str) -> Optional[str]:
        """词性到实体类型的映射"""
        
        if language == "zh":
            pos_mapping = {
                'nr': 'PERSON',      # 人名
                'ns': 'LOCATION',    # 地名
                'nt': 'ORGANIZATION', # 机构名
                'nz': 'OTHER',       # 其他专名
                'eng': 'TECH_TERM',  # 英文词汇（通常是技术术语）
                'x': 'TECH_TERM'     # 未知词（可能是技术术语）
            }
            return pos_mapping.get(pos_flag)
        
        return None
    
    def _merge_overlapping_entities(self, entities: List[QueryEntity]) -> List[QueryEntity]:
        """合并重叠的实体"""
        
        if not entities:
            return entities
        
        # 按起始位置排序
        entities.sort(key=lambda x: x.start)
        
        merged = []
        current = entities[0]
        
        for next_entity in entities[1:]:
            # 检查是否重叠
            if current.end > next_entity.start:
                # 重叠，选择置信度更高或更长的实体
                if (next_entity.confidence > current.confidence or 
                    len(next_entity.text) > len(current.text)):
                    current = next_entity
            else:
                # 不重叠，添加当前实体
                merged.append(current)
                current = next_entity
        
        merged.append(current)
        return merged
    
    def _classify_intent(self, query: str) -> QueryIntent:
        """分类查询意图"""
        
        query_lower = query.lower()
        
        # 计算每种意图的匹配分数
        intent_scores = {}
        
        for intent, patterns in self.intent_patterns.items():
            score = 0
            for pattern in patterns:
                if pattern.lower() in query_lower:
                    score += 1
            intent_scores[intent] = score
        
        # 选择得分最高的意图
        if intent_scores:
            best_intent = max(intent_scores.items(), key=lambda x: x[1])
            if best_intent[1] > 0:
                return best_intent[0]
        
        return QueryIntent.UNKNOWN
    
    async def _extract_keywords(self, query: str, language: str) -> List[str]:
        """提取关键词"""
        
        keywords = []
        
        if language == "zh":
            # 使用jieba提取中文关键词
            words = jieba.cut_for_search(query)
            keywords = [word for word in words if len(word) > 1 and word not in self._get_stopwords("zh")]
        else:
            # 使用spaCy提取英文关键词
            doc = self.en_nlp(query)
            keywords = [
                token.text.lower() for token in doc
                if not token.is_stop and not token.is_punct and len(token.text) > 2
            ]
        
        # 去重并保持顺序
        unique_keywords = []
        for keyword in keywords:
            if keyword not in unique_keywords:
                unique_keywords.append(keyword)
        
        return unique_keywords[:10]  # 限制关键词数量
    
    def _get_stopwords(self, language: str) -> Set[str]:
        """获取停用词列表"""
        
        if language == "zh":
            return {
                "的", "了", "在", "是", "我", "有", "和", "就", "不", "人",
                "都", "一", "一个", "上", "也", "很", "到", "说", "要", "去"
            }
        else:
            return {
                "the", "a", "an", "and", "or", "but", "in", "on", "at",
                "to", "for", "of", "with", "by", "is", "are", "was", "were"
            }
    
    async def _expand_synonyms(self, keywords: List[str], language: str) -> List[str]:
        """扩展同义词"""
        
        synonyms = []
        
        for keyword in keywords:
            # 从自定义词典获取同义词
            if keyword.lower() in self.synonym_dict:
                synonyms.extend(self.synonym_dict[keyword.lower()])
            
            # 使用WordNet获取英文同义词
            if language == "en":
                try:
                    synsets = wordnet.synsets(keyword)
                    for synset in synsets[:2]:  # 限制synset数量
                        for lemma in synset.lemmas()[:3]:  # 限制每个synset的词汇数
                            synonym = lemma.name().replace('_', ' ')
                            if synonym != keyword and synonym not in synonyms:
                                synonyms.append(synonym)
                except Exception as e:
                    self.logger.debug(f"WordNet lookup failed for {keyword}: {e}")
        
        return list(set(synonyms))[:20]  # 去重并限制数量
    
    async def _find_related_terms(
        self,
        keywords: List[str],
        entities: List[QueryEntity],
        language: str
    ) -> List[str]:
        """查找相关词汇"""
        
        related_terms = []
        
        # 基于实体类型添加相关词汇
        for entity in entities:
            if entity.entity_type == "TECH_TERM":
                if language == "zh":
                    related_terms.extend(["技术", "实现", "应用", "开发"])
                else:
                    related_terms.extend(["technology", "implementation", "application", "development"])
        
        # 基于关键词添加领域相关词汇
        for keyword in keywords:
            if keyword.lower() in ["machine", "learning", "机器", "学习"]:
                if language == "zh":
                    related_terms.extend(["算法", "模型", "训练", "预测", "数据"])
                else:
                    related_terms.extend(["algorithm", "model", "training", "prediction", "data"])
            elif keyword.lower() in ["api", "接口"]:
                if language == "zh":
                    related_terms.extend(["接口", "调用", "参数", "返回值"])
                else:
                    related_terms.extend(["interface", "call", "parameter", "response"])
        
        return list(set(related_terms))[:15]  # 去重并限制数量
    
    async def _generate_semantic_variations(
        self,
        query: str,
        entities: List[QueryEntity],
        intent: QueryIntent,
        language: str
    ) -> List[str]:
        """生成语义变化"""
        
        variations = []
        
        # 基于意图生成变化
        if intent == QueryIntent.FACTUAL:
            if language == "zh":
                variations.append(f"解释一下{query}")
                variations.append(f"{query}的定义是什么")
            else:
                variations.append(f"explain {query}")
                variations.append(f"what does {query} mean")
        
        elif intent == QueryIntent.PROCEDURAL:
            if language == "zh":
                variations.append(f"{query}的步骤")
                variations.append(f"怎样{query}")
            else:
                variations.append(f"steps for {query}")
                variations.append(f"how do I {query}")
        
        # 基于实体生成变化
        for entity in entities:
            entity_variations = []
            entity_text = entity.text
            
            if entity.synonyms:
                for synonym in entity.synonyms[:2]:
                    variation = query.replace(entity_text, synonym)
                    entity_variations.append(variation)
            
            variations.extend(entity_variations)
        
        return variations[:10]  # 限制变化数量
    
    def _build_expanded_query(
        self,
        original_query: str,
        keywords: List[str],
        synonyms: List[str],
        related_terms: List[str]
    ) -> str:
        """构建扩展查询"""
        
        # 组合扩展词汇
        expansion_terms = []
        expansion_terms.extend(keywords[:5])  # 限制关键词数量
        expansion_terms.extend(synonyms[:5])  # 限制同义词数量
        expansion_terms.extend(related_terms[:3])  # 限制相关词数量
        
        # 去重
        expansion_terms = list(set(expansion_terms))
        
        # 构建扩展查询
        if expansion_terms:
            expanded = f"{original_query} {' '.join(expansion_terms)}"
        else:
            expanded = original_query
        
        return expanded
    
    def _calculate_expansion_confidence(
        self,
        entities: List[QueryEntity],
        keywords: List[str],
        synonyms: List[str],
        intent: QueryIntent
    ) -> float:
        """计算扩展置信度"""
        
        confidence = 0.5  # 基础置信度
        
        # 实体识别质量
        if entities:
            avg_entity_confidence = sum(e.confidence for e in entities) / len(entities)
            confidence += 0.2 * avg_entity_confidence
        
        # 关键词数量
        if keywords:
            confidence += min(0.1 * len(keywords), 0.2)
        
        # 同义词数量
        if synonyms:
            confidence += min(0.05 * len(synonyms), 0.1)
        
        # 意图识别
        if intent != QueryIntent.UNKNOWN:
            confidence += 0.1
        
        return min(confidence, 1.0)