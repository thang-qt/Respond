"use client"

import { Suspense, useCallback, useEffect, useMemo, useState } from "react"
import Link from "next/link"
import { useRouter, useSearchParams } from "next/navigation"
import { useTranslations } from "next-intl"
import BrandLogo from "@/components/brand-logo"
import { useAuth } from "@/hooks/use-auth"
import { api, ApiError } from "@/lib/api"
import { siteConfig } from "@/lib/config"

export default function VerifyEmailPageWrapper() {
  return (
    <Suspense fallback={<div className="min-h-screen bg-[var(--bg-primary)]" />}>
      <VerifyEmailPage />
    </Suspense>
  )
}

function VerifyEmailPage() {
  const router = useRouter()
  const searchParams = useSearchParams()
  const { status, refresh } = useAuth()
  const t = useTranslations("verifyEmail")
  const tCommon = useTranslations("common")

  const tokenFromUrl = useMemo(() => searchParams.get("token")?.trim() ?? "", [searchParams])

  const [tokenInput, setTokenInput] = useState(tokenFromUrl)
  const [submitting, setSubmitting] = useState(false)
  const [successMessage, setSuccessMessage] = useState<string | null>(null)
  const [errorMessage, setErrorMessage] = useState<string | null>(null)
  const [lastAttemptedToken, setLastAttemptedToken] = useState<string | null>(null)

  useEffect(() => {
    setTokenInput(tokenFromUrl)
  }, [tokenFromUrl])

  const verifyToken = useCallback(
    async (token: string) => {
      const normalized = token.trim()
      if (!normalized || submitting) return

      setSubmitting(true)
      setErrorMessage(null)
      setSuccessMessage(null)

      try {
        const res = await api.post<{ message?: string }>("/auth/verify-email", { token: normalized })
        await refresh()
        setSuccessMessage(res?.message || t("success"))
      } catch (err) {
        if (err instanceof ApiError) {
          setErrorMessage(err.message)
        } else {
          setErrorMessage(tCommon("tryAgain"))
        }
      } finally {
        setSubmitting(false)
      }
    },
    [refresh, submitting, t, tCommon]
  )

  useEffect(() => {
    if (!tokenFromUrl || lastAttemptedToken === tokenFromUrl) return
    setLastAttemptedToken(tokenFromUrl)
    void verifyToken(tokenFromUrl)
  }, [lastAttemptedToken, tokenFromUrl, verifyToken])

  async function onSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const normalized = tokenInput.trim()
    if (!normalized) {
      setErrorMessage(t("required"))
      return
    }
    setLastAttemptedToken(normalized)
    await verifyToken(normalized)
  }

  return (
    <div className="min-h-screen bg-[var(--bg-primary)]">
      <div className="max-w-[820px] mx-auto px-3 sm:px-6 py-8 sm:py-16">
        <div className="max-w-[420px] mx-auto">
          <div className="mb-6">
            <Link href="/" className="inline-flex items-center gap-2 text-[var(--text-primary)] mb-4">
              <BrandLogo size={20} />
              <span className="text-[15px] font-semibold font-sans">{siteConfig.name}</span>
            </Link>
            <h1 className="text-[22px] font-semibold text-[var(--text-primary)] font-sans">{t("title")}</h1>
            <p className="mt-2 text-[13px] text-[var(--text-secondary)] font-sans leading-relaxed">
              {t("description")}
            </p>
          </div>

          <div className="bg-[var(--bg-surface)] border border-[var(--border-default)] rounded-xl p-5 sm:p-6 shadow-[0px_1px_2px_rgba(55,50,47,0.06)]">
            <form onSubmit={onSubmit} className="space-y-4">
              <label className="block">
                <span className="text-[12px] text-[var(--text-secondary)] font-sans">{t("token")}</span>
                <input
                  type="text"
                  value={tokenInput}
                  onChange={(event) => setTokenInput(event.target.value)}
                  placeholder={t("tokenPlaceholder")}
                  className="mt-1 w-full rounded-md border border-[var(--border-default)] bg-[var(--bg-surface)] px-3 py-2 text-[14px] text-[var(--text-secondary)] font-sans"
                  autoFocus={!tokenFromUrl}
                />
              </label>

              {successMessage && (
                <div className="rounded-lg border border-emerald-200 bg-emerald-50 px-3 py-2 text-[12px] text-emerald-700 font-sans">
                  {successMessage}
                </div>
              )}

              {errorMessage && (
                <div className="rounded-lg border border-[var(--error)] bg-[var(--error-light)] px-3 py-2 text-[12px] text-[var(--error)] font-sans">
                  {errorMessage}
                </div>
              )}

              <button
                type="submit"
                disabled={submitting || tokenInput.trim().length === 0}
                className="w-full px-3 py-2 rounded-md text-[12px] font-medium font-sans bg-[var(--text-primary)] text-[var(--bg-primary)] hover:opacity-90 disabled:opacity-60"
              >
                {submitting ? t("submitting") : t("submit")}
              </button>
            </form>

            <div className="mt-4 flex flex-wrap items-center gap-3 text-[12px] font-sans">
              <Link href="/" className="text-[var(--text-secondary)] hover:underline">
                {tCommon("backToHome")}
              </Link>
              {status === "authenticated" && (
                <button type="button" onClick={() => router.push("/settings")} className="text-[var(--text-primary)] font-medium hover:underline">
                  {tCommon("openSettings")}
                </button>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
