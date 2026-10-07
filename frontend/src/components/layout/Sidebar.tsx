import React from 'react'
import { NavLink } from 'react-router-dom'
import {
  LayoutDashboard,
  Box,
  CalendarCheck,
  AlertTriangle,
  Users,
  LogOut,
  Building2,
} from 'lucide-react'
import type { Role } from '../../types'

interface SidebarProps {
  role: Role
  onLogout: () => void
}

export const Sidebar: React.FC<SidebarProps> = ({ role, onLogout }) => {
  const isStaff = role === 'ADMIN' || role === 'TECHNICIAN'
  const isAdmin = role === 'ADMIN'

  const navItems = [
    {
      to: '/dashboard',
      label: 'Dashboard',
      icon: LayoutDashboard,
    },
    {
      to: '/resources',
      label: 'Resources',
      icon: Box,
    },
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
    ...(isAdmin
      ? [
          {
            to: '/users',
            label: 'Users',
            icon: Users,
          },
        ]
      : []),
  ]

  return (
    <aside className="w-64 bg-white border-r border-slate-200 flex flex-col shrink-0 min-h-screen">
      {/* Brand Header */}
      <div className="h-16 flex items-center px-6 border-b border-slate-100 gap-3">
        <div className="w-9 h-9 rounded-lg bg-indigo-600 text-white flex items-center justify-center font-bold shadow-xs">
          <Building2 className="w-5 h-5" />
        </div>
        <div>
          <span className="font-bold text-slate-900 tracking-tight text-lg">CampusHub</span>
          <span className="block text-[11px] font-medium text-slate-400 uppercase tracking-wider">
            {role} Portal
          </span>
        </div>
      </div>

      {/* Nav List */}
      <nav className="flex-1 px-4 py-6 space-y-1">
        {navItems.map((item) => {
          const Icon = item.icon
          return (
            <NavLink
              key={item.to}
              to={item.to}
              className={({ isActive }) =>
                `flex items-center gap-3 px-3.5 py-2.5 rounded-lg text-sm font-medium transition-colors ${
                  isActive
                    ? 'bg-indigo-50 text-indigo-700 font-semibold'
                    : 'text-slate-600 hover:text-slate-900 hover:bg-slate-50'
                }`
              }
            >
              <Icon className="w-4 h-4 shrink-0" />
              {item.label}
            </NavLink>
          )
        })}
      </nav>

      {/* Footer / Logout */}
      <div className="p-4 border-t border-slate-100">
        <button
          onClick={onLogout}
          className="w-full flex items-center gap-3 px-3.5 py-2.5 rounded-lg text-sm font-medium text-slate-600 hover:text-rose-600 hover:bg-rose-50 transition-colors"
        >
          <LogOut className="w-4 h-4" />
          Sign Out
        </button>
      </div>
    </aside>
  )
}
