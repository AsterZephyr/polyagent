import React, { useState } from 'react'
import { motion } from 'framer-motion'
import {
  TrendingUp,
  TrendingDown,
  Activity,
  Users,
  MessageCircle,
  Clock,
  Target,
  Zap,
  Bot,
  Brain,
  AlertCircle,
  CheckCircle,
  MoreVertical,
  Download,
  Filter,
  Calendar,
} from 'lucide-react'

const timeRanges = [
  { label: 'Last 7 days', value: '7d' },
  { label: 'Last 30 days', value: '30d' },
  { label: 'Last 90 days', value: '90d' },
  { label: 'Last year', value: '1y' },
]

const metrics = [
  {
    label: 'Total Requests',
    value: '12,847',
    change: '+12.5%',
    trend: 'up',
    icon: MessageCircle,
    color: 'text-primary-600 dark:text-primary-400',
    bgColor: 'bg-primary-100 dark:bg-primary-900/20',
  },
  {
    label: 'Success Rate',
    value: '94.8%',
    change: '+2.1%',
    trend: 'up',
    icon: CheckCircle,
    color: 'text-success-600 dark:text-success-400',
    bgColor: 'bg-success-100 dark:bg-success-900/20',
  },
  {
    label: 'Avg Response Time',
    value: '1.2s',
    change: '-0.3s',
    trend: 'down',
    icon: Clock,
    color: 'text-warning-600 dark:text-warning-400',
    bgColor: 'bg-warning-100 dark:bg-warning-900/20',
  },
  {
    label: 'Active Users',
    value: '2,456',
    change: '+8.7%',
    trend: 'up',
    icon: Users,
    color: 'text-error-600 dark:text-error-400',
    bgColor: 'bg-error-100 dark:bg-error-900/20',
  },
]

const agentPerformance = [
  {
    id: 'general',
    name: 'General Assistant',
    requests: 5247,
    success_rate: 94.5,
    avg_response: 1.2,
    error_rate: 5.5,
    status: 'excellent',
  },
  {
    id: 'research',
    name: 'Research Specialist',
    requests: 3856,
    success_rate: 97.2,
    avg_response: 2.1,
    error_rate: 2.8,
    status: 'excellent',
  },
  {
    id: 'coder',
    name: 'Code Assistant',
    requests: 2423,
    success_rate: 92.8,
    avg_response: 0.9,
    error_rate: 7.2,
    status: 'good',
  },
  {
    id: 'analyst',
    name: 'Data Analyst',
    requests: 1321,
    success_rate: 96.1,
    avg_response: 3.2,
    error_rate: 3.9,
    status: 'excellent',
  },
]

const systemHealth = [
  {
    component: 'API Gateway',
    status: 'healthy',
    uptime: '99.9%',
    response_time: '45ms',
    last_incident: '3 days ago',
  },
  {
    component: 'AI Processing',
    status: 'healthy',
    uptime: '99.7%',
    response_time: '1.2s',
    last_incident: '1 week ago',
  },
  {
    component: 'Vector Database',
    status: 'healthy',
    uptime: '99.8%',
    response_time: '12ms',
    last_incident: '5 days ago',
  },
  {
    component: 'Knowledge Graph',
    status: 'warning',
    uptime: '98.4%',
    response_time: '234ms',
    last_incident: '2 hours ago',
  },
]

const recentActivities = [
  {
    id: '1',
    type: 'performance',
    title: 'High Response Time Alert',
    description: 'Research Specialist agent showing increased response times',
    timestamp: new Date('2024-01-16T14:30:00'),
    severity: 'warning',
  },
  {
    id: '2',
    type: 'system',
    title: 'Knowledge Graph Update',
    description: 'Successfully updated knowledge graph with 1,247 new nodes',
    timestamp: new Date('2024-01-16T12:15:00'),
    severity: 'info',
  },
  {
    id: '3',
    type: 'usage',
    title: 'Usage Spike Detected',
    description: 'Code Assistant requests increased by 34% in the last hour',
    timestamp: new Date('2024-01-16T11:45:00'),
    severity: 'info',
  },
  {
    id: '4',
    type: 'error',
    title: 'API Rate Limit Reached',
    description: 'OpenAI API rate limit reached for General Assistant',
    timestamp: new Date('2024-01-16T10:20:00'),
    severity: 'error',
  },
]

