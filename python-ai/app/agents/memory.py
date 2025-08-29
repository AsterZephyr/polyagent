"""
Advanced Memory Management System - 先进记忆管理系统
实现短期记忆、长期记忆、语义记忆和程序性记忆
"""

from typing import List, Dict, Any, Optional, Union, Tuple, Set
import asyncio
import json
import time
import hashlib
from dataclasses import dataclass, field
from datetime import datetime, timedelta
from enum import Enum
from collections import defaultdict
import numpy as np

from app.core.logging import LoggerMixin
from app.core.exceptions import AgentException


class MemoryType(Enum):
    """记忆类型"""
    WORKING = "working"        # 工作记忆（短期）
    EPISODIC = "episodic"      # 情节记忆（对话历史）
    SEMANTIC = "semantic"      # 语义记忆（知识库）
    PROCEDURAL = "procedural"  # 程序性记忆（技能和模式）


class MemoryImportance(Enum):
    """记忆重要性"""
    LOW = 1
    MEDIUM = 2
    HIGH = 3
    CRITICAL = 4


@dataclass
class MemoryItem:
    """记忆项"""
    memory_id: str
    content: str
    memory_type: MemoryType
    importance: MemoryImportance = MemoryImportance.MEDIUM
    created_at: datetime = field(default_factory=datetime.now)
    last_accessed: datetime = field(default_factory=datetime.now)
    access_count: int = 0
    decay_rate: float = 0.1
    tags: Set[str] = field(default_factory=set)
    metadata: Dict[str, Any] = field(default_factory=dict)
    embedding: Optional[List[float]] = None
    
    def __post_init__(self):
        """初始化后处理"""
        if not self.memory_id:
            self.memory_id = self._generate_memory_id()
    
    def _generate_memory_id(self) -> str:
        """生成记忆ID"""
        content_hash = hashlib.md5(self.content.encode()).hexdigest()[:8]
        timestamp = int(time.time())
        return f"mem_{self.memory_type.value}_{timestamp}_{content_hash}"
    
    @property
    def current_strength(self) -> float:
        """当前记忆强度"""
        time_since_created = (datetime.now() - self.created_at).total_seconds()
        time_since_accessed = (datetime.now() - self.last_accessed).total_seconds()
        
        # 基础强度基于重要性
        base_strength = self.importance.value / 4.0
        
        # 访问频率加成
        frequency_bonus = min(self.access_count * 0.1, 0.5)
        
        # 时间衰减
        decay_factor = np.exp(-self.decay_rate * time_since_accessed / 3600)  # 小时为单位
        
        # 最近访问奖励
        recency_bonus = max(0, 1 - time_since_accessed / (24 * 3600))  # 24小时内的奖励
        
        return min((base_strength + frequency_bonus) * decay_factor + recency_bonus * 0.2, 1.0)
    
    def access(self):
        """访问记忆"""
        self.last_accessed = datetime.now()
        self.access_count += 1


@dataclass
class ConversationMemory:
    """对话记忆"""
    conversation_id: str
    user_id: str
    messages: List[Dict[str, str]] = field(default_factory=list)
    summary: Optional[str] = None
    key_facts: List[str] = field(default_factory=list)
    emotional_tone: Optional[str] = None
    topics: Set[str] = field(default_factory=set)
    created_at: datetime = field(default_factory=datetime.now)
    updated_at: datetime = field(default_factory=datetime.now)


@dataclass
class UserProfile:
    """用户画像"""
    user_id: str
    preferences: Dict[str, Any] = field(default_factory=dict)
    interests: Set[str] = field(default_factory=set)
    communication_style: Optional[str] = None
    expertise_areas: Set[str] = field(default_factory=set)
    interaction_patterns: Dict[str, Any] = field(default_factory=dict)
    created_at: datetime = field(default_factory=datetime.now)
    updated_at: datetime = field(default_factory=datetime.now)


