import { AuthContext, type AuthContextValue } from "./authContext";
import { useEffect, useState, type PropsWithChildren } from "react";
import { api, getTokens, setTokens, type Schema } from "../../api/client";
import type { UserProfile } from "../../types/user";
function profile(p: Schema["Profile"]): UserProfile {
  return {
    id: p.id,
    email: p.email,
    nickname: p.nickname,
    name: p.name || p.nickname,
    timezone: p.timezone,
    avatarUrl: p.avatar_url,
    version: p.version,
    roles: p.roles.map((r) => (r === "referee" ? "judge" : r)),
    notificationSettings: {
      email: p.email_notifications,
      inApp: true,
      matchChanges: true,
      applicationUpdates: true,
      resultConfirmation: true,
    },
  };
}
export function AuthProvider({ children }: PropsWithChildren) {
  const [user, setUser] = useState<UserProfile | null>(null);
  const [loading, setLoading] = useState(true);
  useEffect(() => {
    let current = true;
    const clear = () => {
      if (!getTokens()) setUser(null);
    };
    window.addEventListener("authchange", clear);
    window.addEventListener("storage", clear);
    if (getTokens())
      api<Schema["Profile"]>("/users/me")
        .then((p) => {
          if (current) setUser(profile(p));
        })
        .catch(() => {
          if (current) setUser(null);
        })
        .finally(() => {
          if (current) setLoading(false);
        });
    else setLoading(false);
    return () => {
      current = false;
      window.removeEventListener("authchange", clear);
      window.removeEventListener("storage", clear);
    };
  }, []);
  const value: AuthContextValue = {
    user,
    loading,
    isAuthenticated: Boolean(user),
    async login(email, password) {
      setTokens(
        await api<Schema["Tokens"]>(
          "/auth/login",
          { method: "POST", body: { email, password } },
          false,
        ),
      );
      setUser(profile(await api<Schema["Profile"]>("/users/me")));
    },
    async logout() {
      const token = getTokens()?.refresh_token;
      try {
        if (token)
          await api("/auth/logout", { method: "POST", body: { token } }, false);
      } finally {
        setTokens(null);
        setUser(null);
      }
    },
    register: (payload) =>
      api("/auth/register", { method: "POST", body: payload }, false),
    async updateProfile(p, avatar) {
      let avatar_file_id: string | undefined;
      if (avatar?.size) {
        const body = new FormData();
        body.set("purpose", "avatar");
        body.set("file", avatar);
        avatar_file_id = (
          await api<Schema["File"]>("/files", { method: "POST", body })
        ).id;
      }
      const updated = await api<Schema["Profile"]>("/users/me", {
        method: "PATCH",
        body: {
          version: p.version,
          nickname: p.nickname,
          name: p.name,
          timezone: p.timezone,
          email_notifications: p.notificationSettings.email,
          avatar_file_id,
        },
      });
      setUser(profile(updated));
    },
  };
  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}
