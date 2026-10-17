import { apiClient } from './client'
import type { Booking } from '../types'

export interface CreateBookingPayload {
  resource_id: string
  title: string
  course_code?: string
  start_time: string // ISO-8601
  end_time: string   // ISO-8601
}

export function normalizeBooking(raw: any): Booking {
  if (!raw || typeof raw !== 'object') {
    return {} as Booking
  }
  return {
    ID: raw.ID || raw.id || '',
    TenantID: raw.TenantID || raw.tenant_id || '',
    ResourceID: raw.ResourceID || raw.resource_id || '',
    UserID: raw.UserID || raw.user_id || '',
    Title: raw.Title || raw.title || 'Academic Session',
    CourseCode: raw.CourseCode || raw.course_code || undefined,
    StartTime: raw.StartTime || raw.start_time || '',
    EndTime: raw.EndTime || raw.end_time || '',
    Status: (raw.Status || raw.status || 'CONFIRMED') as any,
    CreatedAt: raw.CreatedAt || raw.created_at || '',
  }
}

export async function createBooking(payload: CreateBookingPayload): Promise<Booking> {
  const res = await apiClient.post<any>('/bookings', payload)
  return normalizeBooking(res.data)
}

export async function getMyBookings(): Promise<Booking[]> {
  const res = await apiClient.get<any[]>('/bookings/mine')
  if (!Array.isArray(res.data)) return []
  return res.data.map(normalizeBooking)
}

export async function getAllBookings(): Promise<Booking[]> {
  const res = await apiClient.get<any[]>('/bookings')
  if (!Array.isArray(res.data)) return []
  return res.data.map(normalizeBooking)
}

export async function cancelBooking(id: string): Promise<void> {
  await apiClient.delete(`/bookings/${id}`)
}
