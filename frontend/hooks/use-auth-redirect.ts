"use client"

import { usePathname, useRouter, useSearchParams } from "next/navigation"
import { useAuth } from "@/hooks/use-auth"

export function useAuthRedirect() {
  const { status } = useAuth()
  const router = useRouter()
  const pathname = usePathname()
  const searchParams = useSearchParams()
  const query = searchParams.toString()
  const redirectTo = `${pathname}${query ? `?${query}` : ""}`
  const loginUrl = `/auth/login?redirect=${encodeURIComponent(redirectTo)}`

  function requireAuth() {
    if (status === "authenticated") {
      return true
    }
    router.push(loginUrl)
    return false
  }

  return { requireAuth, loginUrl, status }
}
