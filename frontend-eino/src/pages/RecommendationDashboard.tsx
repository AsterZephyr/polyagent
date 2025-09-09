import React, { useState, useEffect } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Badge } from '@/components/ui/Badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/Tabs'
import { BarChart3, Database, Brain, Activity, TrendingUp } from 'lucide-react'
import { recommendationApi } from '@/services/recommendation'

interface SystemMetrics {
  total_agents: number
  active_agents: number
  queued_tasks: number
  processing_tasks: number
  total_tasks_today: number
  success_rate_today: number
  average_latency: number
  timestamp: string
}

export function RecommendationDashboard() {
  const [metrics, setMetrics] = useState<SystemMetrics | null>(null)
  const [loading, setLoading] = useState(false)
  const [agents, setAgents] = useState<any[]>([])

  const fetchSystemMetrics = async () => {
    try {
      setLoading(true)
      const data = await recommendationApi.getSystemMetrics()
      setMetrics(data)
    } catch (error) {
      console.error('Failed to fetch system metrics:', error)
    } finally {
      setLoading(false)
    }
  }

  const fetchAgents = async () => {
    try {
      const agentsData = await recommendationApi.getAgents()
      setAgents(agentsData)
    } catch (error) {
      console.error('Failed to fetch agents:', error)
    }
  }

  useEffect(() => {
    fetchSystemMetrics()
    fetchAgents()
    const interval = setInterval(() => {
      fetchSystemMetrics()
      fetchAgents()
    }, 10000) // 10秒更新一次
    return () => clearInterval(interval)
  }, [])

  const handleDataCollection = async () => {
    try {
      setLoading(true)
      const result = await recommendationApi.collectData({
        collector: 'user_behavior',
        timerange: 'last_7_days'
      })
      console.log('Data collection result:', result)
    } catch (error) {
      console.error('Data collection failed:', error)
    } finally {
      setLoading(false)
    }
  }

  const handleModelTraining = async () => {
    try {
      setLoading(true)
      const result = await recommendationApi.trainModel({
        algorithm: 'collaborative_filtering',
        hyperparameters: {
          learning_rate: 0.001,
          num_factors: 64
        }
      })
      console.log('Model training result:', result)
    } catch (error) {
      console.error('Model training failed:', error)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">推荐业务控制台</h1>
          <p className="text-muted-foreground">监控和管理推荐系统的数据处理、模型训练和业务指标</p>
        </div>
        <Badge variant="outline" className="flex items-center gap-2">
          <Activity className="h-4 w-4" />
          系统运行中
        </Badge>
      </div>

      {/* System Metrics Cards */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">总代理数</CardTitle>
            <Database className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{metrics?.total_agents || 0}</div>
            <p className="text-xs text-muted-foreground">
              活跃代理: {metrics?.active_agents || 0}
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">任务队列</CardTitle>
            <BarChart3 className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{metrics?.queued_tasks || 0}</div>
            <p className="text-xs text-muted-foreground">
              处理中: {metrics?.processing_tasks || 0}
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">今日成功率</CardTitle>
            <TrendingUp className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {metrics?.success_rate_today ? (metrics.success_rate_today * 100).toFixed(1) : 0}%
            </div>
            <p className="text-xs text-muted-foreground">
              今日任务: {metrics?.total_tasks_today || 0}
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">平均延迟</CardTitle>
            <Activity className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{metrics?.average_latency || 0}ms</div>
            <p className="text-xs text-muted-foreground">
              系统响应时间
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Main Content */}
      <Tabs defaultValue="operations" className="space-y-4">
        <TabsList>
          <TabsTrigger value="operations">业务操作</TabsTrigger>
          <TabsTrigger value="monitoring">监控面板</TabsTrigger>
          <TabsTrigger value="models">模型管理</TabsTrigger>
        </TabsList>

        <TabsContent value="operations" className="space-y-4">
          <div className="grid gap-4 md:grid-cols-2">
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Database className="h-5 w-5" />
                  数据操作
                </CardTitle>
                <CardDescription>
                  管理推荐系统的数据采集和特征工程
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <Button 
                  onClick={handleDataCollection}
                  disabled={loading}
                  className="w-full"
                >
                  开始数据采集
                </Button>
                <Button variant="outline" className="w-full">
                  特征工程
                </Button>
                <Button variant="outline" className="w-full">
                  数据验证
                </Button>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Brain className="h-5 w-5" />
                  模型操作
                </CardTitle>
                <CardDescription>
                  训练、评估和部署推荐模型
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <Button 
                  onClick={handleModelTraining}
                  disabled={loading}
                  className="w-full"
                >
                  开始模型训练
                </Button>
                <Button variant="outline" className="w-full">
                  模型评估
                </Button>
                <Button variant="outline" className="w-full">
                  超参数优化
                </Button>
              </CardContent>
            </Card>
          </div>
        </TabsContent>

        <TabsContent value="monitoring" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>实时监控</CardTitle>
              <CardDescription>
                推荐系统实时性能指标和状态监控
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="grid gap-4 md:grid-cols-3">
                <div className="space-y-2">
                  <p className="text-sm font-medium">代理状态</p>
                  <div className="text-2xl font-bold text-green-600">
                    {metrics?.active_agents || 0} / {metrics?.total_agents || 0}
                  </div>
                  <p className="text-xs text-muted-foreground">活跃/总数</p>
                </div>
                <div className="space-y-2">
                  <p className="text-sm font-medium">任务处理</p>
                  <div className="text-2xl font-bold text-blue-600">
                    {metrics?.processing_tasks || 0}
                  </div>
                  <p className="text-xs text-muted-foreground">正在处理</p>
                </div>
                <div className="space-y-2">
                  <p className="text-sm font-medium">队列长度</p>
                  <div className="text-2xl font-bold text-orange-600">
                    {metrics?.queued_tasks || 0}
                  </div>
                  <p className="text-xs text-muted-foreground">等待处理</p>
                </div>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="models" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>模型管理</CardTitle>
              <CardDescription>
                管理推荐算法模型的训练、部署和版本控制
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                <div className="grid gap-4 md:grid-cols-2">
                  <div className="space-y-2">
                    <h4 className="text-sm font-medium">协同过滤</h4>
                    <Badge variant="secondary">已训练</Badge>
                    <p className="text-xs text-muted-foreground">基于用户-物品交互的推荐算法</p>
                  </div>
                  <div className="space-y-2">
                    <h4 className="text-sm font-medium">内容推荐</h4>
                    <Badge variant="outline">待训练</Badge>
                    <p className="text-xs text-muted-foreground">基于物品特征的推荐算法</p>
                  </div>
                  <div className="space-y-2">
                    <h4 className="text-sm font-medium">矩阵分解</h4>
                    <Badge variant="outline">待训练</Badge>
                    <p className="text-xs text-muted-foreground">基于矩阵分解的推荐算法</p>
                  </div>
                  <div className="space-y-2">
                    <h4 className="text-sm font-medium">深度学习</h4>
                    <Badge variant="outline">待训练</Badge>
                    <p className="text-xs text-muted-foreground">基于深度神经网络的推荐算法</p>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  )
}