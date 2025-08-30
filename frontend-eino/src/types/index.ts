export interface Message {
  id: string
  role: 'user' | 'assistant' | 'system'
  content: string
  timestamp: Date
  metadata?: Record<string, any>
}

export interface ChatRequest {
  message: string
  session_id?: string
  agent_id?: string
  stream?: boolean
  metadata?: Record<string, any>
}

export interface ChatResponse {
  response: string
  session_id: string
  agent_id: string
  tokens_used: number
  cost: number
  latency: number
  metadata: Record<string, any>
}

export interface Agent {
  id: string
  name: string
  type: 'conversational' | 'task_oriented' | 'workflow_based'
  system_prompt: string
  model: string
  temperature: number
  max_tokens: number
  tools_enabled: boolean
  memory_enabled: boolean
  safety_filters: string[]
  metadata: Record<string, any>
  created_at?: Date
  updated_at?: Date
}

export interface AgentCreateRequest {
  name: string
  type: string
  system_prompt: string
  model?: string
  temperature?: number
  max_tokens?: number
  tools_enabled?: boolean
  memory_enabled?: boolean
  safety_filters?: string[]
  metadata?: Record<string, any>
}

export interface ModelHealth {
  available: boolean
  latency: number
  error_rate: number
  requests: number
  last_check: Date
  cost_per_1k: number
  priority: number
}

export interface SessionHistory {
  session_id: string
  messages: Message[]
  created_at: Date
  updated_at: Date
}

export interface ChatSession {
  id: string
  name?: string
  agent_id: string
  messages: Message[]
  created_at: Date
  updated_at: Date
}

export interface User {
  id: string
  email: string
  name?: string
  roles: string[]
  created_at: Date
}

export interface ApiError {
  error: string
  code?: string
  details?: Record<string, any>
}