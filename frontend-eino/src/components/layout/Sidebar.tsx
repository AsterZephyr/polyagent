import React from 'react'
import { NavLink } from 'react-router-dom'
import {
  MessageSquare,
  Bot,
  Settings,
  BarChart3,
  ChevronLeft,
  ChevronRight,
  Plus,
  History
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { useStore } from '@/stores/useStore'
import { cn } from '@/lib/utils'

const navItems = [
  { to: '/', icon: MessageSquare, label: '对话' },
  { to: '/agents', icon: Bot, label: '智能体' },
  { to: '/analytics', icon: BarChart3, label: '分析' },
  { to: '/settings', icon: Settings, label: '设置' },
]

export function Sidebar() {
  const {
    sidebarOpen,
    setSidebarOpen,
    sessions,
    currentSession,
    setCurrentSession,
    addSession
  } = useStore()

  const handleNewChat = () => {
    const newSession = {
      id: `session_${Date.now()}`,
      name: '新对话',
      agent_id: 'default',
      messages: [],
      created_at: new Date(),
      updated_at: new Date()
    }
    addSession(newSession)
    setCurrentSession(newSession.id)
  }

  return (
    <div
      className={cn(
        "fixed left-0 top-0 h-full bg-card border-r transition-all duration-300 z-50",
        sidebarOpen ? "w-64" : "w-16"
      )}
    >
      {/* Header */}
      <div className="flex items-center justify-between p-4 border-b">
        {sidebarOpen && (
          <h1 className="text-lg font-semibold">PolyAgent</h1>
        )}
        <Button
          variant="ghost"
          size="icon"
          onClick={() => setSidebarOpen(!sidebarOpen)}
        >
          {sidebarOpen ? (
            <ChevronLeft className="w-4 h-4" />
          ) : (
            <ChevronRight className="w-4 h-4" />
          )}
        </Button>
      </div>

      {/* New Chat Button */}
      <div className="p-4 border-b">
        <Button
          onClick={handleNewChat}
          className="w-full"
          variant="default"
        >
          <Plus className="w-4 h-4" />
          {sidebarOpen && <span className="ml-2">新对话</span>}
        </Button>
      </div>

      {/* Navigation */}
      <div className="flex-1 overflow-y-auto">
        <nav className="p-4 space-y-2">
          {navItems.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              className={({ isActive }) =>
                cn(
                  "flex items-center px-3 py-2 rounded-md transition-colors",
                  isActive
                    ? "bg-primary text-primary-foreground"
                    : "hover:bg-accent hover:text-accent-foreground"
                )
              }
            >
              <item.icon className="w-5 h-5" />
              {sidebarOpen && <span className="ml-3">{item.label}</span>}
            </NavLink>
          ))}
        </nav>

        {/* Recent Sessions */}
        {sidebarOpen && sessions.length > 0 && (
          <div className="p-4">
            <div className="flex items-center gap-2 mb-3 text-sm font-medium text-muted-foreground">
              <History className="w-4 h-4" />
              最近对话
            </div>
            <div className="space-y-1 max-h-64 overflow-y-auto">
              {sessions.slice(-10).reverse().map((session) => (
                <button
                  key={session.id}
                  onClick={() => setCurrentSession(session.id)}
                  className={cn(
                    "w-full text-left px-3 py-2 rounded-md text-sm transition-colors truncate",
                    currentSession === session.id
                      ? "bg-accent"
                      : "hover:bg-accent/50"
                  )}
                >
                  {session.name || '未命名对话'}
                </button>
              ))}
            </div>
          </div>
        )}
      </div>
    </div>
  )
}