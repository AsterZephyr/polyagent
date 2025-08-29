import React from 'react'
import { Link } from 'react-router-dom'
import { motion } from 'framer-motion'
import {
  ArrowRight,
  Brain,
  Zap,
  Shield,
  Globe,
  MessageCircle,
  BookOpen,
  Bot,
  Activity,
  Users,
  Database,
  Search,
  Layers,
  Target,
} from 'lucide-react'
import { useTheme } from '../hooks/useTheme'

const features = [
  {
    icon: Brain,
    title: 'Advanced Reasoning',
    description: 'Chain-of-Thought, Plan-Execute, and Self-Reflection reasoning patterns for complex problem solving.',
  },
  {
    icon: Database,
    title: 'Hybrid RAG System',
    description: 'Vector + Graph + Keyword retrieval with semantic reranking and query expansion.',
  },
  {
    icon: Bot,
    title: 'Multi-Agent Support',
    description: 'Create and manage multiple specialized AI agents for different tasks and domains.',
  },
  {
    icon: Zap,
    title: 'Tool Integration',
    description: 'Seamless integration with external tools and APIs for enhanced capabilities.',
  },
  {
    icon: Layers,
    title: 'Memory Management',
    description: 'Advanced memory systems with short-term, long-term, and semantic memory.',
  },
  {
    icon: Globe,
    title: 'Multi-Provider AI',
    description: 'Support for OpenAI, Anthropic Claude, and other leading AI providers.',
  },
]

const stats = [
  { label: 'AI Providers', value: '5+', icon: Globe },
  { label: 'Reasoning Modes', value: '6', icon: Brain },
  { label: 'Tool Integrations', value: '20+', icon: Zap },
  { label: 'Response Accuracy', value: '95%', icon: Target },
]

const quickActions = [
  {
    title: 'Start Chat',
    description: 'Begin a conversation with our advanced AI agents',
    href: '/chat',
    icon: MessageCircle,
    color: 'bg-primary-500',
  },
  {
    title: 'Manage Agents',
    description: 'Create and configure AI agents for specific tasks',
    href: '/agents',
    icon: Bot,
    color: 'bg-success-500',
  },
  {
    title: 'Knowledge Base',
    description: 'Upload and manage your knowledge documents',
    href: '/knowledge',
    icon: BookOpen,
    color: 'bg-warning-500',
  },
  {
    title: 'Analytics',
    description: 'Monitor performance and system insights',
    href: '/analytics',
    icon: Activity,
    color: 'bg-error-500',
  },
]

