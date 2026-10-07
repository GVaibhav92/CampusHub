import React, { useState, useEffect } from 'react'
import { useNavigate, useSearchParams, Link } from 'react-router-dom'
import { createIssue } from '../api/issues'
import { useResources } from '../hooks/useResources'
import type { IssuePriority } from '../types'
import {
  AlertTriangle,
  ArrowLeft,
  AlertCircle,
  Loader2,
  CheckCircle2,
} from 'lucide-react'

export const CreateIssue: React.FC = () => {
  const [searchParams] = useSearchParams()
  const initialResourceId = searchParams.get('resource_id') || ''

  const navigate = useNavigate()
  const { resources, isLoading: isLoadingResources } = useResources()

  const [resourceId, setResourceId] = useState(initialResourceId)
  const [category, setCategory] = useState('ELECTRICAL')
  const [priority, setPriority] = useState<IssuePriority>('MEDIUM')
  const [description, setDescription] = useState('')

  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [success, setSuccess] = useState(false)

  useEffect(() => {
    if (!resourceId && resources.length > 0) {
      setResourceId(resources[0].ID)
    }
  }, [resources, resourceId])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError(null)
    setSuccess(false)

    if (!resourceId) {
      setError('Please select a facility resource.')
      return
    }

    if (!description.trim()) {
      setError('Please provide a detailed description of the issue.')
      return
    }

    setIsSubmitting(true)

    try {
      await createIssue({
        resource_id: resourceId,
        category,
        priority,
        description,
      })
      setSuccess(true)
      setTimeout(() => {
        navigate('/issues/mine')
      }, 1500)
    } catch (err: unknown) {
      if (err instanceof Error) {
        setError(err.message)
      } else {
        setError('Failed to report issue.')
      }
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <div className="max-w-2xl space-y-6">
      <div>
        <Link
          to="/issues/mine"
          className="inline-flex items-center gap-2 text-sm font-semibold text-slate-500 hover:text-slate-900 transition"
        >
          <ArrowLeft className="w-4 h-4" /> Back to Issues
        </Link>
      </div>

      <div className="bg-white rounded-2xl border border-slate-200/80 shadow-xs p-8 space-y-6">
        <div className="flex items-center gap-3 pb-6 border-b border-slate-100">
          <div className="w-10 h-10 rounded-xl bg-amber-50 text-amber-600 flex items-center justify-center font-bold">
            <AlertTriangle className="w-5 h-5" />
          </div>
          <div>
            <h1 className="text-xl font-bold text-slate-900 tracking-tight">
              Report Campus Issue
            </h1>
            <p className="text-xs text-slate-500">
              Submit facility or equipment faults for maintenance technician triage
            </p>
          </div>
        </div>

        {error && (
          <div className="p-4 rounded-xl bg-rose-50 border border-rose-200 flex items-start gap-3">
            <AlertCircle className="w-5 h-5 text-rose-600 shrink-0 mt-0.5" />
            <p className="text-sm font-medium text-rose-800">{error}</p>
          </div>
        )}

        {success && (
          <div className="p-4 rounded-xl bg-emerald-50 border border-emerald-200 flex items-center gap-3">
            <CheckCircle2 className="w-5 h-5 text-emerald-600 shrink-0" />
            <p className="text-sm font-medium text-emerald-800">
              Issue reported successfully! Maintenance team has been notified.
            </p>
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-5">
          {/* Facility Selector */}
          <div>
            <label className="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-1.5">
              Impacted Facility / Resource
            </label>
            {isLoadingResources ? (
              <div className="flex items-center gap-2 text-xs text-slate-400 py-2">
                <Loader2 className="w-3.5 h-3.5 animate-spin" />
                Loading facilities...
              </div>
            ) : (
              <select
                value={resourceId}
                onChange={(e) => setResourceId(e.target.value)}
                required
                className="w-full px-3.5 py-2.5 bg-white border border-slate-200 rounded-lg text-sm text-slate-900 focus:outline-none focus:ring-2 focus:ring-indigo-600 focus:border-transparent transition"
              >
                {resources.map((r) => (
                  <option key={r.ID} value={r.ID}>
                    {r.Name} ({r.Type}) - {r.Location}
                  </option>
                ))}
              </select>
            )}
          </div>

          {/* Category & Priority Grid */}
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-1.5">
                Category
              </label>
              <select
                value={category}
                onChange={(e) => setCategory(e.target.value)}
                className="w-full px-3.5 py-2.5 bg-white border border-slate-200 rounded-lg text-sm text-slate-900 focus:outline-none focus:ring-2 focus:ring-indigo-600 focus:border-transparent transition"
              >
                <option value="ELECTRICAL">Electrical / Lighting</option>
                <option value="PLUMBING">Plumbing / Water</option>
                <option value="PROJECTOR_AV">Audio / Video / Projector</option>
                <option value="HVAC">Air Conditioning / HVAC</option>
                <option value="FURNITURE">Furniture / Seating</option>
                <option value="IT_NETWORK">IT / Network / PC</option>
                <option value="OTHER">Other Facility Maintenance</option>
              </select>
            </div>

            <div>
              <label className="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-1.5">
                Priority Level
              </label>
              <select
                value={priority}
                onChange={(e) => setPriority(e.target.value as IssuePriority)}
                className="w-full px-3.5 py-2.5 bg-white border border-slate-200 rounded-lg text-sm text-slate-900 focus:outline-none focus:ring-2 focus:ring-indigo-600 focus:border-transparent transition"
              >
                <option value="LOW">Low - Minor defect</option>
                <option value="MEDIUM">Medium - Impairs normal usage</option>
                <option value="HIGH">High - Urgent / Facility unusable</option>
              </select>
            </div>
          </div>

          {/* Description */}
          <div>
            <label className="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-1.5">
              Issue Description
            </label>
            <textarea
              required
              rows={4}
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Describe the defect, exact room location, or error message..."
              className="w-full px-3.5 py-2.5 bg-white border border-slate-200 rounded-lg text-sm text-slate-900 placeholder:text-slate-400 focus:outline-none focus:ring-2 focus:ring-indigo-600 focus:border-transparent transition"
            />
          </div>

          <div className="pt-4 border-t border-slate-100 flex items-center justify-end gap-3">
            <Link
              to="/issues/mine"
              className="px-4 py-2 text-sm font-medium text-slate-700 bg-slate-100 hover:bg-slate-200 rounded-lg transition"
            >
              Cancel
            </Link>
            <button
              type="submit"
              disabled={isSubmitting || success}
              className="px-5 py-2.5 text-sm font-semibold text-white bg-indigo-600 hover:bg-indigo-700 rounded-lg transition flex items-center gap-2 disabled:opacity-50 shadow-xs"
            >
              {isSubmitting && <Loader2 className="w-4 h-4 animate-spin" />}
              <span>Submit Report</span>
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
