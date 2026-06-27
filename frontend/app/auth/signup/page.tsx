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


export default function SignupPageWrapper() {
  return (
    <Suspense fallback={<div className="min-h-screen bg-[var(--bg-primary)]" />}>
      <SignupPage />
    </Suspense>
  )
}

function SignupPage() {
  const router = useRouter()
  const searchParams = useSearchParams()
  const redirectTo = searchParams.get("redirect") || "/"
  const inviteFromQuery = searchParams.get("invite") || searchParams.get("invite_token") || ""
  const invitedEmailFromQuery = searchParams.get("email") || ""
  const { signup, status } = useAuth()
  const t = useTranslations("auth.signup")
  const tCommon = useTranslations("common")
  const tFooter = useTranslations("footer")

  const [email, setEmail] = useState("")
  const [username, setUsername] = useState("")
  const [password, setPassword] = useState("")
  const [confirmPassword, setConfirmPassword] = useState("")
  const [inviteToken, setInviteToken] = useState(inviteFromQuery)
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    if (invitedEmailFromQuery) {
      setEmail(invitedEmailFromQuery)
    }
    if (inviteFromQuery) {
      setInviteToken(inviteFromQuery)
    }
  }, [inviteFromQuery, invitedEmailFromQuery])

  useEffect(() => {
    if (status === "authenticated") {
      router.replace(redirectTo)
    }
  }, [status, router, redirectTo])

  async function onSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setError(null)
    setLoading(true)

    if (username.trim().length < 5) {
      setError(t("errors.usernameShort"))
      setLoading(false)
      return
    }

    if (password !== confirmPassword) {
      setError(t("errors.passwordMismatch"))
      setLoading(false)
      return
    }

    if (inviteToken.trim().length === 0) {
      setError(t("errors.inviteOnly"))
      setLoading(false)
      return
    }

    try {
      await signup({
        email: email.trim(),
        username: username.trim(),
        password,
        invite_token: inviteToken.trim() || undefined,
      })
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
            <h1 className="text-[22px] font-semibold text-[var(--text-primary)] font-sans">{t("title")}</h1>
            <p className="mt-2 text-[13px] text-[var(--text-secondary)] font-sans leading-relaxed">
              {t("subtitle")}
            </p>
          </div>

          {/* Form card */}
          <div className="bg-[var(--bg-surface)] border border-[var(--border-default)] rounded-xl p-5 sm:p-6 shadow-[0px_1px_2px_rgba(55,50,47,0.06)]">
            <form onSubmit={onSubmit} className="space-y-4">
              <div className="rounded-md border border-[var(--border-default)] bg-[var(--bg-surface-alt)] px-3 py-2 text-[12px] font-sans">
                {inviteToken.trim().length > 0 ? (
                  <span className="text-emerald-700">{t("inviteDetected")}</span>
                ) : (
                  <span className="text-[var(--text-secondary)]">{t("inviteRequired")}</span>
                )}
              </div>

              <div className="space-y-2">
                <Label htmlFor="email" className="text-[12px] text-[var(--text-secondary)] font-sans">
                  {t("email")}
                </Label>
                <Input
                  id="email"
                  type="email"
                  autoComplete="email"
                  value={email}
                  onChange={(event) => setEmail(event.target.value)}
                  placeholder="you@example.com"
                  required
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="username" className="text-[12px] text-[var(--text-secondary)] font-sans">
                  {t("username")}
                </Label>
                <Input
                  id="username"
                  type="text"
                  autoComplete="username"
                  value={username}
                  onChange={(event) => setUsername(event.target.value)}
                  placeholder="your_username"
                  minLength={5}
                  maxLength={20}
                  pattern="[A-Za-z0-9_]+"
                  title={t("usernameTitle")}
                  required
                />
                <p className="text-[11px] text-[var(--text-muted)] font-sans">
                  {t("usernameHelp")}
                </p>
              </div>

              <div className="space-y-2">
                <Label htmlFor="password" className="text-[12px] text-[var(--text-secondary)] font-sans">
                  {t("password")}
                </Label>
                <Input
                  id="password"
                  type="password"
                  autoComplete="new-password"
                  value={password}
                  onChange={(event) => setPassword(event.target.value)}
                  placeholder={t("passwordPlaceholder")}
                  minLength={8}
                  maxLength={128}
                  required
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="confirm-password" className="text-[12px] text-[var(--text-secondary)] font-sans">
                  {t("confirmPassword")}
                </Label>
                <Input
                  id="confirm-password"
                  type="password"
                  autoComplete="new-password"
                  value={confirmPassword}
                  onChange={(event) => setConfirmPassword(event.target.value)}
                  placeholder={t("confirmPasswordPlaceholder")}
                  minLength={8}
                  maxLength={128}
                  required
                />
              </div>

              {error && (
                <div className="rounded-lg border border-[var(--error)] bg-[var(--error-light)] px-3 py-2 text-[12px] text-[var(--error)] font-sans">
                  {error}
                </div>
              )}

              <Button type="submit" className="w-full" disabled={loading || status === "loading" || inviteToken.trim().length === 0}>
                {loading ? t("submitting") : t("submit")}
              </Button>
            </form>
          </div>

          {/* Footer links */}
          <div className="mt-5 flex items-center justify-between text-[12px] text-[var(--text-secondary)] font-sans px-1">
            <span>{t("alreadyHaveAccount")}</span>
            <Link
              href={`/auth/login?redirect=${encodeURIComponent(redirectTo)}`}
              className="font-medium text-[var(--text-primary)] hover:opacity-80"
            >
              {tCommon("signIn")}
            </Link>
          </div>
          <div className="mt-4 flex items-center gap-4 text-[12px] text-[var(--text-muted)] font-sans px-1">
            <Link href="/guidelines" className="hover:text-[var(--text-secondary)] transition-colors">
              {tFooter("guidelines")}
            </Link>
            <Link href="/philosophy" className="hover:text-[var(--text-secondary)] transition-colors">
              {tFooter("philosophy")}
            </Link>
            <Link href="/about" className="hover:text-[var(--text-secondary)] transition-colors">
              {tFooter("about")}
            </Link>
          </div>
        </div>
      </div>
    </div>
  )
}
