import React from 'react'
import { User as UserIcon, Shield } from 'lucide-react'
import type { User } from '../../types'

interface HeaderProps {
  user: User
}

export const Header: React.FC<HeaderProps> = ({ user }) => {
  return (
    <header className="h-16 bg-white border-b border-slate-200 px-8 flex items-center justify-between shrink-0">
      <div>
        <h2 className="text-sm font-medium text-slate-500">
          Campus Facility & Resource System
        </h2>
      </div>

      <div className="flex items-center gap-4">
        <div className="flex items-center gap-2 px-3 py-1 bg-slate-100 rounded-full text-xs font-medium text-slate-700">
          <Shield className="w-3.5 h-3.5 text-indigo-600" />
          <span>Tenant ID:</span>
          <span className="font-mono text-slate-500">{user.tenant_id.slice(0, 8)}...</span>
        </div>

        <div className="flex items-center gap-2.5 pl-2 border-l border-slate-200">
          <div className="w-8 h-8 rounded-full bg-indigo-50 border border-indigo-200 text-indigo-700 flex items-center justify-center font-bold text-xs">
            <UserIcon className="w-4 h-4" />
          </div>
          <div className="flex flex-col text-left">
            <span className="text-xs font-semibold text-slate-800 leading-tight">
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
