import React from 'react'
import { motion } from 'framer-motion'
import { 
  Bot, 
  Plus, 
  Settings, 
  Activity, 
  Brain,
  MessageCircle,
  BookOpen,
  Zap,
  Users,
  MoreVertical,
} from 'lucide-react'

const agents = [
  {
    id: 'default',
    name: 'General Assistant',
    description: 'Multi-purpose AI assistant with advanced reasoning capabilities',
    mode: 'Auto',
    status: 'active',
    requests: 1247,
    success_rate: 94.5,
    avg_response: 1.2,
    icon: Brain,
    color: 'bg-primary-500',
  },
  {
    id: 'research',
    name: 'Research Specialist',
    description: 'Specialized in knowledge search and analysis',
    mode: 'RAG',
    status: 'active',
    requests: 856,
    success_rate: 97.2,
    avg_response: 2.1,
    icon: BookOpen,
    color: 'bg-success-500',
  },
  {
    id: 'coder',
    name: 'Code Assistant',
    description: 'Expert in programming and software development',
    mode: 'Tools',
    status: 'active',
    requests: 423,
    success_rate: 92.8,
    avg_response: 0.9,
    icon: Zap,
    color: 'bg-warning-500',
  },
  {
    id: 'analyst',
    name: 'Data Analyst',
    description: 'Advanced analytics and reasoning for complex problems',
    mode: 'Reasoning',
    status: 'idle',
    requests: 234,
    success_rate: 96.1,
    avg_response: 3.2,
    icon: Activity,
    color: 'bg-error-500',
  },
]

