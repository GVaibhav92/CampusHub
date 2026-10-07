import { apiClient } from './client'
import type { Booking } from '../types'

export interface CreateBookingPayload {
  resource_id: string
  start_time: string // ISO-8601
  end_time: string   // ISO-8601
}

export async function createBooking(payload: CreateBookingPayload): Promise<Booking> {
  const res = await apiClient.post<Booking>('/bookings', payload)
  return res.data
}

export async function getMyBookings(): Promise<Booking[]> {
  const res = await apiClient.get<Booking[]>('/bookings/mine')
  return res.data
}

export async function getAllBookings(): Promise<Booking[]> {
  const res = await apiClient.get<Booking[]>('/bookings')
  return res.data
}

export async function cancelBooking(id: string): Promise<void> {
  await apiClient.delete(`/bookings/${id}`)
}
