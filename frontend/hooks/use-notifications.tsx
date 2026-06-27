"use client"

import { createContext, useCallback, useContext, useEffect, useState } from "react"
import { usePathname } from "next/navigation"
import { toast } from "sonner"
import { useAuth } from "@/hooks/use-auth"
import { useNotificationSocket } from "@/hooks/use-notification-socket"
import { emitChallengesRefresh } from "@/lib/challenges-events"
import { joinDebate, respondChallenge } from "@/lib/debates-api"
import { fetchNotifications, markAllNotificationsRead, markNotificationRead } from "@/lib/notifications-api"
import type { NotificationItem } from "@/lib/notifications"
import type { WSEvent } from "@/lib/websocket"

const PER_PAGE = 20

interface NotificationContextValue {
  notifications: NotificationItem[]
  unreadCount: number
  loading: boolean
  loadingMore: boolean
  hasMore: boolean
  error: string | null
  refresh: () => Promise<void>
  loadMore: () => Promise<void>
  markRead: (id: string) => Promise<void>
  markAllRead: () => Promise<void>
}

const NotificationContext = createContext<NotificationContextValue>({
  notifications: [],
  unreadCount: 0,
  loading: false,
  loadingMore: false,
  hasMore: false,
  error: null,
  refresh: async () => {},
  loadMore: async () => {},
  markRead: async () => {},
  markAllRead: async () => {},
})