class MemoryConsolidation:
    """记忆整合"""
    
    @staticmethod
    def should_consolidate(memories: List[MemoryItem]) -> bool:
        """判断是否需要整合记忆"""
        if len(memories) < 3:
            return False
        
        # 检查相似度
        similar_count = 0
        for i, mem1 in enumerate(memories):
            for mem2 in memories[i+1:]:
                if MemoryConsolidation._calculate_similarity(mem1, mem2) > 0.7:
                    similar_count += 1
        
        return similar_count >= 2
    
    @staticmethod
    def _calculate_similarity(mem1: MemoryItem, mem2: MemoryItem) -> float:
        """计算记忆相似度"""
        # 简化的相似度计算
        content_similarity = len(set(mem1.content.lower().split()) & 
                                set(mem2.content.lower().split())) / len(set(mem1.content.lower().split()) | 
                                                                          set(mem2.content.lower().split()))
        
        tag_similarity = len(mem1.tags & mem2.tags) / max(len(mem1.tags | mem2.tags), 1)
        
        return (content_similarity + tag_similarity) / 2
    
    @staticmethod
    def consolidate_memories(memories: List[MemoryItem]) -> MemoryItem:
        """整合记忆"""
        if not memories:
            raise ValueError("Cannot consolidate empty memory list")
        
        # 合并内容
        consolidated_content = "Consolidated memory:\n"
        for i, memory in enumerate(memories):
            consolidated_content += f"{i+1}. {memory.content}\n"
        
        # 合并标签
        consolidated_tags = set()
        for memory in memories:
            consolidated_tags.update(memory.tags)
        
        # 取最高重要性
        max_importance = max(memory.importance for memory in memories)
        
        # 创建整合后的记忆
        consolidated_memory = MemoryItem(
            memory_id="",
            content=consolidated_content.strip(),
            memory_type=memories[0].memory_type,
            importance=max_importance,
            tags=consolidated_tags,
            metadata={
                "consolidated_from": [mem.memory_id for mem in memories],
                "consolidation_time": datetime.now().isoformat()
            }
        )
        
        return consolidated_memory


