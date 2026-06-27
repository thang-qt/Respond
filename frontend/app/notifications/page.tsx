"use client"

import { useEffect } from "react"
import Link from "next/link"
import { useRouter } from "next/navigation"
import { Bell, Check, CheckCircle } from "@phosphor-icons/react"
import { useTranslations } from "next-intl"
import { ApiError } from "@/lib/api"
import { emitChallengesRefresh } from "@/lib/challenges-events"
import { joinDebate, respondChallenge } from "@/lib/debates-api"
import { useAuth } from "@/hooks/use-auth"
import { useNotifications } from "@/hooks/use-notifications"
import { formatTimeAgo } from "@/lib/utils"
import { useState } from "react"

export default function NotificationsPage() {
  const router = useRouter()
  const { status } = useAuth()
  const { notifications, unreadCount, loading, loadingMore, hasMore, error, refresh, loadMore, markAllRead, markRead } = useNotifications()
  const t = useTranslations("notificationsPage")
  const [actioningID, setActioningID] = useState<string | null>(null)
  const [actionError, setActionError] = useState<string | null>(null)

  useEffect(() => {
    if (status === "unauthenticated") {
      router.replace("/auth/login?redirect=/notifications")
    }
  }, [status, router])

  useEffect(() => {
    if (status === "authenticated") {
      void refresh()
    }
  }, [status, refresh])

  if (status === "loading") {
    return (
      <div className="min-h-screen bg-[var(--bg-primary)]">
        <div className="max-w-[820px] mx-auto px-3 sm:px-6 py-8">
          <div className="text-[var(--text-muted)] text-[13px] font-sans">{t("loading")}</div>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-[var(--bg-primary)]">
      {/* Header */}
      <div className="sticky top-0 z-20 bg-[var(--bg-primary)]/95 backdrop-blur-sm border-b border-[var(--border-subtle)]">
        <div className="max-w-[820px] mx-auto px-3 sm:px-6">
          <div className="flex items-center justify-between py-3">
            <div className="flex items-center gap-2">
              <Bell size={18} className="text-[var(--text-primary)]" />
              <h1 className="text-[15px] font-semibold text-[var(--text-primary)] font-sans">{t("title")}</h1>
              {unreadCount > 0 && (
                <span className="min-w-5 h-5 px-1.5 rounded-full bg-[var(--text-primary)] text-[11px] text-[var(--bg-primary)] font-semibold flex items-center justify-center">
                  {unreadCount}
                </span>
              )}
            </div>
            <button
              type="button"
              onClick={() => void markAllRead()}
              disabled={unreadCount < 1}
              className="flex items-center gap-1.5 text-[12px] font-medium text-[var(--text-secondary)] hover:text-[var(--text-primary)] disabled:opacity-50 transition-colors font-sans"
            >
              <CheckCircle size={14} />
              {t("markAllRead")}
            </button>
          </div>
        </div>
      </div>

      {/* Notification list */}
      <div className="max-w-[820px] mx-auto px-3 sm:px-6 py-4">
        {loading ? (
          <div className="py-16 text-center">
            <div className="text-[var(--text-muted)] text-[13px] font-sans">{t("loadingNotifications")}</div>
          </div>
        ) : error ? (
          <div className="py-16 text-center">
            <div className="text-[var(--text-secondary)] text-[13px] font-sans">{error}</div>
          </div>
        ) : notifications.length === 0 ? (
          <div className="py-16 text-center">
            <Bell size={32} className="mx-auto text-[var(--text-muted)] mb-3" />
            <div className="text-[var(--text-secondary)] text-[14px] font-sans font-medium mb-1">{t("emptyTitle")}</div>
            <div className="text-[var(--text-muted)] text-[13px] font-sans">
              {t("emptyBody")}
            </div>
          </div>
        ) : (
          <div className="flex flex-col gap-1">
            {actionError && (
              <div className="mb-2 rounded-md border border-[var(--error)] bg-[var(--error-light)] px-3 py-2 text-[12px] text-[var(--error)]">
                {actionError}
              </div>
            )}
            {notifications.map((item) => {
              const isUnread = !item.read
              const baseDebateHref = item.debate_slug
                ? `/debate/${item.debate_slug}`
                : item.debate_id
                ? `/debate/${item.debate_id}`
                : null
              const debateHref =
                baseDebateHref && typeof item.turn_number === "number"
                  ? `${baseDebateHref}?turn=${item.turn_number}`
                  : baseDebateHref

              const classes = `px-3 py-3 rounded-lg transition-colors ${
                isUnread
                  ? "bg-[var(--bg-surface)] border border-[var(--border-default)] shadow-[0px_1px_2px_rgba(55,50,47,0.06)]"
                  : "hover:bg-[var(--border-subtle)]"
              }`

              const textContent = (
                <>
                  <div
                    className={`text-[13px] leading-snug font-sans ${
                      isUnread ? "font-medium text-[var(--text-primary)]" : "text-[var(--text-secondary)]"
                    }`}
                  >
                    {item.message}
                  </div>
                  <div className="text-[11px] text-[var(--text-muted)] font-sans mt-1">{formatTimeAgo(item.created_at)}</div>
                </>
              )

              return (
                <div key={item.id} className={classes}>
                  <div className="flex gap-3 items-start">
                    <span
                      className={`mt-1.5 h-2 w-2 rounded-full shrink-0 ${
                        isUnread ? "bg-[var(--text-primary)]" : "bg-transparent"
                      }`}
                    />

                    <div className="flex-1 min-w-0">
                      {debateHref ? (
                        <Link href={debateHref} onClick={() => void markRead(item.id)} className="block">
                          {textContent}
                        </Link>
                      ) : (
                        <button
                          type="button"
                          onClick={() => {
                            if (isUnread) {
                              void markRead(item.id)
                            }
                          }}
                          className="w-full text-left"
                        >
                          {textContent}
                        </button>
                      )}

                      {item.type === "challenge_received" && item.debate_id && isUnread && (
                        <div className="mt-2 flex items-center gap-2">
                          <button
                            type="button"
                            onClick={async () => {
                              setActionError(null)
                              setActioningID(item.id)
                              try {
                                await respondChallenge(item.debate_id as string, true)
                                await markRead(item.id)
                                emitChallengesRefresh()
                                router.push(debateHref ?? `/debate/${item.debate_id}`)
                              } catch (err) {
                                setActionError(err instanceof ApiError ? err.message : t("acceptError"))
                                void refresh()
                              } finally {
                                setActioningID(null)
                              }
                            }}
                            disabled={actioningID === item.id}
                            className="h-7 px-2.5 rounded-md text-[11px] font-medium bg-[var(--text-primary)] text-[var(--bg-primary)] disabled:opacity-60"
                          >
                            {actioningID === item.id ? t("working") : t("accept")}
                          </button>
                          <button
                            type="button"
                            onClick={async () => {
                              setActionError(null)
                              setActioningID(item.id)
                              try {
                                await respondChallenge(item.debate_id as string, false)
                                await markRead(item.id)
                                emitChallengesRefresh()
                                await refresh()
                              } catch (err) {
                                setActionError(err instanceof ApiError ? err.message : t("declineError"))
                                void refresh()
                              } finally {
                                setActioningID(null)
                              }
                            }}
                            disabled={actioningID === item.id}
                            className="h-7 px-2.5 rounded-md text-[11px] font-medium border border-[var(--border-default)] text-[var(--text-secondary)] hover:bg-[var(--bg-surface-alt)] disabled:opacity-60"
                          >
                            {t("decline")}
                          </button>
                        </div>
                      )}

                      {item.type === "debate_invited" && item.debate_id && isUnread && (
                        <div className="mt-2 flex items-center gap-2">
                          <button
                            type="button"
                            onClick={async () => {
                              setActionError(null)
                              setActioningID(item.id)
                              try {
                                await joinDebate(item.debate_id as string)
                                await markRead(item.id)
                                router.push(debateHref ?? `/debate/${item.debate_id}`)
                              } catch (err) {
                                setActionError(err instanceof ApiError ? err.message : t("joinError"))
                                void refresh()
                              } finally {
                                setActioningID(null)
                              }
                            }}
                            disabled={actioningID === item.id}
                            className="h-7 px-2.5 rounded-md text-[11px] font-medium bg-[var(--text-primary)] text-[var(--bg-primary)] disabled:opacity-60"
                          >
                            {actioningID === item.id ? t("working") : t("join")}
                          </button>
                        </div>
                      )}
                    </div>

                    {isUnread && (
                      <button
                        type="button"
                        onClick={() => void markRead(item.id)}
                        className="mt-0.5 h-7 w-7 rounded-md flex items-center justify-center text-[var(--text-secondary)] hover:text-[var(--text-primary)] hover:bg-[var(--border-subtle)] shrink-0"
                        aria-label={t("markRead")}
                        title={t("markRead")}
                      >
                        <Check size={14} />
                      </button>
                    )}
                  </div>
                </div>
              )
            })}

            {hasMore && (
              <div className="pt-3 flex justify-center">
                <button
                  type="button"
                  onClick={() => void loadMore()}
                  disabled={loadingMore}
                  className="px-4 py-2 rounded-md text-[12px] font-medium font-sans text-[var(--text-secondary)] border border-[var(--border-default)] hover:bg-[var(--bg-surface-alt)] hover:text-[var(--text-primary)] disabled:opacity-60"
                >
                  {loadingMore ? t("loadingMore") : t("loadMore")}
                </button>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  )
}
