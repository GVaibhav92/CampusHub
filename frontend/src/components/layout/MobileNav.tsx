import { useEffect } from 'react'
import { NavLink } from 'react-router-dom'
import {
  LayoutDashboard,
  Box,
  CalendarCheck,
  AlertTriangle,
  Users,
  LogOut,
  Building2,
  X,
} from 'lucide-react'
import type { Role } from '../../types'

interface MobileNavProps {
  role: Role
  onLogout: () => void
  onClose: () => void
}

export function MobileNav({ role, onLogout, onClose }: MobileNavProps) {
  const isStaff = role === 'ADMIN' || role === 'TECHNICIAN'
  const isAdmin = role === 'ADMIN'

  const navItems = [
    { to: '/dashboard', label: 'Dashboard', icon: LayoutDashboard },
    { to: '/resources', label: 'Resources', icon: Box },
    {
      to: isStaff ? '/bookings/all' : '/bookings/mine',
      label: isStaff ? 'All Bookings' : 'My Bookings',
      icon: CalendarCheck,
    },
    {
      to: isStaff ? '/issues/all' : '/issues/mine',
      label: isStaff ? 'Issue Triage' : 'My Issues',
      icon: AlertTriangle,
    },
    ...(isAdmin ? [{ to: '/users', label: 'Users', icon: Users }] : []),
  ]

  // Close on ESC key
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose()
    }
    document.addEventListener('keydown', handleKeyDown)
    return () => document.removeEventListener('keydown', handleKeyDown)
  }, [onClose])

  // Lock body scroll while open
  useEffect(() => {
    document.body.style.overflow = 'hidden'
    return () => {
      document.body.style.overflow = ''
    }
  }, [])

  return (
    <>
      {/* Backdrop */}
      <div
        className="fixed inset-0 z-40 bg-black/40 backdrop-blur-xs lg:hidden"
        onClick={onClose}
        aria-hidden="true"
      />

      {/* Drawer */}
      <aside
        className="fixed inset-y-0 left-0 z-50 w-72 bg-white shadow-xl flex flex-col lg:hidden"
        role="dialog"
        aria-modal="true"
        aria-label="Mobile Navigation Menu"
      >
        {/* Header */}
        <div className="h-16 flex items-center px-6 border-b border-slate-100 gap-3">
          <div className="w-9 h-9 rounded-lg bg-indigo-600 text-white flex items-center justify-center font-bold shadow-xs">
            <Building2 className="w-5 h-5" aria-hidden="true" />
          </div>
          <div className="flex-1">
            <span className="font-bold text-slate-900 tracking-tight text-lg">CampusHub</span>
            <span className="block text-[11px] font-medium text-slate-400 uppercase tracking-wider">
              {role} Portal
            </span>
          </div>
          <button
            onClick={onClose}
            aria-label="Close navigation menu"
            className="p-1.5 rounded-lg text-slate-400 hover:text-slate-700 hover:bg-slate-100 transition-colors"
          >
            <X className="w-5 h-5" aria-hidden="true" />
          </button>
        </div>

        {/* Navigation Items */}
        <nav className="flex-1 px-4 py-6 space-y-1 overflow-y-auto" aria-label="Mobile Navigation">
          {navItems.map((item) => {
            const Icon = item.icon
            return (
              <NavLink
                key={item.to}
                to={item.to}
                onClick={onClose}
                className={({ isActive }) =>
                  `flex items-center gap-3 px-3.5 py-2.5 rounded-lg text-sm font-medium transition-colors ${
                    isActive
                      ? 'bg-indigo-50 text-indigo-700 font-semibold'
                      : 'text-slate-600 hover:text-slate-900 hover:bg-slate-50'
                  }`
                }
              >
                <Icon className="w-4 h-4 shrink-0" aria-hidden="true" />
                {item.label}
              </NavLink>
            )
          })}
        </nav>

        {/* Footer / Logout */}
        <div className="p-4 border-t border-slate-100">
          <button
            onClick={() => {
              onLogout()
              onClose()
            }}
            className="w-full flex items-center gap-3 px-3.5 py-2.5 rounded-lg text-sm font-medium text-slate-600 hover:text-rose-600 hover:bg-rose-50 transition-colors"
            aria-label="Sign out"
          >
            <LogOut className="w-4 h-4" aria-hidden="true" />
            Sign Out
          </button>
        </div>
      </aside>
    </>
  )
}
