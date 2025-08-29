import { useState, useEffect, useCallback } from 'react'
import { apiService } from '../services/api'

interface UseApiState<T> {
  data: T | null
  loading: boolean
  error: string | null
}

interface UseApiOptions {
  immediate?: boolean
  onSuccess?: (data: any) => void
  onError?: (error: string) => void
}

export function useApi<T = any>(
  apiCall: () => Promise<T>,
  dependencies: any[] = [],
  options: UseApiOptions = {}
): UseApiState<T> & { refetch: () => Promise<void> } {
  const [state, setState] = useState<UseApiState<T>>({
    data: null,
    loading: false,
    error: null,
  })

  const { immediate = true, onSuccess, onError } = options

  const fetchData = useCallback(async () => {
    setState(prev => ({ ...prev, loading: true, error: null }))

    try {
      const data = await apiCall()
      setState({ data, loading: false, error: null })
      onSuccess?.(data)
    } catch (error: any) {
      const errorMessage = error.response?.data?.message || error.message || 'An error occurred'
      setState({ data: null, loading: false, error: errorMessage })
      onError?.(errorMessage)
    }
  }, [apiCall, onSuccess, onError])

  useEffect(() => {
    if (immediate) {
      fetchData()
    }
  }, [fetchData, immediate, ...dependencies])

  return {
    ...state,
    refetch: fetchData,
  }
}

