import React from 'react'
import { useLocation } from 'react-router-dom'
import { motion } from 'framer-motion'
import {
  Sun,
  Moon,
  Bell,
  Search,
  Settings,
  User,
  ChevronDown,
  Zap,
  Activity,
  HelpCircle,
} from 'lucide-react'
import { useTheme } from '../../hooks/useTheme'

const pageNames: Record<string, string> = {
  '/dashboard': 'Dashboard',
  '/chat': 'AI Chat',
  '/agents': 'Agent Management',
  '/knowledge': 'Knowledge Base',
  '/analytics': 'Analytics',
  '/users': 'User Management',
  '/settings': 'Settings',
}

export default function Header() {
  const location = useLocation()
  const { theme, toggleTheme } = useTheme()
  const pageName = pageNames[location.pathname] || 'PolyAgent'

  return (
    <header className="bg-white dark:bg-secondary-800 border-b border-secondary-200 dark:border-secondary-700 px-6 py-4">
      <div className="flex items-center justify-between">
        {/* Left section */}
        <div className="flex items-center space-x-4">
          <div>
            <h1 className="text-2xl font-bold text-secondary-900 dark:text-white">
              {pageName}
            </h1>
            {location.pathname === '/chat' && (
              <p className="text-sm text-secondary-500 dark:text-secondary-400 mt-1">
                Intelligent AI assistance with advanced reasoning
              </p>
            )}
          </div>
        </div>

        {/* Center section - Search */}
        <div className="flex-1 max-w-lg mx-8">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-secondary-400 w-5 h-5" />
            <input
              type="text"
              placeholder="Search agents, knowledge, or ask a question..."
              className="w-full pl-10 pr-4 py-2 bg-secondary-50 dark:bg-secondary-700 border border-secondary-200 dark:border-secondary-600 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent transition-all duration-200 text-secondary-900 dark:text-white placeholder-secondary-500"
            />
          </div>
        </div>

        {/* Right section */}
        <div className="flex items-center space-x-3">
          {/* System Status */}
          <div className="hidden lg:flex items-center space-x-2 px-3 py-2 bg-success-50 dark:bg-success-900/20 rounded-lg">
            <Activity className="w-4 h-4 text-success-600 dark:text-success-400" />
            <span className="text-sm font-medium text-success-700 dark:text-success-300">
              Online
            </span>
            <div className="w-2 h-2 bg-success-500 rounded-full animate-pulse-slow"></div>
          </div>

          {/* Quick Actions */}
          <button className="p-2 text-secondary-600 dark:text-secondary-400 hover:text-secondary-900 dark:hover:text-white hover:bg-secondary-100 dark:hover:bg-secondary-700 rounded-lg transition-colors duration-200">
            <Zap className="w-5 h-5" />
          </button>

          {/* Theme Toggle */}
          <button
            onClick={toggleTheme}
            className="p-2 text-secondary-600 dark:text-secondary-400 hover:text-secondary-900 dark:hover:text-white hover:bg-secondary-100 dark:hover:bg-secondary-700 rounded-lg transition-colors duration-200"
          >
            {theme === 'light' ? (
              <Moon className="w-5 h-5" />
            ) : (
              <Sun className="w-5 h-5" />
            )}
          </button>

          {/* Help */}
          <button className="p-2 text-secondary-600 dark:text-secondary-400 hover:text-secondary-900 dark:hover:text-white hover:bg-secondary-100 dark:hover:bg-secondary-700 rounded-lg transition-colors duration-200">
            <HelpCircle className="w-5 h-5" />
          </button>

          {/* Notifications */}
          <button className="relative p-2 text-secondary-600 dark:text-secondary-400 hover:text-secondary-900 dark:hover:text-white hover:bg-secondary-100 dark:hover:bg-secondary-700 rounded-lg transition-colors duration-200">
            <Bell className="w-5 h-5" />
            <span className="absolute -top-1 -right-1 w-3 h-3 bg-primary-500 rounded-full"></span>
          </button>

          {/* Settings */}
          <button className="p-2 text-secondary-600 dark:text-secondary-400 hover:text-secondary-900 dark:hover:text-white hover:bg-secondary-100 dark:hover:bg-secondary-700 rounded-lg transition-colors duration-200">
            <Settings className="w-5 h-5" />
          </button>

          {/* User Menu */}
          <div className="relative">
            <button className="flex items-center space-x-2 px-3 py-2 bg-secondary-100 dark:bg-secondary-700 hover:bg-secondary-200 dark:hover:bg-secondary-600 rounded-lg transition-colors duration-200 group">
              <div className="w-8 h-8 bg-gradient-primary rounded-full flex items-center justify-center">
                <User className="w-5 h-5 text-white" />
              </div>
              <div className="hidden md:block text-left">
                <p className="text-sm font-medium text-secondary-900 dark:text-white">
                  Demo User
                </p>
                <p className="text-xs text-secondary-500 dark:text-secondary-400">
                  Administrator
                </p>
              </div>
              <ChevronDown className="w-4 h-4 text-secondary-500 dark:text-secondary-400 group-hover:text-secondary-700 dark:group-hover:text-secondary-200 transition-colors" />
            </button>
          </div>
        </div>
      </div>
    </header>
  )
}