import React from 'react'
import { NavLink, useLocation } from 'react-router-dom'
import { motion } from 'framer-motion'
import {
  MessageCircle,
  Brain,
  BookOpen,
  Settings,
  Bot,
  Home,
  Activity,
  Users,
  Database,
  Zap,
} from 'lucide-react'

const navigation = [
  {
    name: 'Overview',
    href: '/dashboard',
    icon: Home,
    description: 'System overview and analytics',
  },
  {
    name: 'Chat',
    href: '/chat',
    icon: MessageCircle,
    description: 'Interact with AI agents',
  },
  {
    name: 'Agents',
    href: '/agents',
    icon: Bot,
    description: 'Manage AI agents',
  },
  {
    name: 'Knowledge',
    href: '/knowledge',
    icon: BookOpen,
    description: 'Knowledge base management',
  },
  {
    name: 'Analytics',
    href: '/analytics',
    icon: Activity,
    description: 'Performance analytics',
  },
  {
    name: 'Users',
    href: '/users',
    icon: Users,
    description: 'User management',
  },
  {
    name: 'Settings',
    href: '/settings',
    icon: Settings,
    description: 'System configuration',
  },
]

const quickActions = [
  {
    name: 'New Chat',
    icon: MessageCircle,
    action: () => window.open('/chat', '_blank'),
  },
  {
    name: 'Upload Knowledge',
    icon: Database,
    action: () => window.open('/knowledge', '_self'),
  },
  {
    name: 'Create Agent',
    icon: Zap,
    action: () => window.open('/agents', '_self'),
  },
]

export default function Sidebar() {
  const location = useLocation()

  return (
    <div className="w-64 bg-white dark:bg-secondary-800 border-r border-secondary-200 dark:border-secondary-700 flex flex-col">
      {/* Logo */}
      <div className="flex items-center px-6 py-4 border-b border-secondary-200 dark:border-secondary-700">
        <div className="flex items-center">
          <div className="w-8 h-8 bg-gradient-primary rounded-lg flex items-center justify-center">
            <Brain className="w-5 h-5 text-white" />
          </div>
          <div className="ml-3">
            <h1 className="text-lg font-bold text-secondary-900 dark:text-white">
              PolyAgent
            </h1>
            <p className="text-xs text-secondary-500 dark:text-secondary-400">
              AI Platform
            </p>
          </div>
        </div>
      </div>

      {/* Navigation */}
      <nav className="flex-1 px-4 py-4 space-y-2">
        {navigation.map((item) => {
          const isActive = location.pathname === item.href ||
            (item.href !== '/dashboard' && location.pathname.startsWith(item.href))

          return (
            <NavLink
              key={item.name}
              to={item.href}
              className={({ isActive: linkActive }) =>
                `group flex items-center px-3 py-2 text-sm font-medium rounded-lg transition-all duration-200 ${
                  linkActive || isActive
                    ? 'bg-primary-50 dark:bg-primary-900/20 text-primary-700 dark:text-primary-300 border-l-2 border-primary-500'
                    : 'text-secondary-600 dark:text-secondary-400 hover:bg-secondary-50 dark:hover:bg-secondary-700 hover:text-secondary-900 dark:hover:text-white'
                }`
              }
            >
              {({ isActive: linkActive }) => (
                <>
                  <item.icon
                    className={`mr-3 h-5 w-5 transition-colors ${
                      linkActive || isActive
                        ? 'text-primary-500'
                        : 'text-secondary-400 group-hover:text-secondary-600 dark:group-hover:text-secondary-300'
                    }`}
                  />
                  <span className="flex-1">{item.name}</span>
                  {(linkActive || isActive) && (
                    <motion.div
                      layoutId="activeTab"
                      className="w-1 h-1 bg-primary-500 rounded-full"
                      initial={false}
                      transition={{ type: "spring", stiffness: 500, damping: 30 }}
                    />
                  )}
                </>
              )}
            </NavLink>
          )
        })}
      </nav>

      {/* Quick Actions */}
      <div className="px-4 py-4 border-t border-secondary-200 dark:border-secondary-700">
        <h3 className="text-xs font-semibold text-secondary-500 dark:text-secondary-400 uppercase tracking-wide mb-3">
          Quick Actions
        </h3>
        <div className="space-y-1">
          {quickActions.map((action) => (
            <button
              key={action.name}
              onClick={action.action}
              className="w-full flex items-center px-3 py-2 text-sm text-secondary-600 dark:text-secondary-400 hover:bg-secondary-50 dark:hover:bg-secondary-700 hover:text-secondary-900 dark:hover:text-white rounded-lg transition-colors duration-200 group"
            >
              <action.icon className="mr-3 h-4 w-4 text-secondary-400 group-hover:text-secondary-600 dark:group-hover:text-secondary-300" />
              <span>{action.name}</span>
            </button>
          ))}
        </div>
      </div>

      {/* Status */}
      <div className="px-4 py-3 border-t border-secondary-200 dark:border-secondary-700">
        <div className="flex items-center">
          <div className="w-2 h-2 bg-success-500 rounded-full animate-pulse-slow"></div>
          <span className="ml-2 text-xs text-secondary-500 dark:text-secondary-400">
            System Online
          </span>
        </div>
      </div>
    </div>
  )
}