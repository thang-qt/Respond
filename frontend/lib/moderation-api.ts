import { api } from "@/lib/api"
import type { ApiListResponse, ApiResponse } from "@/lib/api-types"
import type {
  HiddenContentItem,
  ReportDetail,
  ReportItem,
  ReportReason,
  ReportResolution,
  ReportTargetType,
  UserCapability,
  UserEnforcementAction,
  UserEnforcementActionType,
  UserEnforcementStatus,
} from "@/lib/moderation"

export async function createReport(payload: {
  target_type: ReportTargetType
  target_id: string
  reason: ReportReason
  details?: string
}) {
  return api.post<ApiResponse<ReportItem>>("/reports", payload)
}

export async function listAdminReports(params: {
  status?: "open" | "dismissed" | "actioned" | "all"
  target_type?: ReportTargetType
  page?: number
  per_page?: number
}) {
  const search = new URLSearchParams()
  if (params.status) search.set("status", params.status)
  if (params.target_type) search.set("target_type", params.target_type)
  if (params.page) search.set("page", params.page.toString())
  if (params.per_page) search.set("per_page", params.per_page.toString())
  const query = search.toString()

  return api.get<ApiListResponse<ReportItem>>(`/admin/reports${query ? `?${query}` : ""}`)
}

export async function getAdminReport(id: string) {
  return api.get<ApiResponse<ReportDetail>>(`/admin/reports/${id}`)
}

export async function resolveAdminReport(id: string, payload: { resolution: ReportResolution; note?: string }) {
  return api.post<ApiResponse<{ id: string; status: string; resolution: ReportResolution; note?: string; reviewed_at: string }>>(
    `/admin/reports/${id}/resolve`,
    payload
  )
}

export async function updateAdminUserRole(id: string, role: "user" | "moderator" | "admin") {
  return api.post<ApiResponse<{ id: string; role: "user" | "moderator" | "admin" }>>(`/admin/users/${id}/role`, { role })
}

export async function listAdminHiddenContent(params: { target_type?: ReportTargetType; limit?: number } = {}) {
  const search = new URLSearchParams()
  if (params.target_type) search.set("target_type", params.target_type)
  if (params.limit) search.set("limit", params.limit.toString())
  const query = search.toString()
  return api.get<ApiResponse<HiddenContentItem[]>>(`/admin/content/hidden${query ? `?${query}` : ""}`)
}

export async function restoreAdminHiddenContent(targetType: ReportTargetType, targetID: string, note: string) {
  return api.post<ApiResponse<{ target_type: ReportTargetType; target_id: string; restored: true }>>(
    `/admin/content/${targetType}/${targetID}/restore`,
    { note }
  )
}

export async function moderateAdminContent(targetType: ReportTargetType, targetID: string, action: "hide" | "restore", note: string) {
  return api.post<ApiResponse<{ target_type: ReportTargetType; target_id: string; resolution: "hide" | "restore" }>>(
    `/admin/content/${targetType}/${targetID}/${action}`,
    { note }
  )
}

export async function createAdminUserEnforcementAction(
  userID: string,
  payload: {
    action: Exclude<UserEnforcementActionType, "revoke">
    capabilities?: UserCapability[]
    expires_at?: string
    note: string
  }
) {
  return api.post<ApiResponse<UserEnforcementAction>>(`/admin/users/${userID}/enforcement-actions`, payload)
}

export async function revokeAdminUserEnforcementAction(userID: string, actionID: string, note: string) {
  return api.post<ApiResponse<{ id: string; status: "revoked"; revoked_at: string }>>(
    `/admin/users/${userID}/enforcement-actions/${actionID}/revoke`,
    { note }
  )
}

export async function listAdminUserEnforcementActions(params: {
  user_id: string
  status?: UserEnforcementStatus | "all"
  page?: number
  per_page?: number
}) {
  const search = new URLSearchParams()
  if (params.status) search.set("status", params.status)
  if (params.page) search.set("page", params.page.toString())
  if (params.per_page) search.set("per_page", params.per_page.toString())
  const query = search.toString()

  return api.get<ApiListResponse<UserEnforcementAction>>(`/admin/users/${params.user_id}/enforcement-actions${query ? `?${query}` : ""}`)
}
