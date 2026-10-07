import React, { createContext, useState, useEffect, useCallback, useMemo } from 'react'
import type { Resource } from '../types'
import {
  getResources,
  createResource,
  updateResource,
  deleteResource,
  type CreateResourcePayload,
  type UpdateResourcePayload,
} from '../api/resources'
import { useAuth } from '../hooks/useAuth'

interface ResourceContextType {
  resources: Resource[]
  resourceMap: Map<string, Resource>
  isLoading: boolean
  error: string | null
  refreshResources: () => Promise<void>
  addResource: (payload: CreateResourcePayload) => Promise<Resource>
  editResource: (id: string, payload: UpdateResourcePayload) => Promise<Resource>
  removeResource: (id: string) => Promise<void>
  getResourceById: (id: string) => Resource | undefined
}

export const ResourceContext = createContext<ResourceContextType | undefined>(undefined)

export const ResourceProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { isAuthenticated } = useAuth()
  const [resources, setResources] = useState<Resource[]>([])
  const [isLoading, setIsLoading] = useState<boolean>(false)
  const [error, setError] = useState<string | null>(null)

  const refreshResources = useCallback(async () => {
    if (!isAuthenticated) return
    setIsLoading(true)
    setError(null)
    try {
      const data = await getResources()
      setResources(data)
    } catch (err: unknown) {
      if (err instanceof Error) {
        setError(err.message)
      } else {
        setError('Failed to fetch resources.')
      }
    } finally {
      setIsLoading(false)
    }
  }, [isAuthenticated])

  useEffect(() => {
    if (isAuthenticated) {
      refreshResources()
    } else {
      setResources([])
    }
  }, [isAuthenticated, refreshResources])

  const resourceMap = useMemo(() => {
    const map = new Map<string, Resource>()
    for (const r of resources) {
      map.set(r.ID, r)
    }
    return map
  }, [resources])

  const getResourceById = useCallback(
    (id: string) => resourceMap.get(id),
    [resourceMap]
  )

  const addResource = async (payload: CreateResourcePayload) => {
    const created = await createResource(payload)
    setResources((prev) => [created, ...prev])
    return created
  }

  const editResource = async (id: string, payload: UpdateResourcePayload) => {
    const updated = await updateResource(id, payload)
    setResources((prev) => prev.map((r) => (r.ID === id ? updated : r)))
    return updated
  }

  const removeResource = async (id: string) => {
    await deleteResource(id)
    setResources((prev) => prev.filter((r) => r.ID !== id))
  }

  return (
    <ResourceContext.Provider
      value={{
        resources,
        resourceMap,
        isLoading,
        error,
        refreshResources,
        addResource,
        editResource,
        removeResource,
        getResourceById,
      }}
    >
      {children}
    </ResourceContext.Provider>
  )
}
