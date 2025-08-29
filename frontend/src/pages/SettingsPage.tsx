import React, { useState } from 'react'
import { motion } from 'framer-motion'
import {
  Settings,
  User,
  Key,
  Globe,
  Brain,
  Database,
  Shield,
  Bell,
  Palette,
  Code,
  Save,
  RefreshCw,
  AlertTriangle,
  CheckCircle,
  Info,
  Eye,
  EyeOff,
} from 'lucide-react'

interface SettingSection {
  id: string
  title: string
  description: string
  icon: any
}

const settingSections: SettingSection[] = [
  {
    id: 'profile',
    title: 'Profile & Account',
    description: 'Manage your account information and preferences',
    icon: User,
  },
  {
    id: 'ai-providers',
    title: 'AI Providers',
    description: 'Configure AI model providers and API keys',
    icon: Brain,
  },
  {
    id: 'knowledge',
    title: 'Knowledge Base',
    description: 'Vector database and knowledge graph settings',
    icon: Database,
  },
  {
    id: 'security',
    title: 'Security & Privacy',
    description: 'Authentication and data protection settings',
    icon: Shield,
  },
  {
    id: 'notifications',
    title: 'Notifications',
    description: 'Configure alerts and notification preferences',
    icon: Bell,
  },
  {
    id: 'appearance',
    title: 'Appearance',
    description: 'Theme and display customization',
    icon: Palette,
  },
  {
    id: 'advanced',
    title: 'Advanced',
    description: 'System configuration and developer options',
    icon: Code,
  },
]

const aiProviders = [
  {
    id: 'openai',
    name: 'OpenAI',
    description: 'GPT-4, GPT-3.5 Turbo, and embedding models',
    enabled: true,
    models: ['gpt-4', 'gpt-3.5-turbo', 'text-embedding-ada-002'],
    status: 'connected',
  },
  {
    id: 'anthropic',
    name: 'Anthropic Claude',
    description: 'Claude 3 family models for advanced reasoning',
    enabled: true,
    models: ['claude-3-opus', 'claude-3-sonnet', 'claude-3-haiku'],
    status: 'connected',
  },
  {
    id: 'google',
    name: 'Google AI',
    description: 'Gemini models and PaLM API integration',
    enabled: false,
    models: ['gemini-pro', 'gemini-pro-vision'],
    status: 'disconnected',
  },
]

