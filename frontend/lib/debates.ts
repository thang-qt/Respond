export type DebateStatus = "waiting" | "active" | "pending_extension" | "waiting_replacement" | "finished" | "expired"
export type TimeMode = "marathon" | "standard" | "rapid" | "blitz"
export type DebateOutcome = "concession" | "walkover" | "draw" | "turn_limit" | null
export type DebateSide = "a" | "b"
export type DebateFeed = "trending" | "new" | "live" | "needs_challenger" | "following" | "following_tags"

export interface UserSummary {
  username: string
  rating: number
}

export interface Tag {
  id: string
  slug: string
  name: string
}

export interface DebateSideInfo {
  anonymous_id: string | null
  revealed: boolean
  user: UserSummary | null
}

export interface DebateFeedItem {
  id: string
  slug: string
  topic: string
  is_challenge: boolean
  invited_username?: string | null
  challenge_identity_visible: boolean
  tags: Tag[]
  time_mode: TimeMode
  turn_limit: number
  context: string | null
  latest_turn?: DebateTurn | null
  status: DebateStatus
  side_a: DebateSideInfo
  side_b: DebateSideInfo
  outcome: DebateOutcome
  winner_side: DebateSide | null
  current_turn_side: DebateSide
  turn_count: number
  turn_deadline: string | null
  draw_proposed_by: DebateSide | null
  open_side: DebateSide | null
  extension_deadline: string | null
  extension_a_accepted: boolean | null
  extension_b_accepted: boolean | null
  upvote_count: number
  viewer_has_upvoted?: boolean
  is_following?: boolean
  viewer_is_participant?: boolean
  spectator_count: number
  comment_count: number
  is_daily_prompt: boolean
  created_at: string
  started_at: string | null
  ended_at: string | null
  side_a_rating_delta?: number | null
  side_b_rating_delta?: number | null
  moderation_pending: boolean
  hidden: boolean
}

export interface DebateTurn {
  id: string
  turn_number: number
  side: DebateSide
  anonymous_id: string
  content: string
  hidden: boolean
  ai_assisted: boolean
  ai_note: string | null
  is_system: boolean
  created_at: string
}

export interface DebateEvent {
  id: string
  event_type: string
  side: DebateSide | null
  payload_json: Record<string, unknown>
  created_at: string
}

export interface DebateTimelineItem {
  type: "turn" | "event"
  created_at: string
  turn?: DebateTurn
  event?: DebateEvent
}

export interface DebateSeatStint {
  id: string
  side: DebateSide
  anonymous_id: string
  stint_index: number
  joined_at: string
  left_at: string | null
  left_reason: string | null
  replaced_by_stint_id: string | null
}

export interface DebateParticipantHistory {
  side_a: DebateSeatStint[]
  side_b: DebateSeatStint[]
}

export interface CommentUser {
  username: string
  rating: number
}

export interface DebateComment {
  id: string
  debate_id: string
  parent_id: string | null
  user: CommentUser
  content: string
  is_reflection: boolean
  is_debater: boolean
  debater_side: DebateSide | null
  debater_anonymous_id: string | null
  hidden: boolean
  upvote_count: number
  viewer_has_upvoted: boolean
  is_author: boolean
  created_at: string
  updated_at: string | null
  replies?: DebateComment[]
}

export interface CommentVoteResult {
  comment_id: string
  voted: boolean
  upvote_count: number
}

export interface DebateVoteResult {
  debate_id: string
  voted: boolean
  upvote_count: number
}

export interface DebateViewer {
  is_participant: boolean
  side: DebateSide | null
  has_upvoted: boolean
  is_following: boolean
  reveal_choice: boolean | null
}

export interface DebateDetail extends DebateFeedItem {
  turns: DebateTurn[]
  timeline: DebateTimelineItem[]
  participant_history: DebateParticipantHistory
  viewer: DebateViewer | null
}

export interface JoinDebateResult {
  debate_id: string
  side: "b"
  anonymous_id: string
  status: "active"
  turn_deadline: string | null
}

export interface DebateEndResult {
  debate_id: string
  status: string
  outcome: string | null
  winner_side: string | null
  ended_at: string | null
}

export interface DrawProposeResult {
  debate_id: string
  proposed_by: string
  status: string
}

export interface DrawRespondResult {
  debate_id: string
  draw_status?: string
  status?: string
  outcome?: string | null
  winner_side: string | null
  ended_at?: string | null
}

export interface ReplaceDebateResult {
  debate_id: string
  side: DebateSide
  anonymous_id: string
  status: "active"
  turn_deadline: string | null
}

export interface RevealIdentityResult {
  debate_id: string
  side: DebateSide
  revealed: boolean
  user: UserSummary | null
}

export interface ExtensionRespondResult {
  debate_id: string
  status: string
  turn_limit?: number
  outcome?: string | null
  winner_side?: string | null
  ended_at?: string | null
}

export interface ChallengeListItem {
  debate_id: string
  debate_slug: string
  topic: string
  time_mode: TimeMode
  turn_limit: number
  status: DebateStatus
  created_at: string
  challenge_expires_at: string | null
  challenger_username: string
  invited_username: string
}

export interface RespondChallengeResult {
  debate_id: string
  accepted: boolean
  status: DebateStatus
  side?: DebateSide
  anonymous_id?: string
  turn_deadline?: string | null
}

export interface DebateInviteResult {
  debate_id: string
  invited_username: string
}

export interface ChallengeLobbyEntry {
  username: string
  bio: string
  rating: number
  wins: number
  losses: number
  draws: number
  bio_note: string
  tags: Tag[]
  created_at: string
  updated_at: string
}

export const timeModeLabels: Record<TimeMode, string> = {
  marathon: "Marathon (7d)",
  standard: "Standard (48h)",
  rapid: "Rapid (12h)",
  blitz: "Blitz (2h)",
}

export const debateStatusLabels: Record<DebateStatus, string> = {
  waiting: "Needs Challenger",
  active: "Live",
  pending_extension: "Extension Vote",
  waiting_replacement: "Open Seat",
  finished: "Completed",
  expired: "Expired",
}

/**
 * Returns the resolved username for a side if the user has revealed their
 * identity, otherwise null. Use this when you need to decide whether to
 * render a profile link.
 */
export function resolveUsername(side: DebateSideInfo): string | null {
  return side.revealed ? (side.user?.username ?? null) : null
}

/**
 * Returns the best available display string for a side: real username when
 * revealed, anonymous ID otherwise, with a plain-text fallback.
 */
export function resolveDisplayName(side: DebateSideInfo, fallback: string): string {
  return resolveUsername(side) ?? side.anonymous_id ?? fallback
}
