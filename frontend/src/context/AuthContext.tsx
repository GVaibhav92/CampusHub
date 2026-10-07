import React, { createContext, useState, useEffect, useCallback } from 'react'
import type { User } from '../types'
import { getMe, loginUser, registerUser, type LoginPayload, type RegisterPayload } from '../api/auth'

interface AuthContextType {
  user: User | null
  token: string | null
  isLoading: boolean
  isAuthenticated: boolean
  login: (credentials: LoginPayload) => Promise<void>
  register: (payload: RegisterPayload) => Promise<void>
  logout: () => void
}

export const AuthContext = createContext<AuthContextType | undefined>(undefined)

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [user, setUser] = useState<User | null>(null)
  const [token, setToken] = useState<string | null>(() => localStorage.getItem('campushub_token'))
  const [isLoading, setIsLoading] = useState<boolean>(true)

  const logout = useCallback(() => {
    localStorage.removeItem('campushub_token')
    setToken(null)
    setUser(null)
  }, [])

  // Restore session on mount or token change
  useEffect(() => {
    let isMounted = true

    async function restoreSession() {
      const storedToken = localStorage.getItem('campushub_token')
      if (!storedToken) {
        if (isMounted) {
          setIsLoading(false)
        }
        return
      }

      try {
        const me = await getMe()
        if (isMounted) {
          setUser(me)
          setToken(storedToken)
        }
      } catch (err) {
        console.warn('Session restoration failed:', err)
        if (isMounted) {
          logout()
        }
      } finally {
        if (isMounted) {
          setIsLoading(false)
        }
      }
    }

    restoreSession()

    return () => {
      isMounted = false
    }
  }, [logout])

  const login = async (credentials: LoginPayload) => {
    const { token: receivedToken } = await loginUser(credentials)
    localStorage.setItem('campushub_token', receivedToken)
    setToken(receivedToken)
    const me = await getMe()
    setUser(me)
  }

  const register = async (payload: RegisterPayload) => {
    await registerUser(payload)
    // Auto-login after registration
    await login({
      email: payload.email,
      password: payload.password,
    })
  }

  return (
    <AuthContext.Provider
      value={{
        user,
        token,
        isLoading,
        isAuthenticated: !!user,
        login,
        register,
        logout,
      }}
    >
      {children}
    </AuthContext.Provider>
  )
}
