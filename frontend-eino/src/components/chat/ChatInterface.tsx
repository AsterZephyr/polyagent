import React, { useState, useRef, useEffect } from 'react'
import { Send, Bot, User, Loader2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { cn } from '@/lib/utils'
import { useStore } from '@/stores/useStore'
import { apiService } from '@/services/api'
import { Message } from '@/types'
import { generateSessionId } from '@/lib/utils'

interface ChatInterfaceProps {
  className?: string
}

export function ChatInterface({ className }: ChatInterfaceProps) {
  const [message, setMessage] = useState('')
  const [isLoading, setIsLoading] = useState(false)
  const [streamingResponse, setStreamingResponse] = useState('')
  const messagesEndRef = useRef<HTMLDivElement>(null)
  
  const {
    currentSession,
    currentAgent,
    sessions,
    addMessage,
    addSession,
    setCurrentSession,
    addNotification
  } = useStore()

  const currentSessionData = sessions.find(s => s.id === currentSession)
  const messages = currentSessionData?.messages || []

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }

  useEffect(() => {
    scrollToBottom()
  }, [messages, streamingResponse])

  const handleSendMessage = async () => {
    if (!message.trim() || isLoading) return

    const userMessage: Message = {
      id: `msg_${Date.now()}`,
      role: 'user',
      content: message,
      timestamp: new Date()
    }

    let sessionId = currentSession
    if (!sessionId) {
      sessionId = generateSessionId()
      const newSession = {
        id: sessionId,
        name: message.length > 50 ? message.slice(0, 50) + '...' : message,
        agent_id: currentAgent || 'default',
        messages: [],
        created_at: new Date(),
        updated_at: new Date()
      }
      addSession(newSession)
      setCurrentSession(sessionId)
    }

    addMessage(sessionId, userMessage)
    setMessage('')
    setIsLoading(true)
    setStreamingResponse('')

    try {
      let fullResponse = ''
      
      await apiService.streamChat(
        {
          message: userMessage.content,
          session_id: sessionId,
          agent_id: currentAgent || 'default',
          stream: true
        },
        (chunk) => {
          fullResponse += chunk
          setStreamingResponse(fullResponse)
        },
        () => {
          const assistantMessage: Message = {
            id: `msg_${Date.now()}`,
            role: 'assistant',
            content: fullResponse,
            timestamp: new Date()
          }
          addMessage(sessionId, assistantMessage)
          setStreamingResponse('')
          setIsLoading(false)
        },
        (error) => {
          console.error('Stream error:', error)
          addNotification({
            type: 'error',
            title: '发送失败',
            message: error.message
          })
          setIsLoading(false)
          setStreamingResponse('')
        }
      )
    } catch (error) {
      console.error('Chat error:', error)
      addNotification({
        type: 'error',
        title: '发送失败',
        message: error instanceof Error ? error.message : '未知错误'
      })
      setIsLoading(false)
    }
  }

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSendMessage()
    }
  }

  return (
    <div className={cn('flex flex-col h-full', className)}>
      {/* Messages */}
      <div className="flex-1 overflow-y-auto p-4 space-y-4">
        {messages.length === 0 && !streamingResponse && (
          <div className="text-center text-muted-foreground py-8">
            <Bot className="w-16 h-16 mx-auto mb-4 opacity-50" />
            <h3 className="text-lg font-medium mb-2">开始对话</h3>
            <p className="text-sm">向AI助手发送消息开始对话</p>
          </div>
        )}
        
        {messages.map((msg) => (
          <div
            key={msg.id}
            className={cn(
              'flex gap-3 p-4 rounded-lg',
              msg.role === 'user' 
                ? 'bg-primary/10 ml-auto max-w-[80%]' 
                : 'bg-muted max-w-[80%]'
            )}
          >
            <div className="flex-shrink-0">
              {msg.role === 'user' ? (
                <User className="w-6 h-6" />
              ) : (
                <Bot className="w-6 h-6" />
              )}
            </div>
            <div className="flex-1">
              <div className="font-medium text-sm mb-1">
                {msg.role === 'user' ? '您' : 'AI助手'}
              </div>
              <div className="prose prose-sm max-w-none">
                {msg.content}
              </div>
            </div>
          </div>
        ))}

        {streamingResponse && (
          <div className="flex gap-3 p-4 rounded-lg bg-muted max-w-[80%]">
            <div className="flex-shrink-0">
              <Bot className="w-6 h-6" />
            </div>
            <div className="flex-1">
              <div className="font-medium text-sm mb-1">AI助手</div>
              <div className="prose prose-sm max-w-none">
                {streamingResponse}
                <span className="inline-block w-2 h-4 bg-primary animate-pulse ml-1" />
              </div>
            </div>
          </div>
        )}

        <div ref={messagesEndRef} />
      </div>

      {/* Input */}
      <div className="border-t p-4">
        <div className="flex gap-2">
          <Input
            value={message}
            onChange={(e) => setMessage(e.target.value)}
            onKeyPress={handleKeyPress}
            placeholder="输入消息..."
            disabled={isLoading}
            className="flex-1"
          />
          <Button
            onClick={handleSendMessage}
            disabled={!message.trim() || isLoading}
            size="icon"
          >
            {isLoading ? (
              <Loader2 className="w-4 h-4 animate-spin" />
            ) : (
              <Send className="w-4 h-4" />
            )}
          </Button>
        </div>
      </div>
    </div>
  )
}