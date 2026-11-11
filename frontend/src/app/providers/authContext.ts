import { createContext } from "react";
import type { Schema } from "../../api/client";
import type { UserProfile } from "../../types/user";
export interface AuthContextValue {
  user: UserProfile | null;
  loading: boolean;
  isAuthenticated: boolean;
  login: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  register: (payload: Schema["RegisterRequest"]) => Promise<void>;
  updateProfile: (profile: UserProfile, avatar?: File) => Promise<void>;
}
export const AuthContext = createContext<AuthContextValue | null>(null);
