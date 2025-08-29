import React, { useState } from 'react'
import { motion } from 'framer-motion'
import {
  Upload,
  Search,
  Filter,
  Plus,
  MoreVertical,
  FileText,
  File,
  Globe,
  Database,
  Bot,
  Activity,
  Clock,
  Users,
  CheckCircle,
  AlertCircle,
  Loader2,
  Trash2,
  Edit3,
  Eye,
} from 'lucide-react'

interface KnowledgeItem {
  id: string
  title: string
  description: string
  type: 'document' | 'url' | 'text' | 'database'
  size: string
  status: 'processing' | 'indexed' | 'error'
  chunks: number
  embeddings: number
  last_updated: Date
  created_by: string
  usage_count: number
}

const knowledgeData: KnowledgeItem[] = [
  {
    id: '1',
    title: 'AI Research Papers Collection',
    description: 'Latest research papers on machine learning and AI advancements',
    type: 'document',
    size: '24.5 MB',
    status: 'indexed',
    chunks: 1247,
    embeddings: 1247,
    last_updated: new Date('2024-01-15'),
    created_by: 'Research Team',
    usage_count: 89,
  },
  {
    id: '2',
    title: 'Company Documentation Portal',
    description: 'Internal documentation and knowledge base',
    type: 'url',
    size: '156.2 MB',
    status: 'indexed',
    chunks: 5634,
    embeddings: 5634,
    last_updated: new Date('2024-01-14'),
    created_by: 'Admin',
    usage_count: 234,
  },
  {
    id: '3',
    title: 'Technical Specifications',
    description: 'Product specifications and technical requirements',
    type: 'document',
    size: '12.8 MB',
    status: 'processing',
    chunks: 456,
    embeddings: 423,
    last_updated: new Date('2024-01-16'),
    created_by: 'Engineering',
    usage_count: 45,
  },
  {
    id: '4',
    title: 'Customer Support FAQ',
    description: 'Frequently asked questions and support documentation',
    type: 'text',
    size: '8.3 MB',
    status: 'indexed',
    chunks: 892,
    embeddings: 892,
    last_updated: new Date('2024-01-13'),
    created_by: 'Support Team',
    usage_count: 167,
  },
]

const stats = [
  { label: 'Total Knowledge Items', value: '247', icon: Database, trend: '+12' },
  { label: 'Indexed Chunks', value: '48.2K', icon: FileText, trend: '+2.4K' },
  { label: 'Vector Embeddings', value: '47.8K', icon: Bot, trend: '+2.3K' },
  { label: 'Storage Used', value: '1.2 GB', icon: Activity, trend: '+156 MB' },
]

