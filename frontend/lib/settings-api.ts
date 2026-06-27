import { api } from "@/lib/api"
import type { ApiResponse } from "@/lib/api-types"
import type { ProfileUpdateResult } from "@/lib/users"

export interface NotificationSettings {
  email_your_turn: boolean
  email_debate_joined: boolean
  email_debate_ended: boolean
  email_turn_expiring: boolean
  email_seat_open: boolean
  email_draw_proposed: boolean
}

export interface EmailUpdateResult {
  email: string
  email_verified: boolean
}

export interface MessageResult {
  message: string
}

export interface InviteRecord {
  id: string
  email: string
  status: "pending" | "accepted" | "revoked" | "expired"
  expires_at: string
  accepted_at?: string
  created_at: string
}

export interface InviteRevokeResult {
  id: string
  status: string
}

export async function fetchMyNotificationSettings() {
  return api.get<ApiResponse<NotificationSettings>>("/users/me/settings/notifications")
}

export async function updateMyNotificationSettings(payload: NotificationSettings) {
  return api.put<ApiResponse<NotificationSettings>>("/users/me/settings/notifications", payload)
}

export async function updateMyProfileSettings(payload: { bio?: string; default_reveal?: boolean; locale?: "en" | "vi" }) {
  return api.put<ApiResponse<ProfileUpdateResult>>("/users/me", payload)
}

export async function updateMyPassword(payload: { current_password: string; new_password: string }) {
  return api.put<void>("/users/me/password", payload)
}

export async function updateMyEmail(payload: { email: string; password: string }) {
  return api.put<ApiResponse<EmailUpdateResult>>("/users/me/email", payload)
}

export async function resendVerificationEmail() {
  return api.post<ApiResponse<MessageResult>>("/auth/resend-verification")
}

export async function createMyInvite(payload: { email: string }) {
  return api.post<ApiResponse<InviteRecord>>("/users/me/invites", payload)
}

export async function fetchMyInvites(params?: { status?: "pending" | "accepted" | "revoked" | "expired" | "all"; page?: number; per_page?: number }) {
  const query = new URLSearchParams()
  if (params?.status) query.set("status", params.status)
  if (params?.page) query.set("page", String(params.page))
  if (params?.per_page) query.set("per_page", String(params.per_page))
  const suffix = query.toString() ? `?${query.toString()}` : ""
  return api.get<ApiResponse<InviteRecord[]>>("/users/me/invites" + suffix)
}

export async function revokeMyInvite(id: string) {
  return api.post<ApiResponse<InviteRevokeResult>>(`/users/me/invites/${id}/revoke`)
}
