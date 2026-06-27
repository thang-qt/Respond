import { api } from "@/lib/api"
import type { ApiListResponse } from "@/lib/api-types"
import type { NotificationItem } from "@/lib/notifications"

export async function fetchNotifications(params: {
  unreadOnly?: boolean
  page?: number
  perPage?: number
} = {}) {
  const search = new URLSearchParams()
  if (params.unreadOnly !== undefined) search.set("unread_only", params.unreadOnly ? "true" : "false")
  if (params.page) search.set("page", params.page.toString())
  if (params.perPage) search.set("per_page", params.perPage.toString())
  const query = search.toString()
  return api.get<ApiListResponse<NotificationItem>>(`/users/me/notifications${query ? `?${query}` : ""}`)
}

export async function markNotificationRead(id: string) {
  return api.put<void>(`/users/me/notifications/${id}/read`)
}

export async function markAllNotificationsRead() {
  return api.put<void>("/users/me/notifications/read-all")
}
