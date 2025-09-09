import React, { useEffect } from 'react'
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom'
import { Layout } from '@/components/layout/Layout'
import { AgentDashboard } from '@/pages/AgentDashboard'
import { useStore } from '@/stores/useStore'
import { recommendationApi } from '@/services/recommendation'

function App() {
  const { theme, addNotification } = useStore()

  useEffect(() => {
    // Apply theme
    const root = window.document.documentElement
    root.classList.remove('light', 'dark')

    if (theme === 'system') {
      const systemTheme = window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
      root.classList.add(systemTheme)
    } else {
      root.classList.add(theme)
    }
  }, [theme])

  useEffect(() => {
    // Check recommendation system health
    const checkHealth = async () => {
      try {
        await recommendationApi.healthCheck()
        addNotification({
          type: 'success',
          title: '推荐系统已连接',
          message: '推荐Agent系统运行正常'
        })
      } catch (error) {
        console.error('Failed to connect to recommendation system:', error)
        addNotification({
          type: 'error',
          title: '推荐系统连接失败',
          message: '请确保后端服务器正在运行'
        })
      }
    }

    checkHealth()
  }, [addNotification])

  return (
    <Router>
      <div className="min-h-screen bg-background font-sans antialiased">
        <Routes>
          <Route path="/" element={<Layout />}>
            <Route index element={<AgentDashboard />} />
            <Route path="data" element={<div className="p-4">数据管理页面</div>} />
            <Route path="models" element={<div className="p-4">模型管理页面</div>} />
            <Route path="analytics" element={<div className="p-4">业务分析页面</div>} />
            <Route path="settings" element={<div className="p-4">系统设置页面</div>} />
          </Route>
        </Routes>
      </div>
    </Router>
  )
}

export default App