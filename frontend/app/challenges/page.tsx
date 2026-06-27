"use client"

import { useCallback, useEffect, useMemo, useState, Suspense } from "react"
import { FunnelSimple, Megaphone } from "@phosphor-icons/react"
import Link from "next/link"
import { useRouter, useSearchParams } from "next/navigation"
import { useTranslations } from "next-intl"
import { ApiError } from "@/lib/api"
import { emitChallengesRefresh } from "@/lib/challenges-events"
import { fetchMyChallenges, respondChallenge, fetchTags } from "@/lib/debates-api"
import type { ChallengeListItem, Tag } from "@/lib/debates"
import { useAuth } from "@/hooks/use-auth"
import { Button } from "@/components/ui/button"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { formatTimeAgo } from "@/lib/utils"
import LobbyTab from "./lobby-tab"
import TagFilterDialog from "@/components/tag-filter-dialog"

type ChallengeBox = "inbox" | "outbox"
type ChallengeStatus = "pending" | "accepted" | "declined" | "expired" | "all"
type DashboardView = "lobby" | "my-challenges"

const perPage = 20

function ChallengesPageInner() {
  const router = useRouter()
  const searchParams = useSearchParams()
  const { status, user } = useAuth()
  const t = useTranslations("challenges")
  const initialTab = searchParams.get("tab")

  const [view, setView] = useState<DashboardView>(
    initialTab === "inbox" || initialTab === "outbox"
      ? "my-challenges"
      : "lobby"
  )
  const [box, setBox] = useState<ChallengeBox>(initialTab === "outbox" ? "outbox" : "inbox")
  const [filter, setFilter] = useState<ChallengeStatus>(initialTab === "outbox" ? "all" : "pending")
  const [items, setItems] = useState<ChallengeListItem[]>([])
  const [page, setPage] = useState(1)
  const [totalPages, setTotalPages] = useState(1)
  const [loading, setLoading] = useState(true)
  const [loadingMore, setLoadingMore] = useState(false)
  const [actioningID, setActioningID] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [showStandingChallenge, setShowStandingChallenge] = useState(false)
  const [tagFilterOpen, setTagFilterOpen] = useState(false)
  const [filterTagSlugs, setFilterTagSlugs] = useState<string[]>([])
  const [allTags, setAllTags] = useState<Tag[]>([])

  const canRespond = useMemo(() => box === "inbox" && filter === "pending", [box, filter])

  const load = useCallback(async (nextPage = 1, append = false) => {
    if (status !== "authenticated") return
    if (append) {
      setLoadingMore(true)
    } else {
      setLoading(true)
    }
    setError(null)
    try {
      const res = await fetchMyChallenges({
        box,
        status: filter,
        page: nextPage,
        perPage,
      })
      setItems((prev) => (append ? [...prev, ...(res.data ?? [])] : (res.data ?? [])))
      setPage(res.meta?.page ?? nextPage)
      setTotalPages(res.meta?.total_pages ?? 1)
    } catch (err) {
      setError(err instanceof Error ? err.message : t("loadError"))
    } finally {
      setLoading(false)
      setLoadingMore(false)
    }
  }, [status, box, filter])

  useEffect(() => {
    let active = true
    fetchTags()
      .then((res) => {
        if (!active) return
        setAllTags(res.data ?? [])
      })
      .catch(() => {
        if (!active) return
        setAllTags([])
      })
    return () => {
      active = false
    }
  }, [])

  useEffect(() => {
    if (status === "unauthenticated") {
      router.replace("/auth/login?redirect=/challenges")
      return
    }
    if (status === "authenticated" && view === "my-challenges") {
      void load(1, false)
    }
  }, [status, load, router, view])

  async function handleRespond(item: ChallengeListItem, accept: boolean) {
    setActioningID(item.debate_id)
    setError(null)
    try {
      const res = await respondChallenge(item.debate_id, accept)
      emitChallengesRefresh()
      if (accept && res.data.accepted) {
        router.push(`/debate/${res.data.debate_id}`)
        return
      }
      await load(1, false)
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message)
      } else {
        setError(t("respondError"))
      }
    } finally {
      setActioningID(null)
    }
  }

  if (status !== "authenticated") {
    return <div className="min-h-screen bg-[var(--bg-primary)]" />
  }

  const selectedFilterTagNames = filterTagSlugs.map((slug) => allTags.find((tag) => tag.slug === slug)?.name ?? slug)

  return (
    <div className="min-h-screen bg-[var(--bg-primary)]">
      <div className="sticky top-0 z-20 bg-[var(--bg-primary)]/95 backdrop-blur-sm border-b border-[var(--border-subtle)]">
        <div className="max-w-[820px] mx-auto px-3 sm:px-6">
          <div className="flex items-center gap-1 overflow-x-auto">
            <div className="flex items-center gap-1 overflow-x-auto">
              <button
                type="button"
                onClick={() => setView("lobby")}
                className={`px-3 py-3 text-[13px] font-medium whitespace-nowrap border-b-2 ${view === "lobby"
                  ? "text-[var(--text-primary)] border-[var(--text-primary)]"
                  : "text-[var(--text-muted)] border-transparent hover:text-[var(--text-secondary)]"
                  }`}
              >
                {t("tabs.open")}
              </button>
              <button
                type="button"
                onClick={() => setView("my-challenges")}
                className={`px-3 py-3 text-[13px] font-medium whitespace-nowrap border-b-2 ${view === "my-challenges"
                  ? "text-[var(--text-primary)] border-[var(--text-primary)]"
                  : "text-[var(--text-muted)] border-transparent hover:text-[var(--text-secondary)]"
                  }`}
              >
                {t("tabs.mine")}
              </button>
            </div>

            {view === "lobby" && (
              <div className="ml-auto flex items-center gap-2 pl-2">
                <button
                  type="button"
                  onClick={() => setTagFilterOpen(true)}
                  className={`inline-flex h-8 items-center gap-1.5 rounded-full border px-3 text-[11px] font-medium font-sans whitespace-nowrap ${filterTagSlugs.length > 0
                    ? "border-[var(--text-primary)] bg-[var(--text-primary)] text-[var(--bg-primary)]"
                    : "border-[var(--border-default)] bg-[var(--bg-surface)] text-[var(--text-secondary)] hover:border-[var(--border-strong)]"
                    }`}
                >
                  <FunnelSimple size={12} />
                  {t("filter")}{filterTagSlugs.length > 0 ? ` (${filterTagSlugs.length})` : ""}
                </button>
                <button
                  type="button"
                  onClick={() => setShowStandingChallenge((current) => !current)}
                  className={`inline-flex h-8 items-center gap-1.5 rounded-full border px-3 text-[11px] font-medium font-sans whitespace-nowrap ${showStandingChallenge
                    ? "border-[var(--text-primary)] bg-[var(--bg-surface)] text-[var(--text-primary)]"
                    : "border-[var(--border-default)] bg-[var(--bg-surface)] text-[var(--text-secondary)] hover:border-[var(--border-strong)]"
                    }`}
                >
                  <Megaphone size={12} />
                  {t("standing")}
                </button>
              </div>
            )}

            {view === "my-challenges" && (
              <div className="ml-auto flex items-center gap-2 pl-2">
                <button
                  type="button"
                  onClick={() => {
                    setBox("inbox")
                    setFilter("pending")
                  }}
                  className={`inline-flex h-8 items-center rounded-full border px-3 text-[11px] font-medium font-sans whitespace-nowrap ${box === "inbox"
                    ? "border-[var(--text-primary)] text-[var(--text-primary)]"
                    : "border-[var(--border-default)] text-[var(--text-secondary)] hover:border-[var(--border-strong)]"
                    }`}
                >
                  {t("inbox")}
                </button>
                <button
                  type="button"
                  onClick={() => {
                    setBox("outbox")
                    setFilter("all")
                  }}
                  className={`inline-flex h-8 items-center rounded-full border px-3 text-[11px] font-medium font-sans whitespace-nowrap ${box === "outbox"
                    ? "border-[var(--text-primary)] text-[var(--text-primary)]"
                    : "border-[var(--border-default)] text-[var(--text-secondary)] hover:border-[var(--border-strong)]"
                    }`}
                >
                  {t("outbox")}
                </button>
                <Select value={filter} onValueChange={(value) => setFilter(value as ChallengeStatus)}>
                  <SelectTrigger
                    size="sm"
                    className="h-8 min-w-[110px] rounded-full border-[var(--border-default)] bg-[var(--bg-surface)] px-3 text-[11px] font-medium font-sans text-[var(--text-secondary)]"
                  >
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="pending">{t("statuses.pending")}</SelectItem>
                    <SelectItem value="accepted">{t("statuses.accepted")}</SelectItem>
                    <SelectItem value="declined">{t("statuses.declined")}</SelectItem>
                    <SelectItem value="expired">{t("statuses.expired")}</SelectItem>
                    <SelectItem value="all">{t("statuses.all")}</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            )}
          </div>
        </div>
      </div>

      <div className="max-w-[820px] mx-auto px-3 sm:px-6 py-4">
        {view === "lobby" && selectedFilterTagNames.length > 0 && (
          <div className="mb-4 flex flex-wrap items-center gap-1.5">
            {selectedFilterTagNames.map((name, index) => (
              <button
                key={`${name}-${index}`}
                type="button"
                onClick={() => {
                  const next = [...filterTagSlugs]
                  next.splice(index, 1)
                  setFilterTagSlugs(next)
                }}
                className="px-2.5 py-1 text-[11px] font-medium rounded-full border border-[var(--text-primary)] bg-[var(--text-primary)] text-[var(--bg-primary)] whitespace-nowrap font-sans"
              >
                {name} x
              </button>
            ))}
            <button
              type="button"
              onClick={() => setFilterTagSlugs([])}
              className="px-2 py-1 text-[11px] text-[var(--text-secondary)] hover:text-[var(--text-primary)] font-sans"
            >
              {t("clearAll")}
            </button>
          </div>
        )}

        {view === "lobby" && (
          <LobbyTab
            mode="all"
            showSectionHeaders={false}
            showMyEntrySection={showStandingChallenge}
            filterTagSlugs={filterTagSlugs}
            defaultMyEntryOpen={showStandingChallenge}
            showMyEntryToggle={false}
          />
        )}

        {view === "my-challenges" && (
          <>
            {error && (
              <div className="mb-4 rounded-md border border-[var(--error)] bg-[var(--error-light)] px-3 py-2 text-[12px] text-[var(--error)]">
                {error}
              </div>
            )}

            {loading ? (
              <div className="text-[13px] text-[var(--text-muted)] font-sans">{t("loading")}</div>
            ) : items.length === 0 ? (
              <div className="bg-[var(--bg-surface)] border border-[var(--border-default)] rounded-lg p-4 text-[13px] text-[var(--text-secondary)] font-sans">
                {t("empty")}
              </div>
            ) : (
              <div className="space-y-2">
                {items.map((item) => {
                  const isPending = item.status === "waiting"
                  return (
                    <div key={item.debate_id} className="bg-[var(--bg-surface)] border border-[var(--border-default)] rounded-lg p-4">
                      <div className="flex items-start justify-between gap-3">
                        <div className="min-w-0">
                          <Link href={`/debate/${item.debate_slug || item.debate_id}`} className="text-[14px] font-semibold text-[var(--text-primary)] hover:underline">
                            {item.topic}
                          </Link>
                          <div className="mt-1 text-[12px] text-[var(--text-secondary)] font-sans">
                            {box === "inbox"
                              ? t("from", { username: item.challenger_username })
                              : t("to", { username: item.invited_username })}
                            {` • ${item.time_mode} • ${t("turns", { count: item.turn_limit })}`}
                          </div>
                          <div className="mt-1 text-[11px] text-[var(--text-muted)] font-sans">
                            {t("sent", { time: formatTimeAgo(item.created_at) })}
                            {item.challenge_expires_at ? ` • ${t("expires", { time: formatTimeAgo(item.challenge_expires_at) })}` : ""}
                          </div>
                        </div>
                        <span className="text-[11px] px-2 py-1 rounded-full border border-[var(--border-default)] text-[var(--text-secondary)]">
                          {isPending ? t("statuses.waiting") : item.status}
                        </span>
                      </div>

                      {canRespond && isPending && (
                        <div className="mt-3 flex items-center gap-2">
                          <Button
                            size="sm"
                            onClick={() => void handleRespond(item, true)}
                            disabled={actioningID === item.debate_id || !user?.email_verified}
                          >
                            {actioningID === item.debate_id ? t("working") : t("accept")}
                          </Button>
                          <Button
                            size="sm"
                            variant="outline"
                            onClick={() => void handleRespond(item, false)}
                            disabled={actioningID === item.debate_id}
                          >
                            {t("decline")}
                          </Button>
                        </div>
                      )}
                    </div>
                  )
                })}

                {page < totalPages && (
                  <div className="pt-2 flex justify-center">
                    <button
                      type="button"
                      onClick={() => void load(page + 1, true)}
                      disabled={loadingMore}
                      className="px-4 py-2 rounded-md text-[12px] font-medium font-sans text-[var(--text-secondary)] border border-[var(--border-default)] hover:bg-[var(--bg-surface-alt)] hover:text-[var(--text-primary)] disabled:opacity-60"
                    >
                      {loadingMore ? t("loading") : t("loadMore")}
                    </button>
                  </div>
                )}
              </div>
            )}
          </>
        )}
      </div>

      <TagFilterDialog
        open={tagFilterOpen}
        onOpenChange={setTagFilterOpen}
        tags={allTags}
        selectedSlugs={filterTagSlugs}
        onApply={setFilterTagSlugs}
        title={t("filterTitle")}
      />
    </div>
  )
}

export default function ChallengesPage() {
  return (
    <Suspense fallback={<div className="min-h-screen bg-[var(--bg-primary)]" />}>
      <ChallengesPageInner />
    </Suspense>
  )
}
