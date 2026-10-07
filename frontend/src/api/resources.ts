import { apiClient } from './client'
import type { Resource, ResourceStatus } from '../types'

export interface CreateResourcePayload {
  name: string
  type: string
  capacity: number
  location: string
}

export interface UpdateResourcePayload {
  name: string
  type: string
  capacity: number
  location: string
  status: ResourceStatus
}

export async function getResources(): Promise<Resource[]> {
  const res = await apiClient.get<Resource[]>('/resources')
  return res.data
}

export async function getResource(id: string): Promise<Resource> {
  const res = await apiClient.get<Resource>(`/resources/${id}`)
  return res.data
}

export async function createResource(payload: CreateResourcePayload): Promise<Resource> {
  const res = await apiClient.post<Resource>('/resources', payload)
  return res.data
}

export async function updateResource(id: string, payload: UpdateResourcePayload): Promise<Resource> {
  const res = await apiClient.put<Resource>(`/resources/${id}`, payload)
  return res.data
}

export async function deleteResource(id: string): Promise<void> {
  await apiClient.delete(`/resources/${id}`)
}
