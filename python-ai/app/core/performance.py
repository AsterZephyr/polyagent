"""
Performance Optimization Module - 性能优化模块
提供缓存、连接池、异步处理等性能优化功能
"""

import asyncio
import time
import pickle
import hashlib
from typing import Any, Dict, List, Optional, Callable, Union, TypeVar, Generic
from collections import defaultdict, OrderedDict
from dataclasses import dataclass, field
from datetime import datetime, timedelta
import threading
import weakref
from functools import wraps, partial
import logging

logger = logging.getLogger(__name__)

T = TypeVar('T')

@dataclass
class CacheEntry(Generic[T]):
    """缓存条目"""
    value: T
    created_at: datetime
    last_accessed: datetime
    access_count: int = 0
    ttl: Optional[int] = None  # 秒
    
    def is_expired(self) -> bool:
        """检查是否过期"""
        if self.ttl is None:
            return False
        return (datetime.now() - self.created_at).total_seconds() > self.ttl
    
    def touch(self):
        """更新访问时间"""
        self.last_accessed = datetime.now()
        self.access_count += 1


class LRUCache(Generic[T]):
    """LRU缓存实现"""
    
    def __init__(self, max_size: int = 1000, default_ttl: Optional[int] = None):
        self.max_size = max_size
        self.default_ttl = default_ttl
        self._cache: OrderedDict[str, CacheEntry[T]] = OrderedDict()
        self._lock = threading.RLock()
        self._stats = {
            'hits': 0,
            'misses': 0,
            'evictions': 0,
            'expired': 0
        }
    
    def get(self, key: str) -> Optional[T]:
        """获取缓存项"""
        with self._lock:
            if key not in self._cache:
                self._stats['misses'] += 1
                return None
            
            entry = self._cache[key]
            
            # 检查是否过期
            if entry.is_expired():
                del self._cache[key]
                self._stats['expired'] += 1
                self._stats['misses'] += 1
                return None
            
            # 更新访问信息并移到末尾（最近使用）
            entry.touch()
            self._cache.move_to_end(key)
            self._stats['hits'] += 1
            
            return entry.value
    
    def set(self, key: str, value: T, ttl: Optional[int] = None) -> None:
        """设置缓存项"""
        with self._lock:
            now = datetime.now()
            
            if key in self._cache:
                # 更新现有项
                entry = self._cache[key]
                entry.value = value
                entry.created_at = now
                entry.last_accessed = now
                entry.ttl = ttl or self.default_ttl
                self._cache.move_to_end(key)
            else:
                # 添加新项
                if len(self._cache) >= self.max_size:
                    # 删除最久未使用的项
                    oldest_key = next(iter(self._cache))
                    del self._cache[oldest_key]
                    self._stats['evictions'] += 1
                
                entry = CacheEntry(
                    value=value,
                    created_at=now,
                    last_accessed=now,
                    ttl=ttl or self.default_ttl
                )
                self._cache[key] = entry
    
    def delete(self, key: str) -> bool:
        """删除缓存项"""
        with self._lock:
            if key in self._cache:
                del self._cache[key]
                return True
            return False
    
    def clear(self) -> None:
        """清空缓存"""
        with self._lock:
            self._cache.clear()
            self._stats = {k: 0 for k in self._stats}
    
    def cleanup_expired(self) -> int:
        """清理过期项"""
        with self._lock:
            expired_keys = [
                key for key, entry in self._cache.items()
                if entry.is_expired()
            ]
            
            for key in expired_keys:
                del self._cache[key]
            
            self._stats['expired'] += len(expired_keys)
            return len(expired_keys)
    
    def get_stats(self) -> Dict[str, Any]:
        """获取缓存统计"""
        with self._lock:
            total_requests = self._stats['hits'] + self._stats['misses']
            hit_rate = self._stats['hits'] / total_requests if total_requests > 0 else 0
            
            return {
                'size': len(self._cache),
                'max_size': self.max_size,
                'hit_rate': hit_rate,
                **self._stats
            }


class AsyncCache:
    """异步缓存装饰器"""
    
    def __init__(
        self,
        max_size: int = 1000,
        ttl: Optional[int] = None,
        key_func: Optional[Callable[..., str]] = None
    ):
        self.cache = LRUCache(max_size, ttl)
        self.key_func = key_func or self._default_key_func
        self._pending: Dict[str, asyncio.Future] = {}
        self._lock = asyncio.Lock()
    
    def _default_key_func(self, *args, **kwargs) -> str:
        """默认键生成函数"""
        key_data = str(args) + str(sorted(kwargs.items()))
        return hashlib.md5(key_data.encode()).hexdigest()
    
    def __call__(self, func: Callable) -> Callable:
        """装饰器实现"""
        @wraps(func)
        async def wrapper(*args, **kwargs):
            # 生成缓存键
            cache_key = self.key_func(*args, **kwargs)
            
            # 检查缓存
            cached_result = self.cache.get(cache_key)
            if cached_result is not None:
                return cached_result
            
            # 防止重复计算
            async with self._lock:
                if cache_key in self._pending:
                    return await self._pending[cache_key]
                
                # 创建Future并开始计算
                future = asyncio.create_task(func(*args, **kwargs))
                self._pending[cache_key] = future
            
            try:
                # 等待计算完成
                result = await future
                
                # 存储结果到缓存
                self.cache.set(cache_key, result)
                
                return result
            finally:
                # 清理pending字典
                async with self._lock:
                    self._pending.pop(cache_key, None)
        
        return wrapper


