import React, { useState, useRef, useEffect } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { 
  Send, 
  Paperclip, 
  MoreVertical, 
  Bot, 
  User, 
  Brain,
  Zap,
  BookOpen,
  Loader2,
  MessageCircle,
  Plus,
} from 'lucide-react'
import ReactMarkdown from 'react-markdown'
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter'
import { oneDark } from 'react-syntax-highlighter/dist/esm/styles/prism'
import remarkGfm from 'remark-gfm'

interface Message {
  id: string
  role: 'user' | 'assistant' | 'system'
  content: string
  timestamp: Date
  mode?: string
  reasoning_chain?: any
  tools_used?: string[]
  rag_sources?: any[]
  confidence_score?: number
  processing_time?: number
}

const agentModes = [
  { id: 'auto', label: 'Auto', icon: Brain, description: 'Let AI choose the best approach' },
  { id: 'chat', label: 'Chat', icon: MessageCircle, description: 'Simple conversation mode' },
  { id: 'reasoning', label: 'Reasoning', icon: Brain, description: 'Deep analytical thinking' },
  { id: 'rag', label: 'Knowledge', icon: BookOpen, description: 'Search knowledge base' },
  { id: 'tool_using', label: 'Tools', icon: Zap, description: 'Use external tools' },
]

