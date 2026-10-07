import React from 'react'
import { Outlet } from 'react-router-dom'
import { Sidebar } from './Sidebar'
import { Header } from './Header'
import type { User } from '../../types'

interface AppShellProps {
  user: User
  onLogout: () => void
}

export const AppShell: React.FC<AppShellProps> = ({ user, onLogout }) => {
  return (
    <div className="flex h-screen bg-slate-50 overflow-hidden">
      <Sidebar role={user.role} onLogout={onLogout} />
      <div className="flex-1 flex flex-col min-w-0 overflow-hidden">
        <Header user={user} />
        <main className="flex-1 overflow-y-auto p-8">
          <Outlet />
        </main>
      </div>
    </div>
  )
}
