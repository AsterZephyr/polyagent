"use client";

import { Database, Brain, Zap, BarChart3, Target, TrendingUp } from "lucide-react";
import { GlowingEffect } from "@/components/ui/glowing-effect";
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

  useEffect(() => {
    const fetchData = async () => {
      try {
        const [metricsData, agentsData] = await Promise.all([
          recommendationApi.getSystemMetrics(),
          recommendationApi.getAgents()
        ]);
        setMetrics(metricsData);
        setAgents(agentsData);
      } catch (error) {
        console.error('Failed to fetch data:', error);
      }
    };

    fetchData();
    const interval = setInterval(fetchData, 3000);
    return () => clearInterval(interval);
  }, []);

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-950 via-gray-900 to-slate-900 p-6">
      <div className="max-w-7xl mx-auto space-y-8">
        {/* Header */}
        <div className="text-center space-y-4">
          <h1 className="text-4xl font-bold bg-gradient-to-r from-white to-gray-300 bg-clip-text text-transparent">
            推荐业务智能体系统
          </h1>
          <p className="text-gray-400 text-lg max-w-2xl mx-auto">
            基于Agent4Rec架构的专业推荐业务闭环，从数据采集到实时推荐的完整AI驱动解决方案
          </p>
        </div>

        {/* Agent Grid */}
        <ul className="grid grid-cols-1 grid-rows-none gap-6 md:grid-cols-12 md:grid-rows-3 lg:gap-6 xl:max-h-[40rem] xl:grid-rows-2">
          <GridItem
            area="md:[grid-area:1/1/2/7] xl:[grid-area:1/1/2/5]"
            icon={<Database className="h-6 w-6" />}
            title="数据采集Agent"
            subtitle="DataAgent"
            description={
              <div className="space-y-2">
                <p>专业数据采集和特征工程智能体</p>
                <div className="flex items-center gap-4 text-sm">
                  <span className="text-green-400">活跃状态</span>
                  <span className="text-blue-400">{metrics?.queued_tasks || 0} 队列任务</span>
                </div>
              </div>
            }
            bgGradient="from-blue-500/20 to-cyan-500/20"
            iconBg="bg-blue-500/20 border-blue-500/30"
          />
          
          <GridItem
            area="md:[grid-area:1/7/2/13] xl:[grid-area:2/1/3/5]"
            icon={<Brain className="h-6 w-6" />}
            title="模型训练Agent"
            subtitle="ModelAgent"
            description={
              <div className="space-y-2">
                <p>智能推荐算法训练与优化</p>
                <div className="flex items-center gap-4 text-sm">
                  <span className="text-purple-400">协同过滤</span>
                  <span className="text-pink-400">深度学习</span>
                </div>
              </div>
            }
            bgGradient="from-purple-500/20 to-pink-500/20"
            iconBg="bg-purple-500/20 border-purple-500/30"
          />
          
          <GridItem
            area="md:[grid-area:2/1/3/7] xl:[grid-area:1/5/3/8]"
            icon={<Zap className="h-6 w-6" />}
            title="实时推荐引擎"
            subtitle="ServiceAgent"
            description={
              <div className="space-y-2">
                <p>高性能推荐服务与预测</p>
                <div className="flex items-center gap-4 text-sm">
                  <span className="text-yellow-400">{metrics?.average_latency || 0}ms 延迟</span>
                  <span className="text-green-400">
                    {metrics?.success_rate_today ? (metrics.success_rate_today * 100).toFixed(1) : '0'}% 成功率
                  </span>
                </div>
              </div>
            }
            bgGradient="from-yellow-500/20 to-orange-500/20"
            iconBg="bg-yellow-500/20 border-yellow-500/30"
          />
          
          <GridItem
            area="md:[grid-area:2/7/3/13] xl:[grid-area:1/8/2/13]"
            icon={<BarChart3 className="h-6 w-6" />}
            title="效果评估Agent"
            subtitle="EvalAgent"
            description={
              <div className="space-y-2">
                <p>A/B测试与效果监控分析</p>
                <div className="flex items-center gap-4 text-sm">
                  <span className="text-emerald-400">NDCG@K</span>
                  <span className="text-teal-400">Precision@K</span>
                </div>
              </div>
            }
            bgGradient="from-emerald-500/20 to-teal-500/20"
            iconBg="bg-emerald-500/20 border-emerald-500/30"
          />
          
          <GridItem
            area="md:[grid-area:3/1/4/13] xl:[grid-area:2/8/3/13]"
            icon={<Target className="h-6 w-6" />}
            title="业务指标监控"
            subtitle="BusinessMetrics"
            description={
              <div className="space-y-2">
                <p>点击率、转化率、覆盖率实时监控</p>
                <div className="flex items-center gap-4 text-sm">
                  <span className="text-rose-400">今日 {metrics?.total_tasks_today || 0} 任务</span>
                  <span className="text-orange-400">{metrics?.processing_tasks || 0} 处理中</span>
                </div>
              </div>
            }
            bgGradient="from-rose-500/20 to-red-500/20"
            iconBg="bg-rose-500/20 border-rose-500/30"
          />
        </ul>

        {/* System Status */}
        <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
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
  area: string;
  icon: React.ReactNode;
  title: string;
  subtitle: string;
  description: React.ReactNode;
  bgGradient: string;
  iconBg: string;
}

