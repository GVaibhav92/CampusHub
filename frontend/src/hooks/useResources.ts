import { useContext } from 'react'
import { ResourceContext } from '../context/ResourceContext'

export function useResources() {
  const context = useContext(ResourceContext)
  if (!context) {
    throw new Error('useResources must be used within a ResourceProvider')
  }
  return context
}
