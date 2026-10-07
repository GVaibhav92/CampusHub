import React, { useState, useEffect } from 'react'
import { useParams, Link } from 'react-router-dom'
import { getIssue, updateIssueStatus } from '../api/issues'
import { useResources } from '../hooks/useResources'
import { useAuth } from '../hooks/useAuth'
import type { Issue, IssueStatus } from '../types'
import { StatusBadge } from '../components/common/StatusBadge'
import { LoadingSpinner } from '../components/common/LoadingSpinner'
import { ErrorState } from '../components/common/ErrorState'
import { formatDateTime } from '../lib/utils'
import {
  ArrowLeft,
  Building,
  Clock,
  CheckCircle,
  Loader2,
} from 'lucide-react'


// Exact status transition matrix matching Go backend
const validTransitions: Record<IssueStatus, IssueStatus[]> = {
  OPEN: ['IN_PROGRESS', 'CLOSED'],
  IN_PROGRESS: ['RESOLVED', 'OPEN', 'CLOSED'],
  RESOLVED: ['CLOSED', 'IN_PROGRESS'],
  CLOSED: [],
}

export const IssueDetails: React.FC = () => {
  const { id } = useParams<{ id: string }>()
  const { user } = useAuth()
  const { getResourceById } = useResources()

  const [issue, setIssue] = useState<Issue | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  // Status update state for staff
  const [newStatus, setNewStatus] = useState<IssueStatus | ''>('')
  const [isUpdating, setIsUpdating] = useState(false)
  const [updateError, setUpdateError] = useState<string | null>(null)
  const [updateSuccess, setUpdateSuccess] = useState(false)

  const isStaff = user?.role === 'ADMIN' || user?.role === 'TECHNICIAN'

  useEffect(() => {
    if (!id) return
    let isMounted = true

    async function loadIssue() {
      setIsLoading(true)
      setError(null)
      try {
        const data = await getIssue(id!)
        if (isMounted) {
          setIssue(data)
          const nextAllowed = validTransitions[data.Status] || []
          if (nextAllowed.length > 0) {
            setNewStatus(nextAllowed[0])
          }
        }
      } catch (err: unknown) {
        if (isMounted) {
          if (err instanceof Error) {
            setError(err.message)
          } else {
            setError('Failed to fetch issue details.')
          }
        }
      } finally {
        if (isMounted) {
          setIsLoading(false)
        }
      }
    }

    loadIssue()

    return () => {
      isMounted = false
    }
  }, [id])

  const handleStatusSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!issue || !newStatus) return

    setIsUpdating(true)
    setUpdateError(null)
    setUpdateSuccess(false)

    try {
      const updated = await updateIssueStatus(issue.ID, {
        status: newStatus as IssueStatus,
      })
      setIssue(updated)
      setUpdateSuccess(true)
      const nextAllowed = validTransitions[updated.Status] || []
      if (nextAllowed.length > 0) {
        setNewStatus(nextAllowed[0])
      } else {
        setNewStatus('')
      }
    } catch (err: unknown) {
      if (err instanceof Error) {
        setUpdateError(err.message)
      } else {
        setUpdateError('Failed to update status.')
      }
    } finally {
      setIsUpdating(false)
    }
  }

  if (isLoading) {
    return <LoadingSpinner message="Loading issue report..." />
  }

  if (error || !issue) {
    return (
      <div className="space-y-4">
        <Link
          to="/issues/mine"
          className="inline-flex items-center gap-2 text-sm font-semibold text-slate-600 hover:text-slate-900"
        >
          <ArrowLeft className="w-4 h-4" /> Back to Issues
        </Link>
        <ErrorState
          title="Issue Not Found"
          message={error || 'Unable to locate this issue report.'}
          onRetry={() => window.location.reload()}
        />
      </div>
    )
  }

  const resource = getResourceById(issue.ResourceID)
  const allowedNextStatuses = validTransitions[issue.Status] || []

  return (
    <div className="max-w-3xl space-y-6">
      <div>
        <Link
          to={isStaff ? '/issues/all' : '/issues/mine'}
          className="inline-flex items-center gap-2 text-sm font-semibold text-slate-500 hover:text-slate-900 transition"
        >
          <ArrowLeft className="w-4 h-4" /> Back to Issues
        </Link>
      </div>

      <div className="bg-white rounded-2xl border border-slate-200/80 shadow-xs p-8 space-y-6">
        {/* Title and Badges */}
        <div className="flex flex-col sm:flex-row sm:items-start sm:justify-between gap-4 pb-6 border-b border-slate-100">
          <div>
            <div className="flex items-center gap-2.5">
              <span className="text-xs font-semibold tracking-wider uppercase text-amber-600 bg-amber-50 px-2.5 py-0.5 rounded-md">
                {issue.Category}
              </span>
              <StatusBadge status={issue.Priority} />
              <StatusBadge status={issue.Status} />
            </div>
            <h1 className="text-xl font-bold text-slate-900 mt-2.5">
              {resource?.Name ? `Fault at ${resource.Name}` : 'Facility Issue Report'}
            </h1>
          </div>
        </div>

        {/* Metadata Grid */}
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div className="p-4 bg-slate-50 rounded-xl border border-slate-100 space-y-1">
            <span className="text-xs font-medium text-slate-500 flex items-center gap-1.5">
              <Building className="w-3.5 h-3.5 text-slate-400" /> Facility
            </span>
            <p className="text-sm font-semibold text-slate-800">
              {resource ? `${resource.Name} (${resource.Location})` : issue.ResourceID}
            </p>
          </div>

          <div className="p-4 bg-slate-50 rounded-xl border border-slate-100 space-y-1">
            <span className="text-xs font-medium text-slate-500 flex items-center gap-1.5">
              <Clock className="w-3.5 h-3.5 text-slate-400" /> Reported On
            </span>
            <p className="text-sm font-semibold text-slate-800">
              {formatDateTime(issue.CreatedAt)}
            </p>
          </div>
        </div>

        {/* Description Body */}
        <div className="space-y-2">
          <h3 className="text-xs font-semibold uppercase tracking-wider text-slate-400">
            Description
          </h3>
          <div className="p-4 bg-slate-50 rounded-xl border border-slate-100 text-sm text-slate-800 leading-relaxed whitespace-pre-wrap">
            {issue.Description}
          </div>
        </div>

        {/* Staff Transition Controls */}
        {isStaff && (
          <div className="pt-6 border-t border-slate-100 space-y-4">
            <div className="flex items-center justify-between">
              <div>
                <h3 className="text-sm font-bold text-slate-900">
                  Update Issue Status
                </h3>
                <p className="text-xs text-slate-500">
                  Transition this issue according to campus maintenance policy
                </p>
              </div>
            </div>

            {updateError && (
              <div className="p-3 bg-rose-50 border border-rose-200 rounded-lg text-xs font-semibold text-rose-800">
                {updateError}
              </div>
            )}

            {updateSuccess && (
              <div className="p-3 bg-emerald-50 border border-emerald-200 rounded-lg text-xs font-semibold text-emerald-800 flex items-center gap-2">
                <CheckCircle className="w-4 h-4" /> Status updated successfully!
              </div>
            )}

            {allowedNextStatuses.length === 0 ? (
              <p className="text-xs font-medium text-slate-500 italic">
                This issue is in a terminal state ({issue.Status}) and cannot be transitioned further.
              </p>
            ) : (
              <form onSubmit={handleStatusSubmit} className="flex items-center gap-3">
                <select
                  value={newStatus}
                  onChange={(e) => setNewStatus(e.target.value as IssueStatus)}
                  className="px-3.5 py-2 bg-white border border-slate-200 rounded-lg text-sm text-slate-900 focus:outline-none focus:ring-2 focus:ring-indigo-600 transition"
                >
                  {allowedNextStatuses.map((s) => (
                    <option key={s} value={s}>
                      Move to: {s.replace('_', ' ')}
                    </option>
                  ))}
                </select>

                <button
                  type="submit"
                  disabled={isUpdating}
                  className="px-4 py-2 bg-indigo-600 text-white rounded-lg text-sm font-semibold hover:bg-indigo-700 transition flex items-center gap-2 disabled:opacity-50"
                >
                  {isUpdating && <Loader2 className="w-4 h-4 animate-spin" />}
                  <span>Save Status</span>
                </button>
              </form>
            )}
          </div>
        )}
      </div>
    </div>
  )
}
