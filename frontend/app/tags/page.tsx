"use client"

import { useEffect, useMemo, useState } from "react"
import Link from "next/link"
import { CheckCircle, MagnifyingGlass } from "@phosphor-icons/react"
import { useTranslations } from "next-intl"
import type { Tag } from "@/lib/debates"
import { fetchMyTagFollows, fetchTags, replaceMyTagFollows } from "@/lib/debates-api"
import { ApiError } from "@/lib/api"
import { useAuth } from "@/hooks/use-auth"

export default function TagsPage() {
  const { status } = useAuth()
  const t = useTranslations("tags")
  const tCommon = useTranslations("common")
  const [tags, setTags] = useState<Tag[]>([])
  const [followedTagIDs, setFollowedTagIDs] = useState<string[]>([])
  const [savingTagID, setSavingTagID] = useState<string | null>(null)
  const [saveError, setSaveError] = useState<ApiError | null>(null)
  const [followsLoaded, setFollowsLoaded] = useState(false)
  const [query, setQuery] = useState("")
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<ApiError | null>(null)

  useEffect(() => {
    let active = true
    setLoading(true)
    fetchTags()
      .then((res) => {
        if (!active) return
        setTags(Array.isArray(res.data) ? res.data : [])
      })
      .catch((err: ApiError) => {
        if (!active) return
        setError(err)
        setTags([])
      })
      .finally(() => {
        if (!active) return
        setLoading(false)
      })

    return () => {
      active = false
    }
  }, [])

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
        setFollowedTagIDs((Array.isArray(res.data) ? res.data : []).map((tag) => tag.id))
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

  const sortedTags = useMemo(() => {
    return [...tags].sort((a, b) => a.name.localeCompare(b.name))
  }, [tags])

  const isAuthenticated = status === "authenticated"
  const followedSet = useMemo(() => new Set(followedTagIDs), [followedTagIDs])
  const normalizedQuery = query.trim().toLowerCase()
  const filteredTags = useMemo(() => {
    if (!normalizedQuery) return sortedTags
    return sortedTags.filter((tag) => {
      const name = tag.name.toLowerCase()
      const slug = tag.slug.toLowerCase()
      return name.includes(normalizedQuery) || slug.includes(normalizedQuery)
    })
  }, [normalizedQuery, sortedTags])
  const visibleTags = filteredTags
  const canToggleFollow = isAuthenticated && followsLoaded

  const toggleFollow = async (tagID: string) => {
    if (savingTagID) return
    if (!canToggleFollow) return
    setSaveError(null)
    const current = new Set(followedTagIDs)
    const isFollowing = current.has(tagID)
    if (!isFollowing && current.size >= 30) return
    if (isFollowing) {
      current.delete(tagID)
    } else {
      current.add(tagID)
    }
    const nextIDs = Array.from(current)
    const previousIDs = followedTagIDs
    setFollowedTagIDs(nextIDs)
    setSavingTagID(tagID)
    try {
      const res = await replaceMyTagFollows(nextIDs)
      setFollowedTagIDs((Array.isArray(res.data) ? res.data : []).map((tag) => tag.id))
    } catch (err) {
      setFollowedTagIDs(previousIDs)
      setSaveError(err instanceof ApiError ? err : new ApiError(500, "UNKNOWN_ERROR", tCommon("tryAgain")))
    } finally {
      setSavingTagID(null)
    }
  }

  return (
    <div className="min-h-screen bg-[var(--bg-primary)]">
      <div className="sticky top-0 z-20 bg-[var(--bg-primary)]/95 backdrop-blur-sm border-b border-[var(--border-subtle)]">
        <div className="max-w-[820px] mx-auto px-3 sm:px-6 py-3">
          <div className="relative">
            <MagnifyingGlass
              size={16}
              className="absolute left-3 top-1/2 -translate-y-1/2 text-[var(--text-muted)]"
              aria-hidden
            />
            <input
              value={query}
              onChange={(event) => setQuery(event.target.value)}
              placeholder={t("searchPlaceholder")}
              className="w-full h-10 rounded-lg border border-[var(--border-default)] bg-[var(--bg-surface)] pl-9 pr-3 text-[14px] text-[var(--text-primary)] outline-none focus:border-[var(--border-strong)] font-sans"
            />
          </div>
        </div>
      </div>

      <div className="max-w-[820px] mx-auto px-3 sm:px-6 py-4">
        <div className="mb-4 flex items-center justify-end gap-3">
          <Link href="/" className="text-[12px] font-semibold text-[var(--text-primary)] font-sans hover:underline">
            {tCommon("backToHome")}
          </Link>
        </div>

        {saveError && <div className="mb-3 text-[12px] text-rose-700 font-sans">{saveError.message}</div>}

        {loading && (
          <div className="py-16 text-center text-[var(--text-secondary)] text-sm font-sans">{tCommon("loadingTags")}</div>
        )}

        {!loading && error && (
          <div className="py-16 text-center">
            <div className="text-[var(--text-secondary)] text-base font-sans mb-1">{t("loadError")}</div>
            <div className="text-[var(--text-muted)] text-sm font-sans">{error.message}</div>
          </div>
        )}

        {!loading && !error && visibleTags.length > 0 && (
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
            {visibleTags.map((tag) => (
              <div key={tag.id} className="rounded-lg border border-[var(--border-default)] bg-[var(--bg-surface)] px-4 py-3">
                <div className="flex items-center gap-3">
                  <div className="min-w-0 flex-1">
                    <Link
                      href={`/tags/${encodeURIComponent(tag.slug)}`}
                      className="text-[14px] font-semibold text-[var(--text-primary)] font-sans hover:underline"
                    >
                      {tag.name}
                    </Link>
                    <div className="mt-0.5 text-[12px] text-[var(--text-muted)] font-sans">/{tag.slug}</div>
                  </div>
                  {isAuthenticated && followsLoaded && (
                    <button
                      onClick={() => void toggleFollow(tag.id)}
                      disabled={(savingTagID === tag.id) || (!followedSet.has(tag.id) && followedTagIDs.length >= 30)}
                      className={`shrink-0 inline-flex items-center gap-1 rounded-md border px-2.5 py-1.5 text-[11px] font-semibold font-sans transition-colors ${
                        followedSet.has(tag.id)
                          ? "border-[var(--text-primary)] bg-[var(--text-primary)] text-[var(--bg-primary)]"
                          : "border-[var(--border-default)] text-[var(--text-secondary)] hover:border-[var(--border-strong)]"
                      } disabled:opacity-50 disabled:cursor-not-allowed`}
                    >
                      {followedSet.has(tag.id) && <CheckCircle size={11} weight="fill" />}
                      {savingTagID === tag.id ? tCommon("saving") : followedSet.has(tag.id) ? tCommon("following") : tCommon("follow")}
                    </button>
                  )}
                </div>
              </div>
            ))}
          </div>
        )}

        {!loading && !error && visibleTags.length === 0 && (
          <div className="py-16 text-center text-[var(--text-secondary)] text-sm font-sans">
            {normalizedQuery ? t("noMatch") : t("none")}
          </div>
        )}
      </div>
    </div>
  )
}
