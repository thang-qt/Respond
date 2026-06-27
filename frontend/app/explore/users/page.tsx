"use client"

import { CheckCircle, FunnelSimple, MagnifyingGlass } from "@phosphor-icons/react"
import Link from "next/link"
import { useCallback, useEffect, useMemo, useState } from "react"
import { useRouter, useSearchParams } from "next/navigation"
import { useTranslations } from "next-intl"
import { ApiError } from "@/lib/api"
import type { Tag } from "@/lib/debates"
import type { UserSearchProfile } from "@/lib/users"
import { fetchTags } from "@/lib/debates-api"
import { fetchExploreUsers } from "@/lib/users-api"
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { useAuth } from "@/hooks/use-auth"

function parseTagSlugs(searchParams: URLSearchParams): string[] {
  const rawTags = (searchParams.get("tags") || "").trim()
  const fromTags = rawTags
    ? rawTags
      .split(",")
      .map((tag) => tag.trim().toLowerCase())
      .filter(Boolean)
    : []

  if (fromTags.length > 0) {
    return Array.from(new Set(fromTags))
  }

  const single =
    (searchParams.get("tag") || searchParams.get("category") || searchParams.get("category_id") || "")
      .trim()
      .toLowerCase()
  return single ? [single] : []
}

export default function ExploreUsersPage() {
  const router = useRouter()
  const searchParams = useSearchParams()
  const { status } = useAuth()
  const t = useTranslations("exploreUsers")
  const tCommon = useTranslations("common")
  const tTags = useTranslations("tags")
  const tTagFilter = useTranslations("tagFilter")
  const selectedTagSlugs = useMemo(() => parseTagSlugs(searchParams), [searchParams])

  const [users, setUsers] = useState<UserSearchProfile[]>([])
  const [usersLoading, setUsersLoading] = useState(true)
  const [usersError, setUsersError] = useState<ApiError | null>(null)
  const [usersMeta, setUsersMeta] = useState<{ total: number } | null>(null)
  const [tags, setTags] = useState<Tag[]>([])
  const [tagFilterOpen, setTagFilterOpen] = useState(false)
  const [tagFilterQuery, setTagFilterQuery] = useState("")
  const [draftTagSlugs, setDraftTagSlugs] = useState<string[]>([])

  const navigateTags = useCallback(
    (tagSlugs: string[]) => {
      const query = new URLSearchParams(searchParams.toString())
      const deduped = Array.from(new Set(tagSlugs.map((tag) => tag.trim().toLowerCase()).filter(Boolean)))

      query.delete("tag")
      query.delete("category")
      query.delete("category_id")
      if (deduped.length > 0) {
        query.set("tags", deduped.join(","))
        query.set("tag_mode", "any")
      } else {
        query.delete("tags")
        query.delete("tag_mode")
      }

      const suffix = query.toString()
      router.replace(`/explore/users${suffix ? `?${suffix}` : ""}`, { scroll: false })
    },
    [router, searchParams]
  )

  useEffect(() => {
    if (!tagFilterOpen) return
    setDraftTagSlugs(selectedTagSlugs)
    setTagFilterQuery("")
  }, [selectedTagSlugs, tagFilterOpen])

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
    if (status === "loading") return

    let active = true
    setUsersLoading(true)
    setUsersError(null)

    fetchExploreUsers({
      tagSlugs: selectedTagSlugs,
      tagMode: "any",
      context: "explore",
      page: 1,
      perPage: 20,
    })
      .then((res) => {
        if (!active) return
        setUsers(Array.isArray(res.data) ? res.data : [])
        setUsersMeta(res.meta ? { total: res.meta.total } : null)
      })
      .catch((err: ApiError) => {
        if (!active) return
        setUsersError(err)
        setUsers([])
        setUsersMeta(null)
      })
      .finally(() => {
        if (!active) return
        setUsersLoading(false)
      })

    return () => {
      active = false
    }
  }, [selectedTagSlugs, status])

  useEffect(() => {
    if (selectedTagSlugs.length === 0) return
    if (tags.length === 0) return
    const valid = new Set(tags.map((tag) => tag.slug))
    const next = selectedTagSlugs.filter((slug) => valid.has(slug))
    if (next.length !== selectedTagSlugs.length) {
      navigateTags(next)
    }
  }, [navigateTags, selectedTagSlugs, tags])

  const selectedTagNames = useMemo(() => {
    if (selectedTagSlugs.length === 0) return [] as string[]
    const bySlug = new Map(tags.map((tag) => [tag.slug, tag.name]))
    return selectedTagSlugs.map((slug) => bySlug.get(slug) ?? slug)
  }, [selectedTagSlugs, tags])

  const filteredTags = useMemo(() => {
    const q = tagFilterQuery.trim().toLowerCase()
    const base = [...tags].sort((a, b) => a.name.localeCompare(b.name))
    if (!q) return base
    return base.filter((tag) => tag.name.toLowerCase().includes(q) || tag.slug.toLowerCase().includes(q))
  }, [tagFilterQuery, tags])

  return (
    <div className="min-h-screen bg-[var(--bg-primary)]">
      <div className="sticky top-0 z-20 bg-[var(--bg-primary)]/95 backdrop-blur-sm border-b border-[var(--border-subtle)]">
        <div className="max-w-[820px] mx-auto px-3 sm:px-6">
          <div className="flex items-center py-2.5">
            <button
              type="button"
              onClick={() => setTagFilterOpen(true)}
              className={`ml-auto inline-flex h-8 items-center gap-1.5 rounded-full border px-3 text-[11px] font-medium font-sans ${
                selectedTagSlugs.length > 0
                  ? "border-[var(--text-primary)] bg-[var(--text-primary)] text-[var(--bg-primary)]"
                  : "border-[var(--border-default)] bg-[var(--bg-surface)] text-[var(--text-secondary)] hover:border-[var(--border-strong)]"
              }`}
            >
              <FunnelSimple size={12} />
              {tCommon("filter")}{selectedTagSlugs.length > 0 ? ` (${selectedTagSlugs.length})` : ""}
            </button>
          </div>
        </div>
      </div>

      <div className="max-w-[820px] mx-auto px-3 sm:px-6 py-4">
        {selectedTagNames.length > 0 && (
          <div className="mb-4 flex flex-wrap items-center gap-1.5">
            {selectedTagNames.map((name, index) => (
              <button
                key={`${name}-${index}`}
                type="button"
                onClick={() => {
                  const next = [...selectedTagSlugs]
                  next.splice(index, 1)
                  navigateTags(next)
                }}
                className="px-2.5 py-1 text-[11px] font-medium rounded-full border border-[var(--text-primary)] bg-[var(--text-primary)] text-[var(--bg-primary)] whitespace-nowrap font-sans"
              >
                {name} x
              </button>
            ))}
            <button
              type="button"
              onClick={() => navigateTags([])}
              className="px-2 py-1 text-[11px] text-[var(--text-secondary)] hover:text-[var(--text-primary)] font-sans"
            >
              {tCommon("clearAll")}
            </button>
          </div>
        )}

        <div className="flex flex-col gap-2">
          {users.length > 0 ? (
            users.map((candidate) => (
              <div
                key={candidate.id || candidate.username}
                className="rounded-lg border border-[var(--border-default)] bg-[var(--bg-surface)] px-4 py-3"
              >
                <div className="flex items-start justify-between gap-3">
                  <div className="min-w-0">
                    <Link
                      href={`/profile/${encodeURIComponent(candidate.username)}`}
                      className="text-[14px] font-semibold text-[var(--text-primary)] hover:underline underline-offset-2"
                    >
                      @{candidate.username}
                    </Link>
                    <div className="mt-0.5 text-[12px] text-[var(--text-secondary)] font-sans">
                      {t("stats", {
                        rating: candidate.rating,
                        wins: candidate.wins,
                        losses: candidate.losses,
                        draws: candidate.draws,
                        debates: candidate.debates_count,
                      })}
                    </div>
                    {candidate.bio ? (
                      <div className="mt-1 text-[12px] text-[var(--text-muted)] font-sans line-clamp-2">{candidate.bio}</div>
                    ) : null}
                  </div>
                  <Link
                    href={`/profile/${encodeURIComponent(candidate.username)}`}
                    className="inline-flex h-8 shrink-0 items-center justify-center rounded-md border border-[var(--border-default)] bg-[var(--bg-primary)] px-3 text-[11px] font-medium text-[var(--text-primary)] hover:border-[var(--border-strong)]"
                  >
                    {tCommon("viewProfile")}
                  </Link>
                </div>
                {candidate.shared_tags && candidate.shared_tags.length > 0 && (
                  <div className="mt-2 flex flex-wrap gap-1.5">
                    {candidate.shared_tags.slice(0, 3).map((slug) => {
                      const tagName = tags.find((tag) => tag.slug === slug)?.name ?? slug
                      return (
                        <span
                          key={`${candidate.username}-${slug}`}
                          className="rounded-sm bg-[var(--bg-surface-alt)] px-2 py-0.5 text-[10px] tracking-[0.04em] text-[var(--text-secondary)]"
                        >
                          {tagName}
                        </span>
                      )
                    })}
                  </div>
                )}
              </div>
            ))
          ) : (status === "loading" || usersLoading) ? (
            <div className="py-16 text-center text-[var(--text-secondary)] text-base font-sans">{t("loading")}</div>
          ) : usersError ? (
            <div className="py-20 text-center">
              <div className="text-[var(--text-secondary)] text-base font-sans mb-1">{t("loadError")}</div>
              <div className="text-[var(--text-muted)] text-sm font-sans">{usersError.message}</div>
            </div>
          ) : (
            <div className="py-20 text-center">
              <div className="text-[var(--text-secondary)] text-base font-sans mb-1">{t("empty")}</div>
              <div className="text-[var(--text-muted)] text-sm font-sans">{t("emptyHint")}</div>
            </div>
          )}
        </div>

        {users.length > 0 && (
          <div className="py-8 text-center text-[var(--text-muted)] text-sm font-sans">
            {usersMeta?.total
              ? t("showingOf", { count: users.length, total: usersMeta.total })
              : t("showing", { count: users.length })}
          </div>
        )}
      </div>

      <Dialog open={tagFilterOpen} onOpenChange={setTagFilterOpen}>
        <DialogContent className="sm:max-w-[640px] bg-[var(--bg-surface)] border border-[var(--border-default)]">
          <DialogHeader>
            <DialogTitle className="text-[var(--text-primary)] font-sans">{tTagFilter("title")}</DialogTitle>
          </DialogHeader>

          <div className="space-y-3">
            <div className="relative">
              <MagnifyingGlass
                size={16}
                className="absolute left-3 top-1/2 -translate-y-1/2 text-[var(--text-muted)]"
                aria-hidden
              />
              <input
                value={tagFilterQuery}
                onChange={(event) => setTagFilterQuery(event.target.value)}
                placeholder={tTags("searchPlaceholder")}
                className="w-full h-10 rounded-lg border border-[var(--border-default)] bg-[var(--bg-primary)] pl-9 pr-3 text-[14px] text-[var(--text-primary)] outline-none focus:border-[var(--border-strong)] font-sans"
              />
            </div>

            <div className="max-h-[360px] overflow-y-auto rounded-md border border-[var(--border-subtle)] bg-[var(--bg-primary)]">
              {filteredTags.length > 0 ? (
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-2 p-2">
                  {filteredTags.map((tag) => {
                    const selected = draftTagSlugs.includes(tag.slug)
                    return (
                      <button
                        key={tag.id}
                        type="button"
                        onClick={() => {
                          setDraftTagSlugs((prev) =>
                            prev.includes(tag.slug) ? prev.filter((slug) => slug !== tag.slug) : [...prev, tag.slug]
                          )
                        }}
                        className={`rounded-lg border px-3 py-2 text-left transition-colors ${
                          selected
                            ? "border-[var(--text-primary)] bg-[var(--bg-surface)]"
                            : "border-[var(--border-default)] bg-[var(--bg-surface)] hover:border-[var(--border-strong)]"
                        }`}
                      >
                        <div className="flex items-center gap-2">
                          <div className="min-w-0 flex-1">
                            <div className="text-[13px] font-semibold text-[var(--text-primary)] font-sans">{tag.name}</div>
                            <div className="text-[11px] text-[var(--text-muted)] font-sans">/{tag.slug}</div>
                          </div>
                          {selected && <CheckCircle size={14} weight="fill" className="text-[var(--text-primary)]" />}
                        </div>
                      </button>
                    )
                  })}
                </div>
              ) : (
                <div className="px-3 py-8 text-center text-[12px] text-[var(--text-muted)]">{tTags("noMatch")}</div>
              )}
            </div>

            <div className="flex items-center justify-between gap-2">
              <button
                type="button"
                onClick={() => setDraftTagSlugs([])}
                className="h-9 rounded-md border border-[var(--border-default)] px-3 text-[12px] font-medium text-[var(--text-secondary)] hover:border-[var(--border-strong)]"
              >
                {tCommon("clearSelection")}
              </button>
              <button
                type="button"
                onClick={() => {
                  navigateTags(draftTagSlugs)
                  setTagFilterOpen(false)
                }}
                className="h-9 rounded-md bg-[var(--text-primary)] px-3 text-[12px] font-medium text-[var(--bg-primary)] hover:opacity-90"
              >
                {tCommon("applyFilters")}
              </button>
            </div>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  )
}