export function useChat() {
  const [messages, setMessages] = useState<any[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const sendMessage = useCallback(async (message: string, agentId?: string, mode?: string) => {
    setIsLoading(true)
    setError(null)

    try {
      const userMessage = {
        id: Date.now().toString(),
        role: 'user',
        content: message,
        timestamp: new Date(),
      }

      setMessages(prev => [...prev, userMessage])

      const response = await apiService.sendMessage(message, agentId, mode)
      
      const assistantMessage = {
        id: response.id,
        role: 'assistant',
        content: response.content,
        timestamp: new Date(response.timestamp),
        mode: response.mode,
        reasoning_chain: response.reasoning_chain,
        tools_used: response.tools_used,
        rag_sources: response.rag_sources,
        confidence_score: response.confidence_score,
        processing_time: response.processing_time,
      }

      setMessages(prev => [...prev, assistantMessage])
    } catch (error: any) {
      const errorMessage = error.response?.data?.message || error.message || 'Failed to send message'
      setError(errorMessage)

      const errorMessageObj = {
        id: (Date.now() + 1).toString(),
        role: 'system',
        content: `Error: ${errorMessage}`,
        timestamp: new Date(),
      }

      setMessages(prev => [...prev, errorMessageObj])
    } finally {
      setIsLoading(false)
    }
  }, [])

  const loadChatHistory = useCallback(async (sessionId?: string) => {
    try {
      const history = await apiService.getChatHistory(sessionId)
      setMessages(history.messages || [])
    } catch (error: any) {
      setError(error.response?.data?.message || error.message || 'Failed to load chat history')
    }
  }, [])

  const clearChat = useCallback(() => {
    setMessages([])
    setError(null)
  }, [])

  return {
    messages,
    isLoading,
    error,
    sendMessage,
    loadChatHistory,
    clearChat,
  }
}

export function useAgents() {
  const {
    data: agents,
    loading,
    error,
    refetch
  } = useApi(() => apiService.getAgents())

  const createAgent = useCallback(async (agentData: any) => {
    try {
      const newAgent = await apiService.createAgent(agentData)
      await refetch()
      return newAgent
    } catch (error: any) {
      throw new Error(error.response?.data?.message || error.message || 'Failed to create agent')
    }
  }, [refetch])

  const updateAgent = useCallback(async (agentId: string, agentData: any) => {
    try {
      const updatedAgent = await apiService.updateAgent(agentId, agentData)
      await refetch()
      return updatedAgent
    } catch (error: any) {
      throw new Error(error.response?.data?.message || error.message || 'Failed to update agent')
    }
  }, [refetch])

  const deleteAgent = useCallback(async (agentId: string) => {
    try {
      await apiService.deleteAgent(agentId)
      await refetch()
    } catch (error: any) {
      throw new Error(error.response?.data?.message || error.message || 'Failed to delete agent')
    }
  }, [refetch])

  return {
    agents: agents?.agents || [],
    loading,
    error,
    refetch,
    createAgent,
    updateAgent,
    deleteAgent,
  }
}

export function useKnowledge() {
  const {
    data: knowledgeItems,
    loading,
    error,
    refetch
  } = useApi(() => apiService.getKnowledgeItems())

  const uploadFile = useCallback(async (file: File, metadata?: any) => {
    try {
      const result = await apiService.uploadKnowledgeFile(file, metadata)
      await refetch()
      return result
    } catch (error: any) {
      throw new Error(error.response?.data?.message || error.message || 'Failed to upload file')
    }
  }, [refetch])

  const addUrl = useCallback(async (url: string, metadata?: any) => {
    try {
      const result = await apiService.addKnowledgeUrl(url, metadata)
      await refetch()
      return result
    } catch (error: any) {
      throw new Error(error.response?.data?.message || error.message || 'Failed to add URL')
    }
  }, [refetch])

  const deleteItem = useCallback(async (itemId: string) => {
    try {
      await apiService.deleteKnowledgeItem(itemId)
      await refetch()
    } catch (error: any) {
      throw new Error(error.response?.data?.message || error.message || 'Failed to delete item')
    }
  }, [refetch])

  const searchKnowledge = useCallback(async (query: string, limit?: number) => {
    try {
      return await apiService.searchKnowledge(query, limit)
    } catch (error: any) {
      throw new Error(error.response?.data?.message || error.message || 'Failed to search knowledge')
    }
  }, [])

  return {
    knowledgeItems: knowledgeItems?.items || [],
    loading,
    error,
    refetch,
    uploadFile,
    addUrl,
    deleteItem,
    searchKnowledge,
  }
}

export function useAnalytics(timeRange: string = '30d') {
  const {
    data: stats,
    loading: statsLoading,
    error: statsError,
    refetch: refetchStats
  } = useApi(() => apiService.getSystemStats(timeRange), [timeRange])

  const {
    data: agentAnalytics,
    loading: agentLoading,
    error: agentError,
    refetch: refetchAgents
  } = useApi(() => apiService.getAgentAnalytics(timeRange), [timeRange])

  const {
    data: systemHealth,
    loading: healthLoading,
    error: healthError,
    refetch: refetchHealth
  } = useApi(() => apiService.getSystemHealth())

  const {
    data: activities,
    loading: activitiesLoading,
    error: activitiesError,
    refetch: refetchActivities
  } = useApi(() => apiService.getRecentActivities())

  const loading = statsLoading || agentLoading || healthLoading || activitiesLoading
  const error = statsError || agentError || healthError || activitiesError

  const refetchAll = useCallback(async () => {
    await Promise.all([
      refetchStats(),
      refetchAgents(),
      refetchHealth(),
      refetchActivities(),
    ])
  }, [refetchStats, refetchAgents, refetchHealth, refetchActivities])

  return {
    stats: stats?.stats || {},
    agentAnalytics: agentAnalytics?.agents || [],
    systemHealth: systemHealth?.components || [],
    activities: activities?.activities || [],
    loading,
    error,
    refetchAll,
  }
}

export function useWebSocket(url: string, onMessage?: (data: any) => void) {
  const [socket, setSocket] = useState<WebSocket | null>(null)
  const [isConnected, setIsConnected] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const ws = new WebSocket(url)

    ws.onopen = () => {
      setIsConnected(true)
      setError(null)
    }

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)
        onMessage?.(data)
      } catch (error) {
        console.error('Failed to parse WebSocket message:', error)
      }
    }

    ws.onerror = (error) => {
      setError('WebSocket connection error')
      setIsConnected(false)
    }

    ws.onclose = () => {
      setIsConnected(false)
    }

    setSocket(ws)

    return () => {
      ws.close()
    }
  }, [url, onMessage])

  const sendMessage = useCallback((data: any) => {
    if (socket && isConnected) {
      socket.send(JSON.stringify(data))
    }
  }, [socket, isConnected])

  return {
    socket,
    isConnected,
    error,
    sendMessage,
  }
}