export default function AnalyticsPage() {
  const [selectedTimeRange, setSelectedTimeRange] = useState('30d')

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'healthy':
      case 'excellent':
        return 'text-success-600 dark:text-success-400'
      case 'warning':
      case 'good':
        return 'text-warning-600 dark:text-warning-400'
      case 'error':
      case 'poor':
        return 'text-error-600 dark:text-error-400'
      default:
        return 'text-secondary-600 dark:text-secondary-400'
    }
  }

  const getStatusBg = (status: string) => {
    switch (status) {
      case 'healthy':
      case 'excellent':
        return 'bg-success-100 dark:bg-success-900/20'
      case 'warning':
      case 'good':
        return 'bg-warning-100 dark:bg-warning-900/20'
      case 'error':
      case 'poor':
        return 'bg-error-100 dark:bg-error-900/20'
      default:
        return 'bg-secondary-100 dark:bg-secondary-700'
    }
  }

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case 'info':
        return 'text-primary-600 dark:text-primary-400'
      case 'warning':
        return 'text-warning-600 dark:text-warning-400'
      case 'error':
        return 'text-error-600 dark:text-error-400'
      default:
        return 'text-secondary-600 dark:text-secondary-400'
    }
  }

  return (
    <div className="p-6">
      {/* Header */}
      <div className="mb-8">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold text-secondary-900 dark:text-white mb-2">
              Analytics & Monitoring
            </h1>
            <p className="text-secondary-600 dark:text-secondary-300">
              Monitor system performance and agent analytics
            </p>
          </div>
          <div className="flex items-center space-x-3">
            <select
              value={selectedTimeRange}
              onChange={(e) => setSelectedTimeRange(e.target.value)}
              className="px-4 py-2 bg-white dark:bg-secondary-800 border border-secondary-200 dark:border-secondary-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent transition-all duration-200 text-secondary-900 dark:text-white"
            >
              {timeRanges.map((range) => (
                <option key={range.value} value={range.value}>
                  {range.label}
                </option>
              ))}
            </select>
            <button className="flex items-center space-x-2 px-4 py-2 bg-primary-600 hover:bg-primary-700 text-white rounded-lg transition-colors font-medium">
              <Download className="w-4 h-4" />
              <span>Export</span>
            </button>
          </div>
        </div>
      </div>

      {/* Key Metrics */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
        {metrics.map((metric, index) => (
          <motion.div
            key={metric.label}
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.3, delay: index * 0.1 }}
            className="bg-white dark:bg-secondary-800 rounded-xl p-6 border border-secondary-200 dark:border-secondary-700"
          >
            <div className="flex items-center justify-between mb-4">
              <div className={`w-12 h-12 ${metric.bgColor} rounded-xl flex items-center justify-center`}>
                <metric.icon className={`w-6 h-6 ${metric.color}`} />
              </div>
              <div className={`flex items-center text-sm font-medium ${
                metric.trend === 'up' ? 'text-success-600 dark:text-success-400' : 'text-warning-600 dark:text-warning-400'
              }`}>
                {metric.trend === 'up' ? (
                  <TrendingUp className="w-4 h-4 mr-1" />
                ) : (
                  <TrendingDown className="w-4 h-4 mr-1" />
                )}
                {metric.change}
              </div>
            </div>
            <h3 className="font-medium text-secondary-900 dark:text-white mb-1">
              {metric.label}
            </h3>
            <p className="text-2xl font-bold text-secondary-900 dark:text-white">
              {metric.value}
            </p>
          </motion.div>
        ))}
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-8 mb-8">
        {/* Agent Performance */}
        <div className="bg-white dark:bg-secondary-800 rounded-xl border border-secondary-200 dark:border-secondary-700">
          <div className="px-6 py-4 border-b border-secondary-200 dark:border-secondary-700">
            <div className="flex items-center justify-between">
              <h2 className="text-lg font-semibold text-secondary-900 dark:text-white">
                Agent Performance
              </h2>
              <button className="p-2 text-secondary-600 dark:text-secondary-400 hover:bg-secondary-100 dark:hover:bg-secondary-700 rounded-lg transition-colors">
                <MoreVertical className="w-4 h-4" />
              </button>
            </div>
          </div>

          <div className="p-6">
            <div className="space-y-4">
              {agentPerformance.map((agent, index) => (
                <motion.div
                  key={agent.id}
                  initial={{ opacity: 0, x: -20 }}
                  animate={{ opacity: 1, x: 0 }}
                  transition={{ duration: 0.3, delay: index * 0.1 }}
                  className="flex items-center justify-between p-4 bg-secondary-50 dark:bg-secondary-700 rounded-xl"
                >
                  <div className="flex items-center space-x-3">
                    <div className="w-10 h-10 bg-primary-100 dark:bg-primary-900/20 rounded-lg flex items-center justify-center">
                      <Bot className="w-5 h-5 text-primary-600 dark:text-primary-400" />
                    </div>
                    <div>
                      <h3 className="font-medium text-secondary-900 dark:text-white">
                        {agent.name}
                      </h3>
                      <p className="text-sm text-secondary-500 dark:text-secondary-400">
                        {agent.requests.toLocaleString()} requests
                      </p>
                    </div>
                  </div>
                  <div className="text-right">
                    <div className={`inline-flex items-center px-2 py-1 rounded text-xs font-medium ${getStatusBg(agent.status)}`}>
                      <span className={getStatusColor(agent.status)}>
                        {agent.status}
                      </span>
                    </div>
                    <div className="text-sm text-secondary-600 dark:text-secondary-300 mt-1">
                      {agent.success_rate}% • {agent.avg_response}s
                    </div>
                  </div>
                </motion.div>
              ))}
            </div>
          </div>
        </div>

        {/* System Health */}
        <div className="bg-white dark:bg-secondary-800 rounded-xl border border-secondary-200 dark:border-secondary-700">
          <div className="px-6 py-4 border-b border-secondary-200 dark:border-secondary-700">
            <div className="flex items-center justify-between">
              <h2 className="text-lg font-semibold text-secondary-900 dark:text-white">
                System Health
              </h2>
              <button className="p-2 text-secondary-600 dark:text-secondary-400 hover:bg-secondary-100 dark:hover:bg-secondary-700 rounded-lg transition-colors">
                <MoreVertical className="w-4 h-4" />
              </button>
            </div>
          </div>

          <div className="p-6">
            <div className="space-y-4">
              {systemHealth.map((component, index) => (
                <motion.div
                  key={component.component}
                  initial={{ opacity: 0, x: 20 }}
                  animate={{ opacity: 1, x: 0 }}
                  transition={{ duration: 0.3, delay: index * 0.1 }}
                  className="flex items-center justify-between p-4 bg-secondary-50 dark:bg-secondary-700 rounded-xl"
                >
                  <div className="flex items-center space-x-3">
                    <div className={`w-3 h-3 rounded-full ${
                      component.status === 'healthy' ? 'bg-success-500' : 
                      component.status === 'warning' ? 'bg-warning-500' : 'bg-error-500'
                    }`}></div>
                    <div>
                      <h3 className="font-medium text-secondary-900 dark:text-white">
                        {component.component}
                      </h3>
                      <p className="text-sm text-secondary-500 dark:text-secondary-400">
                        {component.uptime} uptime
                      </p>
                    </div>
                  </div>
                  <div className="text-right">
                    <div className="text-sm font-medium text-secondary-900 dark:text-white">
                      {component.response_time}
                    </div>
                    <div className="text-xs text-secondary-500 dark:text-secondary-400">
                      Last incident: {component.last_incident}
                    </div>
                  </div>
                </motion.div>
              ))}
            </div>
          </div>
        </div>
      </div>

      {/* Recent Activities */}
      <div className="bg-white dark:bg-secondary-800 rounded-xl border border-secondary-200 dark:border-secondary-700">
        <div className="px-6 py-4 border-b border-secondary-200 dark:border-secondary-700">
          <div className="flex items-center justify-between">
            <h2 className="text-lg font-semibold text-secondary-900 dark:text-white">
              Recent Activities
            </h2>
            <div className="flex items-center space-x-2">
              <Filter className="w-4 h-4 text-secondary-500" />
              <select className="text-sm bg-transparent text-secondary-600 dark:text-secondary-400 focus:outline-none">
                <option>All Activities</option>
                <option>Performance</option>
                <option>System</option>
                <option>Usage</option>
                <option>Errors</option>
              </select>
            </div>
          </div>
        </div>

        <div className="divide-y divide-secondary-200 dark:divide-secondary-700">
          {recentActivities.map((activity, index) => (
            <motion.div
              key={activity.id}
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.3, delay: index * 0.1 }}
              className="px-6 py-4 hover:bg-secondary-50 dark:hover:bg-secondary-700 transition-colors"
            >
              <div className="flex items-start space-x-3">
                <div className={`w-8 h-8 rounded-full flex items-center justify-center mt-1 ${
                  activity.severity === 'info' ? 'bg-primary-100 dark:bg-primary-900/20' :
                  activity.severity === 'warning' ? 'bg-warning-100 dark:bg-warning-900/20' :
                  'bg-error-100 dark:bg-error-900/20'
                }`}>
                  {activity.severity === 'info' ? (
                    <Activity className={`w-4 h-4 ${getSeverityColor(activity.severity)}`} />
                  ) : activity.severity === 'warning' ? (
                    <AlertCircle className={`w-4 h-4 ${getSeverityColor(activity.severity)}`} />
                  ) : (
                    <AlertCircle className={`w-4 h-4 ${getSeverityColor(activity.severity)}`} />
                  )}
                </div>
                <div className="flex-1">
                  <h3 className="font-medium text-secondary-900 dark:text-white">
                    {activity.title}
                  </h3>
                  <p className="text-sm text-secondary-600 dark:text-secondary-300 mt-1">
                    {activity.description}
                  </p>
                  <p className="text-xs text-secondary-500 dark:text-secondary-400 mt-2">
                    {activity.timestamp.toLocaleString()}
                  </p>
                </div>
              </div>
            </motion.div>
          ))}
        </div>
      </div>
    </div>
  )
}