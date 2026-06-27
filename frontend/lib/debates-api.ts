import { api } from "@/lib/api"
import type { ApiListResponse, ApiResponse } from "@/lib/api-types"
import type {
  Tag,
  ChallengeLobbyEntry,
  CommentVoteResult,
  DebateComment,
  DebateDetail,
  DebateEndResult,
  DebateFeed,
  DebateFeedItem,
  DebateTurn,
  DebateVoteResult,
  DrawProposeResult,
  DrawRespondResult,
  DebateInviteResult,
  ExtensionRespondResult,
  ChallengeListItem,
  JoinDebateResult,
  ReplaceDebateResult,
  RespondChallengeResult,
  RevealIdentityResult,
} from "@/lib/debates"

export async function fetchDebates(params: {
  feed: DebateFeed
  tagSlug?: string | null
  tagSlugs?: string[] | null
  tagMode?: "any" | "all"
  page?: number
  perPage?: number
}) {
  const search = new URLSearchParams()
  if (params.feed) search.set("feed", params.feed)
  if (params.tagSlug) search.set("tag", params.tagSlug)
  if (params.tagSlugs && params.tagSlugs.length > 0) {
    search.set("tags", params.tagSlugs.join(","))
  }
  if (params.tagMode) search.set("tag_mode", params.tagMode)
  if (params.page) search.set("page", params.page.toString())
  if (params.perPage) search.set("per_page", params.perPage.toString())

  const query = search.toString()
  return api.get<ApiListResponse<DebateFeedItem>>(`/debates${query ? `?${query}` : ""}`)
}

export async function fetchExplore(params: {
  sort?: "hot" | "rising" | "new" | "random"
  tagSlug?: string | null
  tagSlugs?: string[] | null
  tagMode?: "any" | "all"
  page?: number
  perPage?: number
}) {
  const search = new URLSearchParams()
  if (params.sort) search.set("sort", params.sort)
  if (params.tagSlug) search.set("tag", params.tagSlug)
  if (params.tagSlugs && params.tagSlugs.length > 0) {
    search.set("tags", params.tagSlugs.join(","))
  }
  if (params.tagMode) search.set("tag_mode", params.tagMode)
  if (params.page) search.set("page", params.page.toString())
  if (params.perPage) search.set("per_page", params.perPage.toString())

  const query = search.toString()
  return api.get<ApiListResponse<DebateFeedItem>>(`/explore${query ? `?${query}` : ""}`)
}

export async function fetchDebate(id: string) {
  return api.get<ApiResponse<DebateDetail>>(`/debates/${id}`)
}

export async function searchDebates(params: {
  q: string
  sort?: "relevance" | "new"
  tagSlug?: string | null
  tagSlugs?: string[] | null
  tagMode?: "any" | "all"
  page?: number
  perPage?: number
}) {
  const search = new URLSearchParams()
  search.set("q", params.q)
  if (params.sort) search.set("sort", params.sort)
  if (params.tagSlug) search.set("tag", params.tagSlug)
  if (params.tagSlugs && params.tagSlugs.length > 0) {
    search.set("tags", params.tagSlugs.join(","))
  }
  if (params.tagMode) search.set("tag_mode", params.tagMode)
  if (params.page) search.set("page", params.page.toString())
  if (params.perPage) search.set("per_page", params.perPage.toString())

  return api.get<ApiListResponse<DebateFeedItem>>(`/debates/search?${search.toString()}`)
}

export async function createDebate(payload: {
  topic: string
  tag_ids: string[]
  time_mode: "marathon" | "standard" | "rapid" | "blitz"
  turn_limit: number
  context?: string
  opening_turn: string
  opening_turn_ai_assisted?: boolean
  opening_turn_ai_note?: string
}) {
  return api.post<ApiResponse<DebateDetail>>("/debates", payload)
}

export async function createChallengeDebate(payload: {
  invited_username: string
  topic: string
  tag_ids: string[]
  time_mode: "marathon" | "standard" | "rapid" | "blitz"
  turn_limit: number
  context?: string
  opening_turn: string
  opening_turn_ai_assisted?: boolean
  opening_turn_ai_note?: string
}) {
  return api.post<ApiResponse<DebateDetail>>("/debates/challenges", payload)
}

export async function createRechallengeDebate(
  sourceDebateID: string,
  payload: {
    topic: string
    tag_ids: string[]
    time_mode: "marathon" | "standard" | "rapid" | "blitz"
    turn_limit: number
    context?: string
    opening_turn: string
    opening_turn_ai_assisted?: boolean
    opening_turn_ai_note?: string
  }
) {
  return api.post<ApiResponse<DebateDetail>>(`/debates/${sourceDebateID}/rechallenge`, payload)
}

export async function fetchMyChallenges(params: {
  box?: "inbox" | "outbox"
  status?: "pending" | "accepted" | "declined" | "expired" | "all"
  page?: number
  perPage?: number
} = {}) {
  const search = new URLSearchParams()
  if (params.box) search.set("box", params.box)
  if (params.status) search.set("status", params.status)
  if (params.page) search.set("page", params.page.toString())
  if (params.perPage) search.set("per_page", params.perPage.toString())
  const query = search.toString()
  return api.get<ApiListResponse<ChallengeListItem>>(`/users/me/challenges${query ? `?${query}` : ""}`)
}

export async function respondChallenge(id: string, accept: boolean) {
  return api.post<ApiResponse<RespondChallengeResult>>(`/debates/${id}/challenge/respond`, { accept })
}