class AdvancedMemorySystem(LoggerMixin):
    """先进记忆管理系统"""
    
    def __init__(
        self,
        max_working_memory: int = 20,
        max_episodic_memory: int = 1000,
        max_semantic_memory: int = 5000,
        consolidation_threshold: int = 50
    ):
        super().__init__()
        
        # 记忆配置
        self.max_working_memory = max_working_memory
        self.max_episodic_memory = max_episodic_memory
        self.max_semantic_memory = max_semantic_memory
        self.consolidation_threshold = consolidation_threshold
        
        # 记忆存储
        self.memories: Dict[str, MemoryItem] = {}
        self.memory_index: Dict[MemoryType, List[str]] = defaultdict(list)
        
        # 对话记忆
        self.conversations: Dict[str, ConversationMemory] = {}
        
        # 用户画像
        self.user_profiles: Dict[str, UserProfile] = {}
        
        # 记忆网络 - 记忆之间的关联关系
        self.memory_associations: Dict[str, Set[str]] = defaultdict(set)
        
        # 访问统计
        self.memory_stats: Dict[str, Dict[str, Any]] = defaultdict(dict)
    
    async def store_memory(
        self,
        content: str,
        memory_type: MemoryType,
        importance: MemoryImportance = MemoryImportance.MEDIUM,
        tags: Optional[Set[str]] = None,
        metadata: Optional[Dict[str, Any]] = None,
        user_id: Optional[str] = None
    ) -> str:
        """存储记忆"""
        
        # 创建记忆项
        memory_item = MemoryItem(
            memory_id="",
            content=content,
            memory_type=memory_type,
            importance=importance,
            tags=tags or set(),
            metadata=metadata or {}
        )
        
        if user_id:
            memory_item.metadata["user_id"] = user_id
        
        # 存储记忆
        self.memories[memory_item.memory_id] = memory_item
        self.memory_index[memory_type].append(memory_item.memory_id)
        
        # 记录统计
        self.memory_stats[memory_item.memory_id] = {
            "created_at": memory_item.created_at,
            "type": memory_type.value,
            "importance": importance.value
        }
        
        # 检查是否需要清理内存
        await self._manage_memory_capacity(memory_type)
        
        # 检查是否需要记忆整合
        if len(self.memory_index[memory_type]) % self.consolidation_threshold == 0:
            await self._consolidate_memories(memory_type)
        
        self.logger.debug(f"Stored {memory_type.value} memory: {memory_item.memory_id}")
        return memory_item.memory_id
    
    async def retrieve_memories(
        self,
        query: Optional[str] = None,
        memory_type: Optional[MemoryType] = None,
        tags: Optional[Set[str]] = None,
        user_id: Optional[str] = None,
        limit: int = 10,
        min_importance: Optional[MemoryImportance] = None,
        time_range: Optional[Tuple[datetime, datetime]] = None
    ) -> List[MemoryItem]:
        """检索记忆"""
        
        candidate_memories = []
        
        # 筛选候选记忆
        memory_ids = []
        if memory_type:
            memory_ids = self.memory_index[memory_type]
        else:
            for type_memories in self.memory_index.values():
                memory_ids.extend(type_memories)
        
        for memory_id in memory_ids:
            memory = self.memories.get(memory_id)
            if not memory:
                continue
            
            # 用户过滤
            if user_id and memory.metadata.get("user_id") != user_id:
                continue
            
            # 重要性过滤
            if min_importance and memory.importance.value < min_importance.value:
                continue
            
            # 时间范围过滤
            if time_range:
                start_time, end_time = time_range
                if not (start_time <= memory.created_at <= end_time):
                    continue
            
            # 标签过滤
            if tags and not tags.intersection(memory.tags):
                continue
            
            candidate_memories.append(memory)
        
        # 计算相关性分数
        if query:
            scored_memories = []
            for memory in candidate_memories:
                relevance_score = self._calculate_relevance(query, memory)
                scored_memories.append((memory, relevance_score))
            
            # 按相关性和记忆强度排序
            scored_memories.sort(
                key=lambda x: (x[1], x[0].current_strength, x[0].importance.value),
                reverse=True
            )
            
            result_memories = [mem for mem, score in scored_memories[:limit]]
        else:
            # 按记忆强度和重要性排序
            candidate_memories.sort(
                key=lambda x: (x.current_strength, x.importance.value, x.last_accessed),
                reverse=True
            )
            result_memories = candidate_memories[:limit]
        
        # 更新访问记录
        for memory in result_memories:
            memory.access()
        
        return result_memories
    
    def _calculate_relevance(self, query: str, memory: MemoryItem) -> float:
        """计算查询与记忆的相关性"""
        query_words = set(query.lower().split())
        memory_words = set(memory.content.lower().split())
        
        # 词汇重叠度
        overlap = len(query_words & memory_words)
        total_words = len(query_words | memory_words)
        
        if total_words == 0:
            return 0.0
        
        lexical_similarity = overlap / total_words
        
        # 标签匹配奖励
        tag_bonus = 0.0
        if memory.tags:
            query_tags = set(query.lower().split())
            tag_matches = len(query_tags & {tag.lower() for tag in memory.tags})
            tag_bonus = tag_matches * 0.2
        
        return min(lexical_similarity + tag_bonus, 1.0)
    
    async def _manage_memory_capacity(self, memory_type: MemoryType):
        """管理记忆容量"""
        
        max_capacity = {
            MemoryType.WORKING: self.max_working_memory,
            MemoryType.EPISODIC: self.max_episodic_memory,
            MemoryType.SEMANTIC: self.max_semantic_memory,
            MemoryType.PROCEDURAL: self.max_semantic_memory  # 共用语义记忆容量
        }.get(memory_type, 1000)
        
        memory_ids = self.memory_index[memory_type]
        
        if len(memory_ids) <= max_capacity:
            return
        
        # 获取所有记忆并按强度排序
        memories = [self.memories[mid] for mid in memory_ids if mid in self.memories]
        memories.sort(key=lambda x: (x.current_strength, x.importance.value))
        
        # 删除最弱的记忆
        to_remove = len(memories) - max_capacity
        for i in range(to_remove):
            memory_to_remove = memories[i]
            self._remove_memory(memory_to_remove.memory_id)
            
            self.logger.debug(f"Removed weak memory: {memory_to_remove.memory_id}")
    
    def _remove_memory(self, memory_id: str):
        """删除记忆"""
        if memory_id in self.memories:
            memory = self.memories[memory_id]
            
            # 从内存中删除
            del self.memories[memory_id]
            
            # 从索引中删除
            if memory_id in self.memory_index[memory.memory_type]:
                self.memory_index[memory.memory_type].remove(memory_id)
            
            # 删除关联关系
            if memory_id in self.memory_associations:
                del self.memory_associations[memory_id]
            
            # 删除统计信息
            if memory_id in self.memory_stats:
                del self.memory_stats[memory_id]
    
    async def _consolidate_memories(self, memory_type: MemoryType):
        """整合记忆"""
        
        memory_ids = self.memory_index[memory_type]
        memories = [self.memories[mid] for mid in memory_ids[-self.consolidation_threshold:] 
                   if mid in self.memories]
        
        if MemoryConsolidation.should_consolidate(memories):
            # 找到相似的记忆组
            similar_groups = self._find_similar_memory_groups(memories)
            
            for group in similar_groups:
                if len(group) >= 3:  # 至少3个相似记忆才整合
                    consolidated = MemoryConsolidation.consolidate_memories(group)
                    
                    # 删除原始记忆
                    for memory in group:
                        self._remove_memory(memory.memory_id)
                    
                    # 存储整合后的记忆
                    self.memories[consolidated.memory_id] = consolidated
                    self.memory_index[memory_type].append(consolidated.memory_id)
                    
                    self.logger.info(f"Consolidated {len(group)} memories into {consolidated.memory_id}")
    
    def _find_similar_memory_groups(self, memories: List[MemoryItem]) -> List[List[MemoryItem]]:
        """找到相似的记忆组"""
        
        groups = []
        used = set()
        
        for i, memory1 in enumerate(memories):
            if memory1.memory_id in used:
                continue
            
            group = [memory1]
            used.add(memory1.memory_id)
            
            for j, memory2 in enumerate(memories[i+1:], i+1):
                if memory2.memory_id in used:
                    continue
                
                if MemoryConsolidation._calculate_similarity(memory1, memory2) > 0.7:
                    group.append(memory2)
                    used.add(memory2.memory_id)
            
            if len(group) >= 2:
                groups.append(group)
        
        return groups
    
    async def add_conversation_memory(
        self,
        conversation_id: str,
        user_id: str,
        message: Dict[str, str],
        auto_summarize: bool = True
    ):
        """添加对话记忆"""
        
        if conversation_id not in self.conversations:
            self.conversations[conversation_id] = ConversationMemory(
                conversation_id=conversation_id,
                user_id=user_id
            )
        
        conv_memory = self.conversations[conversation_id]
        conv_memory.messages.append(message)
        conv_memory.updated_at = datetime.now()
        
        # 提取话题
        if message.get("content"):
            topics = self._extract_topics(message["content"])
            conv_memory.topics.update(topics)
        
        # 自动总结长对话
        if auto_summarize and len(conv_memory.messages) % 20 == 0:
            await self._summarize_conversation(conversation_id)
        
        # 存储为情节记忆
        episode_content = f"Conversation {conversation_id}: {message.get('role', 'user')}: {message.get('content', '')}"
        await self.store_memory(
            content=episode_content,
            memory_type=MemoryType.EPISODIC,
            importance=MemoryImportance.LOW,
            tags=conv_memory.topics,
            metadata={"conversation_id": conversation_id, "user_id": user_id}
        )
    
    def _extract_topics(self, content: str) -> Set[str]:
        """提取话题关键词"""
        # 简化的话题提取
        words = content.lower().split()
        
        # 过滤停用词和短词
        stop_words = {'的', '了', '在', '是', '我', '你', '他', '她', '它', '我们', '你们', '他们',
                     'the', 'a', 'an', 'and', 'or', 'but', 'in', 'on', 'at', 'to', 'for', 'of', 'with', 'by'}
        
        topics = {word for word in words 
                 if len(word) > 2 and word not in stop_words and word.isalpha()}
        
        return topics
    
    async def _summarize_conversation(self, conversation_id: str):
        """总结对话"""
        
        conv_memory = self.conversations.get(conversation_id)
        if not conv_memory:
            return
        
        # 简化的对话总结
        recent_messages = conv_memory.messages[-10:]  # 最近10条消息
        
        summary_content = []
        for msg in recent_messages:
            role = msg.get("role", "unknown")
            content = msg.get("content", "")[:100]  # 截取前100字符
            summary_content.append(f"{role}: {content}")
        
        conv_memory.summary = "\n".join(summary_content)
        
        # 提取关键事实
        key_facts = []
        for msg in recent_messages:
            content = msg.get("content", "")
            if any(keyword in content.lower() for keyword in ['记住', '重要', '关键', 'remember', 'important', 'key']):
                key_facts.append(content[:200])
        
        conv_memory.key_facts.extend(key_facts)
    
    async def update_user_profile(
        self,
        user_id: str,
        preferences: Optional[Dict[str, Any]] = None,
        interests: Optional[Set[str]] = None,
        communication_style: Optional[str] = None
    ):
        """更新用户画像"""
        
        if user_id not in self.user_profiles:
            self.user_profiles[user_id] = UserProfile(user_id=user_id)
        
        profile = self.user_profiles[user_id]
        
        if preferences:
            profile.preferences.update(preferences)
        
        if interests:
            profile.interests.update(interests)
        
        if communication_style:
            profile.communication_style = communication_style
        
        profile.updated_at = datetime.now()
        
        # 存储为语义记忆
        profile_content = f"User {user_id} profile: preferences={profile.preferences}, interests={profile.interests}"
        await self.store_memory(
            content=profile_content,
            memory_type=MemoryType.SEMANTIC,
            importance=MemoryImportance.HIGH,
            tags={"user_profile", user_id},
            metadata={"user_id": user_id, "profile_update": True}
        )
    
    async def get_context_for_user(
        self,
        user_id: str,
        query: Optional[str] = None,
        limit: int = 10
    ) -> Dict[str, Any]:
        """获取用户上下文"""
        
        context = {
            "user_profile": self.user_profiles.get(user_id),
            "recent_memories": [],
            "relevant_memories": [],
            "conversation_history": []
        }
        
        # 获取最近的记忆
        recent_memories = await self.retrieve_memories(
            user_id=user_id,
            limit=limit//2,
            time_range=(datetime.now() - timedelta(days=7), datetime.now())
        )
        context["recent_memories"] = recent_memories
        
        # 获取相关记忆
        if query:
            relevant_memories = await self.retrieve_memories(
                query=query,
                user_id=user_id,
                limit=limit//2
            )
            context["relevant_memories"] = relevant_memories
        
        # 获取对话历史
        user_conversations = [conv for conv in self.conversations.values() 
                            if conv.user_id == user_id]
        user_conversations.sort(key=lambda x: x.updated_at, reverse=True)
        context["conversation_history"] = user_conversations[:3]  # 最近3个对话
        
        return context
    
    def get_memory_statistics(self) -> Dict[str, Any]:
        """获取记忆统计信息"""
        
        stats = {
            "total_memories": len(self.memories),
            "by_type": {},
            "by_importance": defaultdict(int),
            "average_strength": 0.0,
            "total_conversations": len(self.conversations),
            "total_users": len(self.user_profiles)
        }
        
        # 按类型统计
        for memory_type in MemoryType:
            stats["by_type"][memory_type.value] = len(self.memory_index[memory_type])
        
        # 按重要性统计
        total_strength = 0.0
        for memory in self.memories.values():
            stats["by_importance"][memory.importance.value] += 1
            total_strength += memory.current_strength
        
        if self.memories:
            stats["average_strength"] = total_strength / len(self.memories)
        
        return stats
    
    async def cleanup_expired_memories(self, max_age_days: int = 90):
        """清理过期记忆"""
        
        cutoff_time = datetime.now() - timedelta(days=max_age_days)
        expired_memories = []
        
        for memory_id, memory in self.memories.items():
            # 跳过重要记忆
            if memory.importance == MemoryImportance.CRITICAL:
                continue
            
            # 检查记忆强度和年龄
            if (memory.created_at < cutoff_time and 
                memory.current_strength < 0.1 and
                memory.access_count < 2):
                expired_memories.append(memory_id)
        
        # 删除过期记忆
        for memory_id in expired_memories:
            self._remove_memory(memory_id)
        
        self.logger.info(f"Cleaned up {len(expired_memories)} expired memories")
        return len(expired_memories)