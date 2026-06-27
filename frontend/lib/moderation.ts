export type ReportTargetType = "debate" | "turn" | "comment"
export type ReportReason = "hate" | "harassment" | "spam" | "off_topic" | "illegal" | "other"
export type ReportStatus = "open" | "dismissed" | "actioned"

export interface ReportUserRef {
  id: string
  username: string
}

export interface ReportItem {
  id: string
  target_type: ReportTargetType
  target_id: string
  reason: ReportReason
  details?: string
  status: ReportStatus
  resolution?: ReportResolution
  trusted_report: boolean
  created_at: string
  reporter?: ReportUserRef
  target_author?: ReportUserRef
  debate_id?: string
  debate_slug?: string
  turn_number?: number
}

export interface ReportDetail extends ReportItem {
  reviewed_by_user_id?: string
  reviewed_at?: string
  resolution_note?: string
  target: {
    hidden: boolean
    content: string
  }
}

export type ReportResolution = "dismiss" | "hide" | "restore"

export interface HiddenContentItem {
  target_type: ReportTargetType
  target_id: string
  debate_id?: string
  debate_slug?: string
  turn_number?: number
  target_author?: ReportUserRef
  content: string
  hidden_at?: string
}

export type UserEnforcementActionType = "warning" | "restriction" | "suspension" | "ban" | "revoke"
export type UserCapability = "create_debate" | "comment" | "vote" | "follow" | "report" | "invite"
export type UserEnforcementStatus = "active" | "expired" | "revoked"

export interface UserEnforcementAction {
  id: string
  target_user_id: string
  actor_user_id: string
  action: UserEnforcementActionType
  capabilities: UserCapability[]
  expires_at?: string
  revoked_at?: string
  note: string
  created_at: string
  status?: UserEnforcementStatus
  created_by?: ReportUserRef
}
