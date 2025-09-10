"use client";

import { Database, Brain, Zap, BarChart3, Target, TrendingUp } from "lucide-react";
import { cn } from "@/lib/utils";
import { useEffect, useState } from "react";
import { recommendationApi } from "@/services/recommendation";

interface SystemMetrics {
  total_agents: number;
  active_agents: number;
  queued_tasks: number;
  processing_tasks: number;
  total_tasks_today: number;
  success_rate_today: number;
  average_latency: number;
  timestamp: string;
}

export function RecommendationAgentGrid() {
  const [metrics, setMetrics] = useState<SystemMetrics | null>(null);
  const [agents, setAgents] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchData = async () => {
      try {
        setError(null);
        const [metricsData, agentsData] = await Promise.all([
          recommendationApi.getSystemMetrics(),
          recommendationApi.getAgents()
        ]);
        setMetrics(metricsData);
        setAgents(agentsData);
        setLoading(false);
      } catch (error) {
        console.error('Failed to fetch data:', error);
        setError('无法连接到推荐系统，请检查后端服务是否运行');
        setLoading(false);
      }
    };

    fetchData();
    const interval = setInterval(fetchData, 3000);
    return () => clearInterval(interval);
  }, []);

  if (loading) {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center">
        <div className="text-center space-y-4">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto"></div>
          <p className="text-muted-foreground">正在加载推荐系统数据...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center">
        <div className="text-center space-y-4 max-w-md">
          <div className="text-destructive text-6xl">⚠️</div>
          <h2 className="text-2xl font-bold text-foreground">连接失败</h2>
          <p className="text-muted-foreground">{error}</p>
          <button 
            onClick={() => window.location.reload()} 
            className="px-4 py-2 bg-primary text-primary-foreground rounded-md hover:bg-primary/90 transition-colors"
          >
            重试
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-background">
      <div className="container mx-auto px-4 py-8 space-y-8">
        {/* Header */}
        <div className="text-center space-y-4">
          <h1 className="text-4xl font-bold text-foreground">
            推荐业务智能体系统
          </h1>
          <p className="text-muted-foreground text-lg max-w-2xl mx-auto">
            基于Agent4Rec架构的专业推荐业务闭环，从数据采集到实时推荐的完整AI驱动解决方案
          </p>
        </div>

        {/* Agent Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
          <GridItem
            icon={<Database className="h-6 w-6" />}
            title="数据采集Agent"
            subtitle="DataAgent"
            description={
              <div className="space-y-2">
                <p>专业数据采集和特征工程智能体</p>
                <div className="flex items-center gap-4 text-sm">
                  <span className="text-green-600 dark:text-green-400">活跃状态</span>
                  <span className="text-blue-600 dark:text-blue-400">{metrics?.queued_tasks || 0} 队列任务</span>
                </div>
              </div>
            }
          />
          
          <GridItem
            icon={<Brain className="h-6 w-6" />}
            title="模型训练Agent"
            subtitle="ModelAgent"
            description={
              <div className="space-y-2">
                <p>智能推荐算法训练与优化</p>
                <div className="flex items-center gap-4 text-sm">
                  <span className="text-purple-600 dark:text-purple-400">协同过滤</span>
                  <span className="text-pink-600 dark:text-pink-400">深度学习</span>
                </div>
              </div>
            }
          />
          
          <GridItem
            icon={<Zap className="h-6 w-6" />}
            title="实时推荐引擎"
            subtitle="ServiceAgent"
            description={
              <div className="space-y-2">
                <p>高性能推荐服务与预测</p>
                <div className="flex items-center gap-4 text-sm">
                  <span className="text-yellow-600 dark:text-yellow-400">{metrics?.average_latency || 0}ms 延迟</span>
                  <span className="text-green-600 dark:text-green-400">
                    {metrics?.success_rate_today ? (metrics.success_rate_today * 100).toFixed(1) : '0'}% 成功率
                  </span>
                </div>
              </div>
            }
          />
          
          <GridItem
            icon={<BarChart3 className="h-6 w-6" />}
            title="效果评估Agent"
            subtitle="EvalAgent"
            description={
              <div className="space-y-2">
                <p>A/B测试与效果监控分析</p>
                <div className="flex items-center gap-4 text-sm">
                  <span className="text-emerald-600 dark:text-emerald-400">NDCG@K</span>
                  <span className="text-teal-600 dark:text-teal-400">Precision@K</span>
                </div>
              </div>
            }
          />
          
          <GridItem
            icon={<Target className="h-6 w-6" />}
            title="业务指标监控"
            subtitle="BusinessMetrics"
            description={
              <div className="space-y-2">
                <p>点击率、转化率、覆盖率实时监控</p>
                <div className="flex items-center gap-4 text-sm">
                  <span className="text-rose-600 dark:text-rose-400">今日 {metrics?.total_tasks_today || 0} 任务</span>
                  <span className="text-orange-600 dark:text-orange-400">{metrics?.processing_tasks || 0} 处理中</span>
                </div>
              </div>
            }
          />
        </div>

        {/* System Status */}
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
          <StatusCard 
            title="总代理数"
            value={metrics?.total_agents || 0}
            subtitle={`${metrics?.active_agents || 0} 活跃中`}
            color="blue"
          />
          <StatusCard 
            title="任务队列" 
            value={metrics?.queued_tasks || 0}
            subtitle={`${metrics?.processing_tasks || 0} 处理中`}
            color="purple"
          />
          <StatusCard 
            title="成功率"
            value={`${metrics?.success_rate_today ? (metrics.success_rate_today * 100).toFixed(1) : '0'}%`}
            subtitle="今日成功率"
            color="green"
          />
          <StatusCard 
            title="平均延迟"
            value={`${metrics?.average_latency || 0}ms`}
            subtitle="推荐响应时间"
            color="orange"
          />
        </div>
      </div>
    </div>
  );
}