class ConnectionPool:
    """连接池管理器"""
    
    def __init__(
        self,
        create_connection: Callable,
        max_connections: int = 10,
        min_connections: int = 2,
        max_idle_time: int = 300,  # 5分钟
        connection_timeout: int = 30
    ):
        self.create_connection = create_connection
        self.max_connections = max_connections
        self.min_connections = min_connections
        self.max_idle_time = max_idle_time
        self.connection_timeout = connection_timeout
        
        self._available_connections: List[Any] = []
        self._busy_connections: weakref.WeakSet = weakref.WeakSet()
        self._connection_created_times: Dict[int, datetime] = {}
        
        self._lock = asyncio.Lock()
        self._condition = asyncio.Condition(self._lock)
        
        # 初始化最小连接数
        asyncio.create_task(self._initialize_connections())
    
    async def _initialize_connections(self):
        """初始化最小连接数"""
        for _ in range(self.min_connections):
            try:
                conn = await self.create_connection()
                self._available_connections.append(conn)
                self._connection_created_times[id(conn)] = datetime.now()
            except Exception as e:
                logger.warning(f"Failed to create initial connection: {e}")
    
    async def acquire(self) -> Any:
        """获取连接"""
        async with self._condition:
            # 等待可用连接或创建新连接
            while True:
                # 检查是否有可用连接
                if self._available_connections:
                    conn = self._available_connections.pop(0)
                    
                    # 检查连接是否过期
                    if self._is_connection_expired(conn):
                        await self._close_connection(conn)
                        continue
                    
                    self._busy_connections.add(conn)
                    return conn
                
                # 检查是否可以创建新连接
                total_connections = len(self._available_connections) + len(self._busy_connections)
                if total_connections < self.max_connections:
                    try:
                        conn = await asyncio.wait_for(
                            self.create_connection(),
                            timeout=self.connection_timeout
                        )
                        self._connection_created_times[id(conn)] = datetime.now()
                        self._busy_connections.add(conn)
                        return conn
                    except Exception as e:
                        logger.error(f"Failed to create connection: {e}")
                        raise
                
                # 等待连接释放
                await self._condition.wait()
    
    async def release(self, connection: Any):
        """释放连接"""
        async with self._condition:
            if connection in self._busy_connections:
                self._busy_connections.discard(connection)
                
                # 检查连接是否仍然有效
                if not self._is_connection_expired(connection):
                    self._available_connections.append(connection)
                else:
                    await self._close_connection(connection)
                
                # 通知等待的协程
                self._condition.notify()
    
    def _is_connection_expired(self, connection: Any) -> bool:
        """检查连接是否过期"""
        conn_id = id(connection)
        if conn_id not in self._connection_created_times:
            return True
        
        created_time = self._connection_created_times[conn_id]
        age = (datetime.now() - created_time).total_seconds()
        
        return age > self.max_idle_time
    
    async def _close_connection(self, connection: Any):
        """关闭连接"""
        try:
            if hasattr(connection, 'close'):
                if asyncio.iscoroutinefunction(connection.close):
                    await connection.close()
                else:
                    connection.close()
        except Exception as e:
            logger.warning(f"Error closing connection: {e}")
        finally:
            conn_id = id(connection)
            self._connection_created_times.pop(conn_id, None)
    
    async def cleanup(self):
        """清理过期连接"""
        async with self._lock:
            expired_connections = [
                conn for conn in self._available_connections
                if self._is_connection_expired(conn)
            ]
            
            for conn in expired_connections:
                self._available_connections.remove(conn)
                await self._close_connection(conn)
            
            return len(expired_connections)
    
    async def close_all(self):
        """关闭所有连接"""
        async with self._lock:
            # 关闭可用连接
            for conn in self._available_connections:
                await self._close_connection(conn)
            self._available_connections.clear()
            
            # 关闭忙碌连接
            for conn in list(self._busy_connections):
                await self._close_connection(conn)
            self._busy_connections.clear()
    
    def get_stats(self) -> Dict[str, Any]:
        """获取连接池统计"""
        return {
            'available_connections': len(self._available_connections),
            'busy_connections': len(self._busy_connections),
            'total_connections': len(self._available_connections) + len(self._busy_connections),
            'max_connections': self.max_connections,
            'min_connections': self.min_connections
        }


