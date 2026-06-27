"use client"

import { FunnelSimple } from "@phosphor-icons/react"
import { useCallback, useEffect, useMemo, useState } from "react"
import { useParams, useRouter, useSearchParams } from "next/navigation"
import { useTranslations } from "next-intl"
import { ApiError } from "@/lib/api"
import type { DebateFeedItem, Tag } from "@/lib/debates"
import { fetchExplore, fetchTags, toggleDebateVote } from "@/lib/debates-api"
import DebateCard from "@/components/debate-card"
import TagFilterDialog from "@/components/tag-filter-dialog"
import { useAuth } from "@/hooks/use-auth"

type ExploreSort = "hot" | "rising" | "new" | "random"

function normalizeSort(input: string | undefined): ExploreSort | null {
  if (input === "hot" || input === "rising" || input === "new" || input === "random") {
    return input
  }
  return null
}

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

export default function ExploreDebatesPage() {
  const params = useParams<{ sort?: string }>()
  const router = useRouter()
  const searchParams = useSearchParams()
  const { status } = useAuth()
  const t = useTranslations("explore")
  const tCommon = useTranslations("common")
  const tHome = useTranslations("home")

  const pathSort = normalizeSort(Array.isArray(params.sort) ? params.sort[0] : params.sort)
  const selectedTagSlugs = useMemo(() => parseTagSlugs(searchParams), [searchParams])
  const activeSort: ExploreSort = pathSort ?? "hot"

  const [debates, setDebates] = useState<DebateFeedItem[]>([])
  const [tags, setTags] = useState<Tag[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<ApiError | null>(null)
  const [meta, setMeta] = useState<{ total: number } | null>(null)
  const [tagFilterOpen, setTagFilterOpen] = useState(false)
  const [sortTotals, setSortTotals] = useState<Record<ExploreSort, number | null>>({
    hot: null,
    rising: null,
    new: null,
    random: null,
  })

  const navigateSort = useCallback(
    (sort: ExploreSort) => {
      const query = new URLSearchParams(searchParams.toString())
      const suffix = query.toString()
      router.replace(`/explore/debates/${sort}${suffix ? `?${suffix}` : ""}`, { scroll: false })
    },
    [router, searchParams]
  )

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
      router.replace(`/explore/debates/${activeSort}${suffix ? `?${suffix}` : ""}`, { scroll: false })
    },
    [activeSort, router, searchParams]
  )

  useEffect(() => {
    if (pathSort !== null) return
    const suffix = searchParams.toString()
    router.replace(`/explore/debates/hot${suffix ? `?${suffix}` : ""}`, { scroll: false })
  }, [pathSort, router, searchParams])

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
    setLoading(true)
    setError(null)

    fetchExplore({
      sort: activeSort,
      tagSlugs: selectedTagSlugs,
      tagMode: "any",
      page: 1,
      perPage: 20,
    })
      .then((res) => {
        if (!active) return
        setDebates(Array.isArray(res.data) ? res.data : [])
        setMeta(res.meta ? { total: res.meta.total } : null)
        if (res.meta?.total !== undefined) {
          setSortTotals((prev) => ({ ...prev, [activeSort]: res.meta!.total }))
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
  }, [activeSort, selectedTagSlugs, status])

  useEffect(() => {
    if (selectedTagSlugs.length === 0) return
    if (tags.length === 0) return
    const valid = new Set(tags.map((tag) => tag.slug))
    const next = selectedTagSlugs.filter((slug) => valid.has(slug))
    if (next.length !== selectedTagSlugs.length) {
      navigateTags(next)
    }
  }, [navigateTags, selectedTagSlugs, tags])

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

  const sortTabs: { key: ExploreSort; label: string; count: number }[] = useMemo(
    () => [
      { key: "hot", label: t("sort.hot"), count: sortTotals.hot ?? 0 },
      { key: "rising", label: t("sort.rising"), count: sortTotals.rising ?? 0 },
      { key: "new", label: t("sort.new"), count: sortTotals.new ?? 0 },
      { key: "random", label: t("sort.random"), count: sortTotals.random ?? 0 },
    ],
    [sortTotals, t]
  )

  const selectedTagNames = useMemo(() => {
    if (selectedTagSlugs.length === 0) return [] as string[]
    const bySlug = new Map(tags.map((tag) => [tag.slug, tag.name]))
    return selectedTagSlugs.map((slug) => bySlug.get(slug) ?? slug)
  }, [selectedTagSlugs, tags])

  return (
    <div className="min-h-screen bg-[var(--bg-primary)]">
      <div className="sticky top-0 z-20 bg-[var(--bg-primary)]/95 backdrop-blur-sm border-b border-[var(--border-subtle)]">
        <div className="max-w-[820px] mx-auto px-3 sm:px-6">
          <div className="flex items-center gap-0 overflow-x-auto">
            {sortTabs.map((tab) => (
              <button
                key={tab.key}
                onClick={() => navigateSort(tab.key)}
                className={`px-3 sm:px-4 py-3 text-[13px] font-medium font-sans whitespace-nowrap transition-colors relative ${
                  activeSort === tab.key
                    ? "text-[var(--text-primary)]"
                    : "text-[var(--text-muted)] hover:text-[var(--text-secondary)]"
                }`}
              >
                {tab.label}
                {activeSort === tab.key && (
                  <div className="absolute bottom-0 left-2 right-2 h-[2px] bg-[var(--text-primary)] rounded-full" />
                )}
              </button>
            ))}

            <button
              type="button"
              onClick={() => setTagFilterOpen(true)}
              className={`ml-auto mr-1 inline-flex h-8 items-center gap-1.5 rounded-full border px-3 text-[11px] font-medium font-sans ${
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
          {debates.length > 0 ? (
            debates.map((debate) => <DebateCard key={debate.id} debate={debate} onToggleUpvote={handleToggleUpvote} />)
          ) : loading ? (
            <div className="py-16 text-center text-[var(--text-secondary)] text-base font-sans">{tCommon("loadingDebates")}</div>
          ) : error ? (
            <div className="py-20 text-center">
              <div className="text-[var(--text-secondary)] text-base font-sans mb-1">{t("loadError")}</div>
              <div className="text-[var(--text-muted)] text-sm font-sans">{error.message}</div>
            </div>
          ) : (
            <div className="py-20 text-center">
              <div className="text-[var(--text-secondary)] text-base font-sans mb-1">{t("noMatches")}</div>
              <div className="text-[var(--text-muted)] text-sm font-sans">{t("tryDifferent")}</div>
            </div>
          )}
        </div>

        {debates.length > 0 && (
          <div className="py-8 text-center text-[var(--text-muted)] text-sm font-sans">
            {meta?.total
              ? tHome("empty.showingOf", { count: debates.length, total: meta.total })
              : tHome("empty.showing", { count: debates.length })}
          </div>
        )}
      </div>

      <TagFilterDialog
        open={tagFilterOpen}
        onOpenChange={setTagFilterOpen}
        tags={tags}
        selectedSlugs={selectedTagSlugs}
        onApply={navigateTags}
      />
    </div>
  )
}