export default function HomePage() {
  const { toggleTheme, isDark } = useTheme()

  return (
    <div className="min-h-screen bg-white dark:bg-secondary-900">
      {/* Navigation */}
      <nav className="px-6 py-4">
        <div className="flex items-center justify-between max-w-7xl mx-auto">
          <div className="flex items-center space-x-3">
            <div className="w-10 h-10 bg-gradient-primary rounded-xl flex items-center justify-center">
              <Brain className="w-6 h-6 text-white" />
            </div>
            <div>
              <h1 className="text-xl font-bold text-secondary-900 dark:text-white">
                PolyAgent
              </h1>
              <p className="text-xs text-secondary-500 dark:text-secondary-400">
                AI Platform
              </p>
            </div>
          </div>
          <div className="flex items-center space-x-4">
            <button
              onClick={toggleTheme}
              className="p-2 text-secondary-600 dark:text-secondary-400 hover:bg-secondary-100 dark:hover:bg-secondary-700 rounded-lg transition-colors"
            >
              {isDark ? '☀️' : '🌙'}
            </button>
            <Link
              to="/chat"
              className="btn-primary"
            >
              Get Started
            </Link>
          </div>
        </div>
      </nav>

      {/* Hero Section */}
      <section className="px-6 py-20">
        <div className="max-w-7xl mx-auto">
          <motion.div
            initial={{ opacity: 0, y: 30 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.8 }}
            className="text-center"
          >
            <h1 className="text-5xl lg:text-7xl font-bold text-secondary-900 dark:text-white mb-6">
              Next-Generation
              <br />
              <span className="text-gradient">AI Agent Platform</span>
            </h1>
            <p className="text-xl text-secondary-600 dark:text-secondary-300 max-w-3xl mx-auto mb-12">
              Experience the future of AI with advanced reasoning, hybrid RAG systems, 
              and intelligent memory management. Built for enterprise-scale applications.
            </p>
            <div className="flex flex-col sm:flex-row items-center justify-center space-y-4 sm:space-y-0 sm:space-x-6">
              <Link
                to="/chat"
                className="inline-flex items-center px-8 py-4 bg-primary-600 hover:bg-primary-700 text-white font-semibold rounded-xl transition-colors duration-200 shadow-lg hover:shadow-xl"
              >
                Start Chat
                <ArrowRight className="ml-2 w-5 h-5" />
              </Link>
              <Link
                to="/agents"
                className="inline-flex items-center px-8 py-4 border-2 border-secondary-300 dark:border-secondary-600 text-secondary-900 dark:text-white hover:bg-secondary-50 dark:hover:bg-secondary-800 font-semibold rounded-xl transition-colors duration-200"
              >
                Explore Agents
              </Link>
            </div>
          </motion.div>

          {/* Stats */}
          <motion.div
            initial={{ opacity: 0, y: 30 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.8, delay: 0.2 }}
            className="grid grid-cols-2 lg:grid-cols-4 gap-8 mt-20"
          >
            {stats.map((stat, index) => (
              <div key={stat.label} className="text-center">
                <div className="flex items-center justify-center w-12 h-12 bg-primary-100 dark:bg-primary-900/20 rounded-xl mx-auto mb-4">
                  <stat.icon className="w-6 h-6 text-primary-600 dark:text-primary-400" />
                </div>
                <div className="text-3xl font-bold text-secondary-900 dark:text-white mb-2">
                  {stat.value}
                </div>
                <div className="text-secondary-600 dark:text-secondary-400">
                  {stat.label}
                </div>
              </div>
            ))}
          </motion.div>
        </div>
      </section>

      {/* Features Section */}
      <section className="px-6 py-20 bg-secondary-50 dark:bg-secondary-800">
        <div className="max-w-7xl mx-auto">
          <motion.div
            initial={{ opacity: 0, y: 30 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.8 }}
            className="text-center mb-16"
          >
            <h2 className="text-4xl font-bold text-secondary-900 dark:text-white mb-4">
              Powerful Features
            </h2>
            <p className="text-xl text-secondary-600 dark:text-secondary-300 max-w-2xl mx-auto">
              Built with cutting-edge AI technologies for maximum performance and reliability
            </p>
          </motion.div>

          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
            {features.map((feature, index) => (
              <motion.div
                key={feature.title}
                initial={{ opacity: 0, y: 30 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.8, delay: index * 0.1 }}
                className="bg-white dark:bg-secondary-700 rounded-2xl p-8 shadow-sm hover:shadow-lg transition-shadow duration-300"
              >
                <div className="flex items-center justify-center w-16 h-16 bg-primary-100 dark:bg-primary-900/20 rounded-2xl mb-6">
                  <feature.icon className="w-8 h-8 text-primary-600 dark:text-primary-400" />
                </div>
                <h3 className="text-xl font-semibold text-secondary-900 dark:text-white mb-4">
                  {feature.title}
                </h3>
                <p className="text-secondary-600 dark:text-secondary-300">
                  {feature.description}
                </p>
              </motion.div>
            ))}
          </div>
        </div>
      </section>

      {/* Quick Actions */}
      <section className="px-6 py-20">
        <div className="max-w-7xl mx-auto">
          <motion.div
            initial={{ opacity: 0, y: 30 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.8 }}
            className="text-center mb-16"
          >
            <h2 className="text-4xl font-bold text-secondary-900 dark:text-white mb-4">
              Get Started Quickly
            </h2>
            <p className="text-xl text-secondary-600 dark:text-secondary-300">
              Jump into the most common tasks with one click
            </p>
          </motion.div>

          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
            {quickActions.map((action, index) => (
              <motion.div
                key={action.title}
                initial={{ opacity: 0, y: 30 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.8, delay: index * 0.1 }}
              >
                <Link
                  to={action.href}
                  className="block bg-white dark:bg-secondary-700 rounded-2xl p-6 shadow-sm hover:shadow-lg transition-all duration-300 transform hover:-translate-y-1"
                >
                  <div className={`inline-flex items-center justify-center w-12 h-12 ${action.color} rounded-xl mb-4`}>
                    <action.icon className="w-6 h-6 text-white" />
                  </div>
                  <h3 className="text-lg font-semibold text-secondary-900 dark:text-white mb-2">
                    {action.title}
                  </h3>
                  <p className="text-secondary-600 dark:text-secondary-300 text-sm">
                    {action.description}
                  </p>
                </Link>
              </motion.div>
            ))}
          </div>
        </div>
      </section>

      {/* Footer */}
      <footer className="px-6 py-12 bg-secondary-900 dark:bg-black">
        <div className="max-w-7xl mx-auto text-center">
          <div className="flex items-center justify-center space-x-3 mb-6">
            <div className="w-8 h-8 bg-gradient-primary rounded-lg flex items-center justify-center">
              <Brain className="w-5 h-5 text-white" />
            </div>
            <span className="text-xl font-bold text-white">PolyAgent</span>
          </div>
          <p className="text-secondary-400 mb-6">
            Next-Generation AI Agent Platform
          </p>
          <div className="flex items-center justify-center space-x-6 text-sm text-secondary-500">
            <span>Built with precision engineering</span>
            <span>•</span>
            <span>Enterprise Ready</span>
            <span>•</span>
            <span>Open Source</span>
          </div>
        </div>
      </footer>
    </div>
  )
}