class BatchProcessor:
    """批处理器"""
    
    def __init__(
        self,
        process_func: Callable[[List[Any]], List[Any]],
        batch_size: int = 10,
        max_wait_time: float = 1.0,
        max_queue_size: int = 1000
    ):
        self.process_func = process_func
        self.batch_size = batch_size
        self.max_wait_time = max_wait_time
        self.max_queue_size = max_queue_size
        
        self._queue: List[tuple] = []
        self._futures: List[asyncio.Future] = []
        self._lock = asyncio.Lock()
        self._processing = False
        
        # 启动批处理任务
        self._batch_task = asyncio.create_task(self._batch_processor())
    
    async def add_item(self, item: Any) -> Any:
        """添加项目到批处理队列"""
        if len(self._queue) >= self.max_queue_size:
            raise RuntimeError("Batch queue is full")
        
        future = asyncio.Future()
        
        async with self._lock:
            self._queue.append((item, future))
            self._futures.append(future)
            
            # 如果队列满了，立即处理
            if len(self._queue) >= self.batch_size:
                await self._process_batch()
        
        return await future
    
    async def _batch_processor(self):
        """批处理主循环"""
        while True:
            try:
                await asyncio.sleep(self.max_wait_time)
                
                async with self._lock:
                    if self._queue and not self._processing:
                        await self._process_batch()
                        
            except asyncio.CancelledError:
                break
            except Exception as e:
                logger.error(f"Batch processor error: {e}")
    
    async def _process_batch(self):
        """处理当前批次"""
        if not self._queue or self._processing:
            return
        
        self._processing = True
        
        try:
            # 提取当前批次
            current_batch = self._queue[:self.batch_size]
            self._queue = self._queue[self.batch_size:]
            
            items = [item for item, _ in current_batch]
            futures = [future for _, future in current_batch]
            
            # 处理批次
            try:
                if asyncio.iscoroutinefunction(self.process_func):
                    results = await self.process_func(items)
                else:
                    results = self.process_func(items)
                
                # 设置结果
                for future, result in zip(futures, results):
                    if not future.cancelled():
                        future.set_result(result)
                        
            except Exception as e:
                # 所有Future都设置为异常
                for future in futures:
                    if not future.cancelled():
                        future.set_exception(e)
                        
        finally:
            self._processing = False
    
    async def close(self):
        """关闭批处理器"""
        self._batch_task.cancel()
        
        # 处理剩余项目
        async with self._lock:
            if self._queue:
                await self._process_batch()


class PerformanceMonitor:
    """性能监控器"""
    
    def __init__(self):
        self.metrics: Dict[str, List[float]] = defaultdict(list)
        self.counters: Dict[str, int] = defaultdict(int)
        self.timers: Dict[str, datetime] = {}
        self._lock = threading.Lock()
    
    def record_time(self, metric_name: str, duration: float):
        """记录时间指标"""
        with self._lock:
            self.metrics[metric_name].append(duration)
            # 保持最近1000个记录
            if len(self.metrics[metric_name]) > 1000:
                self.metrics[metric_name] = self.metrics[metric_name][-1000:]
    
    def increment_counter(self, counter_name: str, value: int = 1):
        """增加计数器"""
        with self._lock:
            self.counters[counter_name] += value
    
    def start_timer(self, timer_name: str):
        """启动计时器"""
        self.timers[timer_name] = datetime.now()
    
    def stop_timer(self, timer_name: str) -> float:
        """停止计时器并返回持续时间"""
        if timer_name not in self.timers:
            return 0.0
        
        duration = (datetime.now() - self.timers[timer_name]).total_seconds()
        del self.timers[timer_name]
        
        self.record_time(timer_name, duration)
        return duration
    
    def get_stats(self) -> Dict[str, Any]:
        """获取性能统计"""
        with self._lock:
            stats = {
                'counters': dict(self.counters),
                'metrics': {}
            }
            
            for metric_name, values in self.metrics.items():
                if values:
                    stats['metrics'][metric_name] = {
                        'count': len(values),
                        'avg': sum(values) / len(values),
                        'min': min(values),
                        'max': max(values),
                        'recent_avg': sum(values[-10:]) / min(len(values), 10)
                    }
            
            return stats


# 全局性能监控器实例
performance_monitor = PerformanceMonitor()


def monitor_performance(metric_name: str):
    """性能监控装饰器"""
    def decorator(func: Callable) -> Callable:
        @wraps(func)
        async def async_wrapper(*args, **kwargs):
            performance_monitor.start_timer(metric_name)
            try:
                result = await func(*args, **kwargs)
                return result
            finally:
                performance_monitor.stop_timer(metric_name)
        
        @wraps(func)
        def sync_wrapper(*args, **kwargs):
            performance_monitor.start_timer(metric_name)
            try:
                result = func(*args, **kwargs)
                return result
            finally:
                performance_monitor.stop_timer(metric_name)
        
        return async_wrapper if asyncio.iscoroutinefunction(func) else sync_wrapper
    
    return decorator


# 全局缓存实例
memory_cache = LRUCache(max_size=10000, default_ttl=3600)  # 1小时TTL


# 便捷的缓存装饰器
def cache_result(ttl: Optional[int] = None, key_func: Optional[Callable] = None):
    """结果缓存装饰器"""
    return AsyncCache(ttl=ttl, key_func=key_func)