import { apiClient } from './client'
import type { Issue, IssuePriority, IssueStatus } from '../types'

export interface CreateIssuePayload {
  resource_id: string
  category: string
  priority?: IssuePriority
  description: string
}

export interface UpdateIssueStatusPayload {
  status: IssueStatus
  assigned_to?: string
}

export async function createIssue(payload: CreateIssuePayload): Promise<Issue> {
  const res = await apiClient.post<Issue>('/issues', payload)
  return res.data
}

export async function getMyIssues(): Promise<Issue[]> {
  const res = await apiClient.get<Issue[]>('/issues/mine')
  return res.data
}

export async function getAllIssues(): Promise<Issue[]> {
  const res = await apiClient.get<Issue[]>('/issues')
  return res.data
}

export async function getIssue(id: string): Promise<Issue> {
  const res = await apiClient.get<Issue>(`/issues/${id}`)
  return res.data
}

export async function updateIssueStatus(
  id: string,
  payload: UpdateIssueStatusPayload
): Promise<Issue> {
  const res = await apiClient.patch<Issue>(`/issues/${id}/status`, payload)
  return res.data
}