export default function KnowledgePage() {
  const [searchQuery, setSearchQuery] = useState('')
  const [selectedFilter, setSelectedFilter] = useState('all')
  const [isUploading, setIsUploading] = useState(false)

  const getTypeIcon = (type: string) => {
    switch (type) {
      case 'document':
        return FileText
      case 'url':
        return Globe
      case 'text':
        return File
      case 'database':
        return Database
      default:
        return File
    }
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'indexed':
        return 'bg-success-100 dark:bg-success-900/20 text-success-700 dark:text-success-300'
      case 'processing':
        return 'bg-warning-100 dark:bg-warning-900/20 text-warning-700 dark:text-warning-300'
      case 'error':
        return 'bg-error-100 dark:bg-error-900/20 text-error-700 dark:text-error-300'
      default:
        return 'bg-secondary-100 dark:bg-secondary-700 text-secondary-700 dark:text-secondary-300'
    }
  }

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'indexed':
        return CheckCircle
      case 'processing':
        return Loader2
      case 'error':
        return AlertCircle
      default:
        return Clock
    }
  }

  const handleFileUpload = (event: React.ChangeEvent<HTMLInputElement>) => {
    const files = event.target.files
    if (files && files.length > 0) {
      setIsUploading(true)
      setTimeout(() => setIsUploading(false), 3000)
    }
  }

  const filteredData = knowledgeData.filter(item => {
    const matchesSearch = item.title.toLowerCase().includes(searchQuery.toLowerCase()) ||
                         item.description.toLowerCase().includes(searchQuery.toLowerCase())
    const matchesFilter = selectedFilter === 'all' || item.type === selectedFilter
    return matchesSearch && matchesFilter
  })

  return (
    <div className="p-6">
      {/* Header */}
      <div className="mb-8">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold text-secondary-900 dark:text-white mb-2">
              Knowledge Base
            </h1>
            <p className="text-secondary-600 dark:text-secondary-300">
              Manage your knowledge sources and vector embeddings
            </p>
          </div>
          <div className="flex items-center space-x-3">
            <input
              type="file"
              id="file-upload"
              multiple
              accept=".pdf,.doc,.docx,.txt,.md"
              onChange={handleFileUpload}
              className="hidden"
            />
            <label
              htmlFor="file-upload"
              className="flex items-center space-x-2 px-4 py-2 bg-primary-600 hover:bg-primary-700 text-white rounded-lg transition-colors font-medium cursor-pointer"
            >
              <Upload className="w-5 h-5" />
              <span>Upload Documents</span>
            </label>
            <button className="flex items-center space-x-2 px-4 py-2 border border-secondary-300 dark:border-secondary-600 text-secondary-700 dark:text-secondary-300 hover:bg-secondary-50 dark:hover:bg-secondary-700 rounded-lg transition-colors font-medium">
              <Plus className="w-5 h-5" />
              <span>Add URL</span>
            </button>
          </div>
        </div>
      </div>

      {/* Stats Overview */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
        {stats.map((stat) => (
          <div
            key={stat.label}
            className="bg-white dark:bg-secondary-800 rounded-xl p-6 border border-secondary-200 dark:border-secondary-700"
          >
            <div className="flex items-center justify-between mb-4">
              <div className="w-12 h-12 bg-primary-100 dark:bg-primary-900/20 rounded-xl flex items-center justify-center">
                <stat.icon className="w-6 h-6 text-primary-600 dark:text-primary-400" />
              </div>
              <span className="text-2xl font-bold text-secondary-900 dark:text-white">
                {stat.value}
              </span>
            </div>
            <h3 className="font-medium text-secondary-900 dark:text-white">
              {stat.label}
            </h3>
            <p className="text-sm text-success-600 dark:text-success-400 mt-1">
              {stat.trend} this week
            </p>
          </div>
        ))}
      </div>

      {/* Search and Filter */}
      <div className="mb-6 flex items-center space-x-4">
        <div className="flex-1 relative">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-secondary-400 w-5 h-5" />
          <input
            type="text"
            placeholder="Search knowledge base..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full pl-10 pr-4 py-3 bg-white dark:bg-secondary-800 border border-secondary-200 dark:border-secondary-700 rounded-xl focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent transition-all duration-200 text-secondary-900 dark:text-white placeholder-secondary-500"
          />
        </div>
        <div className="flex items-center space-x-2">
          <Filter className="w-5 h-5 text-secondary-500" />
          <select
            value={selectedFilter}
            onChange={(e) => setSelectedFilter(e.target.value)}
            className="px-4 py-3 bg-white dark:bg-secondary-800 border border-secondary-200 dark:border-secondary-700 rounded-xl focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent transition-all duration-200 text-secondary-900 dark:text-white"
          >
            <option value="all">All Types</option>
            <option value="document">Documents</option>
            <option value="url">URLs</option>
            <option value="text">Text</option>
            <option value="database">Database</option>
          </select>
        </div>
      </div>

      {/* Upload Progress */}
      {isUploading && (
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          className="mb-6 bg-primary-50 dark:bg-primary-900/20 border border-primary-200 dark:border-primary-700 rounded-xl p-4"
        >
          <div className="flex items-center space-x-3">
            <Loader2 className="w-5 h-5 animate-spin text-primary-600" />
            <div>
              <p className="font-medium text-primary-900 dark:text-primary-100">
                Processing uploaded files...
              </p>
              <p className="text-sm text-primary-700 dark:text-primary-300">
                Extracting text, creating chunks, and generating embeddings
              </p>
            </div>
          </div>
        </motion.div>
      )}

      {/* Knowledge Items List */}
      <div className="bg-white dark:bg-secondary-800 rounded-xl border border-secondary-200 dark:border-secondary-700">
        <div className="px-6 py-4 border-b border-secondary-200 dark:border-secondary-700">
          <h2 className="text-lg font-semibold text-secondary-900 dark:text-white">
            Knowledge Sources ({filteredData.length})
          </h2>
        </div>

        <div className="divide-y divide-secondary-200 dark:divide-secondary-700">
          {filteredData.map((item, index) => {
            const TypeIcon = getTypeIcon(item.type)
            const StatusIcon = getStatusIcon(item.status)

            return (
              <motion.div
                key={item.id}
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.3, delay: index * 0.1 }}
                className="p-6 hover:bg-secondary-50 dark:hover:bg-secondary-700 transition-colors"
              >
                <div className="flex items-center justify-between">
                  <div className="flex items-center space-x-4">
                    <div className="w-12 h-12 bg-secondary-100 dark:bg-secondary-600 rounded-xl flex items-center justify-center">
                      <TypeIcon className="w-6 h-6 text-secondary-600 dark:text-secondary-300" />
                    </div>
                    <div className="flex-1">
                      <h3 className="text-lg font-semibold text-secondary-900 dark:text-white">
                        {item.title}
                      </h3>
                      <p className="text-secondary-600 dark:text-secondary-300 text-sm mb-2">
                        {item.description}
                      </p>
                      <div className="flex items-center space-x-4 text-xs text-secondary-500 dark:text-secondary-400">
                        <span className={`inline-flex items-center px-2 py-1 rounded ${getStatusColor(item.status)}`}>
                          <StatusIcon className={`w-3 h-3 mr-1 ${item.status === 'processing' ? 'animate-spin' : ''}`} />
                          {item.status}
                        </span>
                        <span>{item.size}</span>
                        <span>{item.chunks.toLocaleString()} chunks</span>
                        <span>{item.embeddings.toLocaleString()} embeddings</span>
                      </div>
                    </div>
                  </div>

                  <div className="flex items-center space-x-6">
                    <div className="text-right">
                      <div className="text-sm font-medium text-secondary-900 dark:text-white">
                        {item.usage_count} queries
                      </div>
                      <div className="text-xs text-secondary-500 dark:text-secondary-400">
                        Updated {item.last_updated.toLocaleDateString()}
                      </div>
                      <div className="text-xs text-secondary-500 dark:text-secondary-400">
                        by {item.created_by}
                      </div>
                    </div>

                    <div className="flex items-center space-x-2">
                      <button className="p-2 text-secondary-600 dark:text-secondary-400 hover:bg-secondary-200 dark:hover:bg-secondary-600 rounded-lg transition-colors">
                        <Eye className="w-4 h-4" />
                      </button>
                      <button className="p-2 text-secondary-600 dark:text-secondary-400 hover:bg-secondary-200 dark:hover:bg-secondary-600 rounded-lg transition-colors">
                        <Edit3 className="w-4 h-4" />
                      </button>
                      <button className="p-2 text-secondary-600 dark:text-secondary-400 hover:bg-secondary-200 dark:hover:bg-secondary-600 rounded-lg transition-colors">
                        <Trash2 className="w-4 h-4" />
                      </button>
                      <button className="p-2 text-secondary-600 dark:text-secondary-400 hover:bg-secondary-200 dark:hover:bg-secondary-600 rounded-lg transition-colors">
                        <MoreVertical className="w-4 h-4" />
                      </button>
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