export function NotificationProvider({ children }: { children: React.ReactNode }) {
  const { status } = useAuth()
  const pathname = usePathname()
  const [notifications, setNotifications] = useState<NotificationItem[]>([])
  const [unreadCount, setUnreadCount] = useState(0)
  const [loading, setLoading] = useState(false)
  const [loadingMore, setLoadingMore] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [page, setPage] = useState(1)
  const [totalPages, setTotalPages] = useState(1)

  const refresh = useCallback(async () => {
    if (status !== "authenticated") return
    setLoading(true)
    setError(null)
    try {
      const res = await fetchNotifications({ page: 1, perPage: PER_PAGE })
      setNotifications(res.data ?? [])
      setUnreadCount(res.meta?.unread_count ?? 0)
      setPage(res.meta?.page ?? 1)
      setTotalPages(res.meta?.total_pages ?? 1)
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load notifications.")
    } finally {
      setLoading(false)
    }
  }, [status])

  const loadMore = useCallback(async () => {
    if (status !== "authenticated") return
    if (loading || loadingMore) return
    if (page >= totalPages) return

    setLoadingMore(true)
    setError(null)
    try {
      const nextPage = page + 1
      const res = await fetchNotifications({ page: nextPage, perPage: PER_PAGE })
      setNotifications((prev) => [...prev, ...(res.data ?? [])])
      setUnreadCount(res.meta?.unread_count ?? unreadCount)
      setPage(res.meta?.page ?? nextPage)
      setTotalPages(res.meta?.total_pages ?? totalPages)
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load more notifications.")
    } finally {
      setLoadingMore(false)
    }
  }, [status, loading, loadingMore, page, totalPages, unreadCount])

  useEffect(() => {
    if (status !== "authenticated") {
      setNotifications([])
      setUnreadCount(0)
      setLoading(false)
      setLoadingMore(false)
      setError(null)
      setPage(1)
      setTotalPages(1)
      return
    }
    void refresh()
  }, [status, refresh])

  // Live notifications via WebSocket — single connection for the entire app.
  const handleWSEvent = useCallback((event: WSEvent) => {
    if (event.type === "notification.new") {
      const data = event.data as {
        id?: string
        type: string
        message: string
        debate_id: string | null
        debate_slug?: string
        turn_number?: number
        created_at?: string
      }

      const newNotif: NotificationItem = {
        id: data.id ?? crypto.randomUUID(),
        type: data.type as NotificationItem["type"],
        message: data.message,
        debate_id: data.debate_id,
        debate_slug: data.debate_slug ?? null,
        turn_number: data.turn_number ?? null,
        read: false,
        created_at: data.created_at ?? new Date().toISOString(),
      }

      setNotifications((prev) => [newNotif, ...prev])
      setUnreadCount((prev) => prev + 1)

      if (data.type === "challenge_received") {
        emitChallengesRefresh()
      }

      // Show toast for new notifications
      const baseDebateHref = data.debate_slug
        ? `/debate/${data.debate_slug}`
        : data.debate_id
        ? `/debate/${data.debate_id}`
        : null
      const debateHref =
        baseDebateHref && typeof data.turn_number === "number"
          ? `${baseDebateHref}?turn=${data.turn_number}`
          : baseDebateHref

      const normalizePath = (value: string) => value.replace(/\/+$/, "") || "/"
      const isViewingSameDebate = Boolean(
        baseDebateHref && normalizePath(pathname || "") === normalizePath(baseDebateHref)
      )

      // Suppress toast when already on this debate page.
      if (!isViewingSameDebate) {
        if (data.type === "challenge_received" && data.debate_id) {
          toast(data.message, {
            description: "Accept or decline this challenge",
            action: {
              label: "Accept",
              onClick: () => {
                void (async () => {
                  try {
                    await respondChallenge(data.debate_id as string, true)
                    emitChallengesRefresh()
                    await refresh()
                    if (debateHref) {
                      window.location.href = debateHref
                    }
                  } catch {
                    void refresh()
                    toast.error("Could not accept challenge.")
                  }
                })()
              },
            },
            cancel: {
              label: "Decline",
              onClick: () => {
                void (async () => {
                  try {
                    await respondChallenge(data.debate_id as string, false)
                    emitChallengesRefresh()
                    await refresh()
                  } catch {
                    void refresh()
                    toast.error("Could not decline challenge.")
                  }
                })()
              },
            },
            duration: 8000,
          })
          return
        }

        if (data.type === "debate_invited" && data.debate_id) {
          toast(data.message, {
            description: "Join this open debate as Side B",
            action: {
              label: "Join",
              onClick: () => {
                void (async () => {
                  try {
                    await joinDebate(data.debate_id as string)
                    await refresh()
                    if (debateHref) {
                      window.location.href = debateHref
                    }
                  } catch {
                    void refresh()
                    toast.error("Could not join this debate.")
                  }
                })()
              },
            },
            duration: 7000,
          })
          return
        }

        toast(data.message, {
          description: debateHref ? "Click to view debate" : undefined,
          action: debateHref
            ? {
                label: "View",
                onClick: () => {
                  window.location.href = debateHref
                },
              }
            : undefined,
          duration: 5000,
        })
      }
    }
  }, [pathname, refresh])

  useNotificationSocket(handleWSEvent, status === "authenticated")

  const markRead = useCallback(
    async (id: string) => {
      const current = notifications.find((item) => item.id === id)
      if (current && !current.read) {
        setNotifications((prev) => prev.map((item) => (item.id === id ? { ...item, read: true } : item)))
        setUnreadCount((prev) => (prev > 0 ? prev - 1 : 0))
      }
      try {
        await markNotificationRead(id)
      } catch (_) {
        void refresh()
      }
    },
    [notifications, refresh]
  )

  const markAllRead = useCallback(async () => {
    setNotifications((prev) => prev.map((item) => ({ ...item, read: true })))
    setUnreadCount(0)
    try {
      await markAllNotificationsRead()
    } catch (_) {
      void refresh()
    }
  }, [refresh])

  return (
    <NotificationContext.Provider
      value={{
        notifications,
        unreadCount,
        loading,
        loadingMore,
        hasMore: page < totalPages,
        error,
        refresh,
        loadMore,
        markRead,
        markAllRead,
      }}
    >
      {children}
    </NotificationContext.Provider>
  )
}

export function useNotifications() {
  return useContext(NotificationContext)
}
