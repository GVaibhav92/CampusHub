import axios, { AxiosError } from 'axios'

const baseURL = import.meta.env.VITE_API_BASE_URL

if (!baseURL) {
  console.warn(
    'VITE_API_BASE_URL is not set! Please define it in your .env or Amplify environment settings.'
  )
}

export const apiClient = axios.create({
  baseURL: baseURL || '',
  headers: {
    'Content-Type': 'application/json',
  },
})

// Request Interceptor: attach JWT if available
apiClient.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('campushub_token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => Promise.reject(error)
)

// Response Interceptor: normalize error messages
apiClient.interceptors.response.use(
  (response) => response,
  (error: AxiosError<{ error?: string }>) => {
    if (!error.response) {
      return Promise.reject(new Error('Unable to connect to CampusHub. Please check your connection and try again.'))
    }

    const status = error.response.status
    const serverMessage = error.response.data?.error

    if (status === 401) {
      // Clear token if unauthorized on a protected endpoint (except during login)
      if (!error.config?.url?.includes('/auth/login')) {
        localStorage.removeItem('campushub_token')
      }
      return Promise.reject(new Error(serverMessage || 'Your session has expired. Please sign in again.'))
    }

    if (status === 403) {
      return Promise.reject(new Error(serverMessage || "You don't have permission to perform this action."))
    }

    if (status === 404) {
      return Promise.reject(new Error(serverMessage || 'The requested resource could not be found.'))
    }

    if (status === 409) {
      return Promise.reject(new Error(serverMessage || 'This action conflicts with an existing record.'))
    }

    if (status >= 500) {
      return Promise.reject(new Error('Something went wrong on our server. Please try again later.'))
    }

    return Promise.reject(new Error(serverMessage || 'Please check the information you entered.'))
  }
)
