"use client"

import { useEffect, useMemo, useState } from "react"
import Link from "next/link"
import { useRouter } from "next/navigation"
import { useTranslations } from "next-intl"
import { Bell, Check } from "@phosphor-icons/react"
import { ApiError } from "@/lib/api"
import { emitChallengesRefresh } from "@/lib/challenges-events"
import { joinDebate, respondChallenge } from "@/lib/debates-api"
import { useNotifications } from "@/hooks/use-notifications"
import { cn, formatTimeAgo } from "@/lib/utils"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"

export default function NotificationsDropdown({
  collapsed = false,
  className,
  align = "start",
}: {
  collapsed?: boolean
  className?: string
  align?: "start" | "center" | "end"
}) {
  const [open, setOpen] = useState(false)
  const [actioningID, setActioningID] = useState<string | null>(null)
  const [actionError, setActionError] = useState<string | null>(null)
  const router = useRouter()
  const t = useTranslations("notifications")
  const tCommon = useTranslations("common")
  const { notifications, unreadCount, loading, error, refresh, markAllRead, markRead } = useNotifications()

  useEffect(() => {
    if (open) {
      void refresh()
    }
  }, [open, refresh])

  const badgeText = useMemo(() => {
    if (unreadCount < 1) return null
    return unreadCount > 9 ? "9+" : `${unreadCount}`
  }, [unreadCount])

  return (
    <DropdownMenu open={open} onOpenChange={setOpen}>
      <DropdownMenuTrigger asChild>
        <button
          type="button"
          className={cn(
            `relative flex items-center gap-2.5 px-2.5 py-2 rounded-md transition-colors font-sans text-[13px] font-medium hover:bg-[var(--border-subtle)] ${
              collapsed ? "justify-center" : ""
            } ${unreadCount > 0 ? "text-[var(--text-primary)]" : "text-[var(--text-secondary)]"}`,
            className
          )}
          title={collapsed ? t("title") : undefined}
        >
          <Bell size={18} className="shrink-0" />
          {!collapsed && <span className="truncate">{t("title")}</span>}
          {badgeText && (
            <span className="absolute top-1.5 right-1.5 min-w-4 h-4 px-1 rounded-full bg-[var(--text-primary)] text-[10px] text-[var(--bg-primary)] font-semibold flex items-center justify-center">
              {badgeText}
            </span>
          )}
        </button>
      </DropdownMenuTrigger>
      <DropdownMenuContent
        align={align}
        className="w-[320px] bg-[var(--bg-surface)] border-[var(--border-default)] text-[var(--text-primary)] shadow-[0px_10px_30px_rgba(15,12,10,0.12)] p-2"
      >
        <div className="flex items-center justify-between px-2 py-1">
          <span className="text-[13px] font-semibold font-sans">{t("title")}</span>
          <button
            type="button"
            onClick={() => void markAllRead()}
            disabled={unreadCount < 1}
            className="text-[11px] font-medium text-[var(--text-secondary)] hover:text-[var(--text-primary)] disabled:opacity-50"
          >
            {t("markAllRead")}
          </button>
        </div>
        {actionError && (
          <div className="px-2 py-1 text-[11px] text-[var(--error)]">{actionError}</div>
        )}
        <DropdownMenuSeparator className="bg-[var(--border-subtle)]" />
        <div className="flex flex-col gap-1 max-h-[320px] overflow-y-auto">
          {loading ? (
            <div className="px-2 py-3 text-[12px] text-[var(--text-muted)]">{t("loading")}</div>
          ) : error ? (
            <div className="px-2 py-3 text-[12px] text-[var(--text-muted)]">{error}</div>
          ) : notifications.length === 0 ? (
            <div className="px-2 py-3 text-[12px] text-[var(--text-muted)]">{t("empty")}</div>
          ) : (
            notifications.map((item) => {
              const isUnread = !item.read
              const classes = `flex items-start gap-2 px-2 py-2 rounded-md transition-colors ${
                isUnread
                  ? "bg-[var(--bg-surface-alt)] text-[var(--text-primary)]"
                  : "text-[var(--text-secondary)] hover:bg-[var(--border-subtle)]"
              }`
              const baseDebateHref = item.debate_slug
                ? `/debate/${item.debate_slug}`
                : item.debate_id
                ? `/debate/${item.debate_id}`
                : null
              const debateHref =
                baseDebateHref && typeof item.turn_number === "number"
                  ? `${baseDebateHref}?turn=${item.turn_number}`
                  : baseDebateHref

              return (
                <div key={item.id} className={classes}>
                  <span
                    className={`mt-1 h-1.5 w-1.5 rounded-full shrink-0 ${
                      isUnread ? "bg-[var(--text-primary)]" : "bg-transparent"
                    }`}
                  />

                  <div className="flex-1 min-w-0">
                    {debateHref ? (
                      <Link href={debateHref} onClick={() => void markRead(item.id)} className="block">
                        <div className="text-[12px] leading-snug font-medium">{item.message}</div>
                        <div className="text-[11px] text-[var(--text-muted)] mt-1">{formatTimeAgo(item.created_at)}</div>
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
                        <div className="text-[12px] leading-snug font-medium">{item.message}</div>
                        <div className="text-[11px] text-[var(--text-muted)] mt-1">{formatTimeAgo(item.created_at)}</div>
                      </button>
                    )}

                    {item.type === "challenge_received" && item.debate_id && isUnread && (
                      <div className="mt-2 flex items-center gap-1.5">
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
                              setActionError(err instanceof ApiError ? err.message : t("errors.acceptChallenge"))
                              void refresh()
                            } finally {
                              setActioningID(null)
                            }
                          }}
                          disabled={actioningID === item.id}
                          className="h-6 px-2 rounded-md text-[11px] font-medium bg-[var(--text-primary)] text-[var(--bg-primary)] disabled:opacity-60"
                        >
                          {actioningID === item.id ? "..." : tCommon("accept")}
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
                              setActionError(err instanceof ApiError ? err.message : t("errors.declineChallenge"))
                              void refresh()
                            } finally {
                              setActioningID(null)
                            }
                          }}
                          disabled={actioningID === item.id}
                          className="h-6 px-2 rounded-md text-[11px] font-medium border border-[var(--border-default)] text-[var(--text-secondary)] hover:bg-[var(--bg-surface-alt)] disabled:opacity-60"
                        >
                          {tCommon("decline")}
                        </button>
                      </div>
                    )}

                    {item.type === "debate_invited" && item.debate_id && isUnread && (
                      <div className="mt-2 flex items-center gap-1.5">
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
                              setActionError(err instanceof ApiError ? err.message : t("errors.joinDebate"))
                              void refresh()
                            } finally {
                              setActioningID(null)
                            }
                          }}
                          disabled={actioningID === item.id}
                          className="h-6 px-2 rounded-md text-[11px] font-medium bg-[var(--text-primary)] text-[var(--bg-primary)] disabled:opacity-60"
                        >
                          {actioningID === item.id ? "..." : tCommon("join")}
                        </button>
                      </div>
                    )}
                  </div>

                  {isUnread && (
                    <button
                      type="button"
                      onClick={() => void markRead(item.id)}
                      className="shrink-0 mt-0.5 h-6 w-6 rounded-md flex items-center justify-center text-[var(--text-secondary)] hover:text-[var(--text-primary)] hover:bg-[var(--border-subtle)]"
                      aria-label={t("markAsRead")}
                      title={t("markAsRead")}
                    >
                      <Check size={14} />
                    </button>
                  )}
                </div>
              )
            })
          )}
        </div>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
