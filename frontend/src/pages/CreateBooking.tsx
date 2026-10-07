import React, { useState, useEffect } from 'react'
import { useNavigate, useSearchParams, Link } from 'react-router-dom'
import { createBooking } from '../api/bookings'
import { useResources } from '../hooks/useResources'
import {
  CalendarPlus,
  ArrowLeft,
  AlertCircle,
  Loader2,
  Building,
  CheckCircle2,
} from 'lucide-react'

export const CreateBooking: React.FC = () => {
  const [searchParams] = useSearchParams()
  const initialResourceId = searchParams.get('resource_id') || ''

  const navigate = useNavigate()
  const { resources, getResourceById, isLoading: isLoadingResources } = useResources()

  const [resourceId, setResourceId] = useState(initialResourceId)
  const [bookingDate, setBookingDate] = useState(() => {
    const today = new Date()
    return today.toISOString().split('T')[0]
  })
  const [startTimeStr, setStartTimeStr] = useState('09:00')
  const [endTimeStr, setEndTimeStr] = useState('10:00')

  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [success, setSuccess] = useState(false)

  // Update resourceId if resources loaded later or param changed
  useEffect(() => {
    if (!resourceId && resources.length > 0) {
      const firstAvailable = resources.find((r) => r.Status === 'AVAILABLE')
      if (firstAvailable) {
        setResourceId(firstAvailable.ID)
      }
    }
  }, [resources, resourceId])

  const selectedResource = resourceId ? getResourceById(resourceId) : undefined

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError(null)
    setSuccess(false)

    if (!resourceId) {
      setError('Please select a facility resource.')
      return
    }

    if (selectedResource && selectedResource.Status === 'MAINTENANCE') {
      setError('This facility is currently under maintenance and cannot be reserved.')
      return
    }

    // Build ISO timestamps
    const startIso = new Date(`${bookingDate}T${startTimeStr}:00`).toISOString()
    const endIso = new Date(`${bookingDate}T${endTimeStr}:00`).toISOString()

    if (new Date(startIso) >= new Date(endIso)) {
      setError('End time must be later than start time.')
      return
    }

    if (new Date(startIso) < new Date()) {
      setError('Cannot make a reservation in the past.')
      return
    }

    setIsSubmitting(true)

    try {
      await createBooking({
        resource_id: resourceId,
        start_time: startIso,
        end_time: endIso,
      })
      setSuccess(true)
      setTimeout(() => {
        navigate('/bookings/mine')
      }, 1500)
    } catch (err: unknown) {
      if (err instanceof Error) {
        setError(err.message)
      } else {
        setError('Failed to create booking reservation.')
      }
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <div className="max-w-2xl space-y-6">
      <div>
        <Link
          to="/resources"
          className="inline-flex items-center gap-2 text-sm font-semibold text-slate-500 hover:text-slate-900 transition"
        >
          <ArrowLeft className="w-4 h-4" /> Back to Facilities
        </Link>
      </div>

      <div className="bg-white rounded-2xl border border-slate-200/80 shadow-xs p-8 space-y-6">
        <div className="flex items-center gap-3 pb-6 border-b border-slate-100">
          <div className="w-10 h-10 rounded-xl bg-indigo-50 text-indigo-600 flex items-center justify-center font-bold">
            <CalendarPlus className="w-5 h-5" />
          </div>
          <div>
            <h1 className="text-xl font-bold text-slate-900 tracking-tight">
              Reserve Facility
            </h1>
            <p className="text-xs text-slate-500">
              Schedule time in classrooms, computer labs, or seminar halls
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
              Booking confirmed successfully! Redirecting to your reservations...
            </p>
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-5">
          {/* Resource Selector */}
          <div>
            <label className="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-1.5">
              Select Facility Resource
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
                  <option
                    key={r.ID}
                    value={r.ID}
                    disabled={r.Status === 'MAINTENANCE'}
                  >
                    {r.Name} ({r.Type}) - {r.Location} [Capacity: {r.Capacity}]
                    {r.Status === 'MAINTENANCE' ? ' — UNDER MAINTENANCE' : ''}
                  </option>
                ))}
              </select>
            )}
          </div>

          {selectedResource && (
            <div className="p-3 bg-slate-50 rounded-xl border border-slate-100 flex items-center gap-3 text-xs text-slate-600">
              <Building className="w-4 h-4 text-slate-400" />
              <span>
                Capacity: <strong>{selectedResource.Capacity} persons</strong> · Location:{' '}
                <strong>{selectedResource.Location}</strong>
              </span>
            </div>
          )}

          {/* Date Picker */}
          <div>
            <label className="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-1.5">
              Reservation Date
            </label>
            <input
              type="date"
              required
              value={bookingDate}
              min={new Date().toISOString().split('T')[0]}
              onChange={(e) => setBookingDate(e.target.value)}
              className="w-full px-3.5 py-2.5 bg-white border border-slate-200 rounded-lg text-sm text-slate-900 focus:outline-none focus:ring-2 focus:ring-indigo-600 focus:border-transparent transition"
            />
          </div>

          {/* Time Slot Range */}
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-1.5">
                Start Time
              </label>
              <input
                type="time"
                required
                value={startTimeStr}
                onChange={(e) => setStartTimeStr(e.target.value)}
                className="w-full px-3.5 py-2.5 bg-white border border-slate-200 rounded-lg text-sm text-slate-900 focus:outline-none focus:ring-2 focus:ring-indigo-600 focus:border-transparent transition"
              />
            </div>

            <div>
              <label className="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-1.5">
                End Time
              </label>
              <input
                type="time"
                required
                value={endTimeStr}
                onChange={(e) => setEndTimeStr(e.target.value)}
                className="w-full px-3.5 py-2.5 bg-white border border-slate-200 rounded-lg text-sm text-slate-900 focus:outline-none focus:ring-2 focus:ring-indigo-600 focus:border-transparent transition"
              />
            </div>
          </div>

          <div className="pt-4 border-t border-slate-100 flex items-center justify-end gap-3">
            <Link
              to="/resources"
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
              <span>Confirm Reservation</span>
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
