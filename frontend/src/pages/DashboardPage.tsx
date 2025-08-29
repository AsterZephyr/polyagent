import React from 'react'
import { Link } from 'react-router-dom'
import { motion } from 'framer-motion'
import {
  MessageCircle,
  Bot,
  BookOpen,
  Activity,
  TrendingUp,
  Users,
  Clock,
  Target,
  Zap,
  Database,
  ArrowRight,
  Plus,
  AlertCircle,
  CheckCircle,
  Brain,
} from 'lucide-react'

const stats = [
  {
    label: 'Total Requests Today',
    value: '1,247',
    change: '+12.5%',
    trend: 'up',
    icon: MessageCircle,
    color: 'text-primary-600 dark:text-primary-400',
    bgColor: 'bg-primary-100 dark:bg-primary-900/20',
  },
  {
    label: 'Active Agents',
    value: '8',
    change: '+2',
    trend: 'up',
    icon: Bot,
    color: 'text-success-600 dark:text-success-400',
    bgColor: 'bg-success-100 dark:bg-success-900/20',
  },
  {
    label: 'Knowledge Items',
    value: '2,456',
    change: '+156',
    trend: 'up',
    icon: BookOpen,
    color: 'text-warning-600 dark:text-warning-400',
    bgColor: 'bg-warning-100 dark:bg-warning-900/20',
  },
  {
    label: 'Success Rate',
    value: '94.8%',
    change: '+2.1%',
    trend: 'up',
    icon: Target,
    color: 'text-error-600 dark:text-error-400',
    bgColor: 'bg-error-100 dark:bg-error-900/20',
  },
]

const quickActions = [
  {
    title: 'Start New Chat',
    description: 'Begin a conversation with AI agents',
    href: '/chat',
    icon: MessageCircle,
    color: 'bg-primary-500',
  },
  {
    title: 'Create Agent',
    description: 'Build a specialized AI agent',
    href: '/agents',
    icon: Bot,
    color: 'bg-success-500',
  },
  {
    title: 'Upload Knowledge',
    description: 'Add documents to knowledge base',
    href: '/knowledge',
    icon: Database,
    color: 'bg-warning-500',
  },
  {
    title: 'View Analytics',
    description: 'Monitor system performance',
    href: '/analytics',
    icon: Activity,
    color: 'bg-error-500',
  },
]

const recentActivities = [
  {
    id: '1',
    type: 'chat',
    title: 'New conversation started',
    description: 'Research Specialist agent answered complex technical question',
    timestamp: '2 minutes ago',
    user: 'Demo User',
  },
  {
    id: '2',
    type: 'knowledge',
    title: 'Knowledge base updated',
    description: 'Added 47 new documents to AI research collection',
    timestamp: '15 minutes ago',
    user: 'System',
  },
  {
    id: '3',
    type: 'agent',
    title: 'Agent performance optimized',
    description: 'Code Assistant response time improved by 23%',
    timestamp: '1 hour ago',
    user: 'Admin',
  },
  {
    id: '4',
    type: 'system',
    title: 'System health check',
    description: 'All services operational, 99.9% uptime maintained',
    timestamp: '2 hours ago',
    user: 'System',
  },
]

const topAgents = [
  {
    id: 'research',
    name: 'Research Specialist',
    requests: 856,
    success_rate: 97.2,
    avg_response: 2.1,
    status: 'active',
  },
  {
    id: 'general',
    name: 'General Assistant',
    requests: 724,
    success_rate: 94.5,
    avg_response: 1.2,
    status: 'active',
  },
  {
    id: 'coder',
    name: 'Code Assistant',
    requests: 423,
    success_rate: 92.8,
    avg_response: 0.9,
    status: 'active',
  },
]

