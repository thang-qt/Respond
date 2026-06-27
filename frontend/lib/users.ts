import type { DebateFeedItem } from "@/lib/debates"

export interface UserProfile {
  id?: string
  username: string
  bio: string
  rating: number
  wins: number
  losses: number
  draws: number
  debates_count: number
  created_at: string
}

export interface UserSearchProfile {
  id?: string
  username: string
  bio: string
  rating: number
  wins: number
  losses: number
  draws: number
  debates_count: number
  created_at?: string
  shared_tags?: string[]
}

export interface UserDebatesResult {
  debates: DebateFeedItem[]
  page: number
  totalPages: number
}

export interface BlockedUser {
  username: string
  blocked_at: string
}

export interface ProfileUpdateResult {
  id: string
  email: string
  email_verified: boolean
  username: string
  bio: string
  rating: number
  wins: number
  losses: number
  draws: number
  default_reveal: boolean
  locale: "en" | "vi"
  created_at: string
}
