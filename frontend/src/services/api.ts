import axios, { AxiosInstance, AxiosResponse } from 'axios'

const API_BASE_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080'

class ApiService {
  private client: AxiosInstance

  constructor() {
    this.client = axios.create({
      baseURL: API_BASE_URL,
      timeout: 30000,
      headers: {
        'Content-Type': 'application/json',
      },
    })

    this.setupInterceptors()
  }

  private setupInterceptors() {
    this.client.interceptors.request.use(
      (config) => {
        const token = localStorage.getItem('auth_token')
        if (token) {
          config.headers.Authorization = `Bearer ${token}`
        }
        return config
      },
      (error) => Promise.reject(error)
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

  // Chat & Agent APIs
  async sendMessage(message: string, agentId?: string, mode?: string): Promise<any> {
    const response = await this.client.post('/api/v1/chat/message', {
      message,
      agent_id: agentId,
      mode,
    })
    return response.data
  }

  async getChatHistory(sessionId?: string): Promise<any> {
    const response = await this.client.get(`/api/v1/chat/history${sessionId ? `?session_id=${sessionId}` : ''}`)
    return response.data
  }

  async createChatSession(): Promise<any> {
    const response = await this.client.post('/api/v1/chat/session')
    return response.data
  }

  async deleteChatSession(sessionId: string): Promise<void> {
    await this.client.delete(`/api/v1/chat/session/${sessionId}`)
  }

  // Agent Management APIs
  async getAgents(): Promise<any> {
    const response = await this.client.get('/api/v1/agents')
    return response.data
  }

  async getAgent(agentId: string): Promise<any> {
    const response = await this.client.get(`/api/v1/agents/${agentId}`)
    return response.data
  }

  async createAgent(agentData: any): Promise<any> {
    const response = await this.client.post('/api/v1/agents', agentData)
    return response.data
  }

  async updateAgent(agentId: string, agentData: any): Promise<any> {
    const response = await this.client.put(`/api/v1/agents/${agentId}`, agentData)
    return response.data
  }

  async deleteAgent(agentId: string): Promise<void> {
    await this.client.delete(`/api/v1/agents/${agentId}`)
  }

  async getAgentPerformance(agentId: string, timeRange?: string): Promise<any> {
    const response = await this.client.get(`/api/v1/agents/${agentId}/performance${timeRange ? `?range=${timeRange}` : ''}`)
    return response.data
  }

  // Knowledge Base APIs
  async getKnowledgeItems(filter?: string): Promise<any> {
    const response = await this.client.get(`/api/v1/knowledge${filter ? `?filter=${filter}` : ''}`)
    return response.data
  }

  async uploadKnowledgeFile(file: File, metadata?: any): Promise<any> {
    const formData = new FormData()
    formData.append('file', file)
    if (metadata) {
      formData.append('metadata', JSON.stringify(metadata))
    }

    const response = await this.client.post('/api/v1/knowledge/upload', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    })
    return response.data
  }

  async addKnowledgeUrl(url: string, metadata?: any): Promise<any> {
    const response = await this.client.post('/api/v1/knowledge/url', {
      url,
      metadata,
    })
    return response.data
  }

  async deleteKnowledgeItem(itemId: string): Promise<void> {
    await this.client.delete(`/api/v1/knowledge/${itemId}`)
  }

  async getKnowledgeItemDetails(itemId: string): Promise<any> {
    const response = await this.client.get(`/api/v1/knowledge/${itemId}`)
    return response.data
  }

  async searchKnowledge(query: string, limit?: number): Promise<any> {
    const response = await this.client.post('/api/v1/knowledge/search', {
      query,
      limit,
    })
    return response.data
  }

  // Analytics APIs
  async getSystemStats(timeRange?: string): Promise<any> {
    const response = await this.client.get(`/api/v1/analytics/stats${timeRange ? `?range=${timeRange}` : ''}`)
    return response.data
  }

  async getAgentAnalytics(timeRange?: string): Promise<any> {
    const response = await this.client.get(`/api/v1/analytics/agents${timeRange ? `?range=${timeRange}` : ''}`)
    return response.data
  }

  async getSystemHealth(): Promise<any> {
    const response = await this.client.get('/api/v1/system/health')
    return response.data
  }

  async getRecentActivities(limit?: number): Promise<any> {
    const response = await this.client.get(`/api/v1/analytics/activities${limit ? `?limit=${limit}` : ''}`)
    return response.data
  }

  // User Management APIs
  async getCurrentUser(): Promise<any> {
    const response = await this.client.get('/api/v1/user/profile')
    return response.data
  }

  async updateUserProfile(userData: any): Promise<any> {
    const response = await this.client.put('/api/v1/user/profile', userData)
    return response.data
  }

  async getUsers(): Promise<any> {
    const response = await this.client.get('/api/v1/users')
    return response.data
  }

  async createUser(userData: any): Promise<any> {
    const response = await this.client.post('/api/v1/users', userData)
    return response.data
  }

  async updateUser(userId: string, userData: any): Promise<any> {
    const response = await this.client.put(`/api/v1/users/${userId}`, userData)
    return response.data
  }

  async deleteUser(userId: string): Promise<void> {
    await this.client.delete(`/api/v1/users/${userId}`)
  }

  // Settings APIs
  async getSystemSettings(): Promise<any> {
    const response = await this.client.get('/api/v1/settings')
    return response.data
  }

  async updateSystemSettings(settings: any): Promise<any> {
    const response = await this.client.put('/api/v1/settings', settings)
    return response.data
  }

  async testAIProvider(provider: string, apiKey: string): Promise<any> {
    const response = await this.client.post('/api/v1/settings/test-provider', {
      provider,
      api_key: apiKey,
    })
    return response.data
  }

  // WebSocket connection for real-time chat
  createChatWebSocket(sessionId: string): WebSocket {
    const wsUrl = API_BASE_URL.replace('http', 'ws') + `/ws/chat/${sessionId}`
    return new WebSocket(wsUrl)
  }

  // File download helper
  async downloadFile(url: string, filename: string): Promise<void> {
    const response = await this.client.get(url, {
      responseType: 'blob',
    })

    const blob = new Blob([response.data])
    const downloadUrl = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = downloadUrl
    link.download = filename
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(downloadUrl)
  }
}

export const apiService = new ApiService()
export default apiService