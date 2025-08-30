import React from 'react'
import { Outlet } from 'react-router-dom'
import { Sidebar } from './Sidebar'
import { Header } from './Header'
import { useStore } from '@/stores/useStore'
import { cn } from '@/lib/utils'

export function Layout() {
  const { sidebarOpen } = useStore()

  return (
    <div className="flex h-screen bg-background">
      <Sidebar />
      <div className={cn(
        "flex-1 flex flex-col transition-all duration-300",
        sidebarOpen ? "ml-64" : "ml-16"
      )}>
        <Header />
        <main className="flex-1 overflow-hidden">
          <Outlet />
        </main>
      </div>
    </div>
  )
}