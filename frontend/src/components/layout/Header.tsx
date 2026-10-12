import React from 'react'
import { User as UserIcon, Shield, Menu } from 'lucide-react'
import type { User } from '../../types'

interface HeaderProps {
  user: User
  onOpenMobileNav?: () => void
}

export const Header: React.FC<HeaderProps> = ({ user, onOpenMobileNav }) => {
  return (
    <header className="h-16 bg-white border-b border-slate-200 px-4 sm:px-8 flex items-center justify-between shrink-0">
      <div className="flex items-center gap-3">
        {onOpenMobileNav && (
          <button
            type="button"
            onClick={onOpenMobileNav}
            aria-label="Open mobile navigation menu"
            className="lg:hidden p-2 rounded-lg text-slate-500 hover:text-slate-700 hover:bg-slate-100 transition-colors"
          >
            <Menu className="w-5 h-5" aria-hidden="true" />
          </button>
        )}
        <h2 className="text-xs sm:text-sm font-medium text-slate-500 truncate">
          Campus Facility & Resource System
        </h2>
      </div>

      <div className="flex items-center gap-3 sm:gap-4">
        <div className="hidden sm:flex items-center gap-2 px-3 py-1 bg-slate-100 rounded-full text-xs font-medium text-slate-700">
          <Shield className="w-3.5 h-3.5 text-indigo-600" aria-hidden="true" />
          <span>Tenant:</span>
          <span className="font-mono text-slate-500">{user.tenant_id.slice(0, 8)}...</span>
        </div>

        <div className="flex items-center gap-2.5 sm:pl-2 sm:border-l sm:border-slate-200">
          <div className="w-8 h-8 rounded-full bg-indigo-50 border border-indigo-200 text-indigo-700 flex items-center justify-center font-bold text-xs">
            <UserIcon className="w-4 h-4" aria-hidden="true" />
          </div>
          <div className="flex flex-col text-left">
            <span className="text-xs font-semibold text-slate-800 leading-tight max-w-[120px] sm:max-w-[180px] truncate">
              {user.email}
            </span>
            <span className="text-[10px] font-medium text-slate-400 capitalize">
              {user.role.toLowerCase()}
            </span>
          </div>
        </div>
      </div>
    </header>
  )
}
