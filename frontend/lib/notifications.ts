export type NotificationType =
  | "debate_invited"
  | "challenge_received"
  | "challenge_accepted"
  | "challenge_declined"
  | "challenge_expired"
  | "your_turn"
  | "debate_joined"
  | "debate_ended"
  | "turn_expiring"
  | "seat_open"
  | "draw_proposed"
  | "comment_on_reflection"
  | "replacement_joined"
  | "content_hidden"
  | "content_restored"

export interface NotificationItem {
  id: string
  type: NotificationType
  message: string
  debate_id: string | null
  debate_slug: string | null
  turn_number: number | null
  read: boolean
  created_at: string
}