export async function inviteToDebate(id: string, invitedUsername: string) {
  return api.post<ApiResponse<DebateInviteResult>>(`/debates/${id}/invites`, { invited_username: invitedUsername })
}

export async function joinDebate(id: string) {
  return api.post<ApiResponse<JoinDebateResult>>(`/debates/${id}/join`)
}

export async function replaceDebate(id: string) {
  return api.post<ApiResponse<ReplaceDebateResult>>(`/debates/${id}/replace`)
}

export async function submitTurn(id: string, payload: { content: string; ai_assisted?: boolean; ai_note?: string }) {
  return api.post<ApiResponse<DebateTurn>>(`/debates/${id}/turns`, payload)
}

export async function toggleDebateVote(id: string) {
  return api.post<ApiResponse<DebateVoteResult>>(`/debates/${id}/vote`)
}

export async function followDebate(id: string) {
  return api.post<void>(`/debates/${id}/follow`)
}

export async function unfollowDebate(id: string) {
  return api.delete<void>(`/debates/${id}/follow`)
}

export async function fetchDebateComments(
  id: string,
  params: {
    sort?: "newest" | "top"
    page?: number
    perPage?: number
  } = {}
) {
  const search = new URLSearchParams()
  if (params.sort) search.set("sort", params.sort)
  if (params.page) search.set("page", params.page.toString())
  if (params.perPage) search.set("per_page", params.perPage.toString())
  const query = search.toString()
  return api.get<ApiListResponse<DebateComment>>(`/debates/${id}/comments${query ? `?${query}` : ""}`)
}

export async function postDebateComment(
  id: string,
  payload: {
    content: string
    parent_id?: string | null
    is_reflection?: boolean
  }
) {
  return api.post<ApiResponse<DebateComment>>(`/debates/${id}/comments`, payload)
}

export async function updateDebateComment(
  debateId: string,
  commentId: string,
  payload: {
    content: string
  }
) {
  return api.put<ApiResponse<{ id: string; content: string; updated_at: string | null }>>(
    `/debates/${debateId}/comments/${commentId}`,
    payload
  )
}

export async function deleteDebateComment(debateId: string, commentId: string) {
  return api.delete<void>(`/debates/${debateId}/comments/${commentId}`)
}

export async function toggleCommentVote(commentId: string) {
  return api.post<ApiResponse<CommentVoteResult>>(`/comments/${commentId}/vote`)
}

export async function concedeDebate(id: string) {
  return api.post<ApiResponse<DebateEndResult>>(`/debates/${id}/concede`)
}

export async function resignDebate(id: string) {
  return api.post<ApiResponse<DebateEndResult>>(`/debates/${id}/resign`)
}

export async function proposeDraw(id: string) {
  return api.post<ApiResponse<DrawProposeResult>>(`/debates/${id}/draw/propose`)
}

export async function respondDraw(id: string, accept: boolean) {
  return api.post<ApiResponse<DrawRespondResult>>(`/debates/${id}/draw/respond`, { accept })
}

export async function revealDebateIdentity(id: string, reveal: boolean) {
  return api.post<ApiResponse<RevealIdentityResult>>(`/debates/${id}/reveal`, { reveal })
}

export async function respondExtension(id: string, accept: boolean) {
  return api.post<ApiResponse<ExtensionRespondResult>>(`/debates/${id}/extend/respond`, { accept })
}

export async function fetchTags() {
  return api.get<ApiResponse<Tag[]>>("/tags")
}

export async function fetchMyTagFollows() {
  return api.get<ApiResponse<Tag[]>>("/users/me/tag-follows")
}

export async function searchTags(params: { q: string; limit?: number }) {
  const search = new URLSearchParams()
  search.set("q", params.q)
  if (params.limit) search.set("limit", params.limit.toString())

  return api.get<ApiResponse<Tag[]>>(`/tags/search?${search.toString()}`)
}

export async function replaceMyTagFollows(tagIDs: string[]) {
  return api.put<ApiResponse<Tag[]>>("/users/me/tag-follows", { tag_ids: tagIDs })
}

// ── Challenge Lobby ──────────────────────────────────────────────────────────

export async function fetchLobbyEntries(params: {
  tagSlugs?: string[]
  tagMode?: "any" | "all"
  page?: number
  perPage?: number
} = {}) {
  const search = new URLSearchParams()
  if (params.tagSlugs && params.tagSlugs.length > 0) search.set("tags", params.tagSlugs.join(","))
  if (params.tagMode) search.set("tag_mode", params.tagMode)
  if (params.page) search.set("page", params.page.toString())
  if (params.perPage) search.set("per_page", params.perPage.toString())
  const query = search.toString()
  return api.get<ApiListResponse<ChallengeLobbyEntry>>(`/lobby/challenges${query ? `?${query}` : ""}`)
}

export async function getMyLobbyEntry() {
  return api.get<ApiResponse<ChallengeLobbyEntry>>("/users/me/lobby")
}

export async function getUserLobbyEntry(username: string) {
  return api.get<ApiResponse<ChallengeLobbyEntry>>(`/users/${username}/lobby`)
}

export async function upsertMyLobbyEntry(bioNote: string, tagIds: string[]) {
  return api.put<ApiResponse<ChallengeLobbyEntry>>("/users/me/lobby", { bio_note: bioNote, tag_ids: tagIds })
}

export async function deleteMyLobbyEntry() {
  return api.delete<void>("/users/me/lobby")
}
