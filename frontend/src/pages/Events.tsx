import React, { useState, useEffect, useMemo, useCallback } from 'react'
import { getEvents, getMyEnrolledEvents, enrollInEvent, unenrollFromEvent } from '../api/events'
import { useAuth } from '../hooks/useAuth'
import type { EventSession } from '../types'
import { formatDateTime } from '../lib/utils'
import { EventRosterModal } from '../components/roster/EventRosterModal'
import {
  CalendarDays,
  Search,
  RefreshCw,
  Clock,
  Users,
  User,
  CheckCircle2,
  XCircle,
  Loader2,
  AlertCircle,
  Building,
} from 'lucide-react'
import { toast } from 'sonner'

export const Events: React.FC = () => {
  const { user } = useAuth()
  const isStudent = user?.role === 'STUDENT'
  const isFacultyOrAdmin = user?.role === 'FACULTY' || user?.role === 'ADMIN'

  const [events, setEvents] = useState<EventSession[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [searchTerm, setSearchTerm] = useState('')
  const [tab, setTab] = useState<'ALL' | 'ENROLLED'>('ALL')
  const [actionLoadingId, setActionLoadingId] = useState<string | null>(null)

  // Roster Modal for Faculty/Admin
  const [rosterEvent, setRosterEvent] = useState<EventSession | null>(null)

  const fetchEvents = useCallback(async () => {
    setIsLoading(true)
    setError(null)
    try {
      const data = tab === 'ENROLLED' && isStudent ? await getMyEnrolledEvents() : await getEvents()
      setEvents(Array.isArray(data) ? data : [])
    } catch (err: unknown) {
      if (err instanceof Error) {
        setError(err.message)
      } else {
        setError('Failed to load events.')
      }
    } finally {
      setIsLoading(false)
    }
  }, [tab, isStudent])

  useEffect(() => {
    fetchEvents()
  }, [fetchEvents])

  const filteredEvents = useMemo(() => {
    return events.filter((ev) => {
      const term = searchTerm.toLowerCase()
      const title = (ev.title || '').toLowerCase()
      const course = (ev.course_code || '').toLowerCase()
      const faculty = (ev.faculty_email || '').toLowerCase()
      const location = (ev.resource_location || '').toLowerCase()
      const resource = (ev.resource_name || '').toLowerCase()

      return (
        title.includes(term) ||
        course.includes(term) ||
        faculty.includes(term) ||
        location.includes(term) ||
        resource.includes(term)
      )
    })
  }, [events, searchTerm])

  const handleEnroll = async (event: EventSession) => {
    setActionLoadingId(event.id)
    try {
      await enrollInEvent(event.id)
      toast.success(`Successfully enrolled in ${event.title}!`)
      // Refresh event list to get updated count and enrollment flag
      await fetchEvents()
    } catch (err: unknown) {
      if (err instanceof Error) {
        toast.error(err.message)
      } else {
        toast.error('Failed to enroll.')
      }
    } finally {
      setActionLoadingId(null)
    }
  }

  const handleUnenroll = async (event: EventSession) => {
    const confirmed = window.confirm(
      `Cancel your RSVP for "${event.title}"? Your seat will be made available to other students.`
    )
    if (!confirmed) return

    setActionLoadingId(event.id)
    try {
      await unenrollFromEvent(event.id)
      toast.success(`RSVP cancelled for ${event.title}`)
      await fetchEvents()
    } catch (err: unknown) {
      if (err instanceof Error) {
        toast.error(err.message)
      } else {
        toast.error('Failed to cancel enrollment.')
      }
    } finally {
      setActionLoadingId(null)
    }
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-slate-900">
            {isStudent ? 'Events & Lectures' : 'Campus Events & Timetable'}
          </h1>
          <p className="mt-1 text-sm text-slate-500">
            {isStudent
              ? 'Discover upcoming classes, guest lectures, and reserve your seat'
              : 'Overview of scheduled academic sessions across campus facilities'}
          </p>
        </div>

        <div className="flex items-center gap-3">
          <button
            onClick={() => fetchEvents()}
            title="Refresh events"
            className="p-2.5 bg-white border border-slate-200 rounded-lg text-slate-600 hover:text-slate-900 hover:bg-slate-50 transition"
          >
            <RefreshCw className="w-4 h-4" />
          </button>
        </div>
      </div>

      {/* Tabs & Search Bar */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        {/* Search Input */}
        <div className="relative flex-1 max-w-md">
          <Search className="w-4 h-4 text-slate-400 absolute left-3.5 top-1/2 -translate-y-1/2" />
          <input
            type="text"
            placeholder="Search by event title, course code, or professor..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="w-full pl-10 pr-4 py-2 bg-white border border-slate-200 rounded-lg text-sm text-slate-900 focus:outline-none focus:ring-2 focus:ring-indigo-600 focus:border-transparent transition"
          />
        </div>

        {/* Filter Tabs (for Students) */}
        {isStudent && (
          <div className="flex items-center gap-1.5 p-1 bg-slate-100/80 rounded-xl border border-slate-200/60 text-xs font-semibold">
            <button
              onClick={() => setTab('ALL')}
              className={`px-3 py-1.5 rounded-lg transition ${
                tab === 'ALL'
                  ? 'bg-white text-indigo-700 shadow-xs'
                  : 'text-slate-600 hover:text-slate-900'
              }`}
            >
              All Events
            </button>
            <button
              onClick={() => setTab('ENROLLED')}
              className={`px-3 py-1.5 rounded-lg transition ${
                tab === 'ENROLLED'
                  ? 'bg-white text-indigo-700 shadow-xs'
                  : 'text-slate-600 hover:text-slate-900'
              }`}
            >
              My Enrolled
            </button>
          </div>
        )}
      </div>

      {error && (
        <div className="p-4 rounded-xl bg-rose-50 border border-rose-200 flex items-start gap-3">
          <AlertCircle className="w-5 h-5 text-rose-600 shrink-0 mt-0.5" />
          <p className="text-sm font-semibold text-rose-800">{error}</p>
        </div>
      )}

      {/* Events Grid */}
      {isLoading ? (
        <div className="py-16 flex flex-col items-center justify-center gap-2 text-slate-400">
          <Loader2 className="w-8 h-8 animate-spin text-indigo-600" />
          <p className="text-sm font-medium">Loading scheduled campus sessions...</p>
        </div>
      ) : filteredEvents.length === 0 ? (
        <div className="py-16 bg-white rounded-2xl border border-slate-200/80 text-center space-y-3">
          <CalendarDays className="w-10 h-10 mx-auto text-slate-300" />
          <h3 className="text-base font-bold text-slate-800">No scheduled events found</h3>
          <p className="text-xs text-slate-400 max-w-sm mx-auto">
            {tab === 'ENROLLED'
              ? "You haven't enrolled in any upcoming sessions yet. Switch to 'All Events' to browse."
              : 'There are currently no active events matching your search criteria.'}
          </p>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {filteredEvents.map((ev) => {
            const isFull = ev.active_enrolled >= ev.capacity
            const percentFilled =
              ev.capacity > 0 ? Math.min(100, Math.round((ev.active_enrolled / ev.capacity) * 100)) : 0
            const isActionLoading = actionLoadingId === ev.id

            return (
              <div
                key={ev.id}
                className={`bg-white rounded-2xl border transition flex flex-col justify-between overflow-hidden shadow-xs hover:border-indigo-200 ${
                  ev.is_enrolled ? 'border-indigo-300 ring-2 ring-indigo-50' : 'border-slate-200/80'
                }`}
              >
                <div className="p-6 space-y-4">
                  {/* Top Badges */}
                  <div className="flex items-start justify-between gap-2">
                    <div className="space-y-1">
                      <div className="flex flex-wrap items-center gap-1.5">
                        {ev.course_code && (
                          <span className="text-[11px] font-bold text-indigo-700 bg-indigo-50 px-2 py-0.5 rounded-md uppercase">
                            {ev.course_code}
                          </span>
                        )}
                        <span className="text-[11px] font-medium text-slate-500 bg-slate-100 px-2 py-0.5 rounded-md">
                          Confirmed Session
                        </span>
                      </div>
                      <h3 className="text-lg font-bold text-slate-900 tracking-tight pt-1">
                        {ev.title}
                      </h3>
                    </div>

                    {ev.is_enrolled && (
                      <span className="inline-flex items-center gap-1 text-[11px] font-bold text-emerald-700 bg-emerald-50 px-2.5 py-1 rounded-full shrink-0">
                        <CheckCircle2 className="w-3.5 h-3.5" /> Enrolled
                      </span>
                    )}
                  </div>

                  {/* Event Details */}
                  <div className="space-y-2.5 text-xs text-slate-600 pt-1">
                    <div className="flex items-center gap-2">
                      <User className="w-3.5 h-3.5 text-slate-400 shrink-0" />
                      <span className="font-medium text-slate-700 truncate">{ev.faculty_email}</span>
                    </div>

                    <div className="flex items-center gap-2">
                      <Building className="w-3.5 h-3.5 text-slate-400 shrink-0" />
                      <span className="truncate">
                        <strong>{ev.resource_name}</strong> · {ev.resource_location}
                      </span>
                    </div>

                    <div className="flex items-center gap-2">
                      <Clock className="w-3.5 h-3.5 text-slate-400 shrink-0" />
                      <span>
                        {formatDateTime(ev.start_time)} – {new Date(ev.end_time).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                      </span>
                    </div>
                  </div>

                  {/* Capacity Bar */}
                  <div className="pt-2">
                    <div className="flex items-center justify-between text-xs font-semibold mb-1">
                      <span className="text-slate-500 flex items-center gap-1">
                        <Users className="w-3.5 h-3.5 text-slate-400" /> Seats Available
                      </span>
                      <span className={isFull ? 'text-rose-600 font-bold' : 'text-slate-700'}>
                        {ev.active_enrolled} / {ev.capacity} filled {isFull && '(FULL)'}
                      </span>
                    </div>
                    <div className="w-full bg-slate-100 rounded-full h-1.5 overflow-hidden">
                      <div
                        className={`h-full rounded-full transition-all duration-300 ${
                          isFull
                            ? 'bg-rose-500'
                            : percentFilled >= 80
                            ? 'bg-amber-500'
                            : 'bg-indigo-600'
                        }`}
                        style={{ width: `${percentFilled}%` }}
                      />
                    </div>
                  </div>
                </div>

                {/* Card Actions Footer */}
                <div className="px-6 py-3.5 bg-slate-50 border-t border-slate-100 flex items-center justify-between">
                  {isStudent && (
                    <>
                      {ev.is_enrolled ? (
                        <button
                          onClick={() => handleUnenroll(ev)}
                          disabled={isActionLoading}
                          className="w-full inline-flex items-center justify-center gap-1.5 py-2 px-4 rounded-lg text-xs font-semibold text-rose-600 hover:text-rose-700 hover:bg-rose-50 border border-rose-200 transition disabled:opacity-50"
                        >
                          {isActionLoading ? (
                            <Loader2 className="w-3.5 h-3.5 animate-spin" />
                          ) : (
                            <XCircle className="w-3.5 h-3.5" />
                          )}
                          Cancel RSVP
                        </button>
                      ) : isFull ? (
                        <button
                          disabled
                          className="w-full py-2 px-4 rounded-lg text-xs font-semibold text-slate-400 bg-slate-100 cursor-not-allowed text-center"
                        >
                          Capacity Full
                        </button>
                      ) : (
                        <button
                          onClick={() => handleEnroll(ev)}
                          disabled={isActionLoading}
                          className="w-full inline-flex items-center justify-center gap-1.5 py-2 px-4 rounded-lg text-xs font-semibold text-white bg-indigo-600 hover:bg-indigo-700 transition shadow-xs disabled:opacity-50"
                        >
                          {isActionLoading ? (
                            <Loader2 className="w-3.5 h-3.5 animate-spin" />
                          ) : (
                            <CheckCircle2 className="w-3.5 h-3.5" />
                          )}
                          Enroll in Event
                        </button>
                      )}
                    </>
                  )}

                  {isFacultyOrAdmin && (
                    <button
                      onClick={() => setRosterEvent(ev)}
                      className="w-full inline-flex items-center justify-center gap-1.5 py-2 px-4 rounded-lg text-xs font-semibold text-indigo-700 bg-indigo-50 hover:bg-indigo-100 border border-indigo-200 transition"
                    >
                      <Users className="w-3.5 h-3.5" />
                      View Attendee Roster
                    </button>
                  )}
                </div>
              </div>
            )
          })}
        </div>
      )}

      {/* Roster Modal for Faculty */}
      {rosterEvent && (
        <EventRosterModal
          isOpen={!!rosterEvent}
          onClose={() => setRosterEvent(null)}
          bookingId={rosterEvent.id}
          bookingTitle={rosterEvent.title}
          capacity={rosterEvent.capacity}
        />
      )}
    </div>
  )
}
