import React from 'react'
import { Routes, Route } from 'react-router-dom'
import { motion } from 'framer-motion'
import Layout from './components/Layout/Layout'
import HomePage from './pages/HomePage'
import ChatPage from './pages/ChatPage'
import KnowledgePage from './pages/KnowledgePage'
import AgentsPage from './pages/AgentsPage'
import AnalyticsPage from './pages/AnalyticsPage'
import DashboardPage from './pages/DashboardPage'
import SettingsPage from './pages/SettingsPage'
import { useTheme } from './hooks/useTheme'

function App() {
  const { theme } = useTheme()

  return (
    <div className={theme}>
      <div className="min-h-screen bg-white dark:bg-secondary-900 transition-colors duration-200">
        <Layout>
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5 }}
          >
            <Routes>
              <Route path="/" element={<HomePage />} />
              <Route path="/dashboard" element={<DashboardPage />} />
              <Route path="/chat" element={<ChatPage />} />
              <Route path="/chat/:sessionId" element={<ChatPage />} />
              <Route path="/knowledge" element={<KnowledgePage />} />
              <Route path="/agents" element={<AgentsPage />} />
              <Route path="/analytics" element={<AnalyticsPage />} />
              <Route path="/settings" element={<SettingsPage />} />
            </Routes>
          </motion.div>
        </Layout>
      </div>
    </div>
  )
}

export default App