export default function DashboardPage() {
  const getActivityIcon = (type: string) => {
    switch (type) {
      case 'chat':
        return MessageCircle
      case 'knowledge':
        return BookOpen
      case 'agent':
        return Bot
      case 'system':
        return Activity
      default:
        return Activity
    }
  }

  const getActivityColor = (type: string) => {
    switch (type) {
      case 'chat':
        return 'bg-primary-100 dark:bg-primary-900/20 text-primary-600 dark:text-primary-400'
      case 'knowledge':
        return 'bg-warning-100 dark:bg-warning-900/20 text-warning-600 dark:text-warning-400'
      case 'agent':
        return 'bg-success-100 dark:bg-success-900/20 text-success-600 dark:text-success-400'
      case 'system':
        return 'bg-error-100 dark:bg-error-900/20 text-error-600 dark:text-error-400'
      default:
        return 'bg-secondary-100 dark:bg-secondary-700 text-secondary-600 dark:text-secondary-400'
    }
  }

  return (
    <div className="p-6">
      {/* Header */}
      <div className="mb-8">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold text-secondary-900 dark:text-white mb-2">
              Dashboard
            </h1>
            <p className="text-secondary-600 dark:text-secondary-300">
              Welcome back! Here's what's happening with your AI agents today.
            </p>
          </div>
          <div className="flex items-center space-x-3">
            <div className="flex items-center space-x-2 px-3 py-2 bg-success-50 dark:bg-success-900/20 rounded-lg">
              <div className="w-2 h-2 bg-success-500 rounded-full animate-pulse"></div>
              <span className="text-sm font-medium text-success-700 dark:text-success-300">
                All Systems Online
              </span>
            </div>
            <Link
              to="/chat"
              className="flex items-center space-x-2 px-4 py-2 bg-primary-600 hover:bg-primary-700 text-white rounded-lg transition-colors font-medium"
            >
              <Plus className="w-5 h-5" />
              <span>New Chat</span>
            </Link>
          </div>
        </div>
      </div>

      {/* Stats Overview */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
        {stats.map((stat, index) => (
          <motion.div
            key={stat.label}
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.3, delay: index * 0.1 }}
            className="bg-white dark:bg-secondary-800 rounded-xl p-6 border border-secondary-200 dark:border-secondary-700"
          >
            <div className="flex items-center justify-between mb-4">
              <div className={`w-12 h-12 ${stat.bgColor} rounded-xl flex items-center justify-center`}>
                <stat.icon className={`w-6 h-6 ${stat.color}`} />
              </div>
              <div className="flex items-center text-success-600 dark:text-success-400 text-sm font-medium">
                <TrendingUp className="w-4 h-4 mr-1" />
                {stat.change}
              </div>
            </div>
            <h3 className="font-medium text-secondary-900 dark:text-white mb-1">
              {stat.label}
            </h3>
            <p className="text-2xl font-bold text-secondary-900 dark:text-white">
              {stat.value}
            </p>
          </motion.div>
        ))}
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8 mb-8">
        {/* Quick Actions */}
        <div className="lg:col-span-2">
          <div className="bg-white dark:bg-secondary-800 rounded-xl border border-secondary-200 dark:border-secondary-700 p-6">
            <div className="flex items-center justify-between mb-6">
              <h2 className="text-lg font-semibold text-secondary-900 dark:text-white">
                Quick Actions
              </h2>
              <Zap className="w-5 h-5 text-secondary-500" />
            </div>
            
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {quickActions.map((action, index) => (
                <motion.div
                  key={action.title}
                  initial={{ opacity: 0, scale: 0.95 }}
                  animate={{ opacity: 1, scale: 1 }}
                  transition={{ duration: 0.3, delay: index * 0.1 }}
                >
                  <Link
                    to={action.href}
                    className="block p-4 bg-secondary-50 dark:bg-secondary-700 rounded-xl hover:bg-secondary-100 dark:hover:bg-secondary-600 transition-all duration-200 transform hover:scale-105"
                  >
                    <div className="flex items-center space-x-3">
                      <div className={`w-10 h-10 ${action.color} rounded-lg flex items-center justify-center`}>
                        <action.icon className="w-5 h-5 text-white" />
                      </div>
                      <div>
                        <h3 className="font-medium text-secondary-900 dark:text-white">
                          {action.title}
                        </h3>
                        <p className="text-sm text-secondary-600 dark:text-secondary-300">
                          {action.description}
                        </p>
                      </div>
                      <ArrowRight className="w-4 h-4 text-secondary-400 ml-auto" />
                    </div>
                  </Link>
                </motion.div>
              ))}
            </div>
          </div>
        </div>

        {/* Top Performing Agents */}
        <div>
          <div className="bg-white dark:bg-secondary-800 rounded-xl border border-secondary-200 dark:border-secondary-700 p-6">
            <div className="flex items-center justify-between mb-6">
              <h2 className="text-lg font-semibold text-secondary-900 dark:text-white">
                Top Agents
              </h2>
              <Link
                to="/agents"
                className="text-primary-600 dark:text-primary-400 hover:text-primary-700 dark:hover:text-primary-300 text-sm font-medium"
              >
                View All
              </Link>
            </div>
            
            <div className="space-y-4">
              {topAgents.map((agent, index) => (
                <motion.div
                  key={agent.id}
                  initial={{ opacity: 0, x: 20 }}
                  animate={{ opacity: 1, x: 0 }}
                  transition={{ duration: 0.3, delay: index * 0.1 }}
                  className="flex items-center justify-between p-3 bg-secondary-50 dark:bg-secondary-700 rounded-lg"
                >
                  <div className="flex items-center space-x-3">
                    <div className="w-8 h-8 bg-primary-100 dark:bg-primary-900/20 rounded-lg flex items-center justify-center">
                      <Brain className="w-4 h-4 text-primary-600 dark:text-primary-400" />
                    </div>
                    <div>
                      <h3 className="font-medium text-secondary-900 dark:text-white text-sm">
                        {agent.name}
                      </h3>
                      <p className="text-xs text-secondary-500 dark:text-secondary-400">
                        {agent.requests} requests
                      </p>
                    </div>
                  </div>
                  <div className="text-right">
                    <div className="text-sm font-medium text-secondary-900 dark:text-white">
                      {agent.success_rate}%
                    </div>
                    <div className="text-xs text-secondary-500 dark:text-secondary-400">
                      {agent.avg_response}s avg
                    </div>
                  </div>
                </motion.div>
              ))}
            </div>
          </div>
        </div>
      </div>

      {/* Recent Activity */}
      <div className="bg-white dark:bg-secondary-800 rounded-xl border border-secondary-200 dark:border-secondary-700">
        <div className="px-6 py-4 border-b border-secondary-200 dark:border-secondary-700">
          <div className="flex items-center justify-between">
            <h2 className="text-lg font-semibold text-secondary-900 dark:text-white">
              Recent Activity
            </h2>
            <Link
              to="/analytics"
              className="text-primary-600 dark:text-primary-400 hover:text-primary-700 dark:hover:text-primary-300 text-sm font-medium"
            >
              View Analytics
            </Link>
          </div>
        </div>

        <div className="divide-y divide-secondary-200 dark:divide-secondary-700">
          {recentActivities.map((activity, index) => {
            const ActivityIcon = getActivityIcon(activity.type)
            return (
              <motion.div
                key={activity.id}
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.3, delay: index * 0.1 }}
                className="px-6 py-4 hover:bg-secondary-50 dark:hover:bg-secondary-700 transition-colors"
              >
                <div className="flex items-start space-x-3">
                  <div className={`w-8 h-8 rounded-full flex items-center justify-center ${getActivityColor(activity.type)}`}>
                    <ActivityIcon className="w-4 h-4" />
                  </div>
                  <div className="flex-1">
                    <h3 className="font-medium text-secondary-900 dark:text-white">
                      {activity.title}
                    </h3>
                    <p className="text-sm text-secondary-600 dark:text-secondary-300 mt-1">
                      {activity.description}
                    </p>
                    <div className="flex items-center space-x-4 mt-2 text-xs text-secondary-500 dark:text-secondary-400">
                      <span>{activity.timestamp}</span>
                      <span>•</span>
                      <span>{activity.user}</span>
                    </div>
                  </div>
                </div>
              </motion.div>
            )
          })}
        </div>
      </div>
    </div>
  )
}