export default function ChatPage() {
  const [messages, setMessages] = useState<Message[]>([
    {
      id: '1',
      role: 'system',
      content: 'Welcome to PolyAgent! I\'m your advanced AI assistant with reasoning, knowledge search, and tool capabilities. How can I help you today?',
      timestamp: new Date(),
    }
  ])
  const [inputMessage, setInputMessage] = useState('')
  const [isLoading, setIsLoading] = useState(false)
  const [selectedMode, setSelectedMode] = useState('auto')
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const inputRef = useRef<HTMLTextAreaElement>(null)

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }

  useEffect(() => {
    scrollToBottom()
  }, [messages])

  const handleSendMessage = async () => {
    if (!inputMessage.trim() || isLoading) return

    const userMessage: Message = {
      id: Date.now().toString(),
      role: 'user',
      content: inputMessage,
      timestamp: new Date(),
    }

    setMessages(prev => [...prev, userMessage])
    setInputMessage('')
    setIsLoading(true)

    try {
      // Simulate API call
      await new Promise(resolve => setTimeout(resolve, 1500))
      
      const assistantMessage: Message = {
        id: (Date.now() + 1).toString(),
        role: 'assistant',
        content: `I understand your question: "${inputMessage}"\n\nThis is a demo response showing how PolyAgent would process your request using the **${selectedMode}** mode. In a real implementation, this would connect to the backend API and provide intelligent responses based on:\n\n- Advanced reasoning capabilities\n- Knowledge base search\n- Tool integrations\n- Multi-provider AI models\n\nThe system would analyze your query, select the best processing approach, and return a comprehensive response with sources and reasoning chains.`,
        timestamp: new Date(),
        mode: selectedMode,
        confidence_score: 0.87,
        processing_time: 1.2,
      }

      setMessages(prev => [...prev, assistantMessage])
    } catch (error) {
      console.error('Error sending message:', error)
      const errorMessage: Message = {
        id: (Date.now() + 2).toString(),
        role: 'system',
        content: 'Sorry, I encountered an error processing your request. Please try again.',
        timestamp: new Date(),
      }
      setMessages(prev => [...prev, errorMessage])
    } finally {
      setIsLoading(false)
    }
  }

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSendMessage()
    }
  }

  const getMessageIcon = (role: string) => {
    switch (role) {
      case 'user':
        return <User className="w-5 h-5" />
      case 'assistant':
        return <Bot className="w-5 h-5" />
      case 'system':
        return <MessageCircle className="w-5 h-5" />
      default:
        return <MessageCircle className="w-5 h-5" />
    }
  }

  const formatTimestamp = (timestamp: Date) => {
    return timestamp.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  }

  return (
    <div className="flex h-full">
      {/* Main Chat Area */}
      <div className="flex-1 flex flex-col">
        {/* Chat Header */}
        <div className="px-6 py-4 border-b border-secondary-200 dark:border-secondary-700 bg-white dark:bg-secondary-800">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-3">
              <div className="w-10 h-10 bg-gradient-primary rounded-lg flex items-center justify-center">
                <Bot className="w-6 h-6 text-white" />
              </div>
              <div>
                <h2 className="text-lg font-semibold text-secondary-900 dark:text-white">
                  PolyAgent Assistant
                </h2>
                <p className="text-sm text-secondary-500 dark:text-secondary-400">
                  Mode: {agentModes.find(m => m.id === selectedMode)?.label} • Online
                </p>
              </div>
            </div>
            <div className="flex items-center space-x-2">
              <button className="p-2 text-secondary-600 dark:text-secondary-400 hover:bg-secondary-100 dark:hover:bg-secondary-700 rounded-lg transition-colors">
                <Plus className="w-5 h-5" />
              </button>
              <button className="p-2 text-secondary-600 dark:text-secondary-400 hover:bg-secondary-100 dark:hover:bg-secondary-700 rounded-lg transition-colors">
                <MoreVertical className="w-5 h-5" />
              </button>
            </div>
          </div>
        </div>

        {/* Messages Area */}
        <div className="flex-1 overflow-y-auto px-6 py-4 space-y-6">
          <AnimatePresence>
            {messages.map((message) => (
              <motion.div
                key={message.id}
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.3 }}
                className={`flex ${message.role === 'user' ? 'justify-end' : 'justify-start'}`}
              >
                <div className={`flex items-start space-x-3 max-w-[85%] ${message.role === 'user' ? 'flex-row-reverse space-x-reverse' : ''}`}>
                  {/* Avatar */}
                  <div className={`flex-shrink-0 w-8 h-8 rounded-full flex items-center justify-center ${
                    message.role === 'user' 
                      ? 'bg-primary-600 text-white' 
                      : message.role === 'system'
                      ? 'bg-secondary-500 text-white'
                      : 'bg-success-600 text-white'
                  }`}>
                    {getMessageIcon(message.role)}
                  </div>

                  {/* Message Content */}
                  <div className={`flex-1 ${message.role === 'user' ? 'text-right' : ''}`}>
                    <div className={`inline-block px-4 py-3 rounded-2xl ${
                      message.role === 'user'
                        ? 'bg-primary-600 text-white'
                        : message.role === 'system'
                        ? 'bg-secondary-100 dark:bg-secondary-700 text-secondary-900 dark:text-white'
                        : 'bg-white dark:bg-secondary-800 border border-secondary-200 dark:border-secondary-700 text-secondary-900 dark:text-white'
                    }`}>
                      {message.role === 'assistant' ? (
                        <ReactMarkdown
                          remarkPlugins={[remarkGfm]}
                          components={{
                            code({ node, inline, className, children, ...props }) {
                              const match = /language-(\w+)/.exec(className || '')
                              return !inline && match ? (
                                <SyntaxHighlighter
                                  style={oneDark}
                                  language={match[1]}
                                  PreTag="div"
                                  className="rounded-lg my-2"
                                  {...props}
                                >
                                  {String(children).replace(/\n$/, '')}
                                </SyntaxHighlighter>
                              ) : (
                                <code className="bg-secondary-200 dark:bg-secondary-600 px-1 py-0.5 rounded text-sm" {...props}>
                                  {children}
                                </code>
                              )
                            }
                          }}
                        >
                          {message.content}
                        </ReactMarkdown>
                      ) : (
                        <p className="whitespace-pre-wrap">{message.content}</p>
                      )}
                    </div>

                    {/* Message Metadata */}
                    <div className={`mt-2 flex items-center text-xs text-secondary-500 dark:text-secondary-400 space-x-4 ${
                      message.role === 'user' ? 'justify-end' : 'justify-start'
                    }`}>
                      <span>{formatTimestamp(message.timestamp)}</span>
                      {message.mode && (
                        <span className="px-2 py-1 bg-secondary-100 dark:bg-secondary-700 rounded">
                          {message.mode}
                        </span>
                      )}
                      {message.confidence_score && (
                        <span className="px-2 py-1 bg-success-100 dark:bg-success-900/20 text-success-700 dark:text-success-300 rounded">
                          {Math.round(message.confidence_score * 100)}% confidence
                        </span>
                      )}
                      {message.processing_time && (
                        <span>{message.processing_time}s</span>
                      )}
                    </div>

                    {/* Additional Info */}
                    {(message.tools_used?.length || message.rag_sources?.length) && (
                      <div className="mt-2 text-xs text-secondary-600 dark:text-secondary-400">
                        {message.tools_used?.length && (
                          <div className="mb-1">
                            <span className="font-medium">Tools used:</span> {message.tools_used.join(', ')}
                          </div>
                        )}
                        {message.rag_sources?.length && (
                          <div>
                            <span className="font-medium">Knowledge sources:</span> {message.rag_sources.length} documents
                          </div>
                        )}
                      </div>
                    )}
                  </div>
                </div>
              </motion.div>
            ))}
          </AnimatePresence>

          {/* Loading Indicator */}
          {isLoading && (
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              className="flex justify-start"
            >
              <div className="flex items-start space-x-3">
                <div className="w-8 h-8 bg-success-600 rounded-full flex items-center justify-center">
                  <Bot className="w-5 h-5 text-white" />
                </div>
                <div className="bg-white dark:bg-secondary-800 border border-secondary-200 dark:border-secondary-700 rounded-2xl px-4 py-3">
                  <div className="flex items-center space-x-2">
                    <Loader2 className="w-4 h-4 animate-spin text-primary-600" />
                    <span className="text-secondary-600 dark:text-secondary-400">Thinking...</span>
                  </div>
                </div>
              </div>
            </motion.div>
          )}

          <div ref={messagesEndRef} />
        </div>

        {/* Input Area */}
        <div className="px-6 py-4 border-t border-secondary-200 dark:border-secondary-700 bg-white dark:bg-secondary-800">
          {/* Mode Selector */}
          <div className="flex items-center space-x-2 mb-4">
            <span className="text-sm font-medium text-secondary-700 dark:text-secondary-300">Mode:</span>
            <div className="flex space-x-1">
              {agentModes.map((mode) => (
                <button
                  key={mode.id}
                  onClick={() => setSelectedMode(mode.id)}
                  className={`flex items-center space-x-2 px-3 py-1.5 rounded-lg text-sm font-medium transition-colors ${
                    selectedMode === mode.id
                      ? 'bg-primary-100 dark:bg-primary-900/20 text-primary-700 dark:text-primary-300'
                      : 'text-secondary-600 dark:text-secondary-400 hover:bg-secondary-100 dark:hover:bg-secondary-700'
                  }`}
                  title={mode.description}
                >
                  <mode.icon className="w-4 h-4" />
                  <span>{mode.label}</span>
                </button>
              ))}
            </div>
          </div>

          {/* Input Field */}
          <div className="flex items-end space-x-3">
            <div className="flex-1">
              <div className="relative">
                <textarea
                  ref={inputRef}
                  value={inputMessage}
                  onChange={(e) => setInputMessage(e.target.value)}
                  onKeyPress={handleKeyPress}
                  placeholder="Type your message... (Press Enter to send, Shift+Enter for new line)"
                  className="w-full px-4 py-3 pr-12 bg-secondary-50 dark:bg-secondary-700 border border-secondary-200 dark:border-secondary-600 rounded-xl focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent resize-none transition-all duration-200 text-secondary-900 dark:text-white placeholder-secondary-500"
                  rows={1}
                  style={{ minHeight: '52px', maxHeight: '120px' }}
                />
                <button className="absolute right-3 top-1/2 transform -translate-y-1/2 p-1.5 text-secondary-400 hover:text-secondary-600 dark:hover:text-secondary-300 rounded-lg hover:bg-secondary-200 dark:hover:bg-secondary-600 transition-colors">
                  <Paperclip className="w-4 h-4" />
                </button>
              </div>
            </div>
            <button
              onClick={handleSendMessage}
              disabled={!inputMessage.trim() || isLoading}
              className="flex items-center justify-center w-12 h-12 bg-primary-600 hover:bg-primary-700 disabled:bg-secondary-300 dark:disabled:bg-secondary-600 text-white rounded-xl transition-colors duration-200 disabled:cursor-not-allowed"
            >
              {isLoading ? (
                <Loader2 className="w-5 h-5 animate-spin" />
              ) : (
                <Send className="w-5 h-5" />
              )}
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}