import { apiClient } from './client'
import type { Role } from '../types'

export interface UserSummary {
  id: string
  email: string
  role: Role
}

export async function getUsers(role?: Role): Promise<UserSummary[]> {
  const params = role ? { role } : {}
  const res = await apiClient.get<UserSummary[]>('/users', { params })
  return res.data
}
