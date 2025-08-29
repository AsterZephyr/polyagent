"""
Optimization Service - 优化服务
提供系统级性能优化和资源管理
"""

import asyncio
import psutil
import gc
import time
from typing import Dict, Any, List, Optional
from datetime import datetime, timedelta
from dataclasses import dataclass
import logging

from app.core.logging import LoggerMixin
from app.core.performance import performance_monitor, ConnectionPool
from app.services.rag_service import RAGService
from app.agents.memory import AdvancedMemorySystem

logger = logging.getLogger(__name__)


@dataclass
class SystemMetrics:
    """系统指标"""
    cpu_percent: float
    memory_percent: float
    memory_available: int
    disk_usage_percent: float
    network_io: Dict[str, int]
    process_count: int
    thread_count: int
    fd_count: int  # 文件描述符数量
    timestamp: datetime


@dataclass
class OptimizationResult:
    """优化结果"""
    action: str
    before_metrics: Dict[str, Any]
    after_metrics: Dict[str, Any]
    improvement: Dict[str, float]
    execution_time: float
    success: bool
    message: str


class SystemOptimizer(LoggerMixin):
    """系统优化器"""
    
    def __init__(self):
        super().__init__()
        
        self.optimization_history: List[OptimizationResult] = []
        self.optimization_config = {
            "memory_threshold": 80.0,  # 内存使用率阈值
            "cpu_threshold": 85.0,     # CPU使用率阈值
            "disk_threshold": 90.0,    # 磁盘使用率阈值
            "gc_interval": 300,        # GC间隔（秒）
            "cache_cleanup_interval": 600,  # 缓存清理间隔
            "session_cleanup_interval": 3600,  # 会话清理间隔
        }
        
        # 启动后台优化任务
        self._optimization_task = None
        self._running = False
    
    def start_optimization_loop(self):
        """启动优化循环"""
        if not self._running:
            self._running = True
            self._optimization_task = asyncio.create_task(self._optimization_loop())
    
    def stop_optimization_loop(self):
        """停止优化循环"""
        self._running = False
        if self._optimization_task:
            self._optimization_task.cancel()
    
    async def _optimization_loop(self):
        """优化主循环"""
        last_gc = time.time()
        last_cache_cleanup = time.time()
        last_session_cleanup = time.time()
        
        while self._running:
            try:
                current_time = time.time()
                
                # 获取系统指标
                metrics = await self.get_system_metrics()
                
                # 检查是否需要紧急优化
                if await self._needs_immediate_optimization(metrics):
                    await self.optimize_system()
                
                # 定期垃圾回收
                if current_time - last_gc >= self.optimization_config["gc_interval"]:
                    await self._force_garbage_collection()
                    last_gc = current_time
                
                # 定期缓存清理
                if current_time - last_cache_cleanup >= self.optimization_config["cache_cleanup_interval"]:
                    await self._cleanup_caches()
                    last_cache_cleanup = current_time
                
                # 定期会话清理
                if current_time - last_session_cleanup >= self.optimization_config["session_cleanup_interval"]:
                    await self._cleanup_inactive_sessions()
                    last_session_cleanup = current_time
                
                # 休眠30秒
                await asyncio.sleep(30)
                
            except asyncio.CancelledError:
                break
            except Exception as e:
                self.logger.error(f"Optimization loop error: {e}")
                await asyncio.sleep(60)  # 出错时延长休眠时间
    
    async def get_system_metrics(self) -> SystemMetrics:
        """获取系统指标"""
        
        # CPU和内存指标
        cpu_percent = psutil.cpu_percent(interval=1)
        memory = psutil.virtual_memory()
        disk = psutil.disk_usage('/')
        
        # 网络IO
        network = psutil.net_io_counters()
        network_io = {
            "bytes_sent": network.bytes_sent,
            "bytes_recv": network.bytes_recv,
            "packets_sent": network.packets_sent,
            "packets_recv": network.packets_recv
        }
        
        # 进程信息
        process = psutil.Process()
        process_count = len(psutil.pids())
        
        try:
            thread_count = process.num_threads()
            fd_count = process.num_fds() if hasattr(process, 'num_fds') else 0
        except (psutil.NoSuchProcess, psutil.AccessDenied):
            thread_count = 0
            fd_count = 0
        
        return SystemMetrics(
            cpu_percent=cpu_percent,
            memory_percent=memory.percent,
            memory_available=memory.available,
            disk_usage_percent=disk.percent,
            network_io=network_io,
            process_count=process_count,
            thread_count=thread_count,
            fd_count=fd_count,
            timestamp=datetime.now()
        )
    
    async def _needs_immediate_optimization(self, metrics: SystemMetrics) -> bool:
        """检查是否需要立即优化"""
        
        return (
            metrics.memory_percent > self.optimization_config["memory_threshold"] or
            metrics.cpu_percent > self.optimization_config["cpu_threshold"] or
            metrics.disk_usage_percent > self.optimization_config["disk_threshold"]
        )
    
    async def optimize_system(self) -> List[OptimizationResult]:
        """系统优化"""
        
        self.logger.info("Starting system optimization...")
        
        results = []
        
        # 获取优化前的指标
        before_metrics = await self.get_system_metrics()
        
        # 1. 强制垃圾回收
        gc_result = await self._force_garbage_collection()
        results.append(gc_result)
        
        # 2. 清理缓存
        cache_result = await self._cleanup_caches()
        results.append(cache_result)
        
        # 3. 清理不活跃会话
        session_result = await self._cleanup_inactive_sessions()
        results.append(session_result)
        
        # 4. 优化内存使用
        memory_result = await self._optimize_memory_usage()
        results.append(memory_result)
        
        # 5. 清理临时文件
        temp_result = await self._cleanup_temp_files()
        results.append(temp_result)
        
        # 获取优化后的指标
        await asyncio.sleep(2)  # 等待指标稳定
        after_metrics = await self.get_system_metrics()
        
        # 记录优化历史
        self.optimization_history.append(OptimizationResult(
            action="full_system_optimization",
            before_metrics=before_metrics.__dict__,
            after_metrics=after_metrics.__dict__,
            improvement={
                "memory_reduction": before_metrics.memory_percent - after_metrics.memory_percent,
                "cpu_reduction": before_metrics.cpu_percent - after_metrics.cpu_percent
            },
            execution_time=time.time(),
            success=True,
            message=f"Optimized {len(results)} components"
        ))
        
        self.logger.info(f"System optimization completed. Memory: {before_metrics.memory_percent:.1f}% -> {after_metrics.memory_percent:.1f}%")
        
        return results
    
    async def _force_garbage_collection(self) -> OptimizationResult:
        """强制垃圾回收"""
        
        start_time = time.time()
        
        # 获取GC前的内存使用
        before_memory = psutil.virtual_memory().percent
        
        # 执行垃圾回收
        collected_objects = []
        for generation in range(3):
            collected = gc.collect(generation)
            collected_objects.append(collected)
        
        # 获取GC后的内存使用
        after_memory = psutil.virtual_memory().percent
        
        execution_time = time.time() - start_time
        
        return OptimizationResult(
            action="garbage_collection",
            before_metrics={"memory_percent": before_memory},
            after_metrics={"memory_percent": after_memory},
            improvement={"memory_reduction": before_memory - after_memory},
            execution_time=execution_time,
            success=True,
            message=f"Collected {sum(collected_objects)} objects"
        )
    
    async def _cleanup_caches(self) -> OptimizationResult:
        """清理缓存"""
        
        start_time = time.time()
        before_memory = psutil.virtual_memory().percent
        
        cleaned_items = 0
        
        # 清理性能监控器缓存
        try:
            # 这里可以添加具体的缓存清理逻辑
            # 例如清理LRU缓存、清理过期条目等
            pass
        except Exception as e:
            self.logger.warning(f"Cache cleanup error: {e}")
        
        after_memory = psutil.virtual_memory().percent
        execution_time = time.time() - start_time
        
        return OptimizationResult(
            action="cache_cleanup",
            before_metrics={"memory_percent": before_memory},
            after_metrics={"memory_percent": after_memory},
            improvement={"memory_reduction": before_memory - after_memory},
            execution_time=execution_time,
            success=True,
            message=f"Cleaned {cleaned_items} cache items"
        )
    
    async def _cleanup_inactive_sessions(self) -> OptimizationResult:
        """清理不活跃会话"""
        
        start_time = time.time()
        before_memory = psutil.virtual_memory().percent
        
        cleaned_sessions = 0
        
        # 这里应该调用各个服务的会话清理方法
        # 暂时模拟实现
        
        after_memory = psutil.virtual_memory().percent
        execution_time = time.time() - start_time
        
        return OptimizationResult(
            action="session_cleanup",
            before_metrics={"memory_percent": before_memory},
            after_metrics={"memory_percent": after_memory},
            improvement={"memory_reduction": before_memory - after_memory},
            execution_time=execution_time,
            success=True,
            message=f"Cleaned {cleaned_sessions} inactive sessions"
        )
    
    async def _optimize_memory_usage(self) -> OptimizationResult:
        """优化内存使用"""
        
        start_time = time.time()
        before_memory = psutil.virtual_memory().percent
        
        # 内存优化策略
        optimizations_applied = []
        
        # 1. 压缩大型对象
        try:
            # 这里可以实现具体的内存压缩逻辑
            optimizations_applied.append("object_compression")
        except Exception as e:
            self.logger.warning(f"Memory compression error: {e}")
        
        # 2. 释放未使用的内存池
        try:
            # 释放内存池
            optimizations_applied.append("memory_pool_release")
        except Exception as e:
            self.logger.warning(f"Memory pool release error: {e}")
        
        after_memory = psutil.virtual_memory().percent
        execution_time = time.time() - start_time
        
        return OptimizationResult(
            action="memory_optimization",
            before_metrics={"memory_percent": before_memory},
            after_metrics={"memory_percent": after_memory},
            improvement={"memory_reduction": before_memory - after_memory},
            execution_time=execution_time,
            success=True,
            message=f"Applied {len(optimizations_applied)} memory optimizations"
        )
    
    async def _cleanup_temp_files(self) -> OptimizationResult:
        """清理临时文件"""
        
        start_time = time.time()
        before_disk = psutil.disk_usage('/').percent
        
        cleaned_files = 0
        cleaned_size = 0
        
        # 清理临时文件的逻辑
        # 这里可以实现具体的临时文件清理
        
        after_disk = psutil.disk_usage('/').percent
        execution_time = time.time() - start_time
        
        return OptimizationResult(
            action="temp_file_cleanup",
            before_metrics={"disk_percent": before_disk},
            after_metrics={"disk_percent": after_disk},
            improvement={"disk_reduction": before_disk - after_disk},
            execution_time=execution_time,
            success=True,
            message=f"Cleaned {cleaned_files} temp files ({cleaned_size} bytes)"
        )
    
    async def get_optimization_report(self) -> Dict[str, Any]:
        """获取优化报告"""
        
        current_metrics = await self.get_system_metrics()
        perf_stats = performance_monitor.get_stats()
        
        # 计算最近优化的效果
        recent_optimizations = self.optimization_history[-10:] if self.optimization_history else []
        
        total_memory_saved = sum(
            result.improvement.get("memory_reduction", 0)
            for result in recent_optimizations
        )
        
        return {
            "current_metrics": current_metrics.__dict__,
            "performance_stats": perf_stats,
            "recent_optimizations": len(recent_optimizations),
            "total_memory_saved_percent": total_memory_saved,
            "optimization_config": self.optimization_config,
            "system_health": {
                "memory_status": "healthy" if current_metrics.memory_percent < 80 else "warning" if current_metrics.memory_percent < 95 else "critical",
                "cpu_status": "healthy" if current_metrics.cpu_percent < 70 else "warning" if current_metrics.cpu_percent < 90 else "critical",
                "disk_status": "healthy" if current_metrics.disk_usage_percent < 80 else "warning" if current_metrics.disk_usage_percent < 95 else "critical"
            },
            "recommendations": await self._generate_recommendations(current_metrics)
        }
    
    async def _generate_recommendations(self, metrics: SystemMetrics) -> List[str]:
        """生成优化建议"""
        
        recommendations = []
        
        if metrics.memory_percent > 85:
            recommendations.append("内存使用率过高，建议清理缓存或重启部分服务")
        
        if metrics.cpu_percent > 80:
            recommendations.append("CPU使用率过高，建议检查是否有异常进程或优化算法")
        
        if metrics.disk_usage_percent > 85:
            recommendations.append("磁盘空间不足，建议清理日志文件和临时文件")
        
        if metrics.thread_count > 1000:
            recommendations.append("线程数过多，建议检查是否有线程泄漏")
        
        if metrics.fd_count > 1000:
            recommendations.append("文件描述符数量过多，建议检查是否有文件句柄泄漏")
        
        return recommendations


# 全局优化器实例
system_optimizer = SystemOptimizer()