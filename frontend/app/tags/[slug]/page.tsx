"use client"

import { useCallback, useEffect, useState, Suspense } from "react"
import Link from "next/link"
import { useParams, useRouter, useSearchParams } from "next/navigation"
import { useTranslations } from "next-intl"
import { ArrowLeft, CheckCircle, Plus } from "@phosphor-icons/react"
import type { DebateFeed, DebateFeedItem, Tag } from "@/lib/debates"
import { fetchDebates, fetchMyTagFollows, fetchTags, replaceMyTagFollows, toggleDebateVote } from "@/lib/debates-api"
import { ApiError } from "@/lib/api"
import DebateCard from "@/components/debate-card"
import { useAuth } from "@/hooks/use-auth"

type Tab = Exclude<DebateFeed, "following_tags">

export default function TagPageWrapper() {
  return (
    <Suspense fallback={<div className="min-h-screen bg-[var(--bg-primary)]" />}>
      <TagPage />
    </Suspense>
  )
}

function TagPage() {
  const params = useParams()
  const slugParam = params?.slug
  const slug = Array.isArray(slugParam) ? slugParam[0] : slugParam

  const searchParams = useSearchParams()
  const router = useRouter()
  const urlTab = searchParams.get("tab") as Tab | null

  const [activeTab, setActiveTab] = useState<Tab>(urlTab || "trending")
  const [debates, setDebates] = useState<DebateFeedItem[]>([])
  const [tag, setTag] = useState<Tag | null>(null)
  const [followedTagIDs, setFollowedTagIDs] = useState<string[]>([])
  const [followsLoaded, setFollowsLoaded] = useState(false)
  const [savingFollow, setSavingFollow] = useState(false)
  const [saveError, setSaveError] = useState<ApiError | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<ApiError | null>(null)
  const [meta, setMeta] = useState<{ total: number } | null>(null)
  const [tabTotals, setTabTotals] = useState<Record<Tab, number | null>>({
    trending: null,
    new: null,
    live: null,
    needs_challenger: null,
    following: null,
  })
  const { status } = useAuth()
  const t = useTranslations("tags")
  const tHome = useTranslations("home")
  const tCommon = useTranslations("common")

  const setTabInUrl = useCallback(
    (tab: Tab) => {
      if (!slug) return
      const params = new URLSearchParams(searchParams.toString())
      params.set("tab", tab)
      const query = params.toString()
      router.replace(`/tags/${slug}${query ? `?${query}` : ""}`, { scroll: false })
    },
    [router, searchParams, slug]
  )

  useEffect(() => {
    let active = true
    fetchTags()
      .then((res) => {
        if (!active) return
        const list = Array.isArray(res.data) ? res.data : []
        setTag(list.find((item) => item.slug === slug) ?? null)
      })
      .catch(() => {
        if (!active) return
        setTag(null)
      })
    return () => {
      active = false
    }
  }, [slug])

  useEffect(() => {
    if (status !== "authenticated") {
      setFollowedTagIDs([])
      setFollowsLoaded(status === "unauthenticated")
      return
    }

    let active = true
    setFollowsLoaded(false)
    fetchMyTagFollows()
      .then((res) => {
        if (!active) return
        setFollowedTagIDs((Array.isArray(res.data) ? res.data : []).map((item) => item.id))
      })
      .catch(() => {
        if (!active) return
        setFollowedTagIDs([])
      })
      .finally(() => {
        if (!active) return
        setFollowsLoaded(true)
      })
    return () => {
      active = false
    }
  }, [status])

  useEffect(() => {
    if (!slug) return
    let active = true
    setLoading(true)
    setError(null)

    if (activeTab === "following" && status !== "authenticated") {
      if (!active) return
      setDebates([])
      setMeta(null)
      setLoading(false)
      return () => {
        active = false
      }
    }

    fetchDebates({ feed: activeTab, tagSlug: slug, page: 1, perPage: 20 })
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
  }, [activeTab, slug, status])

  useEffect(() => {
    if (status === "unauthenticated" && activeTab === "following") {
      setActiveTab("trending")
      setTabInUrl("trending")
    }
  }, [status, activeTab, setTabInUrl])

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

  const isFollowingTag = Boolean(tag && followedTagIDs.includes(tag.id))
  const hasReachedFollowLimit = !isFollowingTag && followedTagIDs.length >= 30

  const handleToggleTagFollow = useCallback(async () => {
    if (!tag || status !== "authenticated" || savingFollow) return
    if (hasReachedFollowLimit) return

    const nextIDs = isFollowingTag
      ? followedTagIDs.filter((id) => id !== tag.id)
      : [...followedTagIDs, tag.id]

    setSavingFollow(true)
    setSaveError(null)
    try {
      const res = await replaceMyTagFollows(nextIDs)
      setFollowedTagIDs((Array.isArray(res.data) ? res.data : []).map((item) => item.id))
    } catch (err) {
      setSaveError(err instanceof ApiError ? err : new ApiError(500, "UNKNOWN_ERROR", tCommon("tryAgain")))
    } finally {
      setSavingFollow(false)
    }
  }, [followedTagIDs, hasReachedFollowLimit, isFollowingTag, savingFollow, status, tag])

  const tabs: { key: Tab; label: string; count: number }[] = [
    { key: "trending", label: tHome("tabs.trending"), count: tabTotals.trending ?? 0 },
    { key: "new", label: tHome("tabs.new"), count: tabTotals.new ?? 0 },
    { key: "live", label: tHome("tabs.live"), count: tabTotals.live ?? 0 },
    { key: "needs_challenger", label: tHome("tabs.open"), count: tabTotals.needs_challenger ?? 0 },
  ]
  if (status === "authenticated") {
    tabs.push({ key: "following", label: tCommon("following"), count: tabTotals.following ?? 0 })
  }

  return (
    <div className="min-h-screen bg-[var(--bg-primary)]">
      <div className="sticky top-0 z-20 bg-[var(--bg-primary)]/95 backdrop-blur-sm border-b border-[var(--border-subtle)]">
        <div className="max-w-[820px] mx-auto px-3 sm:px-6">
          <div className="flex items-center justify-between py-3">
            <Link
              href="/tags"
              className="flex items-center gap-2 text-[12px] font-medium text-[var(--text-muted)] hover:text-[var(--text-primary)] font-sans"
            >
              <ArrowLeft size={14} />
              {tCommon("tags")}
            </Link>
          </div>

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

      <div className="max-w-[820px] mx-auto px-3 sm:px-6 py-6">
        <div className="mb-6 rounded-xl border border-[var(--border-subtle)] bg-[var(--bg-surface)] p-5">
          <div className="flex items-start gap-3">
            <div className="min-w-0">
              <div className="text-[20px] font-semibold text-[var(--text-primary)] font-sans">
                {tag?.name ?? t("titleFallback")}
              </div>
            </div>
            {status === "authenticated" && followsLoaded && tag && (
              <button
                onClick={handleToggleTagFollow}
                disabled={savingFollow || hasReachedFollowLimit}
                className={`ml-auto shrink-0 inline-flex items-center gap-1.5 rounded-full border px-3 py-1.5 text-[11px] font-semibold font-sans transition-colors ${
                  isFollowingTag
                    ? "border-[var(--text-primary)] bg-[var(--text-primary)] text-[var(--bg-primary)]"
                    : "border-[var(--border-default)] text-[var(--text-secondary)] hover:border-[var(--border-strong)]"
                } disabled:opacity-50 disabled:cursor-not-allowed`}
              >
                {isFollowingTag ? <CheckCircle size={12} weight="fill" /> : <Plus size={12} />}
                {savingFollow ? tCommon("saving") : isFollowingTag ? tCommon("following") : tCommon("follow")}
              </button>
            )}
          </div>
          {status === "authenticated" && followsLoaded && hasReachedFollowLimit && (
            <div className="mt-3 text-[11px] text-[var(--text-muted)] font-sans">{t("followLimit")}</div>
          )}
          {saveError && <div className="mt-2 text-[12px] text-rose-700 font-sans">{saveError.message}</div>}
        </div>

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
              <div className="text-[var(--text-secondary)] text-base font-sans mb-1">{tHome("empty.loadError")}</div>
              <div className="text-[var(--text-muted)] text-sm font-sans">{error.message}</div>
            </div>
          ) : (
            <div className="py-20 text-center">
              <div className="text-[var(--text-secondary)] text-base font-sans mb-1">{t("noDebates")}</div>
              <div className="text-[var(--text-muted)] text-sm font-sans">{t("checkBack")}</div>
            </div>
          )}
        </div>

        {debates.length > 0 && (
          <div className="py-8 text-center">
            <div className="text-[var(--text-muted)] text-sm font-sans">
              {meta?.total
                ? tHome("empty.showingOf", { count: debates.length, total: meta.total })
                : tHome("empty.showing", { count: debates.length })}
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
