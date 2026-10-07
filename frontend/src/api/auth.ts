import { apiClient } from './client'
import type { User, Role } from '../types'

export interface RegisterPayload {
  tenant_slug: string
  email: string
  password: string
  role?: 'STUDENT' | 'FACULTY'
}

export interface RegisterResponse {
  id: string
  tenant_id: string
  email: string
  role: Role
}

export interface LoginPayload {
  email: string
  password: string
}

export interface LoginResponse {
  token: string
}

export async function registerUser(payload: RegisterPayload): Promise<RegisterResponse> {
  const res = await apiClient.post<RegisterResponse>('/auth/register', payload)
  return res.data
}

export async function loginUser(payload: LoginPayload): Promise<LoginResponse> {
  const res = await apiClient.post<LoginResponse>('/auth/login', payload)
  return res.data
}

export async function getMe(): Promise<User> {
  const res = await apiClient.get<User>('/me')
  return res.data
}