export default function SettingsPage() {
  const [activeSection, setActiveSection] = useState('profile')
  const [showApiKeys, setShowApiKeys] = useState(false)
  const [isLoading, setIsLoading] = useState(false)

  const handleSave = async () => {
    setIsLoading(true)
    // Simulate save operation
    await new Promise(resolve => setTimeout(resolve, 1500))
    setIsLoading(false)
  }

  const renderProfileSection = () => (
    <div className="space-y-6">
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <div>
          <label className="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-2">
            Full Name
          </label>
          <input
            type="text"
            defaultValue="Demo User"
            className="w-full px-4 py-3 bg-white dark:bg-secondary-800 border border-secondary-200 dark:border-secondary-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent transition-all duration-200 text-secondary-900 dark:text-white"
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-2">
            Email Address
          </label>
          <input
            type="email"
            defaultValue="demo@polyagent.ai"
            className="w-full px-4 py-3 bg-white dark:bg-secondary-800 border border-secondary-200 dark:border-secondary-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent transition-all duration-200 text-secondary-900 dark:text-white"
          />
        </div>
      </div>
      
      <div>
        <label className="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-2">
          Organization
        </label>
        <input
          type="text"
          defaultValue="PolyAgent Team"
          className="w-full px-4 py-3 bg-white dark:bg-secondary-800 border border-secondary-200 dark:border-secondary-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent transition-all duration-200 text-secondary-900 dark:text-white"
        />
      </div>

      <div>
        <label className="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-2">
          Role
        </label>
        <select className="w-full px-4 py-3 bg-white dark:bg-secondary-800 border border-secondary-200 dark:border-secondary-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent transition-all duration-200 text-secondary-900 dark:text-white">
          <option value="admin">Administrator</option>
          <option value="user">User</option>
          <option value="developer">Developer</option>
        </select>
      </div>
    </div>
  )

  const renderAIProvidersSection = () => (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h3 className="text-lg font-medium text-secondary-900 dark:text-white">
          AI Provider Configuration
        </h3>
        <button
          onClick={() => setShowApiKeys(!showApiKeys)}
          className="flex items-center space-x-2 text-sm text-primary-600 dark:text-primary-400 hover:text-primary-700 dark:hover:text-primary-300"
        >
          {showApiKeys ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
          <span>{showApiKeys ? 'Hide' : 'Show'} API Keys</span>
        </button>
      </div>

      {aiProviders.map((provider) => (
        <div
          key={provider.id}
          className="bg-secondary-50 dark:bg-secondary-700 rounded-xl p-6 border border-secondary-200 dark:border-secondary-600"
        >
          <div className="flex items-center justify-between mb-4">
            <div className="flex items-center space-x-3">
              <div className={`w-3 h-3 rounded-full ${
                provider.status === 'connected' ? 'bg-success-500' : 'bg-error-500'
              }`}></div>
              <div>
                <h4 className="font-medium text-secondary-900 dark:text-white">
                  {provider.name}
                </h4>
                <p className="text-sm text-secondary-600 dark:text-secondary-300">
                  {provider.description}
                </p>
              </div>
            </div>
            <label className="relative inline-flex items-center cursor-pointer">
              <input
                type="checkbox"
                defaultChecked={provider.enabled}
                className="sr-only peer"
              />
              <div className="w-11 h-6 bg-secondary-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-primary-300 dark:peer-focus:ring-primary-800 rounded-full peer dark:bg-secondary-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-secondary-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-secondary-600 peer-checked:bg-primary-600"></div>
            </label>
          </div>

          {showApiKeys && (
            <div className="mb-4">
              <label className="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-2">
                API Key
              </label>
              <input
                type="password"
                placeholder="sk-..."
                className="w-full px-4 py-2 bg-white dark:bg-secondary-800 border border-secondary-200 dark:border-secondary-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent transition-all duration-200 text-secondary-900 dark:text-white"
              />
            </div>
          )}

          <div>
            <p className="text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-2">
              Available Models
            </p>
            <div className="flex flex-wrap gap-2">
              {provider.models.map((model) => (
                <span
                  key={model}
                  className="px-3 py-1 bg-primary-100 dark:bg-primary-900/20 text-primary-700 dark:text-primary-300 rounded-full text-sm"
                >
                  {model}
                </span>
              ))}
            </div>
          </div>
        </div>
      ))}
    </div>
  )

  const renderKnowledgeSection = () => (
    <div className="space-y-6">
      <div>
        <h3 className="text-lg font-medium text-secondary-900 dark:text-white mb-4">
          Vector Database Settings
        </h3>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div>
            <label className="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-2">
              Embedding Model
            </label>
            <select className="w-full px-4 py-3 bg-white dark:bg-secondary-800 border border-secondary-200 dark:border-secondary-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent transition-all duration-200 text-secondary-900 dark:text-white">
              <option value="text-embedding-ada-002">OpenAI Ada v2</option>
              <option value="text-embedding-3-small">OpenAI Embedding v3 Small</option>
              <option value="text-embedding-3-large">OpenAI Embedding v3 Large</option>
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-2">
              Vector Dimensions
            </label>
            <input
              type="number"
              defaultValue="1536"
              className="w-full px-4 py-3 bg-white dark:bg-secondary-800 border border-secondary-200 dark:border-secondary-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent transition-all duration-200 text-secondary-900 dark:text-white"
            />
          </div>
        </div>
      </div>

      <div>
        <h3 className="text-lg font-medium text-secondary-900 dark:text-white mb-4">
          Retrieval Settings
        </h3>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          <div>
            <label className="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-2">
              Top K Results
            </label>
            <input
              type="number"
              defaultValue="10"
              className="w-full px-4 py-3 bg-white dark:bg-secondary-800 border border-secondary-200 dark:border-secondary-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent transition-all duration-200 text-secondary-900 dark:text-white"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-2">
              Similarity Threshold
            </label>
            <input
              type="number"
              step="0.01"
              defaultValue="0.75"
              className="w-full px-4 py-3 bg-white dark:bg-secondary-800 border border-secondary-200 dark:border-secondary-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent transition-all duration-200 text-secondary-900 dark:text-white"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-2">
              Rerank Top N
            </label>
            <input
              type="number"
              defaultValue="5"
              className="w-full px-4 py-3 bg-white dark:bg-secondary-800 border border-secondary-200 dark:border-secondary-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent transition-all duration-200 text-secondary-900 dark:text-white"
            />
          </div>
        </div>
      </div>
    </div>
  )

  const renderSecuritySection = () => (
    <div className="space-y-6">
      <div className="bg-warning-50 dark:bg-warning-900/20 border border-warning-200 dark:border-warning-700 rounded-xl p-4">
        <div className="flex items-start space-x-3">
          <AlertTriangle className="w-5 h-5 text-warning-600 dark:text-warning-400 mt-0.5" />
          <div>
            <h4 className="font-medium text-warning-900 dark:text-warning-100">
              Security Notice
            </h4>
            <p className="text-sm text-warning-800 dark:text-warning-200 mt-1">
              Keep your API keys secure and never share them publicly. Rotate keys regularly for optimal security.
            </p>
          </div>
        </div>
      </div>

      <div className="space-y-4">
        <div className="flex items-center justify-between p-4 bg-secondary-50 dark:bg-secondary-700 rounded-lg">
          <div>
            <h4 className="font-medium text-secondary-900 dark:text-white">
              Two-Factor Authentication
            </h4>
            <p className="text-sm text-secondary-600 dark:text-secondary-300">
              Add an extra layer of security to your account
            </p>
          </div>
          <label className="relative inline-flex items-center cursor-pointer">
            <input type="checkbox" className="sr-only peer" />
            <div className="w-11 h-6 bg-secondary-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-primary-300 dark:peer-focus:ring-primary-800 rounded-full peer dark:bg-secondary-600 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-secondary-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-secondary-600 peer-checked:bg-primary-600"></div>
          </label>
        </div>

        <div className="flex items-center justify-between p-4 bg-secondary-50 dark:bg-secondary-700 rounded-lg">
          <div>
            <h4 className="font-medium text-secondary-900 dark:text-white">
              Data Encryption
            </h4>
            <p className="text-sm text-secondary-600 dark:text-secondary-300">
              Encrypt sensitive data at rest and in transit
            </p>
          </div>
          <label className="relative inline-flex items-center cursor-pointer">
            <input type="checkbox" defaultChecked className="sr-only peer" />
            <div className="w-11 h-6 bg-secondary-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-primary-300 dark:peer-focus:ring-primary-800 rounded-full peer dark:bg-secondary-600 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-secondary-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-secondary-600 peer-checked:bg-primary-600"></div>
          </label>
        </div>

        <div className="flex items-center justify-between p-4 bg-secondary-50 dark:bg-secondary-700 rounded-lg">
          <div>
            <h4 className="font-medium text-secondary-900 dark:text-white">
              Audit Logging
            </h4>
            <p className="text-sm text-secondary-600 dark:text-secondary-300">
              Log all system actions for security monitoring
            </p>
          </div>
          <label className="relative inline-flex items-center cursor-pointer">
            <input type="checkbox" defaultChecked className="sr-only peer" />
            <div className="w-11 h-6 bg-secondary-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-primary-300 dark:peer-focus:ring-primary-800 rounded-full peer dark:bg-secondary-600 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-secondary-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-secondary-600 peer-checked:bg-primary-600"></div>
          </label>
        </div>
      </div>
    </div>
  )

  const renderSection = () => {
    switch (activeSection) {
      case 'profile':
        return renderProfileSection()
      case 'ai-providers':
        return renderAIProvidersSection()
      case 'knowledge':
        return renderKnowledgeSection()
      case 'security':
        return renderSecuritySection()
      default:
        return (
          <div className="flex items-center justify-center h-64">
            <div className="text-center">
              <Settings className="w-16 h-16 text-secondary-400 mx-auto mb-4" />
              <p className="text-secondary-600 dark:text-secondary-300">
                Select a setting category to configure
              </p>
            </div>
          </div>
        )
    }
  }

  return (
    <div className="p-6">
      {/* Header */}
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-secondary-900 dark:text-white mb-2">
          Settings
        </h1>
        <p className="text-secondary-600 dark:text-secondary-300">
          Configure your PolyAgent system preferences and integrations
        </p>
      </div>

      <div className="flex gap-8">
        {/* Settings Navigation */}
        <div className="w-64 flex-shrink-0">
          <nav className="space-y-1">
            {settingSections.map((section) => (
              <button
                key={section.id}
                onClick={() => setActiveSection(section.id)}
                className={`w-full text-left flex items-center px-3 py-3 text-sm font-medium rounded-lg transition-all duration-200 ${
                  activeSection === section.id
                    ? 'bg-primary-50 dark:bg-primary-900/20 text-primary-700 dark:text-primary-300 border-l-2 border-primary-500'
                    : 'text-secondary-600 dark:text-secondary-400 hover:bg-secondary-50 dark:hover:bg-secondary-700 hover:text-secondary-900 dark:hover:text-white'
                }`}
              >
                <section.icon className="mr-3 h-5 w-5" />
                <div>
                  <div className="font-medium">{section.title}</div>
                  <div className="text-xs text-secondary-500 dark:text-secondary-400 mt-1">
                    {section.description}
                  </div>
                </div>
              </button>
            ))}
          </nav>
        </div>

        {/* Settings Content */}
        <div className="flex-1">
          <div className="bg-white dark:bg-secondary-800 rounded-xl border border-secondary-200 dark:border-secondary-700 p-8">
            <motion.div
              key={activeSection}
              initial={{ opacity: 0, x: 20 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{ duration: 0.3 }}
            >
              {renderSection()}
            </motion.div>

            {/* Save Button */}
            <div className="mt-8 pt-6 border-t border-secondary-200 dark:border-secondary-700">
              <div className="flex items-center justify-between">
                <div className="flex items-center space-x-2 text-sm text-secondary-600 dark:text-secondary-400">
                  <CheckCircle className="w-4 h-4 text-success-500" />
                  <span>Settings saved automatically</span>
                </div>
                <div className="flex items-center space-x-3">
                  <button className="flex items-center space-x-2 px-4 py-2 text-secondary-600 dark:text-secondary-400 hover:text-secondary-900 dark:hover:text-white border border-secondary-300 dark:border-secondary-600 rounded-lg transition-colors">
                    <RefreshCw className="w-4 h-4" />
                    <span>Reset</span>
                  </button>
                  <button
                    onClick={handleSave}
                    disabled={isLoading}
                    className="flex items-center space-x-2 px-6 py-2 bg-primary-600 hover:bg-primary-700 disabled:bg-primary-400 text-white rounded-lg transition-colors font-medium"
                  >
                    {isLoading ? (
                      <RefreshCw className="w-4 h-4 animate-spin" />
                    ) : (
                      <Save className="w-4 h-4" />
                    )}
                    <span>{isLoading ? 'Saving...' : 'Save Changes'}</span>
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}