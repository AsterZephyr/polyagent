export interface User {
  id: string
  name: string
  email: string
  role: 'admin' | 'user' | 'developer'
  organization?: string
  created_at: string
  updated_at: string
}

export interface Agent {
  id: string
  name: string
  description: string
  type: 'general' | 'research' | 'code' | 'analyst' | 'custom'
  mode: 'auto' | 'chat' | 'reasoning' | 'rag' | 'tool_using'
  status: 'active' | 'inactive' | 'training'
  config: {
    model_provider: string
    model_name: string
    temperature: number
    max_tokens: number
    system_prompt?: string
    tools?: string[]
    knowledge_sources?: string[]
  }
  performance: {
    requests: number
    success_rate: number
    avg_response_time: number
    error_rate: number
  }
  created_by: string
  created_at: string
  updated_at: string
}

export interface Message {
  id: string
  role: 'user' | 'assistant' | 'system'
  content: string
  timestamp: Date
  agent_id?: string
  session_id?: string
  mode?: string
  reasoning_chain?: ReasoningChain
  tools_used?: string[]
  rag_sources?: RAGSource[]
  confidence_score?: number
  processing_time?: number
  metadata?: Record<string, any>
}

export interface ChatSession {
  id: string
  title: string
  agent_id?: string
  user_id: string
  message_count: number
  created_at: string
  updated_at: string
  last_message_at: string
}

export interface ReasoningChain {
  id: string
  mode: 'chain_of_thought' | 'plan_execute' | 'react' | 'self_reflection' | 'tree_of_thoughts'
  steps: ReasoningStep[]
  final_answer: string
  confidence_score: number
  processing_time: number
}

export interface ReasoningStep {
  step_number: number
  type: 'thought' | 'action' | 'observation' | 'plan' | 'reflection'
  content: string
  metadata?: Record<string, any>
}

export interface RAGSource {
  id: string
  title: string
  content: string
  source_type: 'document' | 'url' | 'text' | 'database'
  relevance_score: number
  chunk_id: string
  metadata?: Record<string, any>
}

export interface KnowledgeItem {
  id: string
  title: string
  description: string
  type: 'document' | 'url' | 'text' | 'database'
  source_url?: string
  file_path?: string
  size: number
  status: 'processing' | 'indexed' | 'error' | 'pending'
  processing_progress?: number
  chunks_count: number
  embeddings_count: number
  vector_dimension: number
  created_by: string
  created_at: string
  updated_at: string
  last_accessed_at?: string
  usage_count: number
  metadata?: Record<string, any>
}

export interface SystemStats {
  total_requests: number
  requests_today: number
  success_rate: number
  avg_response_time: number
  active_agents: number
  total_agents: number
  knowledge_items: number
  storage_used: number
  storage_limit: number
  active_users: number
  total_users: number
}

export interface AgentAnalytics {
  agent_id: string
  agent_name: string
  requests: number
  success_rate: number
  avg_response_time: number
  error_rate: number
  user_satisfaction?: number
  popular_queries: string[]
  performance_trend: Array<{
    date: string
    requests: number
    success_rate: number
    avg_response_time: number
  }>
}

export interface SystemHealthComponent {
  name: string
  status: 'healthy' | 'warning' | 'error' | 'maintenance'
  uptime: number
  response_time: number
  last_check: string
  last_incident?: string
  metadata?: Record<string, any>
}

export interface Activity {
  id: string
  type: 'chat' | 'agent' | 'knowledge' | 'system' | 'user' | 'error'
  title: string
  description: string
  user_id?: string
  user_name?: string
  agent_id?: string
  severity: 'info' | 'warning' | 'error' | 'critical'
  timestamp: string
  metadata?: Record<string, any>
}

export interface Settings {
  ai_providers: {
    [key: string]: {
      enabled: boolean
      api_key: string
      models: string[]
      default_model?: string
      rate_limit?: number
    }
  }
  vector_database: {
    provider: string
    embedding_model: string
    dimension: number
    similarity_threshold: number
  }
  retrieval: {
    top_k: number
    rerank_top_n: number
    hybrid_search_weights: {
      vector: number
      keyword: number
      graph: number
    }
  }
  security: {
    encryption_enabled: boolean
    two_factor_auth: boolean
    audit_logging: boolean
    session_timeout: number
  }
  notifications: {
    email_enabled: boolean
    webhook_url?: string
    alert_thresholds: {
      error_rate: number
      response_time: number
      system_load: number
    }
  }
  system: {
    max_concurrent_requests: number
    request_timeout: number
    memory_limit: number
    log_level: 'debug' | 'info' | 'warning' | 'error'
  }
}

export interface ApiResponse<T = any> {
  success: boolean
  data?: T
  message?: string
  error?: string
  timestamp: string
}

export interface PaginatedResponse<T = any> {
  items: T[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

export interface ChatStreamMessage {
  type: 'message' | 'reasoning' | 'tool_use' | 'error' | 'complete'
  content: string
  metadata?: Record<string, any>
}