"use client"

import { Suspense, useEffect, useState } from "react"
import Link from "next/link"
import { useRouter, useSearchParams } from "next/navigation"
import { useTranslations } from "next-intl"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import BrandLogo from "@/components/brand-logo"
import { siteConfig } from "@/lib/config"
import { ApiError } from "@/lib/api"
import { useAuth } from "@/hooks/use-auth"


export default function LoginPageWrapper() {
  return (
    <Suspense fallback={<div className="min-h-screen bg-[var(--bg-primary)]" />}>
      <LoginPage />
    </Suspense>
  )
}

function LoginPage() {
  const router = useRouter()
  const searchParams = useSearchParams()
  const redirectTo = searchParams.get("redirect") || "/"
  const { login, status } = useAuth()
  const t = useTranslations("auth.login")
  const tCommon = useTranslations("common")

  const [identifier, setIdentifier] = useState("")
  const [password, setPassword] = useState("")
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    if (status === "authenticated") {
      router.replace(redirectTo)
    }
  }, [status, router, redirectTo])

  async function onSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setError(null)
    setLoading(true)

    try {
      await login({ identifier: identifier.trim(), password })
      router.push(redirectTo)
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message)
      } else {
        setError(tCommon("tryAgain"))
      }
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen bg-[var(--bg-primary)]">
      <div className="max-w-[820px] mx-auto px-3 sm:px-6 py-8 sm:py-16">
        <div className="max-w-[400px] mx-auto">
          <div className="mb-6">
            <Link href="/" className="inline-flex items-center gap-2 text-[var(--text-primary)] mb-4">
              <BrandLogo size={20} />
              <span className="text-[15px] font-semibold font-sans">{siteConfig.name}</span>
            </Link>
            <h1 className="text-[22px] font-semibold text-[var(--text-primary)] font-sans">{t("title", { siteName: siteConfig.name })}</h1>
            <p className="mt-2 text-[13px] text-[var(--text-secondary)] font-sans leading-relaxed">
              {t("subtitle")}
            </p>
          </div>

          {/* Form card */}
          <div className="bg-[var(--bg-surface)] border border-[var(--border-default)] rounded-xl p-5 sm:p-6 shadow-[0px_1px_2px_rgba(55,50,47,0.06)]">
            <form onSubmit={onSubmit} className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="identifier" className="text-[12px] text-[var(--text-secondary)] font-sans">
                  {t("identifier")}
                </Label>
                <Input
                  id="identifier"
                  type="text"
                  autoComplete="username"
                  value={identifier}
                  onChange={(event) => setIdentifier(event.target.value)}
                  placeholder={t("identifierPlaceholder")}
                  required
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="password" className="text-[12px] text-[var(--text-secondary)] font-sans">
                  {t("password")}
                </Label>
                <Input
                  id="password"
                  type="password"
                  autoComplete="current-password"
                  value={password}
                  onChange={(event) => setPassword(event.target.value)}
                  placeholder="••••••••"
                  required
                />
              </div>

              {error && (
                <div className="rounded-lg border border-[var(--error)] bg-[var(--error-light)] px-3 py-2 text-[12px] text-[var(--error)] font-sans">
                  {error}
                </div>
              )}

              <Button type="submit" className="w-full" disabled={loading || status === "loading"}>
                {loading ? t("submitting") : t("submit")}
              </Button>
            </form>
          </div>

          {/* Footer link */}
          <div className="mt-5 flex items-center justify-between text-[12px] text-[var(--text-secondary)] font-sans px-1">
            <span>{t("newHere")}</span>
            <Link
              href={`/auth/signup?redirect=${encodeURIComponent(redirectTo)}`}
              className="font-medium text-[var(--text-primary)] hover:opacity-80"
            >
              {t("createAccount")}
            </Link>
          </div>
        </div>
      </div>
    </div>
  )
}
