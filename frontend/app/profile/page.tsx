"use client"

import { useEffect } from "react"
import { useRouter } from "next/navigation"
import { useTranslations } from "next-intl"
import { useAuth } from "@/hooks/use-auth"

export default function ProfileRootPage() {
  const router = useRouter()
  const { status, user } = useAuth()
  const t = useTranslations("profile")

  useEffect(() => {
    if (status === "unauthenticated") {
      router.replace("/auth/login?redirect=/profile")
      return
    }

    if (status === "authenticated" && user?.username) {
      router.replace(`/profile/${encodeURIComponent(user.username)}`)
    }
  }, [router, status, user?.username])

  return (
    <div className="min-h-screen bg-[var(--bg-primary)]">
      <div className="max-w-[820px] mx-auto px-3 sm:px-6 py-8">
        <div className="text-[var(--text-muted)] text-[13px] font-sans">{t("loading")}</div>
      </div>
    </div>
  )
}