export default function AgentsPage() {
  return (
    <div className="p-6">
      {/* Header */}
      <div className="mb-8">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold text-secondary-900 dark:text-white mb-2">
              AI Agents
            </h1>
            <p className="text-secondary-600 dark:text-secondary-300">
              Manage and monitor your AI agents
            </p>
          </div>
          <button className="flex items-center space-x-2 px-4 py-2 bg-primary-600 hover:bg-primary-700 text-white rounded-lg transition-colors font-medium">
            <Plus className="w-5 h-5" />
            <span>Create Agent</span>
          </button>
        </div>
      </div>

      {/* Stats Overview */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
        <div className="bg-white dark:bg-secondary-800 rounded-xl p-6 border border-secondary-200 dark:border-secondary-700">
          <div className="flex items-center justify-between mb-4">
            <div className="w-12 h-12 bg-primary-100 dark:bg-primary-900/20 rounded-xl flex items-center justify-center">
              <Bot className="w-6 h-6 text-primary-600 dark:text-primary-400" />
            </div>
            <span className="text-2xl font-bold text-secondary-900 dark:text-white">
              {agents.length}
            </span>
          </div>
          <h3 className="font-medium text-secondary-900 dark:text-white">Total Agents</h3>
          <p className="text-sm text-secondary-500 dark:text-secondary-400 mt-1">
            {agents.filter(a => a.status === 'active').length} active
          </p>
        </div>

        <div className="bg-white dark:bg-secondary-800 rounded-xl p-6 border border-secondary-200 dark:border-secondary-700">
          <div className="flex items-center justify-between mb-4">
            <div className="w-12 h-12 bg-success-100 dark:bg-success-900/20 rounded-xl flex items-center justify-center">
              <MessageCircle className="w-6 h-6 text-success-600 dark:text-success-400" />
            </div>
            <span className="text-2xl font-bold text-secondary-900 dark:text-white">
              2,760
            </span>
          </div>
          <h3 className="font-medium text-secondary-900 dark:text-white">Total Requests</h3>
          <p className="text-sm text-secondary-500 dark:text-secondary-400 mt-1">
            +12% from last week
          </p>
        </div>

        <div className="bg-white dark:bg-secondary-800 rounded-xl p-6 border border-secondary-200 dark:border-secondary-700">
          <div className="flex items-center justify-between mb-4">
            <div className="w-12 h-12 bg-warning-100 dark:bg-warning-900/20 rounded-xl flex items-center justify-center">
              <Activity className="w-6 h-6 text-warning-600 dark:text-warning-400" />
            </div>
            <span className="text-2xl font-bold text-secondary-900 dark:text-white">
              95.2%
            </span>
          </div>
          <h3 className="font-medium text-secondary-900 dark:text-white">Success Rate</h3>
          <p className="text-sm text-secondary-500 dark:text-secondary-400 mt-1">
            Above target
          </p>
        </div>

        <div className="bg-white dark:bg-secondary-800 rounded-xl p-6 border border-secondary-200 dark:border-secondary-700">
          <div className="flex items-center justify-between mb-4">
            <div className="w-12 h-12 bg-error-100 dark:bg-error-900/20 rounded-xl flex items-center justify-center">
              <Zap className="w-6 h-6 text-error-600 dark:text-error-400" />
            </div>
            <span className="text-2xl font-bold text-secondary-900 dark:text-white">
              1.6s
            </span>
          </div>
          <h3 className="font-medium text-secondary-900 dark:text-white">Avg Response</h3>
          <p className="text-sm text-secondary-500 dark:text-secondary-400 mt-1">
            -0.2s improvement
          </p>
        </div>
      </div>

      {/* Agents List */}
      <div className="bg-white dark:bg-secondary-800 rounded-xl border border-secondary-200 dark:border-secondary-700">
        <div className="px-6 py-4 border-b border-secondary-200 dark:border-secondary-700">
          <h2 className="text-lg font-semibold text-secondary-900 dark:text-white">
            Active Agents
          </h2>
        </div>

        <div className="p-6">
          <div className="grid gap-6">
            {agents.map((agent, index) => (
              <motion.div
                key={agent.id}
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.3, delay: index * 0.1 }}
                className="flex items-center justify-between p-6 bg-secondary-50 dark:bg-secondary-700 rounded-xl hover:bg-secondary-100 dark:hover:bg-secondary-600 transition-colors"
              >
                <div className="flex items-center space-x-4">
                  <div className={`w-12 h-12 ${agent.color} rounded-xl flex items-center justify-center`}>
                    <agent.icon className="w-6 h-6 text-white" />
                  </div>
                  <div>
                    <h3 className="text-lg font-semibold text-secondary-900 dark:text-white">
                      {agent.name}
                    </h3>
                    <p className="text-secondary-600 dark:text-secondary-300 text-sm">
                      {agent.description}
                    </p>
                    <div className="flex items-center space-x-4 mt-2">
                      <span className="text-xs px-2 py-1 bg-primary-100 dark:bg-primary-900/20 text-primary-700 dark:text-primary-300 rounded">
                        {agent.mode}
                      </span>
                      <span className={`text-xs px-2 py-1 rounded ${
                        agent.status === 'active' 
                          ? 'bg-success-100 dark:bg-success-900/20 text-success-700 dark:text-success-300'
                          : 'bg-secondary-100 dark:bg-secondary-600 text-secondary-700 dark:text-secondary-300'
                      }`}>
                        {agent.status}
                      </span>
                    </div>
                  </div>
                </div>

                <div className="flex items-center space-x-8">
                  <div className="text-right">
                    <div className="text-sm font-medium text-secondary-900 dark:text-white">
                      {agent.requests.toLocaleString()} requests
                    </div>
                    <div className="text-xs text-secondary-500 dark:text-secondary-400">
                      {agent.success_rate}% success • {agent.avg_response}s avg
                    </div>
                  </div>

                  <div className="flex items-center space-x-2">
                    <button className="p-2 text-secondary-600 dark:text-secondary-400 hover:bg-secondary-200 dark:hover:bg-secondary-600 rounded-lg transition-colors">
                      <Settings className="w-4 h-4" />
                    </button>
                    <button className="p-2 text-secondary-600 dark:text-secondary-400 hover:bg-secondary-200 dark:hover:bg-secondary-600 rounded-lg transition-colors">
                      <MoreVertical className="w-4 h-4" />
                    </button>
                  </div>
                </div>
              </motion.div>
            ))}
          </div>
        </div>
      </div>
    </div>
  )
}