interface GridItemProps {
  icon: React.ReactNode;
  title: string;
  subtitle: string;
  description: React.ReactNode;
}

const GridItem = ({ icon, title, subtitle, description }: GridItemProps) => {
  return (
    <div className="group relative h-full">
      <div className="relative h-full rounded-lg border bg-card p-6 shadow-sm transition-all duration-300 hover:shadow-lg hover:scale-[1.02] hover:border-primary/20">
        <div className="flex h-full flex-col gap-4">
          <div className="flex items-center gap-3">
            <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-primary/10 text-primary transition-colors duration-200 group-hover:bg-primary/20">
              {icon}
            </div>
            <div>
              <div className="text-xs font-medium text-muted-foreground">{subtitle}</div>
              <h3 className="text-lg font-semibold text-card-foreground group-hover:text-primary transition-colors duration-200">
                {title}
              </h3>
            </div>
          </div>
          <div className="flex-1 text-sm text-muted-foreground">
            {description}
          </div>
        </div>
      </div>
    </div>
  );
};

interface StatusCardProps {
  title: string;
  value: string | number;
  subtitle: string;
  color: 'blue' | 'purple' | 'green' | 'orange';
}

const StatusCard = ({ title, value, subtitle, color }: StatusCardProps) => {
  const colorClasses = {
    blue: "border-blue-200 bg-blue-50 text-blue-900 dark:border-blue-800 dark:bg-blue-950 dark:text-blue-100",
    purple: "border-purple-200 bg-purple-50 text-purple-900 dark:border-purple-800 dark:bg-purple-950 dark:text-purple-100", 
    green: "border-green-200 bg-green-50 text-green-900 dark:border-green-800 dark:bg-green-950 dark:text-green-100",
    orange: "border-orange-200 bg-orange-50 text-orange-900 dark:border-orange-800 dark:bg-orange-950 dark:text-orange-100"
  };

  return (
    <div className={cn(
      "rounded-lg border p-6 transition-all duration-300 hover:shadow-lg hover:scale-[1.02] group",
      colorClasses[color]
    )}>
      <div className="space-y-2">
        <div className="text-sm font-medium opacity-70 group-hover:opacity-90 transition-opacity duration-200">{title}</div>
        <div className="text-3xl font-bold group-hover:scale-105 transition-transform duration-200">{value}</div>
        <div className="text-xs opacity-60 group-hover:opacity-80 transition-opacity duration-200">{subtitle}</div>
      </div>
    </div>
  );
};