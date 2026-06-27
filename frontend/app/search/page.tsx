"use client"

import { Suspense, useCallback, useEffect, useMemo, useState } from "react"
import Link from "next/link"
import { useRouter, useSearchParams } from "next/navigation"
import { useTranslations } from "next-intl"
import { MagnifyingGlass } from "@phosphor-icons/react"
import { ApiError } from "@/lib/api"
import type { DebateFeedItem, Tag } from "@/lib/debates"
import type { UserSearchProfile } from "@/lib/users"
import { searchDebates, searchTags, toggleDebateVote } from "@/lib/debates-api"
import { searchUsers } from "@/lib/users-api"
import DebateCard from "@/components/debate-card"
import { useAuth } from "@/hooks/use-auth"

type SearchType = "all" | "debates" | "users" | "tags"
type DebateSearchSort = "relevance" | "new"

export default function SearchPageWrapper() {
  return (
    <Suspense fallback={<div className="min-h-screen bg-[var(--bg-primary)]" />}>
      <SearchPage />
    </Suspense>
  )
}

function SearchPage() {
  const searchParams = useSearchParams()
  const router = useRouter()
  const { status } = useAuth()
  const t = useTranslations("search")
  const tCommon = useTranslations("common")

  const urlType = searchParams.get("type")
  const initialType: SearchType =
    urlType === "debates" || urlType === "users" || urlType === "tags" ? urlType : "all"

  const initialQuery = (searchParams.get("q") || "").trim()

  const [searchType, setSearchType] = useState<SearchType>(initialType)
  const [queryInput, setQueryInput] = useState(initialQuery)
  const [debouncedQuery, setDebouncedQuery] = useState(initialQuery)
  const [debateSort, setDebateSort] = useState<DebateSearchSort>("relevance")

  const [debates, setDebates] = useState<DebateFeedItem[]>([])
  const [users, setUsers] = useState<UserSearchProfile[]>([])
  const [tags, setTags] = useState<Tag[]>([])

  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<ApiError | null>(null)
  const [resultTotal, setResultTotal] = useState<number | null>(null)

  useEffect(() => {
    const timeoutID = window.setTimeout(() => {
      setDebouncedQuery(queryInput.trim())
    }, 250)
    return () => window.clearTimeout(timeoutID)
  }, [queryInput])

  useEffect(() => {
    const params = new URLSearchParams(searchParams.toString())
    if (debouncedQuery) {
      params.set("q", debouncedQuery)
    } else {
      params.delete("q")
    }
    params.set("type", searchType)
    const nextQuery = params.toString()
    if (nextQuery !== searchParams.toString()) {
      router.replace(`/search?${nextQuery}`, { scroll: false })
    }
  }, [debouncedQuery, router, searchParams, searchType])

  useEffect(() => {
    if (!debouncedQuery) {
      setLoading(false)
      setError(null)
      setResultTotal(null)
      setDebates([])
      setUsers([])
      setTags([])
      return
    }

    if (status === "loading") {
      setLoading(true)
      return
    }

    let active = true
    setLoading(true)
    setError(null)

    const run = async () => {
      try {
        if (searchType === "debates") {
          const res = await searchDebates({
            q: debouncedQuery,
            sort: debateSort,
            page: 1,
            perPage: 20,
          })
          if (!active) return
          setDebates(Array.isArray(res.data) ? res.data : [])
          setUsers([])
          setTags([])
          setResultTotal(res.meta?.total ?? null)
          return
        }

        if (searchType === "users") {
          const res = await searchUsers(debouncedQuery, { page: 1, perPage: 20 })
          if (!active) return
          setUsers(Array.isArray(res.data) ? res.data : [])
          setDebates([])
          setTags([])
          setResultTotal(res.meta?.total ?? null)
          return
        }

        if (searchType === "all") {
          const [debateRes, userRes, tagRes] = await Promise.all([
            searchDebates({
              q: debouncedQuery,
              sort: debateSort,
              page: 1,
              perPage: 6,
            }),
            searchUsers(debouncedQuery, { page: 1, perPage: 6 }),
            searchTags({ q: debouncedQuery, limit: 8 }),
          ])
          if (!active) return
          const nextDebates = Array.isArray(debateRes.data) ? debateRes.data : []
          const nextUsers = Array.isArray(userRes.data) ? userRes.data : []
          const nextTags = Array.isArray(tagRes.data) ? tagRes.data : []
          setDebates(nextDebates)
          setUsers(nextUsers)
          setTags(nextTags)
          setResultTotal(nextDebates.length + nextUsers.length + nextTags.length)
          return
        }

        const res = await searchTags({ q: debouncedQuery, limit: 30 })
        if (!active) return
        setTags(Array.isArray(res.data) ? res.data : [])
        setDebates([])
        setUsers([])
        setResultTotal(Array.isArray(res.data) ? res.data.length : null)
      } catch (err) {
        if (!active) return
        setError(err instanceof ApiError ? err : new ApiError(500, "UNKNOWN_ERROR", tCommon("tryAgain")))
        setDebates([])
        setUsers([])
        setTags([])
        setResultTotal(null)
      } finally {
        if (!active) return
        setLoading(false)
      }
    }

    void run()

    return () => {
      active = false
    }
  }, [debateSort, debouncedQuery, searchType, status])

  const handleToggleUpvote = useCallback(async (debateID: string) => {
    const res = await toggleDebateVote(debateID)
    setDebates((prev) =>
      prev.map((debate) =>
        debate.id === res.data.debate_id
          ? { ...debate, upvote_count: res.data.upvote_count, viewer_has_upvoted: res.data.voted }
          : debate
      )
    )
  }, [])

  const typeTabs: { key: SearchType; label: string }[] = useMemo(
    () => [
      { key: "all", label: t("tabs.all") },
      { key: "debates", label: t("tabs.debates") },
      { key: "users", label: t("tabs.users") },
      { key: "tags", label: t("tabs.tags") },
    ],
    [t]
  )

  return (
    <div className="min-h-screen bg-[var(--bg-primary)]">
      <div className="sticky top-0 z-20 bg-[var(--bg-primary)]/95 backdrop-blur-sm border-b border-[var(--border-subtle)]">
        <div className="max-w-[820px] mx-auto px-3 sm:px-6">
          <div className="pt-3 pb-2">
            <div className="relative">
              <MagnifyingGlass
                size={16}
                className="absolute left-3 top-1/2 -translate-y-1/2 text-[var(--text-muted)]"
                aria-hidden
              />
              <input
                value={queryInput}
                onChange={(event) => setQueryInput(event.target.value)}
                placeholder={t("placeholder")}
                className="w-full h-10 rounded-lg border border-[var(--border-default)] bg-[var(--bg-surface)] pl-9 pr-3 text-[14px] text-[var(--text-primary)] outline-none focus:border-[var(--border-strong)] font-sans"
              />
            </div>
          </div>

          <div className="flex items-center gap-0 overflow-x-auto">
            {typeTabs.map((tab) => (
              <button
                key={tab.key}
                onClick={() => setSearchType(tab.key)}
                className={`px-3 sm:px-4 py-3 text-[13px] font-medium font-sans whitespace-nowrap transition-colors relative ${
                  searchType === tab.key
                    ? "text-[var(--text-primary)]"
                    : "text-[var(--text-muted)] hover:text-[var(--text-secondary)]"
                }`}
              >
                {tab.label}
                {searchType === tab.key && (
                  <div className="absolute bottom-0 left-2 right-2 h-[2px] bg-[var(--text-primary)] rounded-full" />
                )}
              </button>
            ))}
            {(searchType === "all" || searchType === "debates") && debouncedQuery && (
              <div className="ml-auto flex items-center gap-2 pl-3">
                <button
                  onClick={() => setDebateSort("relevance")}
                  className={`px-2.5 py-1 text-[11px] font-medium rounded-full border transition-colors font-sans ${
                    debateSort === "relevance"
                      ? "bg-[var(--text-primary)] text-[var(--bg-primary)] border-[var(--text-primary)]"
                      : "bg-[var(--bg-surface)] text-[var(--text-secondary)] border-[var(--border-default)] hover:border-[var(--border-strong)]"
                  }`}
                >
                  {t("sort.relevance")}
                </button>
                <button
                  onClick={() => setDebateSort("new")}
                  className={`px-2.5 py-1 text-[11px] font-medium rounded-full border transition-colors font-sans ${
                    debateSort === "new"
                      ? "bg-[var(--text-primary)] text-[var(--bg-primary)] border-[var(--text-primary)]"
                      : "bg-[var(--bg-surface)] text-[var(--text-secondary)] border-[var(--border-default)] hover:border-[var(--border-strong)]"
                  }`}
                >
                  {t("sort.new")}
                </button>
              </div>
            )}
          </div>
        </div>
      </div>

      <div className="max-w-[820px] mx-auto px-3 sm:px-6 py-4">
        {!debouncedQuery && (
          <div className="py-20 text-center">
            <div className="text-[var(--text-secondary)] text-base font-sans mb-1">{t("introTitle")}</div>
            <div className="text-[var(--text-muted)] text-sm font-sans">{t("introHint")}</div>
          </div>
        )}

        {debouncedQuery && loading && (
          <div className="py-16 text-center text-[var(--text-secondary)] text-base font-sans">{t("searching")}</div>
        )}

        {debouncedQuery && !loading && error && (
          <div className="py-20 text-center">
            <div className="text-[var(--text-secondary)] text-base font-sans mb-1">{t("loadError")}</div>
            <div className="text-[var(--text-muted)] text-sm font-sans">{error.message}</div>
          </div>
        )}

        {debouncedQuery && !loading && !error && searchType === "debates" && debates.length > 0 && (
          <div className="flex flex-col gap-2">
            {debates.map((debate) => (
              <DebateCard key={debate.id} debate={debate} onToggleUpvote={handleToggleUpvote} />
            ))}
          </div>
        )}

        {debouncedQuery && !loading && !error && searchType === "all" && (
          <div className="flex flex-col gap-6">
            {users.length > 0 && (
              <SearchSection
                title={t("section.users")}
                moreHref={`/search?${new URLSearchParams({ q: debouncedQuery, type: "users" }).toString()}`}
                onMoreClick={() => setSearchType("users")}
              >
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
                  {users.map((user) => (
                    <Link
                      key={user.username}
                      href={`/profile/${encodeURIComponent(user.username)}`}
                      className="rounded-lg border border-[var(--border-default)] bg-[var(--bg-surface)] px-4 py-3 hover:border-[var(--border-strong)] transition-colors"
                    >
                      <div className="text-[14px] font-semibold text-[var(--text-primary)] font-sans">@{user.username}</div>
                      <div className="mt-1 text-[12px] text-[var(--text-secondary)] font-sans">{t("rating", { rating: user.rating })}</div>
                      {user.bio && <div className="mt-1 text-[12px] text-[var(--text-muted)] font-sans line-clamp-2">{user.bio}</div>}
                    </Link>
                  ))}
                </div>
              </SearchSection>
            )}
            {tags.length > 0 && (
              <SearchSection
                title={t("section.tags")}
                moreHref={`/search?${new URLSearchParams({ q: debouncedQuery, type: "tags" }).toString()}`}
                onMoreClick={() => setSearchType("tags")}
              >
                <div className="grid gap-2 sm:grid-cols-2">
                  {tags.map((tag) => (
                    <Link
                      key={tag.id}
                      href={`/tags/${encodeURIComponent(tag.slug)}`}
                      className="rounded-lg border border-[var(--border-default)] bg-[var(--bg-surface)] px-4 py-3 hover:border-[var(--border-strong)] transition-colors"
                    >
                      <div className="text-[14px] font-semibold text-[var(--text-primary)] font-sans">{tag.name}</div>
                      <div className="mt-1 text-[12px] text-[var(--text-muted)] font-sans">/{tag.slug}</div>
                    </Link>
                  ))}
                </div>
              </SearchSection>
            )}
            {debates.length > 0 && (
              <SearchSection title={t("section.debates")}>
                <div className="flex flex-col gap-2">
                  {debates.map((debate) => (
                    <DebateCard key={debate.id} debate={debate} onToggleUpvote={handleToggleUpvote} />
                  ))}
                </div>
              </SearchSection>
            )}
          </div>
        )}

        {debouncedQuery && !loading && !error && searchType === "users" && users.length > 0 && (
          <div className="flex flex-col gap-2">
            {users.map((user) => (
              <Link
                key={user.username}
                href={`/profile/${encodeURIComponent(user.username)}`}
                className="rounded-lg border border-[var(--border-default)] bg-[var(--bg-surface)] px-4 py-3 hover:border-[var(--border-strong)] transition-colors"
              >
                <div className="text-[14px] font-semibold text-[var(--text-primary)] font-sans">@{user.username}</div>
                <div className="mt-1 text-[12px] text-[var(--text-secondary)] font-sans">{t("rating", { rating: user.rating })}</div>
                {user.bio && <div className="mt-1 text-[12px] text-[var(--text-muted)] font-sans line-clamp-2">{user.bio}</div>}
              </Link>
            ))}
          </div>
        )}

        {debouncedQuery && !loading && !error && searchType === "tags" && tags.length > 0 && (
          <div className="grid gap-2 sm:grid-cols-2">
            {tags.map((tag) => (
              <Link
                key={tag.id}
                href={`/tags/${encodeURIComponent(tag.slug)}`}
                className="rounded-lg border border-[var(--border-default)] bg-[var(--bg-surface)] px-4 py-3 hover:border-[var(--border-strong)] transition-colors"
              >
                <div className="text-[14px] font-semibold text-[var(--text-primary)] font-sans">{tag.name}</div>
                <div className="mt-1 text-[12px] text-[var(--text-muted)] font-sans">/{tag.slug}</div>
              </Link>
            ))}
          </div>
        )}

        {debouncedQuery && !loading && !error && searchType === "debates" && debates.length === 0 && (
          <EmptyState typeLabel={t("typeLabel.debates")} />
        )}
        {debouncedQuery && !loading && !error && searchType === "users" && users.length === 0 && (
          <EmptyState typeLabel={t("typeLabel.users")} />
        )}
        {debouncedQuery && !loading && !error && searchType === "tags" && tags.length === 0 && (
          <EmptyState typeLabel={t("typeLabel.tags")} />
        )}
        {debouncedQuery && !loading && !error && searchType === "all" && debates.length === 0 && users.length === 0 && tags.length === 0 && (
          <EmptyState typeLabel={t("typeLabel.results")} />
        )}

        {debouncedQuery && !loading && !error && resultTotal !== null && (
          <div className="py-8 text-center text-[var(--text-muted)] text-sm font-sans">
            {t("results", { count: resultTotal })}
          </div>
        )}
      </div>
    </div>
  )
}

function SearchSection({
  title,
  moreHref,
  onMoreClick,
  children,
}: {
  title: string
  moreHref?: string
  onMoreClick?: () => void
  children: React.ReactNode
}) {
  return (
    <section>
      <div className="mb-2 flex items-center justify-between gap-3">
        <h3 className="text-[12px] font-semibold uppercase tracking-[0.14em] text-[var(--text-muted)] font-sans">{title}</h3>
        {moreHref && (
          <Link
            href={moreHref}
            onClick={onMoreClick}
            className="text-[12px] font-semibold text-[var(--text-primary)] font-sans hover:underline"
          >
            {useTranslations("search")("section.more")}
          </Link>
        )}
      </div>
      {children}
    </section>
  )
}

function EmptyState({ typeLabel }: { typeLabel: string }) {
  return (
    <div className="py-20 text-center">
      <div className="text-[var(--text-secondary)] text-base font-sans mb-1">{useTranslations("search")("empty", { typeLabel })}</div>
      <div className="text-[var(--text-muted)] text-sm font-sans">{useTranslations("search")("emptyHint")}</div>
    </div>
  )
}
