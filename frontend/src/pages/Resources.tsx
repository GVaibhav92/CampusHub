import React, { useState, useMemo } from 'react'
import { Link } from 'react-router-dom'
import { useResources } from '../hooks/useResources'
import { useAuth } from '../hooks/useAuth'
import { StatusBadge } from '../components/common/StatusBadge'
import { LoadingSpinner } from '../components/common/LoadingSpinner'
import { EmptyState } from '../components/common/EmptyState'
import { ErrorState } from '../components/common/ErrorState'
import { ResourceModal } from '../components/resources/ResourceModal'
import { ConfirmDialog } from '../components/common/ConfirmDialog'
import type { Resource } from '../types'
import {
  Search,
  Plus,
  Edit2,
  Trash2,
  Users,
  MapPin,
  ArrowRight,
  RefreshCw,
} from 'lucide-react'

export const Resources: React.FC = () => {
  const { user } = useAuth()
  const {
    resources,
    isLoading,
    error,
    refreshResources,
    addResource,
    editResource,
    removeResource,
  } = useResources()

  const isAdmin = user?.role === 'ADMIN'

  // Search & Type Filter State
  const [searchTerm, setSearchTerm] = useState('')
  const [typeFilter, setTypeFilter] = useState('ALL')

  // Modal & Dialog State
  const [isModalOpen, setIsModalOpen] = useState(false)
  const [editingResource, setEditingResource] = useState<Resource | null>(null)

  const [deletingResource, setDeletingResource] = useState<Resource | null>(null)
  const [isDeleting, setIsDeleting] = useState(false)
  const [actionError, setActionError] = useState<string | null>(null)

  const filteredResources = useMemo(() => {
    return resources.filter((r) => {
      const matchesSearch =
        r.Name.toLowerCase().includes(searchTerm.toLowerCase()) ||
        r.Location.toLowerCase().includes(searchTerm.toLowerCase())
      const matchesType = typeFilter === 'ALL' || r.Type === typeFilter
      return matchesSearch && matchesType
    })
  }, [resources, searchTerm, typeFilter])

  const handleCreateOrEdit = async (payload: {
    name: string
    type: string
    capacity: number
    location: string
    status?: any
  }) => {
    if (editingResource) {
      await editResource(editingResource.ID, {
        name: payload.name,
        type: payload.type,
        capacity: payload.capacity,
        location: payload.location,
        status: payload.status || editingResource.Status,
      })
    } else {
      await addResource({
        name: payload.name,
        type: payload.type,
        capacity: payload.capacity,
        location: payload.location,
      })
    }
  }

  const handleDeleteConfirm = async () => {
    if (!deletingResource) return
    setIsDeleting(true)
    setActionError(null)
    try {
      await removeResource(deletingResource.ID)
      setDeletingResource(null)
    } catch (err: unknown) {
      if (err instanceof Error) {
        setActionError(err.message)
      } else {
        setActionError('Failed to delete resource.')
      }
    } finally {
      setIsDeleting(false)
    }
  }

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-slate-900">
            Campus Facilities & Resources
          </h1>
          <p className="mt-1 text-sm text-slate-500">
            Browse and reserve rooms, laboratories, and halls across campus
          </p>
        </div>

        <div className="flex items-center gap-3">
          <button
            onClick={() => refreshResources()}
            title="Refresh list"
            className="p-2.5 bg-white border border-slate-200 rounded-lg text-slate-600 hover:text-slate-900 hover:bg-slate-50 transition"
          >
            <RefreshCw className="w-4 h-4" />
          </button>

          {isAdmin && (
            <button
              onClick={() => {
                setEditingResource(null)
                setIsModalOpen(true)
              }}
              className="inline-flex items-center gap-2 px-4 py-2.5 bg-indigo-600 text-white rounded-lg text-sm font-semibold hover:bg-indigo-700 transition shadow-xs"
            >
              <Plus className="w-4 h-4" />
              Add Resource
            </button>
          )}
        </div>
      </div>

      {actionError && (
        <div className="p-4 bg-rose-50 border border-rose-200 rounded-xl text-sm font-semibold text-rose-800 flex items-center justify-between">
          <span>{actionError}</span>
          <button
            onClick={() => setActionError(null)}
            className="text-xs text-rose-600 underline ml-4 hover:text-rose-800"
          >
            Dismiss
          </button>
        </div>
      )}

      {/* Filter & Search Bar */}
      <div className="bg-white p-4 rounded-xl border border-slate-200/80 shadow-xs flex flex-col sm:flex-row gap-3">
        <div className="relative flex-1">
          <Search className="w-4 h-4 text-slate-400 absolute left-3.5 top-1/2 -translate-y-1/2" />
          <input
            type="text"
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            placeholder="Search resources by name or location..."
            className="w-full pl-9 pr-4 py-2 bg-slate-50 border border-slate-200 rounded-lg text-sm text-slate-900 placeholder:text-slate-400 focus:outline-none focus:ring-2 focus:ring-indigo-600 focus:bg-white transition"
          />
        </div>

        <select
          value={typeFilter}
          onChange={(e) => setTypeFilter(e.target.value)}
          className="px-3 py-2 bg-slate-50 border border-slate-200 rounded-lg text-sm text-slate-700 focus:outline-none focus:ring-2 focus:ring-indigo-600 focus:bg-white transition"
        >
          <option value="ALL">All Facility Types</option>
          <option value="CLASSROOM">Classrooms</option>
          <option value="LAB">Laboratories</option>
          <option value="AUDITORIUM">Auditoriums</option>
          <option value="SEMINAR_HALL">Seminar Halls</option>
          <option value="EQUIPMENT">Equipment</option>
        </select>
      </div>

      {/* Content Area */}
      {isLoading ? (
        <LoadingSpinner message="Loading campus resources..." />
      ) : error ? (
        <ErrorState message={error} onRetry={refreshResources} />
      ) : filteredResources.length === 0 ? (
        <EmptyState
          title="No resources found"
          description={
            searchTerm || typeFilter !== 'ALL'
              ? 'Try modifying your search or type filters.'
              : 'There are currently no resources registered for your campus.'
          }
          action={
            isAdmin
              ? {
                  label: 'Add First Resource',
                  onClick: () => {
                    setEditingResource(null)
                    setIsModalOpen(true)
                  },
                }
              : undefined
          }
        />
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {filteredResources.map((res) => (
            <div
              key={res.ID}
              className="bg-white rounded-2xl border border-slate-200/80 shadow-xs hover:border-indigo-200 transition flex flex-col justify-between overflow-hidden group"
            >
              <div className="p-6 space-y-4">
                <div className="flex items-start justify-between gap-2">
                  <div>
                    <span className="text-[11px] font-semibold tracking-wider uppercase text-indigo-600 bg-indigo-50 px-2 py-0.5 rounded-md">
                      {res.Type}
                    </span>
                    <h3 className="text-lg font-bold text-slate-900 mt-2 tracking-tight group-hover:text-indigo-600 transition">
                      {res.Name}
                    </h3>
                  </div>
                  <StatusBadge status={res.Status} />
                </div>

                <div className="space-y-2 text-xs text-slate-500 pt-1">
                  <div className="flex items-center gap-2">
                    <MapPin className="w-3.5 h-3.5 text-slate-400 shrink-0" />
                    <span className="truncate">{res.Location || 'Campus Main Grounds'}</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <Users className="w-3.5 h-3.5 text-slate-400 shrink-0" />
                    <span>Capacity: {res.Capacity} persons</span>
                  </div>
                </div>
              </div>

              {/* Action Footer */}
              <div className="px-6 py-3.5 bg-slate-50 border-t border-slate-100 flex items-center justify-between">
                <Link
                  to={`/resources/${res.ID}`}
                  className="text-xs font-semibold text-indigo-600 hover:text-indigo-700 flex items-center gap-1 group-hover:translate-x-0.5 transition"
                >
                  View Details <ArrowRight className="w-3.5 h-3.5" />
                </Link>

                {isAdmin && (
                  <div className="flex items-center gap-1">
                    <button
                      onClick={() => {
                        setEditingResource(res)
                        setIsModalOpen(true)
                      }}
                      title="Edit Resource"
                      className="p-1.5 rounded-md text-slate-500 hover:text-indigo-600 hover:bg-white transition"
                    >
                      <Edit2 className="w-3.5 h-3.5" />
                    </button>
                    <button
                      onClick={() => setDeletingResource(res)}
                      title="Delete Resource"
                      className="p-1.5 rounded-md text-slate-500 hover:text-rose-600 hover:bg-white transition"
                    >
                      <Trash2 className="w-3.5 h-3.5" />
                    </button>
                  </div>
                )}
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Admin Create/Edit Modal */}
      <ResourceModal
        isOpen={isModalOpen}
        resource={editingResource}
        onClose={() => setIsModalOpen(false)}
        onSave={handleCreateOrEdit}
      />

      {/* Delete Confirmation */}
      <ConfirmDialog
        isOpen={!!deletingResource}
        title="Delete Resource"
        message={`Are you sure you want to delete "${deletingResource?.Name}"? This action cannot be undone.`}
        confirmText="Delete"
        isDangerous={true}
        isLoading={isDeleting}
        onConfirm={handleDeleteConfirm}
        onCancel={() => setDeletingResource(null)}
      />
    </div>
  )
}
