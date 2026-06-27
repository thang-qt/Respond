import { api } from "@/lib/api"
import type { ApiListResponse, ApiResponse } from "@/lib/api-types"
import type { DebateFeedItem } from "@/lib/debates"
import type { BlockedUser, ProfileUpdateResult, UserProfile, UserSearchProfile } from "@/lib/users"

export async function fetchUserProfile(username: string) {
  return api.get<ApiResponse<UserProfile>>(`/users/${encodeURIComponent(username)}`)
}

export async function fetchUserDebates(
  username: string,
  params: {
    page?: number
    perPage?: number
  } = {}
) {
  const search = new URLSearchParams()
  if (params.page) search.set("page", params.page.toString())
  if (params.perPage) search.set("per_page", params.perPage.toString())
  const query = search.toString()
  return api.get<ApiListResponse<DebateFeedItem>>(
    `/users/${encodeURIComponent(username)}/debates${query ? `?${query}` : ""}`
  )
}

export async function updateMyProfile(payload: { bio?: string; default_reveal?: boolean }) {
  return api.put<ApiResponse<ProfileUpdateResult>>("/users/me", payload)
}

export async function searchUsers(
  q?: string,
  params: {
    page?: number
    perPage?: number
  } = {}
) {
  const search = new URLSearchParams()
  const query = q?.trim()
  if (query) search.set("q", query)
  if (params.page) search.set("page", params.page.toString())
  if (params.perPage) search.set("per_page", params.perPage.toString())
  return api.get<ApiListResponse<UserSearchProfile>>(`/users/search?${search.toString()}`)
}

export async function fetchExploreUsers(
  params: {
    q?: string
    tagSlug?: string | null
    tagSlugs?: string[] | null
    page?: number
    perPage?: number
    tagMode?: "any" | "all"
    context?: "explore" | "invite" | "search"
  } = {}
) {
  const search = new URLSearchParams()
  const query = params.q?.trim()
  if (query) search.set("q", query)
  if (params.tagSlug) search.set("tag", params.tagSlug)
  if (params.tagSlugs && params.tagSlugs.length > 0) {
    search.set("tags", params.tagSlugs.join(","))
  }
  if (params.tagMode) search.set("tag_mode", params.tagMode)
  if (params.context) search.set("context", params.context)
  if (params.page) search.set("page", params.page.toString())
  if (params.perPage) search.set("per_page", params.perPage.toString())
  return api.get<ApiListResponse<UserSearchProfile>>(`/explore/users?${search.toString()}`)
}

export async function fetchMyBlockedUsers() {
  return api.get<ApiResponse<BlockedUser[]>>("/users/me/blocks")
}

export async function blockUser(username: string) {
  return api.post<ApiResponse<{ username: string; blocked: boolean }>>(`/users/${encodeURIComponent(username)}/block`)
}

export async function unblockUser(username: string) {
  return api.delete<void>(`/users/${encodeURIComponent(username)}/block`)
}
