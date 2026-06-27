import { localeCookieName } from "@/i18n/client-locale"
import { clearAccessToken, getAccessToken, setAccessToken } from "@/lib/auth-token"

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api/v1"

export class ApiError extends Error {
  code: string
  status: number

  constructor(status: number, code: string, message: string) {
    super(message)
    this.status = status
    this.code = code
  }
}

async function parseError(res: Response): Promise<ApiError> {
  try {
    const payload = await res.json()
    if (payload?.error?.code && payload?.error?.message) {
      return new ApiError(res.status, payload.error.code, payload.error.message)
    }
  } catch (_) {
    // Ignore JSON parse errors.
  }
  return new ApiError(res.status, "UNKNOWN_ERROR", "Something went wrong. Try again.")
}

function getClientLocale(): string | null {
  if (typeof document === "undefined") return null

  const cookie = document.cookie
    .split(";")
    .map((part) => part.trim())
    .find((part) => part.startsWith(`${localeCookieName}=`))

  if (cookie) {
    return decodeURIComponent(cookie.slice(localeCookieName.length + 1))
  }

  return navigator.language || null
}

async function refreshAccessToken(): Promise<string | null> {
  try {
    const res = await fetch(`${API_BASE}/auth/refresh`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        ...(getClientLocale() ? { "Accept-Language": getClientLocale() as string } : {}),
      },
      credentials: "include",
    })

    if (!res.ok) {
      clearAccessToken()
      return null
    }

    const payload = await res.json()
    const token = payload?.data?.access_token
    if (typeof token === "string" && token.length > 0) {
      setAccessToken(token)
      return token
    }
  } catch (_) {
    clearAccessToken()
  }

  return null
}

async function request<T>(
  path: string,
  options: RequestInit = {},
  retryOnUnauthorized = true
): Promise<T> {
  const token = getAccessToken()
  const locale = getClientLocale()
  const res = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...(locale ? { "Accept-Language": locale } : {}),
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...options.headers,
    },
    credentials: "include",
  })

  if (res.status === 401 && retryOnUnauthorized) {
    const refreshed = await refreshAccessToken()
    if (refreshed) {
      return request<T>(path, options, false)
    }
  }

  if (!res.ok) {
    throw await parseError(res)
  }

  if (res.status === 204) {
    return undefined as T
  }

  return res.json()
}

export const api = {
  get: <T>(path: string) => request<T>(path),
  post: <T>(path: string, body?: unknown) =>
    request<T>(path, {
      method: "POST",
      body: body ? JSON.stringify(body) : undefined,
    }),
  put: <T>(path: string, body?: unknown) =>
    request<T>(path, {
      method: "PUT",
      body: body ? JSON.stringify(body) : undefined,
    }),
  delete: <T>(path: string) => request<T>(path, { method: "DELETE" }),
  refresh: () => refreshAccessToken(),
}
