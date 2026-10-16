import React, { useState, useMemo } from 'react'
import { Link } from 'react-router-dom'
import { useBookings } from '../hooks/useBookings'
import { useResources } from '../hooks/useResources'
import { StatusBadge } from '../components/common/StatusBadge'
import { LoadingSpinner } from '../components/common/LoadingSpinner'
import { EmptyState } from '../components/common/EmptyState'
import { ErrorState } from '../components/common/ErrorState'
import { ConfirmDialog } from '../components/common/ConfirmDialog'
import { EventRosterModal } from '../components/roster/EventRosterModal'
import { formatDateTime } from '../lib/utils'
import type { Booking } from '../types'
import {
  CalendarPlus,
  RefreshCw,
  Search,
  XCircle,
  Building,
  Clock,
  Users,
} from 'lucide-react'

interface BookingsProps {
  isAll?: boolean
}

export const Bookings: React.FC<BookingsProps> = ({ isAll = false }) => {
  const { bookings, isLoading, error, refresh, cancelBooking } = useBookings(isAll)
  const { getResourceById } = useResources()

  const [searchTerm, setSearchTerm] = useState('')
  const [statusFilter, setStatusFilter] = useState<'ALL' | 'CONFIRMED' | 'CANCELLED'>('ALL')

  const [rosterBooking, setRosterBooking] = useState<Booking | null>(null)
  const [cancelingBooking, setCancelingBooking] = useState<Booking | null>(null)
  const [isCanceling, setIsCanceling] = useState(false)
  const [actionError, setActionError] = useState<string | null>(null)

  const filteredBookings = useMemo(() => {
    return bookings.filter((b) => {
      const resource = getResourceById(b.ResourceID)
      const resName = resource?.Name.toLowerCase() || ''
      const title = (b.Title || '').toLowerCase()
      const course = (b.CourseCode || '').toLowerCase()
      const matchesSearch =
        resName.includes(searchTerm.toLowerCase()) ||
        title.includes(searchTerm.toLowerCase()) ||
        course.includes(searchTerm.toLowerCase()) ||
        b.ID.includes(searchTerm)
      const matchesStatus = statusFilter === 'ALL' || b.Status === statusFilter
      return matchesSearch && matchesStatus
    })
  }, [bookings, searchTerm, statusFilter, getResourceById])

  const handleCancelConfirm = async () => {
    if (!cancelingBooking) return
    setIsCanceling(true)
    setActionError(null)
    try {
      await cancelBooking(cancelingBooking.ID)
      setCancelingBooking(null)
    } catch (err: unknown) {
      if (err instanceof Error) {
        setActionError(err.message)
      } else {
        setActionError('Failed to cancel reservation.')
      }
    } finally {
      setIsCanceling(false)
    }
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-slate-900">
            {isAll ? 'Campus Facility Bookings' : 'My Reservations'}
          </h1>
          <p className="mt-1 text-sm text-slate-500">
            {isAll
              ? 'Tenant-wide log of all facility reservations'
              : 'Review upcoming schedules and manage your bookings'}
          </p>
        </div>

        <div className="flex items-center gap-3">
          <button
            onClick={() => refresh()}
            title="Refresh bookings"
            className="p-2.5 bg-white border border-slate-200 rounded-lg text-slate-600 hover:text-slate-900 hover:bg-slate-50 transition"
          >
            <RefreshCw className="w-4 h-4" />
          </button>

          <Link
            to="/bookings/create"
            className="inline-flex items-center gap-2 px-4 py-2.5 bg-indigo-600 text-white rounded-lg text-sm font-semibold hover:bg-indigo-700 transition shadow-xs"
          >
            <CalendarPlus className="w-4 h-4" />
            New Booking
          </Link>
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

      {/* Filter and Search */}
      <div className="bg-white p-4 rounded-xl border border-slate-200/80 shadow-xs flex flex-col sm:flex-row gap-3">
        <div className="relative flex-1">
          <Search className="w-4 h-4 text-slate-400 absolute left-3.5 top-1/2 -translate-y-1/2" />
          <input
            type="text"
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            placeholder="Search by facility name or booking ID..."
            className="w-full pl-9 pr-4 py-2 bg-slate-50 border border-slate-200 rounded-lg text-sm text-slate-900 placeholder:text-slate-400 focus:outline-none focus:ring-2 focus:ring-indigo-600 focus:bg-white transition"
          />
        </div>

        <select
          value={statusFilter}
          onChange={(e) => setStatusFilter(e.target.value as any)}
          className="px-3 py-2 bg-slate-50 border border-slate-200 rounded-lg text-sm text-slate-700 focus:outline-none focus:ring-2 focus:ring-indigo-600 focus:bg-white transition"
        >
          <option value="ALL">All Statuses</option>
          <option value="CONFIRMED">Confirmed</option>
          <option value="CANCELLED">Cancelled</option>
        </select>
      </div>

      {/* Bookings Table / List */}
      {isLoading ? (
        <LoadingSpinner message="Loading reservations..." />
      ) : error ? (
        <ErrorState message={error} onRetry={refresh} />
      ) : filteredBookings.length === 0 ? (
        <EmptyState
          title="No reservations found"
          description={
            searchTerm || statusFilter !== 'ALL'
              ? 'No bookings match your current filter settings.'
              : isAll
              ? 'No facility reservations exist yet for this campus.'
              : "You haven't reserved any campus facilities yet."
          }
          action={{
            label: 'Make a Reservation',
            onClick: () => (window.location.href = '/bookings/create'),
          }}
        />
      ) : (
        <div className="bg-white rounded-2xl border border-slate-200/80 shadow-xs overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full text-left border-collapse text-sm">
              <thead>
                <tr className="bg-slate-50/80 border-b border-slate-200 text-xs font-semibold text-slate-500 uppercase tracking-wider">
                  <th className="py-3.5 px-6">Event / Session</th>
                  <th className="py-3.5 px-6">Facility</th>
                  <th className="py-3.5 px-6">Start Time</th>
                  <th className="py-3.5 px-6">End Time</th>
                  <th className="py-3.5 px-6">Status</th>
                  {isAll && <th className="py-3.5 px-6">Booked By</th>}
                  <th className="py-3.5 px-6 text-right">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100">
                {filteredBookings.map((b) => {
                  const resource = getResourceById(b.ResourceID)
                  const isPast = new Date(b.EndTime) < new Date()
                  const canCancel = b.Status === 'CONFIRMED' && !isPast

                  return (
                    <tr key={b.ID} className="hover:bg-slate-50/50 transition">
                      <td className="py-4 px-6 font-medium text-slate-900">
                        <div className="space-y-0.5">
                          <span className="font-semibold text-slate-900 block">
                            {b.Title || 'Academic Session'}
                          </span>
                          {b.CourseCode && (
                            <span className="inline-flex text-[11px] font-semibold text-indigo-700 bg-indigo-50 px-2 py-0.5 rounded-md">
                              {b.CourseCode}
                            </span>
                          )}
                        </div>
                      </td>

                      <td className="py-4 px-6 font-medium text-slate-700">
                        <div className="flex items-center gap-2.5">
                          <Building className="w-4 h-4 text-indigo-600 shrink-0" />
                          <div>
                            <span className="font-semibold block text-slate-900">
                              {resource?.Name || 'Facility'}
                            </span>
                            <span className="text-xs text-slate-400 font-mono">
                              {resource ? `${resource.Type} · ${resource.Location}` : b.ResourceID.slice(0, 8)}
                            </span>
                          </div>
                        </div>
                      </td>

                      <td className="py-4 px-6 text-slate-600">
                        <div className="flex items-center gap-1.5 text-xs font-medium">
                          <Clock className="w-3.5 h-3.5 text-slate-400 shrink-0" />
                          <span>{formatDateTime(b.StartTime)}</span>
                        </div>
                      </td>

                      <td className="py-4 px-6 text-slate-600">
                        <div className="flex items-center gap-1.5 text-xs font-medium">
                          <Clock className="w-3.5 h-3.5 text-slate-400 shrink-0" />
                          <span>{formatDateTime(b.EndTime)}</span>
                        </div>
                      </td>

                      <td className="py-4 px-6">
                        <StatusBadge status={b.Status} />
                      </td>

                      {isAll && (
                        <td className="py-4 px-6 text-xs text-slate-500 font-mono">
                          {b.UserID.slice(0, 8)}...
                        </td>
                      )}

                      <td className="py-4 px-6 text-right space-x-1.5">
                        {b.Status === 'CONFIRMED' && (
                          <button
                            onClick={() => setRosterBooking(b)}
                            className="inline-flex items-center gap-1 text-xs font-semibold text-indigo-600 hover:text-indigo-700 hover:bg-indigo-50 px-2.5 py-1.5 rounded-lg transition"
                            title="View student attendee roster"
                          >
                            <Users className="w-3.5 h-3.5" />
                            Roster
                          </button>
                        )}
                        {canCancel && (
                          <button
                            onClick={() => setCancelingBooking(b)}
                            className="inline-flex items-center gap-1 text-xs font-semibold text-rose-600 hover:text-rose-700 hover:bg-rose-50 px-2.5 py-1.5 rounded-lg transition"
                          >
                            <XCircle className="w-3.5 h-3.5" />
                            Cancel
                          </button>
                        )}
                      </td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* Event Attendee Roster Modal */}
      {rosterBooking && (
        <EventRosterModal
          isOpen={!!rosterBooking}
          onClose={() => setRosterBooking(null)}
          bookingId={rosterBooking.ID}
          bookingTitle={rosterBooking.Title || 'Academic Session'}
          capacity={getResourceById(rosterBooking.ResourceID)?.Capacity || 0}
        />
      )}

      {/* Cancel Confirmation Dialog */}
      <ConfirmDialog
        isOpen={!!cancelingBooking}
        title="Cancel Reservation"
        message="Are you sure you want to cancel this booking? The slot will immediately be made available for others to reserve."
        confirmText="Cancel Booking"
        isDangerous={true}
        isLoading={isCanceling}
        onConfirm={handleCancelConfirm}
        onCancel={() => setCancelingBooking(null)}
      />
    </div>
  )
}
