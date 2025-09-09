import axios from 'axios'

const API_BASE_URL = 'http://localhost:8080/api/v1/recommendation'

const api = axios.create({
  baseURL: API_BASE_URL,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
})

export interface DataCollectionRequest {
  collector: string
  timerange: string
  config?: Record<string, any>
}

export interface ModelTrainingRequest {
  algorithm: string
  hyperparameters: Record<string, any>
  training_config?: Record<string, any>
}

export interface PredictionRequest {
  user_id: string
  top_k?: number
  context?: Record<string, any>
}

export interface SystemMetrics {
  total_agents: number
  active_agents: number
  queued_tasks: number
  processing_tasks: number
  total_tasks_today: number
  success_rate_today: number
  average_latency: number
  timestamp: string
}

export interface Agent {
  id: string
  type: string
  status: string
  capabilities: string[]
  performance: {
    tasks_completed: number
    success_rate: number
    average_latency: number
    uptime: string
  }
}

export interface Model {
  id: string
  algorithm: string
  status: string
  created_at: string
  training_time?: number
  metrics?: {
    rmse?: number
    mae?: number
    precision_at_k?: Record<string, number>
    recall_at_k?: Record<string, number>
  }
}

export const recommendationApi = {
  // 数据操作
  async collectData(request: DataCollectionRequest) {
    const response = await api.post('/data/collect', request)
    return response.data
  },

  async extractFeatures(request: { dataset_id: string; features: string[] }) {
    const response = await api.post('/data/features', request)
    return response.data
  },

  async validateData(request: { dataset_id: string }) {
    const response = await api.post('/data/validate', request)
    return response.data
  },

  // 模型操作
  async trainModel(request: ModelTrainingRequest) {
    const response = await api.post('/models/train', request)
    return response.data
  },

  async evaluateModel(request: { model_id: string }) {
    const response = await api.post('/models/evaluate', request)
    return response.data
  },

  async optimizeHyperparameters(request: { algorithm: string; search_space: Record<string, any> }) {
    const response = await api.post('/models/optimize', request)
    return response.data
  },

  async deployModel(request: { model_id: string; deployment_config?: Record<string, any> }) {
    const response = await api.post('/models/deploy', request)
    return response.data
  },

  async getModels(): Promise<Model[]> {
    const response = await api.get('/models')
    return response.data
  },

  async getModel(modelId: string): Promise<Model> {
    const response = await api.get(`/models/${modelId}`)
    return response.data
  },

  // 推荐服务
  async predict(request: PredictionRequest) {
    const response = await api.post('/predict', request)
    return response.data
  },

  async recommend(request: PredictionRequest) {
    const response = await api.post('/recommend', request)
    return response.data
  },

  // 系统监控
  async getAgents(): Promise<Agent[]> {
    const response = await api.get('/agents')
    return response.data
  },

  async getAgentStats(agentId: string) {
    const response = await api.get(`/agents/${agentId}/stats`)
    return response.data
  },

  async getSystemMetrics(): Promise<SystemMetrics> {
    const response = await api.get('/system/metrics')
    return response.data
  },

  async healthCheck() {
    const response = await api.get('/health')
    return response.data
  },
}

export default recommendationApi