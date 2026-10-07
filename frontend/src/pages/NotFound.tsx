import { useNavigate } from 'react-router-dom'
import { Home, ArrowLeft } from 'lucide-react'

export function NotFound() {
  const navigate = useNavigate()

  return (
    <div
      role="main"
      aria-label="Page not found"
      className="flex flex-col items-center justify-center min-h-[60vh] text-center px-4"
    >
      <div className="relative mb-6 select-none">
        <span className="text-[120px] font-black text-slate-100 leading-none">404</span>
        <div className="absolute inset-0 flex items-center justify-center">
          <div className="w-16 h-16 rounded-2xl bg-indigo-600 text-white flex items-center justify-center shadow-lg">
            <Home className="w-8 h-8" aria-hidden="true" />
          </div>
        </div>
      </div>

      <h1 className="text-2xl font-bold text-slate-800 mb-2">Page Not Found</h1>
      <p className="text-sm text-slate-500 max-w-xs mb-8">
        The page you are looking for does not exist or has been moved.
      </p>

      <div className="flex flex-col sm:flex-row gap-3">
        <button
          onClick={() => navigate(-1)}
          aria-label="Go back to previous page"
          className="inline-flex items-center gap-2 px-4 py-2.5 rounded-lg border border-slate-200 text-sm font-medium text-slate-700 hover:bg-slate-50 transition-colors"
        >
          <ArrowLeft className="w-4 h-4" aria-hidden="true" />
          Go Back
        </button>
        <button
          onClick={() => navigate('/dashboard')}
          aria-label="Go to dashboard"
          className="inline-flex items-center gap-2 px-4 py-2.5 rounded-lg bg-indigo-600 text-white text-sm font-medium hover:bg-indigo-700 transition-colors shadow-sm"
        >
          <Home className="w-4 h-4" aria-hidden="true" />
          Dashboard
        </button>
      </div>
    </div>
  )
}
