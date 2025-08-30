import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import { Agent, ChatSession, Message, ModelHealth } from '@/types'

interface AppStore {
  // Theme
  theme: 'light' | 'dark' | 'system'
  setTheme: (theme: 'light' | 'dark' | 'system') => void

  // Authentication
  isAuthenticated: boolean
  user: any | null
  setAuthenticated: (status: boolean, user?: any) => void

  // Agents
  agents: Record<string, Agent>
  currentAgent: string | null
  setAgents: (agents: Record<string, Agent>) => void
  setCurrentAgent: (agentId: string | null) => void
  addAgent: (agent: Agent) => void
  removeAgent: (agentId: string) => void

  // Chat Sessions
  sessions: ChatSession[]
  currentSession: string | null
  setCurrentSession: (sessionId: string | null) => void
  addSession: (session: ChatSession) => void
  updateSession: (sessionId: string, updates: Partial<ChatSession>) => void
  removeSession: (sessionId: string) => void
  addMessage: (sessionId: string, message: Message) => void

  // Models
  models: Record<string, ModelHealth>
  setModels: (models: Record<string, ModelHealth>) => void

  // UI State
  sidebarOpen: boolean
  setSidebarOpen: (open: boolean) => void
  
  // Notifications
  notifications: Array<{
    id: string
    type: 'success' | 'error' | 'warning' | 'info'
    title: string
    message: string
    timestamp: Date
  }>
  addNotification: (notification: Omit<AppStore['notifications'][0], 'id' | 'timestamp'>) => void
  removeNotification: (id: string) => void
}

export const useStore = create<AppStore>()(
  persist(
    (set, get) => ({
      // Theme
      theme: 'system',
      setTheme: (theme) => set({ theme }),

      // Authentication
      isAuthenticated: false,
      user: null,
      setAuthenticated: (status, user) => set({ isAuthenticated: status, user }),

      // Agents
      agents: {},
      currentAgent: null,
      setAgents: (agents) => set({ agents }),
      setCurrentAgent: (agentId) => set({ currentAgent: agentId }),
      addAgent: (agent) => set((state) => ({
        agents: { ...state.agents, [agent.id]: agent }
      })),
      removeAgent: (agentId) => set((state) => {
        const { [agentId]: removed, ...rest } = state.agents
        return { agents: rest }
      }),

      // Chat Sessions
      sessions: [],
      currentSession: null,
      setCurrentSession: (sessionId) => set({ currentSession: sessionId }),
      addSession: (session) => set((state) => ({
        sessions: [...state.sessions, session]
      })),
      updateSession: (sessionId, updates) => set((state) => ({
        sessions: state.sessions.map(session =>
          session.id === sessionId ? { ...session, ...updates } : session
        )
      })),
      removeSession: (sessionId) => set((state) => ({
        sessions: state.sessions.filter(session => session.id !== sessionId)
      })),
      addMessage: (sessionId, message) => set((state) => ({
        sessions: state.sessions.map(session =>
          session.id === sessionId
            ? { ...session, messages: [...session.messages, message], updated_at: new Date() }
            : session
        )
      })),

      // Models
      models: {},
      setModels: (models) => set({ models }),

      // UI State
      sidebarOpen: true,
      setSidebarOpen: (open) => set({ sidebarOpen: open }),

      // Notifications
      notifications: [],
      addNotification: (notification) => set((state) => ({
        notifications: [
          ...state.notifications,
          {
            ...notification,
            id: `${Date.now()}_${Math.random()}`,
            timestamp: new Date(),
          }
        ]
      })),
      removeNotification: (id) => set((state) => ({
        notifications: state.notifications.filter(n => n.id !== id)
      })),
    }),
    {
      name: 'polyagent-storage',
      partialize: (state) => ({
        theme: state.theme,
        isAuthenticated: state.isAuthenticated,
        user: state.user,
        currentAgent: state.currentAgent,
        sessions: state.sessions.slice(-50), // Keep last 50 sessions
        sidebarOpen: state.sidebarOpen,
      }),
    }
  )
)