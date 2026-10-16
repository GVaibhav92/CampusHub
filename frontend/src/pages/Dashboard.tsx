import React, { useState, useEffect } from 'react'
import { useAuth } from '../hooks/useAuth'
import { useResources } from '../hooks/useResources'
import { getMyBookings, getAllBookings } from '../api/bookings'
import { getMyIssues, getAllIssues } from '../api/issues'
import { getMyEnrolledEvents } from '../api/events'
import type { Booking, Issue, EventSession } from '../types'
import { StatusBadge } from '../components/common/StatusBadge'
import { LoadingSpinner } from '../components/common/LoadingSpinner'
import { formatDateTime } from '../lib/utils'
import {
  CalendarCheck,
  CalendarDays,
  AlertTriangle,
  Box,
  ArrowUpRight,
  Clock,
  Building,
  CheckCircle2,
} from 'lucide-react'
import { Link } from 'react-router-dom'

export const Dashboard: React.FC = () => {
  const { user } = useAuth()
  const { resources = [], getResourceById, isLoading: isLoadingResources } = useResources()

  const [bookings, setBookings] = useState<Booking[]>([])
  const [enrolledEvents, setEnrolledEvents] = useState<EventSession[]>([])
  const [issues, setIssues] = useState<Issue[]>([])
  const [isLoadingData, setIsLoadingData] = useState(true)

  const isStaff = user?.role === 'ADMIN' || user?.role === 'TECHNICIAN'
  const isStudent = user?.role === 'STUDENT'

  useEffect(() => {
    let isMounted = true

    async function loadDashboardData() {
      setIsLoadingData(true)
      try {
        if (isStudent) {
          const [eventsData, issuesData] = await Promise.all([
            getMyEnrolledEvents(),
            getMyIssues(),
          ])
          if (isMounted) {
            setEnrolledEvents(Array.isArray(eventsData) ? eventsData : [])
            setIssues(Array.isArray(issuesData) ? issuesData : [])
          }
        } else {
          const [bookingsData, issuesData] = await Promise.all([
            isStaff ? getAllBookings() : getMyBookings(),
            isStaff ? getAllIssues() : getMyIssues(),
          ])
          if (isMounted) {
            setBookings(Array.isArray(bookingsData) ? bookingsData : [])
            setIssues(Array.isArray(issuesData) ? issuesData : [])
          }
        }
      } catch (err) {
        console.warn('Dashboard data fetch error:', err)
        if (isMounted) {
          setBookings([])
          setEnrolledEvents([])
          setIssues([])
        }
      } finally {
        if (isMounted) {
          setIsLoadingData(false)
        }
      }
    }

    if (user) {
      loadDashboardData()
    }

    return () => {
      isMounted = false
    }
  }, [user, isStaff, isStudent])

  if (!user) return null

  // Ensure safe arrays in case of null responses from fresh databases
  const safeBookings = Array.isArray(bookings) ? bookings : []
  const safeIssues = Array.isArray(issues) ? issues : []
  const safeResources = Array.isArray(resources) ? resources : []

  // Derived metrics
  const activeBookings = safeBookings.filter((b) => b?.Status === 'CONFIRMED')
  const upcomingBookings = activeBookings
    .filter((b) => b?.EndTime && new Date(b.EndTime) >= new Date())
    .slice(0, 3)

  const upcomingEnrolled = enrolledEvents
    .filter((e) => e?.end_time && new Date(e.end_time) >= new Date())
    .slice(0, 3)

  const openIssues = safeIssues.filter((i) => i?.Status === 'OPEN' || i?.Status === 'IN_PROGRESS')
  const recentIssues = safeIssues.slice(0, 3)

  const availableResourcesCount = safeResources.filter((r) => r?.Status === 'AVAILABLE').length

  return (
    <div className="space-y-8">
      {/* Welcome Banner */}
      <div>
        <h1 className="text-2xl font-bold tracking-tight text-slate-900">
          Good day, {user.email ? user.email.split('@')[0] : 'User'} 👋
        </h1>
        <p className="mt-1 text-sm text-slate-500">
          Campus facility summary and real-time activity
        </p>
      </div>

      {/* Metric Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <Link
          to="/resources"
          className="bg-white p-6 rounded-2xl border border-slate-200/80 shadow-xs hover:border-indigo-200 transition group"
        >
          <div className="flex items-center justify-between">
            <span className="text-xs font-semibold uppercase tracking-wider text-slate-400">
              Campus Facilities
            </span>
            <div className="w-9 h-9 rounded-lg bg-indigo-50 text-indigo-600 flex items-center justify-center group-hover:bg-indigo-600 group-hover:text-white transition">
              <Box className="w-5 h-5" />
            </div>
          </div>
          <div className="mt-4 flex items-baseline justify-between">
            <span className="text-3xl font-extrabold text-slate-900">
              {safeResources.length}
            </span>
            <span className="text-xs text-indigo-600 font-semibold flex items-center gap-0.5">
              Browse <ArrowUpRight className="w-3.5 h-3.5" />
            </span>
          </div>
          <p className="mt-1 text-xs text-slate-500">
            {availableResourcesCount} available for reservation
          </p>
        </Link>

        {isStudent ? (
          <Link
            to="/events"
            className="bg-white p-6 rounded-2xl border border-slate-200/80 shadow-xs hover:border-indigo-200 transition group"
          >
            <div className="flex items-center justify-between">
              <span className="text-xs font-semibold uppercase tracking-wider text-slate-400">
                Enrolled Lectures
              </span>
              <div className="w-9 h-9 rounded-lg bg-indigo-50 text-indigo-600 flex items-center justify-center group-hover:bg-indigo-600 group-hover:text-white transition">
                <CalendarDays className="w-5 h-5" />
              </div>
            </div>
            <div className="mt-4 flex items-baseline justify-between">
              <span className="text-3xl font-extrabold text-slate-900">
                {enrolledEvents.length}
              </span>
              <span className="text-xs text-indigo-600 font-semibold flex items-center gap-0.5">
                View Timetable <ArrowUpRight className="w-3.5 h-3.5" />
              </span>
            </div>
            <p className="mt-1 text-xs text-slate-500">
              {upcomingEnrolled.length} upcoming sessions
            </p>
          </Link>
        ) : (
          <Link
            to={isStaff ? '/bookings/all' : '/bookings/mine'}
            className="bg-white p-6 rounded-2xl border border-slate-200/80 shadow-xs hover:border-emerald-200 transition group"
          >
            <div className="flex items-center justify-between">
              <span className="text-xs font-semibold uppercase tracking-wider text-slate-400">
                {isStaff ? 'Campus Bookings' : 'My Reservations'}
              </span>
              <div className="w-9 h-9 rounded-lg bg-emerald-50 text-emerald-600 flex items-center justify-center group-hover:bg-emerald-600 group-hover:text-white transition">
                <CalendarCheck className="w-5 h-5" />
              </div>
            </div>
            <div className="mt-4 flex items-baseline justify-between">
              <span className="text-3xl font-extrabold text-slate-900">
                {activeBookings.length}
              </span>
              <span className="text-xs text-emerald-600 font-semibold flex items-center gap-0.5">
                Manage <ArrowUpRight className="w-3.5 h-3.5" />
              </span>
            </div>
            <p className="mt-1 text-xs text-slate-500">
              {upcomingBookings.length} active or upcoming
            </p>
          </Link>
        )}

        <Link
          to={isStaff ? '/issues/all' : '/issues/mine'}
          className="bg-white p-6 rounded-2xl border border-slate-200/80 shadow-xs hover:border-amber-200 transition group"
        >
          <div className="flex items-center justify-between">
            <span className="text-xs font-semibold uppercase tracking-wider text-slate-400">
              {isStaff ? 'Open Triage' : 'My Active Issues'}
            </span>
            <div className="w-9 h-9 rounded-lg bg-amber-50 text-amber-600 flex items-center justify-center group-hover:bg-amber-600 group-hover:text-white transition">
              <AlertTriangle className="w-5 h-5" />
            </div>
          </div>
          <div className="mt-4 flex items-baseline justify-between">
            <span className="text-3xl font-extrabold text-slate-900">
              {openIssues.length}
            </span>
            <span className="text-xs text-amber-600 font-semibold flex items-center gap-0.5">
              Triage <ArrowUpRight className="w-3.5 h-3.5" />
            </span>
          </div>
          <p className="mt-1 text-xs text-slate-500">
            {safeIssues.length - openIssues.length} resolved or closed
          </p>
        </Link>
      </div>

      {/* Two Column Section: Upcoming Bookings & Recent Issues */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
        {/* Upcoming Bookings / Sessions Widget */}
        <div className="bg-white p-6 rounded-2xl border border-slate-200/80 shadow-xs space-y-4">
          <div className="flex items-center justify-between pb-3 border-b border-slate-100">
            <h2 className="text-base font-bold text-slate-900 flex items-center gap-2">
              {isStudent ? (
                <>
                  <CalendarDays className="w-4 h-4 text-indigo-600" />
                  <span>My Enrolled Sessions</span>
                </>
              ) : (
                <>
                  <CalendarCheck className="w-4 h-4 text-emerald-600" />
                  <span>Upcoming Reservations</span>
                </>
              )}
            </h2>
            <Link
              to={isStudent ? '/events' : isStaff ? '/bookings/all' : '/bookings/mine'}
              className="text-xs font-semibold text-indigo-600 hover:text-indigo-700"
            >
              View all
            </Link>
          </div>

          {isLoadingData || isLoadingResources ? (
            <LoadingSpinner message="Checking schedule..." size="sm" />
          ) : isStudent ? (
            upcomingEnrolled.length === 0 ? (
              <div className="py-8 text-center text-xs text-slate-400 space-y-2">
                <p>You haven't enrolled in any upcoming lectures or sessions.</p>
                <Link
                  to="/events"
                  className="inline-block text-xs font-semibold text-indigo-600 hover:underline"
                >
                  Browse Campus Events &rarr;
                </Link>
              </div>
            ) : (
              <div className="divide-y divide-slate-100 space-y-1">
                {upcomingEnrolled.map((ev) => (
                  <div key={ev.id} className="py-3 flex items-center justify-between">
                    <div>
                      <div className="flex items-center gap-2">
                        <h4 className="text-sm font-semibold text-slate-800">
                          {ev.title}
                        </h4>
                        {ev.course_code && (
                          <span className="text-[10px] font-bold text-indigo-700 bg-indigo-50 px-1.5 py-0.5 rounded">
                            {ev.course_code}
                          </span>
                        )}
                      </div>
                      <p className="text-xs text-slate-500 flex items-center gap-1 mt-0.5">
                        <Building className="w-3 h-3 text-slate-400" />
                        <span>
                          {ev.resource_name} ({ev.resource_location})
                        </span>
                        <span className="mx-1">&middot;</span>
                        <Clock className="w-3 h-3 text-slate-400" />
                        <span>{formatDateTime(ev.start_time)}</span>
                      </p>
                    </div>
                    <span className="inline-flex items-center gap-1 text-[11px] font-bold text-emerald-700 bg-emerald-50 px-2.5 py-1 rounded-full shrink-0">
                      <CheckCircle2 className="w-3 h-3" /> Enrolled
                    </span>
                  </div>
                ))}
              </div>
            )
          ) : upcomingBookings.length === 0 ? (
            <div className="py-8 text-center text-xs text-slate-400">
              No upcoming reservations scheduled.
            </div>
          ) : (
            <div className="divide-y divide-slate-100 space-y-1">
              {upcomingBookings.map((b) => {
                const resource = getResourceById(b.ResourceID)
                return (
                  <div key={b.ID} className="py-3 flex items-center justify-between">
                    <div>
                      <div className="flex items-center gap-2">
                        <h4 className="text-sm font-semibold text-slate-800">
                          {b.Title || resource?.Name || 'Facility'}
                        </h4>
                        {b.CourseCode && (
                          <span className="text-[10px] font-bold text-indigo-700 bg-indigo-50 px-1.5 py-0.5 rounded">
                            {b.CourseCode}
                          </span>
                        )}
                      </div>
                      <p className="text-xs text-slate-500 flex items-center gap-1 mt-0.5">
                        <Building className="w-3 h-3 text-slate-400" />
                        <span>
                          {resource?.Name || 'Facility'} ({resource?.Location || 'Campus'})
                        </span>
                        <span className="mx-1">&middot;</span>
                        <Clock className="w-3 h-3 text-slate-400" />
                        <span>{formatDateTime(b.StartTime)}</span>
                      </p>
                    </div>
                    <StatusBadge status={b.Status} />
                  </div>
                )
              })}
            </div>
          )}
        </div>

        {/* Recent Issues Widget */}
        <div className="bg-white p-6 rounded-2xl border border-slate-200/80 shadow-xs space-y-4">
          <div className="flex items-center justify-between pb-3 border-b border-slate-100">
            <h2 className="text-base font-bold text-slate-900 flex items-center gap-2">
              <AlertTriangle className="w-4 h-4 text-amber-600" />
              <span>Recent Facility Issues</span>
            </h2>
            <Link
              to={isStaff ? '/issues/all' : '/issues/mine'}
              className="text-xs font-semibold text-indigo-600 hover:text-indigo-700"
            >
              View all
            </Link>
          </div>

          {isLoadingData || isLoadingResources ? (
            <LoadingSpinner message="Checking issues..." size="sm" />
          ) : recentIssues.length === 0 ? (
            <div className="py-8 text-center text-xs text-slate-400">
              No recent defect reports.
            </div>
          ) : (
            <div className="divide-y divide-slate-100 space-y-1">
              {recentIssues.map((item) => {
                const resource = getResourceById(item.ResourceID)
                return (
                  <div key={item.ID} className="py-3 flex items-center justify-between">
                    <div>
                      <h4 className="text-sm font-semibold text-slate-800 flex items-center gap-2">
                        <span>{resource?.Name || 'Facility'}</span>
                        <span className="text-[10px] uppercase font-bold text-amber-600 bg-amber-50 px-1.5 py-0.5 rounded">
                          {item.Category}
                        </span>
                      </h4>
                      <p className="text-xs text-slate-500 truncate max-w-xs mt-0.5">
                        {item.Description}
                      </p>
                    </div>
                    <StatusBadge status={item.Status} />
                  </div>
                )
              })}
            </div>
          )}
        </div>
      </div>

      {/* Multitenancy Footer Banner */}
      <div className="bg-indigo-50/60 border border-indigo-100 rounded-2xl p-6 flex items-center justify-between gap-4">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-xl bg-indigo-600 text-white flex items-center justify-center font-bold shrink-0">
            <Building className="w-5 h-5" />
          </div>
          <div>
            <h3 className="text-sm font-bold text-indigo-950">
              Academic Tenant Scope
            </h3>
            <p className="text-xs text-indigo-700">
              All bookings, issues, and facilities are strictly partitioned for Tenant{' '}
              <span className="font-mono font-semibold">{user.tenant_id}</span>.
            </p>
          </div>
        </div>
        <div className="hidden sm:flex items-center gap-2 text-xs font-semibold text-indigo-800 bg-white px-3.5 py-1.5 rounded-full border border-indigo-200">
          <CheckCircle2 className="w-4 h-4 text-emerald-600" />
          <span>Role: {user.role}</span>
        </div>
      </div>
    </div>
  )
}
