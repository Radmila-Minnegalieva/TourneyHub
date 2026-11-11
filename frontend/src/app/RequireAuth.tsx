import type { PropsWithChildren } from 'react'
import { Navigate, useLocation } from 'react-router-dom'
import { ROUTES } from '../constants/routes'
import { useAuth } from '../hooks/useAuth'

export function RequireAuth({ children }: PropsWithChildren) {
  const { isAuthenticated, loading } = useAuth()
  const location = useLocation()
  if (loading) return <p role="status">Загрузка сессии…</p>
  if (!isAuthenticated) return <Navigate to={ROUTES.login} replace state={{ from: location.pathname + location.search }} />
  return children
}