const GridItem = ({ area, icon, title, subtitle, description, bgGradient, iconBg }: GridItemProps) => {
  return (
    <li className={cn("min-h-[18rem] list-none", area)}>
      <div className="relative h-full rounded-[1.25rem] border border-gray-800/50 p-2 md:rounded-[1.5rem] md:p-3">
        <GlowingEffect
          spread={40}
          glow={true}
          disabled={false}
          proximity={64}
          inactiveZone={0.01}
          borderWidth={2}
        />
        <div className={cn(
          "relative flex h-full flex-col justify-between gap-6 overflow-hidden rounded-xl border border-gray-800/50 p-6 shadow-2xl",
          "bg-gradient-to-br", bgGradient,
          "backdrop-blur-sm"
        )}>
          <div className="relative flex flex-1 flex-col justify-between gap-4">
            <div className={cn(
              "w-fit rounded-xl border p-3",
              iconBg
            )}>
              <div className="text-white">
                {icon}
              </div>
            </div>
            <div className="space-y-4">
              <div>
                <div className="text-xs text-gray-400 font-medium mb-1">{subtitle}</div>
                <h3 className="text-2xl font-bold text-white leading-tight">
                  {title}
                </h3>
              </div>
              <div className="text-gray-300 text-sm leading-relaxed">
                {description}
              </div>
            </div>
          </div>
        </div>
      </div>
    </li>
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
    blue: "from-blue-500/20 to-cyan-500/20 border-blue-500/30",
    purple: "from-purple-500/20 to-pink-500/20 border-purple-500/30", 
    green: "from-green-500/20 to-emerald-500/20 border-green-500/30",
    orange: "from-orange-500/20 to-red-500/20 border-orange-500/30"
  };

  return (
    <div className="relative rounded-2xl border border-gray-800/50 p-1">
      <GlowingEffect
        spread={30}
        glow={true}
        disabled={false}
        proximity={32}
        inactiveZone={0.1}
        borderWidth={1}
      />
      <div className={cn(
        "relative rounded-xl border p-6 backdrop-blur-sm",
        "bg-gradient-to-br", colorClasses[color]
      )}>
        <div className="space-y-2">
          <div className="text-gray-400 text-sm font-medium">{title}</div>
          <div className="text-3xl font-bold text-white">{value}</div>
          <div className="text-gray-300 text-xs">{subtitle}</div>
        </div>
      </div>
    </div>
  );
};