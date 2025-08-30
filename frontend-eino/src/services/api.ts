import axios, { AxiosInstance, AxiosResponse } from 'axios'
import { Agent, AgentCreateRequest, ChatRequest, ChatResponse, ModelHealth, SessionHistory, ApiError } from '@/types'

class ApiService {
  private client: AxiosInstance

  constructor() {
    this.client = axios.create({
      baseURL: '/api/v1',
      timeout: 30000,
      headers: {
        'Content-Type': 'application/json',
      },
    })

    this.client.interceptors.request.use(
      (config) => {
        const token = localStorage.getItem('auth_token')
        if (token) {
          config.headers.Authorization = `Bearer ${token}`
        }
        return config
      },
      (error) => {
        return Promise.reject(error)
      }
    )

    this.client.interceptors.response.use(
      (response) => response,
      (error) => {
        if (error.response?.status === 401) {
          localStorage.removeItem('auth_token')
          window.location.href = '/login'
        }
        return Promise.reject(error)
      }
    )
  }

  // Health endpoints
  async getHealth(): Promise<any> {
    const response = await this.client.get('/health')
    return response.data
  }

  async getModels(): Promise<Record<string, ModelHealth>> {
    const response = await this.client.get('/models')
    return response.data.models
  }

  // Chat endpoints
  async sendMessage(request: ChatRequest): Promise<ChatResponse> {
    const response = await this.client.post('/chat', request)
    return response.data
  }

  async streamChat(
    request: ChatRequest,
    onMessage: (chunk: string) => void,
    onComplete: () => void,
    onError: (error: Error) => void
  ): Promise<void> {
    try {
      const response = await fetch('/api/v1/chat/stream', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('auth_token')}`,
        },
        body: JSON.stringify(request),
      })

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }

      const reader = response.body?.getReader()
      if (!reader) {
        throw new Error('No reader available')
      }

      const decoder = new TextDecoder()
      let buffer = ''

      while (true) {
        const { done, value } = await reader.read()
        
        if (done) break

        buffer += decoder.decode(value, { stream: true })
        const lines = buffer.split('\n')
        buffer = lines.pop() || ''

        for (const line of lines) {
          if (line.trim() === '') continue
          
          if (line.startsWith('data: ')) {
            const data = line.slice(6).trim()
            if (data && data !== '[DONE]') {
              onMessage(data)
            }
          }
        }
      }

      onComplete()
    } catch (error) {
      onError(error instanceof Error ? error : new Error('Stream error'))
    }
  }

  // Agent endpoints
  async createAgent(request: AgentCreateRequest): Promise<{ agent_id: string; config: Agent }> {
    const response = await this.client.post('/agents', request)
    return response.data
  }

  async getAgents(): Promise<Record<string, Agent>> {
    const response = await this.client.get('/agents')
    return response.data.agents
  }

  async getAgent(id: string): Promise<Agent> {
    const response = await this.client.get(`/agents/${id}`)
    return response.data.agent
  }

  async deleteAgent(id: string): Promise<void> {
    await this.client.delete(`/agents/${id}`)
  }

  // Session endpoints
  async getSessionHistory(sessionId: string, agentId: string): Promise<SessionHistory> {
    const response = await this.client.get(`/sessions/${sessionId}/history`, {
      params: { agent_id: agentId }
    })
    return response.data
  }

  // Utility methods
  setAuthToken(token: string): void {
    localStorage.setItem('auth_token', token)
  }

  clearAuthToken(): void {
    localStorage.removeItem('auth_token')
  }

  isAuthenticated(): boolean {
    return !!localStorage.getItem('auth_token')
  }
}

export const apiService = new ApiService()
export default apiService