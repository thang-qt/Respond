"use client"

import { useCallback, useEffect, useMemo, useState, Suspense } from "react"
import Link from "next/link"
import { useRouter, useSearchParams } from "next/navigation"
import { useTranslations } from "next-intl"
import { Star } from "@phosphor-icons/react"
import { ApiError } from "@/lib/api"
import type { DebateFeedItem, Tag } from "@/lib/debates"
import { fetchDebates, fetchMyTagFollows, fetchTags, toggleDebateVote } from "@/lib/debates-api"
import { resendVerificationEmail } from "@/lib/settings-api"
import DebateCard from "@/components/debate-card"
import { useAuth } from "@/hooks/use-auth"

type StatusTab = "trending" | "new" | "live" | "needs_challenger"

export default function HomePageWrapper() {
  return (
    <Suspense fallback={<div className="min-h-screen bg-[var(--bg-primary)]" />}>
      <HomePage />
    </Suspense>
  )
}

function HomePage() {
  const searchParams = useSearchParams()
  const router = useRouter()
  const { status, user, refresh } = useAuth()
  const t = useTranslations("home")
  const tCommon = useTranslations("common")

  const urlTab = searchParams.get("tab")
  const urlTag = searchParams.get("tag") ?? searchParams.get("category") ?? searchParams.get("category_id")

  const initialTab: StatusTab =
    urlTab === "new" || urlTab === "live" || urlTab === "needs_challenger" ? urlTab : "trending"

  const [activeTab, setActiveTab] = useState<StatusTab>(initialTab)
  const [selectedTag, setSelectedTag] = useState<string | null>(urlTag)
  const [debates, setDebates] = useState<DebateFeedItem[]>([])
  const [tags, setTags] = useState<Tag[]>([])
  const [tagFollows, setTagFollows] = useState<Tag[]>([])
  const [followsLoaded, setFollowsLoaded] = useState(false)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<ApiError | null>(null)
  const [meta, setMeta] = useState<{ total: number } | null>(null)
  const [tabTotals, setTabTotals] = useState<Record<StatusTab, number | null>>({
    trending: null,
    new: null,
    live: null,
    needs_challenger: null,
  })
  const [resendSaving, setResendSaving] = useState(false)
  const [resendMessage, setResendMessage] = useState<string | null>(null)
  const [resendError, setResendError] = useState<string | null>(null)

  const setTabInUrl = useCallback(
    (tab: StatusTab) => {
      const params = new URLSearchParams(searchParams.toString())
      params.set("tab", tab)
      router.replace(`/?${params.toString()}`, { scroll: false })
    },
    [router, searchParams]
  )

  useEffect(() => {
    let active = true
    fetchTags()
      .then((res) => {
        if (!active) return
        setTags(res.data)
      })
      .catch(() => {
        if (!active) return
        setTags([])
      })
    return () => {
      active = false
    }
  }, [])

  useEffect(() => {
    if (status !== "authenticated") {
      setTagFollows([])
      setFollowsLoaded(status === "unauthenticated")
      return
    }

    let active = true
    setFollowsLoaded(false)
    fetchMyTagFollows()
      .then((res) => {
        if (!active) return
        setTagFollows(Array.isArray(res.data) ? res.data : [])
      })
      .catch(() => {
        if (!active) return
        setTagFollows([])
      })
      .finally(() => {
        if (!active) return
        setFollowsLoaded(true)
      })

    return () => {
      active = false
    }
  }, [status])

  const followedTagSlugs = useMemo(() => tagFollows.map((tag) => tag.slug), [tagFollows])
  const hasFollowedTags = status === "authenticated" && followsLoaded && followedTagSlugs.length > 0
  const visibleFilterTags = hasFollowedTags ? tagFollows : tags

  useEffect(() => {
    if (!hasFollowedTags || !selectedTag) return
    if (!followedTagSlugs.includes(selectedTag)) {
      setSelectedTag(null)
    }
  }, [followedTagSlugs, hasFollowedTags, selectedTag])

  useEffect(() => {
    let active = true
    setLoading(true)
    setError(null)

    if (status === "authenticated" && !followsLoaded) {
      return () => {
        active = false
      }
    }

    const feed = hasFollowedTags && activeTab === "trending" ? "following_tags" : activeTab
    const scopedTagSlugs =
      hasFollowedTags && activeTab !== "trending" ? (selectedTag ? [selectedTag] : followedTagSlugs) : null

    fetchDebates({
      feed,
      tagSlug: !scopedTagSlugs ? selectedTag : null,
      tagSlugs: scopedTagSlugs,
      page: 1,
      perPage: 20,
    })
      .then((res) => {
        if (!active) return
        setDebates(Array.isArray(res.data) ? res.data : [])
        setMeta(res.meta ? { total: res.meta.total } : null)
        if (res.meta?.total !== undefined) {
          setTabTotals((prev) => ({ ...prev, [activeTab]: res.meta!.total }))
        }
      })
      .catch((err: ApiError) => {
        if (!active) return
        setError(err)
        setDebates([])
        setMeta(null)
      })
      .finally(() => {
        if (!active) return
        setLoading(false)
      })

    return () => {
      active = false
    }
  }, [activeTab, followedTagSlugs, followsLoaded, hasFollowedTags, selectedTag, status])

  const handleToggleUpvote = useCallback(async (debateId: string) => {
    const res = await toggleDebateVote(debateId)
    setDebates((prev) =>
      prev.map((debate) =>
        debate.id === res.data.debate_id
          ? { ...debate, upvote_count: res.data.upvote_count, viewer_has_upvoted: res.data.voted }
          : debate
      )
    )
  }, [])

  const tabs: { key: StatusTab; label: string; count: number }[] = [
    { key: "trending", label: t("tabs.trending"), count: tabTotals.trending ?? 0 },
    { key: "new", label: t("tabs.new"), count: tabTotals.new ?? 0 },
    { key: "live", label: t("tabs.live"), count: tabTotals.live ?? 0 },
    { key: "needs_challenger", label: t("tabs.open"), count: tabTotals.needs_challenger ?? 0 },
  ]

  const showFollowOnboarding = status === "authenticated" && followsLoaded && tagFollows.length === 0
  const showVerifyBanner = status === "authenticated" && !user?.email_verified

  const handleResendVerification = useCallback(async () => {
    setResendSaving(true)
    setResendError(null)
    setResendMessage(null)
    try {
      await resendVerificationEmail()
      setResendMessage(t("verify.accepted"))
    } catch (err) {
      if (err instanceof ApiError && err.code === "AUTH_ALREADY_VERIFIED") {
        await refresh()
        setResendMessage(t("verify.alreadyVerified"))
      } else {
        setResendError(err instanceof Error ? err.message : t("verify.failed"))
      }
    } finally {
      setResendSaving(false)
    }
  }, [refresh, t])

  return (
    <div className="min-h-screen bg-[var(--bg-primary)]">
      <div className="sticky top-0 z-20 bg-[var(--bg-primary)]/95 backdrop-blur-sm border-b border-[var(--border-subtle)]">
        <div className="max-w-[820px] mx-auto px-3 sm:px-6">
          <div className="flex items-center gap-0 overflow-x-auto">
            {tabs.map((tab) => (
              <button
                key={tab.key}
                onClick={() => {
                  setActiveTab(tab.key)
                  setTabInUrl(tab.key)
                }}
                className={`px-3 sm:px-4 py-3 text-[13px] font-medium font-sans whitespace-nowrap transition-colors relative ${
                  activeTab === tab.key
                    ? "text-[var(--text-primary)]"
                    : "text-[var(--text-muted)] hover:text-[var(--text-secondary)]"
                }`}
              >
                {tab.label}
                {tab.key === "live" && (
                  <span className="ml-1.5 inline-block w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse" />
                )}
                {activeTab === tab.key && (
                  <div className="absolute bottom-0 left-2 right-2 h-[2px] bg-[var(--text-primary)] rounded-full" />
                )}
              </button>
            ))}
          </div>
        </div>
      </div>

      <div className="max-w-[820px] mx-auto px-3 sm:px-6 py-4">
        {showVerifyBanner && (
          <div className="mb-4 rounded-lg border border-[var(--border-default)] bg-[var(--bg-surface)] px-4 py-3">
            <div className="text-[12px] text-[var(--text-secondary)] font-sans">
              {t("verify.body")}
            </div>
            <div className="mt-2 flex flex-wrap items-center gap-3">
              <button
                type="button"
                onClick={() => void handleResendVerification()}
                disabled={resendSaving}
                className="text-[12px] font-semibold text-[var(--text-primary)] font-sans hover:underline disabled:opacity-60"
              >
                {resendSaving ? tCommon("sending") : t("verify.resend")}
              </button>
              <Link href="/settings" className="text-[12px] text-[var(--text-secondary)] font-sans hover:underline">
                {tCommon("openSettings")}
              </Link>
            </div>
            {resendMessage && <div className="mt-1 text-[12px] text-emerald-700 font-sans break-words">{resendMessage}</div>}
            {resendError && <div className="mt-1 text-[12px] text-[var(--error)] font-sans break-words">{resendError}</div>}
          </div>
        )}

        {showFollowOnboarding && (
          <div className="mb-4 rounded-lg border border-[var(--border-default)] bg-[var(--bg-surface)] px-4 py-3 flex flex-col sm:flex-row sm:items-center gap-2 sm:gap-3">
            <div className="text-[12px] text-[var(--text-secondary)] font-sans">
              {t("onboarding.body")}
            </div>
            <div className="sm:ml-auto flex items-center gap-3">
              <Link href="/explore" className="text-[12px] font-medium text-[var(--text-secondary)] font-sans hover:underline">
                {t("onboarding.explore")}
              </Link>
              <Link href="/tags" className="text-[12px] font-semibold text-[var(--text-primary)] font-sans hover:underline">
                {t("onboarding.followTags")}
              </Link>
            </div>
          </div>
        )}

        <div className="flex items-center gap-1.5 pb-4 overflow-x-auto scrollbar-hide">
          <button
            onClick={() => setSelectedTag(null)}
            className={`px-2.5 py-1 text-[11px] font-medium rounded-full border transition-colors whitespace-nowrap font-sans ${
              !selectedTag
                ? "bg-[var(--text-primary)] text-[var(--bg-primary)] border-[var(--text-primary)]"
                : "bg-[var(--bg-surface)] text-[var(--text-secondary)] border-[var(--border-default)] hover:border-[var(--border-strong)]"
            }`}
          >
            {tCommon("all")}
          </button>
          {visibleFilterTags.map((tag) => (
            <button
              key={tag.slug}
              onClick={() => setSelectedTag(selectedTag === tag.slug ? null : tag.slug)}
              className={`px-2.5 py-1 text-[11px] font-medium rounded-full border transition-colors whitespace-nowrap font-sans ${
                selectedTag === tag.slug
                  ? "bg-[var(--text-primary)] text-[var(--bg-primary)] border-[var(--text-primary)]"
                  : "bg-[var(--bg-surface)] text-[var(--text-secondary)] border-[var(--border-default)] hover:border-[var(--border-strong)]"
              }`}
            >
              {tag.name}
            </button>
          ))}
        </div>

        {activeTab === "trending" && !selectedTag && !hasFollowedTags && (
          <div className="mb-4 px-4 py-3 bg-[var(--bg-surface)] border border-[var(--border-default)] rounded-lg shadow-[0px_1px_2px_rgba(55,50,47,0.06)]">
            <div className="flex items-center gap-2 mb-1.5">
              <Star size={14} className="text-[var(--text-primary)]" />
              <span className="text-[11px] font-semibold text-[var(--text-primary)] font-sans uppercase tracking-wider">
                {t("dailyPrompt.label")}
              </span>
            </div>
            <div className="text-[var(--text-primary)] text-[15px] font-semibold font-sans leading-snug">
              {t("dailyPrompt.topic")}
            </div>
            <div className="flex items-center gap-3 mt-2">
              <span className="text-[11px] text-[var(--text-muted)] font-sans">{t("dailyPrompt.tag")}</span>
              <span className="text-[11px] text-[var(--text-muted)] font-sans">{t("dailyPrompt.time")}</span>
              <button className="ml-auto text-[12px] font-medium text-[var(--text-primary)] font-sans hover:underline">
                {t("dailyPrompt.cta")}
              </button>
            </div>
          </div>
        )}

        <div className="flex flex-col gap-2">
          {debates.length > 0 ? (
            debates.map((debate) => (
              <DebateCard key={debate.id} debate={debate} onToggleUpvote={handleToggleUpvote} />
            ))
          ) : loading ? (
            <div className="py-16 text-center">
              <div className="text-[var(--text-secondary)] text-base font-sans mb-1">{tCommon("loadingDebates")}</div>
            </div>
          ) : error ? (
            <div className="py-20 text-center">
              <div className="text-[var(--text-secondary)] text-base font-sans mb-1">{t("empty.loadError")}</div>
              <div className="text-[var(--text-muted)] text-sm font-sans">{error.message}</div>
            </div>
          ) : (
            <div className="py-20 text-center">
              <div className="text-[var(--text-secondary)] text-base font-sans mb-1">{t("empty.noMatches")}</div>
              <div className="text-[var(--text-muted)] text-sm font-sans">{t("empty.tryDifferent")}</div>
            </div>
          )}
        </div>

        {debates.length > 0 && (
          <div className="py-8 text-center">
            <div className="text-[var(--text-muted)] text-sm font-sans">
              {meta?.total
                ? t("empty.showingOf", { count: debates.length, total: meta.total })
                : t("empty.showing", { count: debates.length })}
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
