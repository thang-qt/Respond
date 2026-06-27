"use client"

import { useEffect, useState, createContext, useContext } from "react"
import Link from "next/link"
import { usePathname, useRouter, useSearchParams } from "next/navigation"
import { useTranslations } from "next-intl"
import FooterSection from "@/components/footer-section"
import BrandLogo from "@/components/brand-logo"
import { siteConfig } from "@/lib/config"
import { CHALLENGES_REFRESH_EVENT } from "@/lib/challenges-events"
import { fetchMyChallenges } from "@/lib/debates-api"
import { useAuth } from "@/hooks/use-auth"
import { MobileSidebar, DesktopSidebar } from "@/components/app-shell-sidebar"
import { List, X } from "@phosphor-icons/react"

interface SidebarContextValue {
  collapsed: boolean
  setCollapsed: (v: boolean) => void
}

const SidebarContext = createContext<SidebarContextValue>({
  collapsed: false,
  setCollapsed: () => { },
})

export function useSidebar() {
  return useContext(SidebarContext)
}

export default function AppShell({ children }: { children: React.ReactNode }) {
  const [collapsed, setCollapsed] = useState(false)
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false)
  const [pendingChallengesCount, setPendingChallengesCount] = useState(0)
  const pathname = usePathname()
  const router = useRouter()
  const searchParams = useSearchParams()
  const { status, user } = useAuth()
  const t = useTranslations("shell")
  const query = searchParams.toString()
  const redirectTo = `${pathname}${query ? `?${query}` : ""}`
  const loginUrl = `/auth/login?redirect=${encodeURIComponent(redirectTo)}`

  useEffect(() => {
    try {
      const stored = localStorage.getItem("sidebar_collapsed")
      if (stored === "true") {
        setCollapsed(true)
      }
    } catch (_) {
      // Ignore storage errors.
    }
  }, [])

  useEffect(() => {
    try {
      localStorage.setItem("sidebar_collapsed", collapsed ? "true" : "false")
    } catch (_) {
      // Ignore storage errors.
    }
  }, [collapsed])

  const isExpanded = !collapsed

  useEffect(() => {
    let active = true

    async function loadPendingChallenges() {
      if (status !== "authenticated") {
        if (active) setPendingChallengesCount(0)
        return
      }
      try {
        const res = await fetchMyChallenges({
          box: "inbox",
          status: "pending",
          page: 1,
          perPage: 1,
        })
        if (!active) return
        setPendingChallengesCount(res.meta?.total ?? (res.data?.length ?? 0))
      } catch {
        if (!active) return
        setPendingChallengesCount(0)
      }
    }

    const onRefresh = () => {
      void loadPendingChallenges()
    }

    void loadPendingChallenges()
    if (typeof window !== "undefined") {
      window.addEventListener(CHALLENGES_REFRESH_EVENT, onRefresh)
    }

    return () => {
      active = false
      if (typeof window !== "undefined") {
        window.removeEventListener(CHALLENGES_REFRESH_EVENT, onRefresh)
      }
    }
  }, [status, pathname])

  return (
    <SidebarContext.Provider value={{ collapsed: isExpanded, setCollapsed }}>
      <div className="min-h-screen bg-[var(--bg-primary)]">
        {/* Mobile top bar - visible only on small screens */}
        <div className="lg:hidden bg-[var(--bg-primary)] h-12 px-4 sm:px-6 flex items-center justify-between">
          <Link href="/" className="flex items-center gap-2 text-[var(--text-primary)]">
            <BrandLogo size={20} />
            <span className="text-[15px] font-semibold font-sans">{siteConfig.name}</span>
          </Link>
          <button
            onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
            className="w-9 h-9 flex items-center justify-center text-[var(--text-secondary)] hover:text-[var(--text-primary)] transition-colors"
          >
            {mobileMenuOpen ? <X size={20} /> : <List size={20} />}
          </button>
        </div>

        {/* Mobile menu overlay */}
        {mobileMenuOpen && (
          <div className="fixed inset-0 z-40 lg:hidden">
            <button
              type="button"
              aria-label={t("closeMenu")}
              className="absolute inset-0 bg-black/20 animate-in fade-in duration-200 motion-reduce:animate-none"
              onClick={() => setMobileMenuOpen(false)}
            />
            <MobileSidebar
              onClose={() => setMobileMenuOpen(false)}
              redirectTo={redirectTo}
              loginUrl={loginUrl}
              status={status}
              username={user?.username ?? null}
              userRole={user?.role ?? null}
              pendingChallengesCount={pendingChallengesCount}
              onLogout={() => {
                router.push("/")
              }}
            />
          </div>
        )}

        {/* Desktop sidebar - hidden on mobile */}
        <DesktopSidebar
          collapsed={!isExpanded}
          setCollapsed={setCollapsed}
          redirectTo={redirectTo}
          loginUrl={loginUrl}
          status={status}
          username={user?.username ?? null}
          userRole={user?.role ?? null}
          pendingChallengesCount={pendingChallengesCount}
          onLogout={() => {
            router.push("/")
          }}
        />

        <main
          className={`transition-all duration-200 ease-in-out ${collapsed ? "lg:ml-[60px]" : "lg:ml-[240px]"
            }`}
        >
          {status === "authenticated" && user?.account_status && user.account_status !== "active" && (
            <div className="mx-auto max-w-[980px] px-4 sm:px-6 pt-4">
              <div className="rounded-md border border-[var(--warning)]/40 bg-[var(--warning-light)] px-4 py-3 text-[13px] text-[var(--warning)] font-medium">
                {user.account_status === "suspended"
                  ? t("accountSuspended")
                  : t("accountBanned")}
              </div>
            </div>
          )}
          {children}
          <FooterSection />
        </main>
      </div>
    </SidebarContext.Provider>
  )
}
