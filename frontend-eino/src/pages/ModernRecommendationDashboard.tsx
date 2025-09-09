import React, { useState, useEffect } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Badge } from '@/components/ui/Badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/Tabs'
import { 
  BarChart3, 
  Database, 
  Brain, 
  Activity, 
  TrendingUp, 
  Users, 
  Zap,
  ChevronRight,
  Play,
  Pause,
  Settings,
  RefreshCw,
  AlertCircle,
  CheckCircle2,
  Clock,
  Target,
  Sparkles,
  LineChart,
  PieChart
} from 'lucide-react'
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

interface QuickAction {
  id: string
  title: string
  description: string
  icon: React.ElementType
  color: string
  action: () => void
  loading?: boolean
}

export function ModernRecommendationDashboard() {
  const [metrics, setMetrics] = useState<SystemMetrics | null>(null)
  const [loading, setLoading] = useState(false)
  const [agents, setAgents] = useState<any[]>([])
  const [realtimeMode, setRealtimeMode] = useState(true)

  const fetchSystemMetrics = async () => {
    try {
      const data = await recommendationApi.getSystemMetrics()
      setMetrics(data)
    } catch (error) {
      console.error('Failed to fetch system metrics:', error)
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
    
    if (realtimeMode) {
      const interval = setInterval(() => {
        fetchSystemMetrics()
        fetchAgents()
      }, 5000) // 5秒实时更新
      return () => clearInterval(interval)
    }
  }, [realtimeMode])

  const quickActions: QuickAction[] = [
    {
      id: 'data-collect',
      title: '数据采集',
      description: '开始用户行为数据采集',
      icon: Database,
      color: 'from-blue-500 to-blue-600',
      action: async () => {
        setLoading(true)
        try {
          await recommendationApi.collectData({
            collector: 'user_behavior',
            timerange: 'last_7_days'
          })
        } catch (error) {
          console.error('Data collection failed:', error)
        } finally {
          setLoading(false)
        }
      }
    },
    {
      id: 'model-train',
      title: '模型训练',
      description: '训练协同过滤推荐模型',
      icon: Brain,
      color: 'from-purple-500 to-purple-600',
      action: async () => {
        setLoading(true)
        try {
          await recommendationApi.trainModel({
            algorithm: 'collaborative_filtering',
            hyperparameters: { learning_rate: 0.001, num_factors: 64 }
          })
        } catch (error) {
          console.error('Model training failed:', error)
        } finally {
          setLoading(false)
        }
      }
    },
    {
      id: 'feature-extract',
      title: '特征工程',
      description: '提取用户和物品特征',
      icon: Zap,
      color: 'from-orange-500 to-orange-600',
      action: () => console.log('Feature extraction')
    },
    {
      id: 'model-optimize',
      title: '模型优化',
      description: '超参数调优和性能提升',
      icon: Target,
      color: 'from-green-500 to-green-600',
      action: () => console.log('Model optimization')
    }
  ]

  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-50 to-white dark:from-gray-900 dark:to-gray-800">
      <div className="space-y-8 p-6">
        {/* Hero Header */}
        <div className="relative overflow-hidden rounded-2xl bg-gradient-to-r from-indigo-600 via-purple-600 to-blue-600 p-8 text-white">
          <div className="relative z-10">
            <div className="flex items-center justify-between">
              <div>
                <h1 className="text-4xl font-bold mb-2">推荐AI控制台</h1>
                <p className="text-lg opacity-90">智能推荐系统 · 数据驱动 · 实时洞察</p>
              </div>
              <div className="flex items-center gap-4">
                <div className="flex items-center gap-2">
                  <Button 
                    variant="secondary"
                    size="sm"
                    onClick={() => setRealtimeMode(!realtimeMode)}
                    className="bg-white/20 hover:bg-white/30 text-white border-white/20"
                  >
                    {realtimeMode ? <Pause className="h-4 w-4" /> : <Play className="h-4 w-4" />}
                    {realtimeMode ? '暂停实时' : '开启实时'}
                  </Button>
                </div>
                <Badge variant="secondary" className="bg-white/20 text-white border-white/20 px-3 py-2">
                  <Activity className="h-4 w-4 mr-2" />
                  系统运行中
                </Badge>
              </div>
            </div>
          </div>
          <div className="absolute inset-0 bg-gradient-to-r from-black/20 to-transparent"></div>
          <Sparkles className="absolute top-4 right-4 h-8 w-8 opacity-30" />
        </div>

        {/* Real-time Metrics */}
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-4">
          <Card className="relative overflow-hidden border-0 shadow-lg bg-gradient-to-br from-blue-50 to-blue-100 dark:from-blue-950 dark:to-blue-900">
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium text-blue-700 dark:text-blue-300">智能代理</CardTitle>
              <Users className="h-4 w-4 text-blue-600 dark:text-blue-400" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold text-blue-800 dark:text-blue-200">
                {metrics?.active_agents || 0} / {metrics?.total_agents || 0}
              </div>
              <p className="text-xs text-blue-600 dark:text-blue-400 mt-1">
                活跃代理数量
              </p>
            </CardContent>
          </Card>

          <Card className="relative overflow-hidden border-0 shadow-lg bg-gradient-to-br from-green-50 to-green-100 dark:from-green-950 dark:to-green-900">
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium text-green-700 dark:text-green-300">成功率</CardTitle>
              <TrendingUp className="h-4 w-4 text-green-600 dark:text-green-400" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold text-green-800 dark:text-green-200">
                {metrics?.success_rate_today ? (metrics.success_rate_today * 100).toFixed(1) : '0.0'}%
              </div>
              <p className="text-xs text-green-600 dark:text-green-400 mt-1">
                今日任务成功率
              </p>
            </CardContent>
          </Card>

          <Card className="relative overflow-hidden border-0 shadow-lg bg-gradient-to-br from-orange-50 to-orange-100 dark:from-orange-950 dark:to-orange-900">
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium text-orange-700 dark:text-orange-300">任务队列</CardTitle>
              <BarChart3 className="h-4 w-4 text-orange-600 dark:text-orange-400" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold text-orange-800 dark:text-orange-200">
                {metrics?.queued_tasks || 0}
              </div>
              <p className="text-xs text-orange-600 dark:text-orange-400 mt-1">
                等待处理: {metrics?.processing_tasks || 0}
              </p>
            </CardContent>
          </Card>

          <Card className="relative overflow-hidden border-0 shadow-lg bg-gradient-to-br from-purple-50 to-purple-100 dark:from-purple-950 dark:to-purple-900">
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium text-purple-700 dark:text-purple-300">响应时间</CardTitle>
              <Clock className="h-4 w-4 text-purple-600 dark:text-purple-400" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold text-purple-800 dark:text-purple-200">
                {metrics?.average_latency || 0}ms
              </div>
              <p className="text-xs text-purple-600 dark:text-purple-400 mt-1">
                平均响应延迟
              </p>
            </CardContent>
          </Card>
        </div>

        {/* Quick Actions */}
        <Card className="border-0 shadow-lg">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Zap className="h-5 w-5 text-yellow-500" />
              快速操作
            </CardTitle>
            <CardDescription>
              一键执行常用的推荐系统操作
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
              {quickActions.map((action) => (
                <Button
                  key={action.id}
                  onClick={action.action}
                  disabled={loading}
                  className={`h-auto p-4 bg-gradient-to-r ${action.color} hover:opacity-90 transition-all duration-200 group`}
                  size="lg"
                >
                  <div className="flex flex-col items-center text-center space-y-2">
                    <action.icon className="h-6 w-6 group-hover:scale-110 transition-transform" />
                    <div>
                      <div className="font-semibold">{action.title}</div>
                      <div className="text-xs opacity-90">{action.description}</div>
                    </div>
                  </div>
                </Button>
              ))}
            </div>
          </CardContent>
        </Card>

        {/* Advanced Analytics */}
        <Tabs defaultValue="analytics" className="space-y-6">
          <TabsList className="grid w-full grid-cols-3 bg-gray-100 dark:bg-gray-800 rounded-xl p-1">
            <TabsTrigger value="analytics" className="rounded-lg">
              <LineChart className="h-4 w-4 mr-2" />
              数据分析
            </TabsTrigger>
            <TabsTrigger value="models" className="rounded-lg">
              <Brain className="h-4 w-4 mr-2" />
              模型管理
            </TabsTrigger>
            <TabsTrigger value="monitoring" className="rounded-lg">
              <Activity className="h-4 w-4 mr-2" />
              实时监控
            </TabsTrigger>
          </TabsList>

          <TabsContent value="analytics" className="space-y-6">
            <div className="grid gap-6 md:grid-cols-2">
              <Card className="border-0 shadow-lg">
                <CardHeader>
                  <CardTitle className="flex items-center gap-2">
                    <PieChart className="h-5 w-5 text-blue-500" />
                    推荐效果分析
                  </CardTitle>
                  <CardDescription>模型性能与业务指标概览</CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="space-y-4">
                    <div className="flex justify-between items-center">
                      <span className="text-sm font-medium">点击率 (CTR)</span>
                      <Badge variant="secondary">12.8%</Badge>
                    </div>
                    <div className="flex justify-between items-center">
                      <span className="text-sm font-medium">转化率 (CVR)</span>
                      <Badge variant="secondary">3.2%</Badge>
                    </div>
                    <div className="flex justify-between items-center">
                      <span className="text-sm font-medium">用户覆盖率</span>
                      <Badge variant="secondary">89.5%</Badge>
                    </div>
                    <div className="flex justify-between items-center">
                      <span className="text-sm font-medium">物品多样性</span>
                      <Badge variant="secondary">76.3%</Badge>
                    </div>
                  </div>
                </CardContent>
              </Card>

              <Card className="border-0 shadow-lg">
                <CardHeader>
                  <CardTitle className="flex items-center gap-2">
                    <Target className="h-5 w-5 text-green-500" />
                    模型精度指标
                  </CardTitle>
                  <CardDescription>算法性能评估结果</CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="space-y-4">
                    <div className="flex justify-between items-center">
                      <span className="text-sm font-medium">RMSE</span>
                      <Badge variant="outline">0.842</Badge>
                    </div>
                    <div className="flex justify-between items-center">
                      <span className="text-sm font-medium">MAE</span>
                      <Badge variant="outline">0.673</Badge>
                    </div>
                    <div className="flex justify-between items-center">
                      <span className="text-sm font-medium">Precision@10</span>
                      <Badge variant="outline">0.758</Badge>
                    </div>
                    <div className="flex justify-between items-center">
                      <span className="text-sm font-medium">Recall@10</span>
                      <Badge variant="outline">0.634</Badge>
                    </div>
                  </div>
                </CardContent>
              </Card>
            </div>
          </TabsContent>

          <TabsContent value="models" className="space-y-6">
            <Card className="border-0 shadow-lg">
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Brain className="h-5 w-5 text-purple-500" />
                  推荐算法模型
                </CardTitle>
                <CardDescription>管理和监控各种推荐算法的训练状态</CardDescription>
              </CardHeader>
              <CardContent>
                <div className="grid gap-4 md:grid-cols-2">
                  {[
                    { name: '协同过滤', status: 'trained', accuracy: '92.5%' },
                    { name: '内容推荐', status: 'training', accuracy: '88.3%' },
                    { name: '矩阵分解', status: 'pending', accuracy: '90.1%' },
                    { name: '深度学习', status: 'optimizing', accuracy: '94.7%' }
                  ].map((model, index) => (
                    <div key={index} className="flex items-center justify-between p-4 rounded-lg border bg-gradient-to-r from-gray-50 to-white dark:from-gray-800 dark:to-gray-900">
                      <div className="flex items-center gap-3">
                        <div className={`w-3 h-3 rounded-full ${
                          model.status === 'trained' ? 'bg-green-500' :
                          model.status === 'training' ? 'bg-blue-500 animate-pulse' :
                          model.status === 'optimizing' ? 'bg-yellow-500 animate-pulse' :
                          'bg-gray-300'
                        }`}></div>
                        <div>
                          <p className="font-medium">{model.name}</p>
                          <p className="text-sm text-gray-500">精度: {model.accuracy}</p>
                        </div>
                      </div>
                      <Badge variant={
                        model.status === 'trained' ? 'default' :
                        model.status === 'training' ? 'secondary' :
                        model.status === 'optimizing' ? 'secondary' :
                        'outline'
                      }>
                        {model.status === 'trained' ? '已训练' :
                         model.status === 'training' ? '训练中' :
                         model.status === 'optimizing' ? '优化中' :
                         '待训练'}
                      </Badge>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="monitoring" className="space-y-6">
            <Card className="border-0 shadow-lg">
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Activity className="h-5 w-5 text-red-500" />
                  系统健康监控
                </CardTitle>
                <CardDescription>实时监控推荐Agent系统状态</CardDescription>
              </CardHeader>
              <CardContent>
                <div className="grid gap-6 md:grid-cols-3">
                  <div className="space-y-4">
                    <h4 className="font-medium flex items-center gap-2">
                      <Database className="h-4 w-4 text-blue-500" />
                      数据Agent
                    </h4>
                    <div className="space-y-2">
                      <div className="flex items-center justify-between">
                        <span className="text-sm">运行状态</span>
                        <CheckCircle2 className="h-4 w-4 text-green-500" />
                      </div>
                      <div className="flex items-center justify-between">
                        <span className="text-sm">CPU使用率</span>
                        <span className="text-sm font-medium">23%</span>
                      </div>
                      <div className="flex items-center justify-between">
                        <span className="text-sm">内存使用</span>
                        <span className="text-sm font-medium">1.2GB</span>
                      </div>
                    </div>
                  </div>

                  <div className="space-y-4">
                    <h4 className="font-medium flex items-center gap-2">
                      <Brain className="h-4 w-4 text-purple-500" />
                      模型Agent
                    </h4>
                    <div className="space-y-2">
                      <div className="flex items-center justify-between">
                        <span className="text-sm">运行状态</span>
                        <CheckCircle2 className="h-4 w-4 text-green-500" />
                      </div>
                      <div className="flex items-center justify-between">
                        <span className="text-sm">GPU使用率</span>
                        <span className="text-sm font-medium">67%</span>
                      </div>
                      <div className="flex items-center justify-between">
                        <span className="text-sm">显存使用</span>
                        <span className="text-sm font-medium">3.8GB</span>
                      </div>
                    </div>
                  </div>

                  <div className="space-y-4">
                    <h4 className="font-medium flex items-center gap-2">
                      <Activity className="h-4 w-4 text-red-500" />
                      系统负载
                    </h4>
                    <div className="space-y-2">
                      <div className="flex items-center justify-between">
                        <span className="text-sm">服务状态</span>
                        <CheckCircle2 className="h-4 w-4 text-green-500" />
                      </div>
                      <div className="flex items-center justify-between">
                        <span className="text-sm">QPS</span>
                        <span className="text-sm font-medium">1,247</span>
                      </div>
                      <div className="flex items-center justify-between">
                        <span className="text-sm">平均延迟</span>
                        <span className="text-sm font-medium">{metrics?.average_latency || 0}ms</span>
                      </div>
                    </div>
                  </div>
                </div>
              </CardContent>
            </Card>
          </TabsContent>
        </Tabs>
      </div>
    </div>
  )
}