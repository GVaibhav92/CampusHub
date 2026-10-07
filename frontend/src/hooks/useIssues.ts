import { useState, useEffect, useCallback } from 'react'
import type { Issue, IssueStatus } from '../types'
import { getMyIssues, getAllIssues, updateIssueStatus as apiUpdateStatus } from '../api/issues'

export function useIssues(isAll: boolean = false) {
  const [issues, setIssues] = useState<Issue[]>([])
  const [isLoading, setIsLoading] = useState<boolean>(true)
  const [error, setError] = useState<string | null>(null)

  const fetchIssues = useCallback(async () => {
    setIsLoading(true)
    setError(null)
    try {
      const data = isAll ? await getAllIssues() : await getMyIssues()
      setIssues(data)
    } catch (err: unknown) {
      if (err instanceof Error) {
        setError(err.message)
      } else {
        setError('Failed to fetch issues.')
      }
    } finally {
      setIsLoading(false)
    }
  }, [isAll])

  useEffect(() => {
    fetchIssues()
  }, [fetchIssues])

  const changeStatus = async (id: string, newStatus: IssueStatus, assignedTo?: string) => {
    const updated = await apiUpdateStatus(id, { status: newStatus, assigned_to: assignedTo })
    setIssues((prev) => prev.map((item) => (item.ID === id ? updated : item)))
    return updated
  }

  return {
    issues,
    isLoading,
    error,
    refresh: fetchIssues,
    changeStatus,
  }
}
