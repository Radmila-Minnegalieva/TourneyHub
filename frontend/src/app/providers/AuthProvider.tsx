import { createContext, useMemo, useState, type PropsWithChildren } from 'react'
import { mockUser } from '../../mocks/user'
import type { UserProfile } from '../../types/user'

export interface AuthContextValue {
  user: UserProfile | null
  isAuthenticated: boolean
  login: (email: string, password: string) => void
  logout: () => void
  register: (payload: { email: string; nickname: string; name: string }) => void
  updateProfile: (profile: UserProfile) => void
}

export const AuthContext = createContext<AuthContextValue | null>(null)

export function AuthProvider({ children }: PropsWithChildren) {
  const [user, setUser] = useState<UserProfile | null>(mockUser)

  const value = useMemo<AuthContextValue>(() => ({
    user,
    isAuthenticated: Boolean(user),
    login: () => setUser(mockUser),
    logout: () => setUser(null),
    register: ({ email, nickname, name }) => setUser({ ...mockUser, email, nickname, name }),
    updateProfile: setUser,
  }), [user])

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}
