import { apiClient } from './client'
import type { EventSession, RosterStudent } from '../types'

export async function getEvents(): Promise<EventSession[]> {
  const res = await apiClient.get<EventSession[]>('/events')
  return res.data
}

export async function getMyEnrolledEvents(): Promise<EventSession[]> {
  const res = await apiClient.get<EventSession[]>('/events/enrolled')
  return res.data
}

export async function enrollInEvent(eventId: string): Promise<void> {
  await apiClient.post(`/events/${eventId}/enroll`)
}

export async function unenrollFromEvent(eventId: string): Promise<void> {
  await apiClient.delete(`/events/${eventId}/enroll`)
}

export async function getEventRoster(eventId: string): Promise<RosterStudent[]> {
  const res = await apiClient.get<RosterStudent[]>(`/events/${eventId}/roster`)
  return res.data
}

export async function removeStudentFromRoster(eventId: string, studentId: string): Promise<void> {
  await apiClient.delete(`/events/${eventId}/roster/${studentId}`)
}
