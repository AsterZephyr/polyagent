import React from 'react'
import { NavLink } from 'react-router-dom'
import {
  BarChart3,
  Database,
  Brain,
  Settings,
  ChevronLeft,
  ChevronRight,
  Activity
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { useStore } from '@/stores/useStore'
import { cn } from '@/lib/utils'

const navItems = [
  { to: '/', icon: BarChart3, label: '控制台' },
  { to: '/data', icon: Database, label: '数据管理' },
  { to: '/models', icon: Brain, label: '模型管理' },
  { to: '/analytics', icon: Activity, label: '业务分析' },
  { to: '/settings', icon: Settings, label: '系统设置' },
]

export function Sidebar() {
  const {
    sidebarOpen,
    setSidebarOpen
  } = useStore()

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
          <h1 className="text-lg font-semibold">推荐Agent系统</h1>
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

        {/* System Status */}
        {sidebarOpen && (
          <div className="p-4">
            <div className="flex items-center gap-2 mb-3 text-sm font-medium text-muted-foreground">
              <Activity className="w-4 h-4" />
              系统状态
            </div>
            <div className="space-y-2 text-sm">
              <div className="flex items-center justify-between">
                <span>推荐Agent</span>
                <div className="w-2 h-2 bg-green-500 rounded-full"></div>
              </div>
              <div className="flex items-center justify-between">
                <span>数据Agent</span>
                <div className="w-2 h-2 bg-green-500 rounded-full"></div>
              </div>
              <div className="flex items-center justify-between">
                <span>模型Agent</span>
                <div className="w-2 h-2 bg-green-500 rounded-full"></div>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}