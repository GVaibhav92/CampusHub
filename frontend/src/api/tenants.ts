import { apiClient } from './client'
import type { Tenant } from '../types'

export async function getTenants(): Promise<Tenant[]> {
  const res = await apiClient.get<Tenant[]>('/tenants')
  return res.data
}
