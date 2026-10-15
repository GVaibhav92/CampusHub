import { useState, useEffect, useCallback } from 'react'
import type { Booking } from '../types'
import { getMyBookings, getAllBookings, cancelBooking as apiCancelBooking } from '../api/bookings'

export function useBookings(isAll: boolean = false) {
  const [bookings, setBookings] = useState<Booking[]>([])
  const [isLoading, setIsLoading] = useState<boolean>(true)
  const [error, setError] = useState<string | null>(null)

  const fetchBookings = useCallback(async () => {
    setIsLoading(true)
    setError(null)
    try {
      const data = isAll ? await getAllBookings() : await getMyBookings()
      setBookings(Array.isArray(data) ? data : [])
    } catch (err: unknown) {
      setBookings([])
      if (err instanceof Error) {
        setError(err.message)
      } else {
        setError('Failed to fetch reservations.')
      }
    } finally {
      setIsLoading(false)
    }
  }, [isAll])

  useEffect(() => {
    fetchBookings()
  }, [fetchBookings])

  const cancel = async (id: string) => {
    await apiCancelBooking(id)
    setBookings((prev) =>
      Array.isArray(prev) ? prev.map((b) => (b.ID === id ? { ...b, Status: 'CANCELLED' } : b)) : []
    )
  }

  return {
    bookings: Array.isArray(bookings) ? bookings : [],
    isLoading,
    error,
    refresh: fetchBookings,
    cancelBooking: cancel,
  }
}
