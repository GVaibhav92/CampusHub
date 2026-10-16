import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { Toaster } from 'sonner'
import { AuthProvider } from './context/AuthContext'
import { ResourceProvider } from './context/ResourceContext'
import { useAuth } from './hooks/useAuth'
import { ProtectedRoute } from './routes/ProtectedRoute'
import { AppShell } from './components/layout/AppShell'
import { Login } from './pages/auth/Login'
import { Register } from './pages/auth/Register'
import { Dashboard } from './pages/Dashboard'
import { Resources } from './pages/Resources'
import { ResourceDetails } from './pages/ResourceDetails'
import { Bookings } from './pages/Bookings'
import { CreateBooking } from './pages/CreateBooking'
import { Events } from './pages/Events'
import { RoleRoute } from './routes/RoleRoute'
import { Issues } from './pages/Issues'
import { CreateIssue } from './pages/CreateIssue'
import { IssueDetails } from './pages/IssueDetails'
import { Users } from './pages/Users'
import { NotFound } from './pages/NotFound'

function AuthenticatedApp() {
  const { user, logout } = useAuth()

  if (!user) return null

  return (
    <ResourceProvider>
      <AppShell user={user} onLogout={logout} />
    </ResourceProvider>
  )
}

export default function App() {
  return (
    <AuthProvider>
      <Toaster position="top-right" richColors closeButton />
      <BrowserRouter>
        <Routes>
          {/* Public Auth Routes */}
          <Route path="/login" element={<Login />} />
          <Route path="/register" element={<Register />} />

          {/* Protected Routes inside AppShell */}
          <Route
            element={
              <ProtectedRoute>
                <AuthenticatedApp />
              </ProtectedRoute>
            }
          >
            <Route path="/dashboard" element={<Dashboard />} />
            <Route path="/events" element={<Events />} />
            <Route path="/resources" element={<Resources />} />
            <Route path="/resources/:id" element={<ResourceDetails />} />
            <Route path="/bookings/mine" element={<Bookings isAll={false} />} />
            <Route path="/bookings/all" element={<Bookings isAll={true} />} />
            <Route
              path="/bookings/create"
              element={
                <RoleRoute allowedRoles={['FACULTY', 'ADMIN']}>
                  <CreateBooking />
                </RoleRoute>
              }
            />
            <Route path="/issues/mine" element={<Issues isAll={false} />} />
            <Route path="/issues/all" element={<Issues isAll={true} />} />
            <Route path="/issues/create" element={<CreateIssue />} />
            <Route path="/issues/:id" element={<IssueDetails />} />
            <Route
              path="/users"
              element={
                <RoleRoute allowedRoles={['ADMIN']}>
                  <Users />
                </RoleRoute>
              }
            />
            <Route path="/" element={<Navigate to="/dashboard" replace />} />

            {/* In-app fallback */}
            <Route path="*" element={<NotFound />} />
          </Route>

          {/* Root/Global fallback */}
          <Route path="*" element={<NotFound />} />
        </Routes>
      </BrowserRouter>
    </AuthProvider>
  )
}
