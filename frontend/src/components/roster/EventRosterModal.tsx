import React, { useState, useEffect, useCallback } from 'react'
import { getEventRoster, removeStudentFromRoster } from '../../api/events'
import type { RosterStudent } from '../../types'
import { formatDateTime } from '../../lib/utils'
import {
  X,
  Users,
  UserX,
  Loader2,
  AlertCircle,
  CheckCircle2,
  Calendar,
} from 'lucide-react'
import { toast } from 'sonner'

interface EventRosterModalProps {
  isOpen: boolean
  onClose: () => void
  bookingId: string
  bookingTitle: string
  capacity: number
}

export const EventRosterModal: React.FC<EventRosterModalProps> = ({
  isOpen,
  onClose,
  bookingId,
  bookingTitle,
  capacity,
}) => {
  const [roster, setRoster] = useState<RosterStudent[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [removingId, setRemovingId] = useState<string | null>(null)

  const fetchRoster = useCallback(async () => {
    setIsLoading(true)
    setError(null)
    try {
      const data = await getEventRoster(bookingId)
      setRoster(data)
    } catch (err: unknown) {
      if (err instanceof Error) {
        setError(err.message)
      } else {
        setError('Failed to load event attendee roster.')
      }
    } finally {
      setIsLoading(false)
    }
  }, [bookingId])

  useEffect(() => {
    if (isOpen && bookingId) {
      fetchRoster()
    }
  }, [isOpen, bookingId, fetchRoster])

  if (!isOpen) return null

  const activeEnrolled = roster.filter((r) => r.status === 'ENROLLED')
  const percentFilled = capacity > 0 ? Math.min(100, Math.round((activeEnrolled.length / capacity) * 100)) : 0

  const handleRemoveStudent = async (student: RosterStudent) => {
    const confirmed = window.confirm(
      `Are you sure you want to remove ${student.email} from this event? This action is terminal and will prevent the student from re-enrolling.`
    )
    if (!confirmed) return

    setRemovingId(student.student_id)
    try {
      await removeStudentFromRoster(bookingId, student.student_id)
      toast.success(`Removed ${student.email} from event roster`)
      // Refresh or update state in place
      setRoster((prev) =>
        prev.map((s) =>
          s.student_id === student.student_id
            ? { ...s, status: 'CANCELLED_BY_FACULTY' }
            : s
        )
      )
    } catch (err: unknown) {
      if (err instanceof Error) {
        toast.error(err.message)
      } else {
        toast.error('Failed to remove student.')
      }
    } finally {
      setRemovingId(null)
    }
  }

  return (
    <div className="fixed inset-0 z-50 overflow-y-auto bg-slate-900/60 backdrop-blur-xs flex items-center justify-center p-4">
      <div className="relative bg-white rounded-2xl shadow-xl border border-slate-200 max-w-2xl w-full overflow-hidden animate-in fade-in zoom-in-95 duration-150">
        {/* Modal Header */}
        <div className="px-6 py-5 border-b border-slate-100 flex items-start justify-between gap-4">
          <div>
            <div className="flex items-center gap-2">
              <span className="p-1.5 bg-indigo-50 text-indigo-600 rounded-lg">
                <Users className="w-5 h-5" />
              </span>
              <h2 className="text-lg font-bold text-slate-900 tracking-tight">
                Event Attendee Roster
              </h2>
            </div>
            <p className="mt-1 text-xs text-slate-500 font-medium truncate max-w-md">
              Session: <strong>{bookingTitle}</strong>
            </p>
          </div>

          <button
            onClick={onClose}
            className="p-1.5 rounded-lg text-slate-400 hover:text-slate-600 hover:bg-slate-100 transition"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Capacity Bar */}
        <div className="px-6 py-4 bg-slate-50/70 border-b border-slate-100">
          <div className="flex items-center justify-between text-xs font-semibold text-slate-700 mb-1.5">
            <span>Room Capacity Utilization</span>
            <span>
              {activeEnrolled.length} / {capacity} seats ({percentFilled}%)
            </span>
          </div>
          <div className="w-full bg-slate-200 rounded-full h-2 overflow-hidden">
            <div
              className={`h-full rounded-full transition-all duration-300 ${
                percentFilled >= 100
                  ? 'bg-rose-500'
                  : percentFilled >= 75
                  ? 'bg-amber-500'
                  : 'bg-indigo-600'
              }`}
              style={{ width: `${percentFilled}%` }}
            />
          </div>
        </div>

        {/* Modal Body */}
        <div className="p-6 max-h-[60vh] overflow-y-auto">
          {error && (
            <div className="p-4 mb-4 rounded-xl bg-rose-50 border border-rose-200 flex items-start gap-3">
              <AlertCircle className="w-5 h-5 text-rose-600 shrink-0 mt-0.5" />
              <p className="text-xs font-semibold text-rose-800">{error}</p>
            </div>
          )}

          {isLoading ? (
            <div className="py-12 flex flex-col items-center justify-center gap-2 text-slate-400">
              <Loader2 className="w-6 h-6 animate-spin text-indigo-600" />
              <p className="text-xs font-medium">Loading attendee roster...</p>
            </div>
          ) : roster.length === 0 ? (
            <div className="py-12 text-center text-slate-400 space-y-2">
              <Calendar className="w-8 h-8 mx-auto text-slate-300" />
              <p className="text-sm font-semibold text-slate-600">No students enrolled yet</p>
              <p className="text-xs text-slate-400">
                Students will appear here once they register for this lecture.
              </p>
            </div>
          ) : (
            <div className="divide-y divide-slate-100 border border-slate-100 rounded-xl overflow-hidden">
              {roster.map((student) => {
                const isEnrolled = student.status === 'ENROLLED'
                const isRevoked = student.status === 'CANCELLED_BY_FACULTY'
                const isSelfCancelled = student.status === 'CANCELLED_BY_STUDENT'

                return (
                  <div
                    key={student.student_id}
                    className="p-3.5 flex items-center justify-between gap-3 hover:bg-slate-50/60 transition"
                  >
                    <div className="space-y-0.5">
                      <div className="flex items-center gap-2">
                        <span className="text-sm font-semibold text-slate-900">
                          {student.email}
                        </span>
                        {isEnrolled && (
                          <span className="inline-flex items-center gap-1 text-[11px] font-semibold text-emerald-700 bg-emerald-50 px-2 py-0.5 rounded-md">
                            <CheckCircle2 className="w-3 h-3" /> Enrolled
                          </span>
                        )}
                        {isRevoked && (
                          <span className="inline-flex items-center gap-1 text-[11px] font-semibold text-rose-700 bg-rose-50 px-2 py-0.5 rounded-md">
                            Revoked by Faculty
                          </span>
                        )}
                        {isSelfCancelled && (
                          <span className="inline-flex items-center gap-1 text-[11px] font-semibold text-slate-500 bg-slate-100 px-2 py-0.5 rounded-md">
                            Self-Cancelled
                          </span>
                        )}
                      </div>
                      <p className="text-[11px] text-slate-400 font-medium">
                        Registered: {formatDateTime(student.enrolled_at)}
                      </p>
                    </div>

                    {isEnrolled && (
                      <button
                        onClick={() => handleRemoveStudent(student)}
                        disabled={removingId === student.student_id}
                        className="inline-flex items-center gap-1.5 text-xs font-semibold text-rose-600 hover:text-rose-700 hover:bg-rose-50 px-2.5 py-1.5 rounded-lg transition disabled:opacity-50"
                        title="Remove student from this event permanently"
                      >
                        {removingId === student.student_id ? (
                          <Loader2 className="w-3.5 h-3.5 animate-spin" />
                        ) : (
                          <UserX className="w-3.5 h-3.5" />
                        )}
                        Remove
                      </button>
                    )}
                  </div>
                )
              })}
            </div>
          )}
        </div>

        {/* Modal Footer */}
        <div className="px-6 py-3.5 bg-slate-50 border-t border-slate-100 flex items-center justify-end">
          <button
            onClick={onClose}
            className="px-4 py-2 bg-white border border-slate-200 text-slate-700 rounded-lg text-xs font-semibold hover:bg-slate-100 transition"
          >
            Close
          </button>
        </div>
      </div>
    </div>
  )
}
