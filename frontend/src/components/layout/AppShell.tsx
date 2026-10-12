import React, { useState } from 'react'
import { Outlet } from 'react-router-dom'
import { Sidebar } from './Sidebar'
import { Header } from './Header'
import { MobileNav } from './MobileNav'
import type { User } from '../../types'

interface AppShellProps {
  user: User
  onLogout: () => void
}

export const AppShell: React.FC<AppShellProps> = ({ user, onLogout }) => {
  const [mobileNavOpen, setMobileNavOpen] = useState(false)

  return (
    <div className="flex h-screen bg-slate-50 overflow-hidden">
      {/* Desktop Sidebar */}
      <Sidebar role={user.role} onLogout={onLogout} />

      {/* Mobile Drawer */}
      {mobileNavOpen && (
        <MobileNav
          role={user.role}
          onLogout={onLogout}
          onClose={() => setMobileNavOpen(false)}
        />
      )}

      {/* Main Content Area */}
      <div className="flex-1 flex flex-col min-w-0 overflow-hidden">
        <Header user={user} onOpenMobileNav={() => setMobileNavOpen(true)} />
        <main className="flex-1 overflow-y-auto p-4 sm:p-6 lg:p-8">
          <Outlet />
        </main>
      </div>
    </div>
  )
}
