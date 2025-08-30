import React, { useEffect } from 'react'
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom'
import { Layout } from '@/components/layout/Layout'
import { ChatPage } from '@/pages/ChatPage'
import { useStore } from '@/stores/useStore'
import { apiService } from '@/services/api'

function App() {
  const { theme, setAgents, setModels, addNotification } = useStore()

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
    // Load initial data
    const loadData = async () => {
      try {
        // Load agents
        const agents = await apiService.getAgents()
        setAgents(agents)

        // Load models
        const models = await apiService.getModels()
        setModels(models)
      } catch (error) {
        console.error('Failed to load initial data:', error)
        addNotification({
          type: 'warning',
          title: '数据加载失败',
          message: '部分功能可能无法正常使用'
        })
      }
    }

    loadData()
  }, [setAgents, setModels, addNotification])

  return (
    <Router>
      <div className="min-h-screen bg-background font-sans antialiased">
        <Routes>
          <Route path="/" element={<Layout />}>
            <Route index element={<ChatPage />} />
            <Route path="agents" element={<div className="p-4">智能体管理页面</div>} />
            <Route path="analytics" element={<div className="p-4">数据分析页面</div>} />
            <Route path="settings" element={<div className="p-4">设置页面</div>} />
          </Route>
        </Routes>
      </div>
    </Router>
  )
}

export default App