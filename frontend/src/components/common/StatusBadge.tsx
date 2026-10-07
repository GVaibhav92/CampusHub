import React from 'react'

interface StatusBadgeProps {
  status: string
}

export const StatusBadge: React.FC<StatusBadgeProps> = ({ status }) => {
  const getBadgeStyle = (val: string) => {
    switch (val.toUpperCase()) {
      case 'AVAILABLE':
      case 'CONFIRMED':
      case 'RESOLVED':
        return 'bg-emerald-50 text-emerald-700 border-emerald-200'
      case 'MAINTENANCE':
      case 'IN_PROGRESS':
        return 'bg-amber-50 text-amber-700 border-amber-200'
      case 'OPEN':
        return 'bg-blue-50 text-blue-700 border-blue-200'
      case 'CANCELLED':
      case 'CLOSED':
        return 'bg-slate-100 text-slate-600 border-slate-200'
      case 'HIGH':
        return 'bg-rose-50 text-rose-700 border-rose-200'
      case 'MEDIUM':
        return 'bg-amber-50 text-amber-700 border-amber-200'
      case 'LOW':
        return 'bg-slate-50 text-slate-700 border-slate-200'
      default:
        return 'bg-slate-50 text-slate-700 border-slate-200'
    }
  }

  return (
    <span
      className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-semibold border ${getBadgeStyle(
        status
      )}`}
    >
      {status.replace('_', ' ')}
    </span>
  )
}
