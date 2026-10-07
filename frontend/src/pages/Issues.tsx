import React, { useState, useMemo } from 'react'
import { Link } from 'react-router-dom'
import { useIssues } from '../hooks/useIssues'
import { useResources } from '../hooks/useResources'
import { StatusBadge } from '../components/common/StatusBadge'
import { LoadingSpinner } from '../components/common/LoadingSpinner'
import { EmptyState } from '../components/common/EmptyState'
import { ErrorState } from '../components/common/ErrorState'
import { formatDateTime } from '../lib/utils'
import {
  RefreshCw,
  Search,
  Plus,
  ArrowRight,
  Building,
  Clock,
} from 'lucide-react'


interface IssuesProps {
  isAll?: boolean
}

export const Issues: React.FC<IssuesProps> = ({ isAll = false }) => {
  const { issues, isLoading, error, refresh } = useIssues(isAll)
  const { getResourceById } = useResources()

  const [searchTerm, setSearchTerm] = useState('')
  const [statusFilter, setStatusFilter] = useState<string>('ALL')

  const filteredIssues = useMemo(() => {
    return issues.filter((i) => {
      const resource = getResourceById(i.ResourceID)
      const resName = resource?.Name.toLowerCase() || ''
      const matchesSearch =
        resName.includes(searchTerm.toLowerCase()) ||
        i.Description.toLowerCase().includes(searchTerm.toLowerCase()) ||
        i.Category.toLowerCase().includes(searchTerm.toLowerCase())
      const matchesStatus = statusFilter === 'ALL' || i.Status === statusFilter
      return matchesSearch && matchesStatus
    })
  }, [issues, searchTerm, statusFilter, getResourceById])

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-slate-900">
            {isAll ? 'Campus Issue Triage' : 'My Reported Issues'}
          </h1>
          <p className="mt-1 text-sm text-slate-500">
            {isAll
              ? 'Triage, assign, and update equipment and facility defects'
              : 'Track the resolution status of faults you have reported'}
          </p>
        </div>

        <div className="flex items-center gap-3">
          <button
            onClick={() => refresh()}
            title="Refresh issues"
            className="p-2.5 bg-white border border-slate-200 rounded-lg text-slate-600 hover:text-slate-900 hover:bg-slate-50 transition"
          >
            <RefreshCw className="w-4 h-4" />
          </button>

          <Link
            to="/issues/create"
            className="inline-flex items-center gap-2 px-4 py-2.5 bg-indigo-600 text-white rounded-lg text-sm font-semibold hover:bg-indigo-700 transition shadow-xs"
          >
            <Plus className="w-4 h-4" />
            Report Issue
          </Link>
        </div>
      </div>

      {/* Filter and Search */}
      <div className="bg-white p-4 rounded-xl border border-slate-200/80 shadow-xs flex flex-col sm:flex-row gap-3">
        <div className="relative flex-1">
          <Search className="w-4 h-4 text-slate-400 absolute left-3.5 top-1/2 -translate-y-1/2" />
          <input
            type="text"
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            placeholder="Search by facility, category, or description..."
            className="w-full pl-9 pr-4 py-2 bg-slate-50 border border-slate-200 rounded-lg text-sm text-slate-900 placeholder:text-slate-400 focus:outline-none focus:ring-2 focus:ring-indigo-600 focus:bg-white transition"
          />
        </div>

        <select
          value={statusFilter}
          onChange={(e) => setStatusFilter(e.target.value)}
          className="px-3 py-2 bg-slate-50 border border-slate-200 rounded-lg text-sm text-slate-700 focus:outline-none focus:ring-2 focus:ring-indigo-600 focus:bg-white transition"
        >
          <option value="ALL">All Statuses</option>
          <option value="OPEN">Open</option>
          <option value="IN_PROGRESS">In Progress</option>
          <option value="RESOLVED">Resolved</option>
          <option value="CLOSED">Closed</option>
        </select>
      </div>

      {/* Issues Table / Cards */}
      {isLoading ? (
        <LoadingSpinner message="Loading issues..." />
      ) : error ? (
        <ErrorState message={error} onRetry={refresh} />
      ) : filteredIssues.length === 0 ? (
        <EmptyState
          title="No issues found"
          description={
            searchTerm || statusFilter !== 'ALL'
              ? 'No issues match your current search criteria.'
              : isAll
              ? 'No issues currently reported across the campus.'
              : 'You have not submitted any issue reports.'
          }
          action={{
            label: 'Report an Issue',
            onClick: () => (window.location.href = '/issues/create'),
          }}
        />
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {filteredIssues.map((item) => {
            const resource = getResourceById(item.ResourceID)

            return (
              <div
                key={item.ID}
                className="bg-white p-6 rounded-2xl border border-slate-200/80 shadow-xs hover:border-indigo-200 transition flex flex-col justify-between space-y-4 group"
              >
                <div className="space-y-3">
                  <div className="flex items-start justify-between gap-2">
                    <div className="flex items-center gap-2">
                      <span className="text-[11px] font-semibold tracking-wider uppercase text-amber-700 bg-amber-50 px-2.5 py-0.5 rounded-md">
                        {item.Category}
                      </span>
                      <StatusBadge status={item.Priority} />
                    </div>
                    <StatusBadge status={item.Status} />
                  </div>

                  <div>
                    <h3 className="text-base font-bold text-slate-900 group-hover:text-indigo-600 transition flex items-center gap-2">
                      <Building className="w-4 h-4 text-slate-400 shrink-0" />
                      <span>{resource?.Name || 'Facility Issue'}</span>
                    </h3>
                    <p className="mt-1 text-xs text-slate-500 line-clamp-2 leading-relaxed">
                      {item.Description}
                    </p>
                  </div>
                </div>

                <div className="pt-3 border-t border-slate-100 flex items-center justify-between text-xs text-slate-400">
                  <span className="flex items-center gap-1">
                    <Clock className="w-3.5 h-3.5" />
                    {formatDateTime(item.CreatedAt)}
                  </span>
                  <Link
                    to={`/issues/${item.ID}`}
                    className="font-semibold text-indigo-600 hover:text-indigo-700 flex items-center gap-1 group-hover:translate-x-0.5 transition"
                  >
                    View Details <ArrowRight className="w-3.5 h-3.5" />
                  </Link>
                </div>
              </div>
            )
          })}
        </div>
      )}
    </div>
  )
}
