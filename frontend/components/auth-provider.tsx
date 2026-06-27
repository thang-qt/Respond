"use client"

import { createContext, useCallback, useContext, useEffect, useMemo, useState } from "react"
import { setLocaleCookie } from "@/i18n/client-locale"
import { api, ApiError } from "@/lib/api"
import { clearAccessToken, setAccessToken } from "@/lib/auth-token"

export interface AuthUser {
  id: string
  email: string
  email_verified: boolean
  role: "user" | "moderator" | "admin"
  account_status: "active" | "suspended" | "banned"
  username: string
  bio: string
  rating: number
  wins: number
  losses: number
  draws: number
  default_reveal: boolean
  locale: "en" | "vi"
  created_at: string
}

type AuthStatus = "loading" | "authenticated" | "unauthenticated"

interface AuthContextValue {
  user: AuthUser | null
  status: AuthStatus
  login: (params: { identifier: string; password: string }) => Promise<void>
  signup: (params: { email: string; username: string; password: string; invite_token?: string }) => Promise<void>
  logout: () => Promise<void>
  refresh: () => Promise<void>
}

const AuthContext = createContext<AuthContextValue | undefined>(undefined)

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<AuthUser | null>(null)
  const [status, setStatus] = useState<AuthStatus>("loading")

  const login = useCallback(async ({ identifier, password }: { identifier: string; password: string }) => {
    const payload = await api.post<{ data: { user: AuthUser; access_token: string } }>("/auth/login", {
      identifier,
      password,
    })
    setAccessToken(payload.data.access_token)
    setLocaleCookie(payload.data.user.locale)
    setUser(payload.data.user)
    setStatus("authenticated")
  }, [])

  const signup = useCallback(async ({ email, username, password, invite_token }: { email: string; username: string; password: string; invite_token?: string }) => {
    const payload = await api.post<{ data: { user: AuthUser; access_token: string } }>("/auth/signup", {
      email,
      username,
      password,
      invite_token,
    })
    setAccessToken(payload.data.access_token)
    setLocaleCookie(payload.data.user.locale)
    setUser(payload.data.user)
    setStatus("authenticated")
  }, [])

  const logout = useCallback(async () => {
    try {
      await api.post("/auth/logout")
    } catch (_) {
      // Ignore network errors on logout — we clear local state regardless.
    } finally {
      clearAccessToken()
      setUser(null)
      setStatus("unauthenticated")
    }
  }, [])

  const refresh = useCallback(async () => {
    const token = await api.refresh()
    if (token) {
      try {
        const payload = await api.get<{ data: AuthUser }>("/users/me")
        setLocaleCookie(payload.data.locale)
        setUser(payload.data)
        setStatus("authenticated")
        return
      } catch (err) {
        if (err instanceof ApiError && (err.status === 401 || err.status === 403)) {
          clearAccessToken()
          setUser(null)
          setStatus("unauthenticated")
          return
        }
        // Keep the token and user state on transient errors.
        setStatus((prev) => (prev === "loading" ? "unauthenticated" : prev))
        return
      }
    }
    clearAccessToken()
    setUser(null)
    setStatus("unauthenticated")
  }, [])

  useEffect(() => {
    let active = true
    refresh().finally(() => {
      if (!active) return
      // State is already set inside refresh(), nothing extra needed.
    })

    return () => {
      active = false
    }
  }, [refresh])

  const value = useMemo(
    () => ({
      user,
      status,
      login,
      signup,
      logout,
      refresh,
    }),
    [user, status, login, signup, logout, refresh]
  )

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth() {
  const context = useContext(AuthContext)
  if (!context) {
    throw new Error("useAuth must be used within AuthProvider")
  }
  return context
}
