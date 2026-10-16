import React, { useState, useEffect } from 'react'
import { useParams, Link, useNavigate } from 'react-router-dom'
import { getResource } from '../api/resources'
import { useResources } from '../hooks/useResources'
import { useAuth } from '../hooks/useAuth'
import type { Resource } from '../types'
import { StatusBadge } from '../components/common/StatusBadge'
import { LoadingSpinner } from '../components/common/LoadingSpinner'
import { ErrorState } from '../components/common/ErrorState'
import { ResourceModal } from '../components/resources/ResourceModal'
import { ConfirmDialog } from '../components/common/ConfirmDialog'
import {
  ArrowLeft,
  MapPin,
  Users,
  CalendarPlus,
  AlertTriangle,
  Edit2,
  Trash2,
  Building,
} from 'lucide-react'

export const ResourceDetails: React.FC = () => {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { user } = useAuth()
  const { getResourceById, editResource, removeResource } = useResources()

  const [resource, setResource] = useState<Resource | null>(() => (id ? getResourceById(id) || null : null))
  const [isLoading, setIsLoading] = useState(!resource)
  const [error, setError] = useState<string | null>(null)

  // Admin state
  const [isEditOpen, setIsEditOpen] = useState(false)
  const [isDeleteOpen, setIsDeleteOpen] = useState(false)
  const [isDeleting, setIsDeleting] = useState(false)
  const [actionError, setActionError] = useState<string | null>(null)

  const isAdmin = user?.role === 'ADMIN'

  useEffect(() => {
    if (!id) return
    let isMounted = true

    async function fetchDetails() {
      setIsLoading(true)
      setError(null)
      try {
        const data = await getResource(id!)
        if (isMounted) {
          setResource(data)
        }
      } catch (err: unknown) {
        if (isMounted) {
          if (err instanceof Error) {
            setError(err.message)
          } else {
            setError('Failed to fetch resource details.')
          }
        }
      } finally {
        if (isMounted) {
          setIsLoading(false)
        }
      }
    }

    fetchDetails()

    return () => {
      isMounted = false
    }
  }, [id])

  const handleEditSave = async (payload: {
    name: string
    type: string
    capacity: number
    location: string
    status?: any
  }) => {
    if (!resource) return
    const updated = await editResource(resource.ID, {
      name: payload.name,
      type: payload.type,
      capacity: payload.capacity,
      location: payload.location,
      status: payload.status || resource.Status,
    })
    setResource(updated)
  }

  const handleDeleteConfirm = async () => {
    if (!resource) return
    setIsDeleting(true)
    setActionError(null)
    try {
      await removeResource(resource.ID)
      navigate('/resources', { replace: true })
    } catch (err: unknown) {
      if (err instanceof Error) {
        setActionError(err.message)
      } else {
        setActionError('Failed to delete resource.')
      }
      setIsDeleting(false)
      setIsDeleteOpen(false)
    }
  }

  if (isLoading) {
    return <LoadingSpinner message="Loading facility details..." />
  }

  if (error || !resource) {
    return (
      <div className="space-y-4">
        <Link
          to="/resources"
          className="inline-flex items-center gap-2 text-sm font-semibold text-slate-600 hover:text-slate-900"
        >
          <ArrowLeft className="w-4 h-4" /> Back to Resources
        </Link>
        <ErrorState
          title="Facility Not Found"
          message={error || 'The requested resource could not be found.'}
          onRetry={() => window.location.reload()}
        />
      </div>
    )
  }

  return (
    <div className="max-w-4xl space-y-6">
      {/* Back button */}
      <div>
        <Link
          to="/resources"
          className="inline-flex items-center gap-2 text-sm font-semibold text-slate-500 hover:text-slate-900 transition"
        >
          <ArrowLeft className="w-4 h-4" /> Back to Resources
        </Link>
      </div>

      {actionError && (
        <div className="p-4 bg-rose-50 border border-rose-200 rounded-xl text-sm font-semibold text-rose-800">
          {actionError}
        </div>
      )}

      {/* Main Detail Card */}
      <div className="bg-white rounded-2xl border border-slate-200/80 shadow-xs overflow-hidden">
        <div className="p-8 space-y-6">
          <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 pb-6 border-b border-slate-100">
            <div>
              <div className="flex items-center gap-3">
                <span className="text-xs font-semibold tracking-wider uppercase text-indigo-600 bg-indigo-50 px-2.5 py-1 rounded-md">
                  {resource.Type}
                </span>
                <StatusBadge status={resource.Status} />
              </div>
              <h1 className="text-2xl font-bold text-slate-900 mt-2 tracking-tight">
                {resource.Name}
              </h1>
            </div>

            {isAdmin && (
              <div className="flex items-center gap-2">
                <button
                  onClick={() => setIsEditOpen(true)}
                  className="inline-flex items-center gap-2 px-3.5 py-2 text-sm font-semibold text-slate-700 bg-slate-100 hover:bg-slate-200 rounded-lg transition"
                >
                  <Edit2 className="w-4 h-4" /> Edit
                </button>
                <button
                  onClick={() => setIsDeleteOpen(true)}
                  className="inline-flex items-center gap-2 px-3.5 py-2 text-sm font-semibold text-rose-600 bg-rose-50 hover:bg-rose-100 rounded-lg transition"
                >
                  <Trash2 className="w-4 h-4" /> Delete
                </button>
              </div>
            )}
          </div>

          {/* Key Properties Grid */}
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-6">
            <div className="p-4 bg-slate-50 rounded-xl border border-slate-100 space-y-1">
              <span className="text-xs font-medium text-slate-500 flex items-center gap-1.5">
                <MapPin className="w-3.5 h-3.5 text-slate-400" /> Location
              </span>
              <p className="text-sm font-bold text-slate-800">
                {resource.Location || 'Not specified'}
              </p>
            </div>

            <div className="p-4 bg-slate-50 rounded-xl border border-slate-100 space-y-1">
              <span className="text-xs font-medium text-slate-500 flex items-center gap-1.5">
                <Users className="w-3.5 h-3.5 text-slate-400" /> Capacity
              </span>
              <p className="text-sm font-bold text-slate-800">
                {resource.Capacity} seats
              </p>
            </div>

            <div className="p-4 bg-slate-50 rounded-xl border border-slate-100 space-y-1">
              <span className="text-xs font-medium text-slate-500 flex items-center gap-1.5">
                <Building className="w-3.5 h-3.5 text-slate-400" /> Facility ID
              </span>
              <p className="text-xs font-mono font-semibold text-slate-600 truncate">
                {resource.ID}
              </p>
            </div>
          </div>

          {/* Quick Actions Card */}
          <div className="pt-4 border-t border-slate-100 flex flex-wrap gap-4">
            {user?.role !== 'STUDENT' && (
              <Link
                to={`/bookings/create?resource_id=${resource.ID}`}
                className={`inline-flex items-center gap-2 px-5 py-2.5 rounded-lg text-sm font-semibold text-white shadow-xs transition ${
                  resource.Status === 'MAINTENANCE'
                    ? 'bg-slate-400 cursor-not-allowed pointer-events-none'
                    : 'bg-indigo-600 hover:bg-indigo-700'
                }`}
              >
                <CalendarPlus className="w-4 h-4" />
                {resource.Status === 'MAINTENANCE'
                  ? 'Under Maintenance'
                  : 'Reserve this Resource'}
              </Link>
            )}

            <Link
              to={`/issues/create?resource_id=${resource.ID}`}
              className="inline-flex items-center gap-2 px-5 py-2.5 rounded-lg text-sm font-semibold text-slate-700 bg-slate-100 hover:bg-slate-200 transition"
            >
              <AlertTriangle className="w-4 h-4 text-amber-600" />
              Report Fault / Issue
            </Link>
          </div>
        </div>
      </div>

      {/* Admin Edit Modal */}
      <ResourceModal
        isOpen={isEditOpen}
        resource={resource}
        onClose={() => setIsEditOpen(false)}
        onSave={handleEditSave}
      />

      {/* Admin Delete Confirmation */}
      <ConfirmDialog
        isOpen={isDeleteOpen}
        title="Delete Resource"
        message={`Are you sure you want to delete "${resource.Name}"? This action cannot be undone.`}
        confirmText="Delete"
        isDangerous={true}
        isLoading={isDeleting}
        onConfirm={handleDeleteConfirm}
        onCancel={() => setIsDeleteOpen(false)}
      />
    </div>
  )
}
