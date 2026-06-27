"use client"

import { useEffect, useState } from "react"
import { useTranslations } from "next-intl"
import { toast } from "sonner"
import type { HiddenContentItem, ReportDetail, ReportItem, ReportResolution, UserCapability, UserEnforcementAction, UserEnforcementActionType } from "@/lib/moderation"
import {
  createAdminUserEnforcementAction,
  getAdminReport,
  listAdminHiddenContent,
  listAdminReports,
  listAdminUserEnforcementActions,
  resolveAdminReport,
  restoreAdminHiddenContent,
  revokeAdminUserEnforcementAction,
  updateAdminUserRole,
} from "@/lib/moderation-api"
import { useAuth } from "@/hooks/use-auth"
import {
  EnforcementSection,
  HiddenContentSection,
  ReportDetailSection,
  ReportsQueueSection,
  RoleManagementSection,
  type EnforcementFilter,
  type RoleValue,
  type StatusFilter,
} from "@/components/admin/moderation-sections"

function toRFC3339FromLocal(value: string) {
  if (!value) return undefined
  const parsed = new Date(value)
  if (Number.isNaN(parsed.getTime())) return undefined
  return parsed.toISOString()
}

export default function AdminModerationPage() {
  const { status: authStatus, user } = useAuth()
  const t = useTranslations("moderation")
  const [statusFilter, setStatusFilter] = useState<StatusFilter>("open")
  const [items, setItems] = useState<ReportItem[]>([])
  const [meta, setMeta] = useState({ page: 1, total_pages: 1, total: 0 })
  const [page, setPage] = useState(1)
  const [loading, setLoading] = useState(true)
  const [selectedID, setSelectedID] = useState<string | null>(null)
  const [detail, setDetail] = useState<ReportDetail | null>(null)
  const [detailLoading, setDetailLoading] = useState(false)
  const [resolutionNote, setResolutionNote] = useState("")
  const [updating, setUpdating] = useState(false)
  const [hiddenItems, setHiddenItems] = useState<HiddenContentItem[]>([])
  const [hiddenLoading, setHiddenLoading] = useState(true)
  const [roleUserID, setRoleUserID] = useState("")
  const [roleValue, setRoleValue] = useState<RoleValue>("moderator")
  const [enforcementUserID, setEnforcementUserID] = useState("")
  const [enforcementAction, setEnforcementAction] = useState<Exclude<UserEnforcementActionType, "revoke">>("warning")
  const [selectedCapabilities, setSelectedCapabilities] = useState<UserCapability[]>([])
  const [enforcementExpiresAt, setEnforcementExpiresAt] = useState("")
  const [enforcementNote, setEnforcementNote] = useState("")
  const [enforcementFilter, setEnforcementFilter] = useState<EnforcementFilter>("active")
  const [enforcementItems, setEnforcementItems] = useState<UserEnforcementAction[]>([])
  const [enforcementLoading, setEnforcementLoading] = useState(false)

  const isPrivileged = user?.role === "moderator" || user?.role === "admin"

  useEffect(() => {
    if (!isPrivileged) {
      setItems([])
      setMeta({ page: 1, total_pages: 1, total: 0 })
      setSelectedID(null)
      setDetail(null)
      setLoading(false)
      return
    }

    let active = true
    setLoading(true)
    listAdminReports({ status: statusFilter, page, per_page: 50 })
      .then((res) => {
        if (!active) return
        setItems(res.data)
        setMeta({
          page: res.meta?.page ?? 1,
          total_pages: res.meta?.total_pages ?? 1,
          total: res.meta?.total ?? res.data.length,
        })
      })
      .catch((err) => {
        if (!active) return
        const message = err instanceof Error ? err.message : t("failedLoad")
        toast.error(message)
      })
      .finally(() => {
        if (!active) return
        setLoading(false)
      })

    return () => {
      active = false
    }
  }, [isPrivileged, statusFilter, page, t])

  useEffect(() => {
    if (!selectedID || !isPrivileged) {
      setDetail(null)
      return
    }

    let active = true
    setDetailLoading(true)
    getAdminReport(selectedID)
      .then((res) => {
        if (!active) return
        setDetail(res.data)
      })
      .catch((err) => {
        if (!active) return
        const message = err instanceof Error ? err.message : t("failedLoadDetail")
        toast.error(message)
      })
      .finally(() => {
        if (!active) return
        setDetailLoading(false)
      })

    return () => {
      active = false
    }
  }, [selectedID, isPrivileged, t])

  useEffect(() => {
    if (!isPrivileged) {
      setHiddenItems([])
      setHiddenLoading(false)
      return
    }

    let active = true
    setHiddenLoading(true)
    listAdminHiddenContent({ limit: 100 })
      .then((res) => {
        if (!active) return
        setHiddenItems(res.data)
      })
      .catch((err) => {
        if (!active) return
        const message = err instanceof Error ? err.message : t("hiddenLoadError")
        toast.error(message)
      })
      .finally(() => {
        if (!active) return
        setHiddenLoading(false)
      })

    return () => {
      active = false
    }
  }, [isPrivileged, t])

  useEffect(() => {
    if (!isPrivileged || !enforcementUserID.trim()) return
    void loadEnforcementHistory(enforcementUserID)
  }, [isPrivileged, enforcementFilter])

  const refreshReports = async () => {
    const listRes = await listAdminReports({ status: statusFilter, page, per_page: 50 })
    setItems(listRes.data)
    setMeta({
      page: listRes.meta?.page ?? 1,
      total_pages: listRes.meta?.total_pages ?? 1,
      total: listRes.meta?.total ?? listRes.data.length,
    })
  }

  const handleResolve = async (resolution: ReportResolution) => {
    if (!selectedID || !detail || updating) return
    const trimmedNote = resolutionNote.trim()
    if ((resolution === "hide" || resolution === "restore") && trimmedNote.length === 0) {
      toast.error(t("noteRequired"))
      return
    }
    if (trimmedNote.length > 500) {
      toast.error(t("noteMaxLength"))
      return
    }
    setUpdating(true)
    try {
      await resolveAdminReport(selectedID, { resolution, note: trimmedNote || undefined })
      toast.success(t("reportUpdated"))
      const detailRes = await getAdminReport(selectedID)
      await refreshReports()
      setDetail(detailRes.data)
      setResolutionNote("")
    } catch (err) {
      const message = err instanceof Error ? err.message : t("resolveError")
      toast.error(message)
    } finally {
      setUpdating(false)
    }
  }

  const handleRoleUpdate = async () => {
    if (updating) return
    const trimmedID = roleUserID.trim()
    if (!trimmedID) {
      toast.error(t("roleRequired"))
      return
    }

    setUpdating(true)
    try {
      await updateAdminUserRole(trimmedID, roleValue)
      toast.success(t("roleUpdated"))
      setRoleUserID("")
    } catch (err) {
      const message = err instanceof Error ? err.message : t("roleError")
      toast.error(message)
    } finally {
      setUpdating(false)
    }
  }

  const loadEnforcementHistory = async (userID: string) => {
    const trimmedID = userID.trim()
    if (!trimmedID) {
      setEnforcementItems([])
      return
    }

    setEnforcementLoading(true)
    try {
      const res = await listAdminUserEnforcementActions({
        user_id: trimmedID,
        status: enforcementFilter,
        page: 1,
        per_page: 50,
      })
      setEnforcementItems(res.data)
    } catch (err) {
      const message = err instanceof Error ? err.message : t("enforcementLoadError")
      toast.error(message)
    } finally {
      setEnforcementLoading(false)
    }
  }

  const handleCreateEnforcementAction = async () => {
    if (updating) return
    const trimmedID = enforcementUserID.trim()
    const trimmedNote = enforcementNote.trim()
    if (!trimmedID) {
      toast.error(t("enforcementTargetRequired"))
      return
    }
    if (!trimmedNote) {
      toast.error(t("enforcementNoteRequired"))
      return
    }
    if (trimmedNote.length > 500) {
      toast.error(t("noteMaxLength"))
      return
    }
    if (enforcementAction === "restriction" && selectedCapabilities.length === 0) {
      toast.error(t("enforcementCapabilityRequired"))
      return
    }

    setUpdating(true)
    try {
      await createAdminUserEnforcementAction(trimmedID, {
        action: enforcementAction,
        capabilities: enforcementAction === "restriction" ? selectedCapabilities : undefined,
        expires_at: enforcementAction === "restriction" || enforcementAction === "suspension" ? toRFC3339FromLocal(enforcementExpiresAt) : undefined,
        note: trimmedNote,
      })
      toast.success(t("enforcementCreated"))
      setEnforcementNote("")
      if (enforcementAction !== "restriction") setSelectedCapabilities([])
      await loadEnforcementHistory(trimmedID)
    } catch (err) {
      const message = err instanceof Error ? err.message : t("enforcementCreateError")
      toast.error(message)
    } finally {
      setUpdating(false)
    }
  }

  const handleRevokeEnforcement = async (action: UserEnforcementAction) => {
    if (updating) return
    if (!enforcementUserID.trim()) {
      toast.error(t("enforcementTargetRequired"))
      return
    }
    const note = window.prompt(t("enforcementRevokePrompt"))?.trim()
    if (!note) {
      toast.error(t("enforcementRevokeRequired"))
      return
    }
    if (note.length > 500) {
      toast.error(t("noteMaxLength"))
      return
    }

    setUpdating(true)
    try {
      await revokeAdminUserEnforcementAction(enforcementUserID.trim(), action.id, note)
      toast.success(t("enforcementRevoked"))
      await loadEnforcementHistory(enforcementUserID.trim())
    } catch (err) {
      const message = err instanceof Error ? err.message : t("enforcementRevokeError")
      toast.error(message)
    } finally {
      setUpdating(false)
    }
  }

  const handleRestoreHidden = async (item: HiddenContentItem) => {
    if (updating) return
    const note = window.prompt(t("contentRestorePrompt"))?.trim()
    if (!note) {
      toast.error(t("contentRestoreRequired"))
      return
    }
    if (note.length > 500) {
      toast.error(t("noteMaxLength"))
      return
    }

    setUpdating(true)
    try {
      await restoreAdminHiddenContent(item.target_type, item.target_id, note)
      toast.success(t("contentRestored"))
      const hiddenRes = await listAdminHiddenContent({ limit: 100 })
      setHiddenItems(hiddenRes.data)
    } catch (err) {
      const message = err instanceof Error ? err.message : t("contentRestoreError")
      toast.error(message)
    } finally {
      setUpdating(false)
    }
  }

  if (authStatus === "loading") return <div className="min-h-[60vh]" />

  if (!isPrivileged) {
    return (
      <div className="max-w-[820px] mx-auto px-4 sm:px-6 py-8">
        <div className="rounded-lg border border-[var(--border-default)] bg-[var(--bg-surface)] p-5">
          <h1 className="text-[20px] font-semibold font-sans text-[var(--text-primary)]">{t("title")}</h1>
          <p className="mt-2 text-[14px] text-[var(--text-secondary)] font-sans">{t("noAccess")}</p>
        </div>
      </div>
    )
  }

  return (
    <div className="max-w-[1200px] mx-auto px-4 sm:px-6 py-8">
      <div className="grid gap-4 lg:grid-cols-[360px_1fr]">
        <div>
          <ReportsQueueSection
            items={items}
            loading={loading}
            meta={meta}
            page={page}
            selectedID={selectedID}
            statusFilter={statusFilter}
            t={t as any}
            onPageChange={setPage}
            onSelect={setSelectedID}
            onStatusFilterChange={(value) => {
              setStatusFilter(value)
              setPage(1)
            }}
          />
        </div>
        <ReportDetailSection
          detail={detail}
          detailLoading={detailLoading}
          resolutionNote={resolutionNote}
          selectedID={selectedID}
          t={t as any}
          updating={updating}
          onResolve={handleResolve}
          onResolutionNoteChange={setResolutionNote}
        />
      </div>

      <HiddenContentSection
        hiddenItems={hiddenItems}
        hiddenLoading={hiddenLoading}
        t={t as any}
        updating={updating}
        onRestore={handleRestoreHidden}
      />

      <EnforcementSection
        enforcementAction={enforcementAction}
        enforcementExpiresAt={enforcementExpiresAt}
        enforcementFilter={enforcementFilter}
        enforcementItems={enforcementItems}
        enforcementLoading={enforcementLoading}
        enforcementNote={enforcementNote}
        enforcementUserID={enforcementUserID}
        selectedCapabilities={selectedCapabilities}
        t={t as any}
        updating={updating}
        onActionChange={(value) => {
          setEnforcementAction(value)
          if (value !== "restriction") setSelectedCapabilities([])
        }}
        onCapabilitiesChange={setSelectedCapabilities}
        onCreate={handleCreateEnforcementAction}
        onExpiresAtChange={setEnforcementExpiresAt}
        onFilterChange={setEnforcementFilter}
        onLoadHistory={() => void loadEnforcementHistory(enforcementUserID)}
        onNoteChange={setEnforcementNote}
        onRevoke={handleRevokeEnforcement}
        onUserIDChange={setEnforcementUserID}
      />

      {user?.role === "admin" && (
        <RoleManagementSection
          roleUserID={roleUserID}
          roleValue={roleValue}
          t={t as any}
          updating={updating}
          onRoleUserIDChange={setRoleUserID}
          onRoleValueChange={setRoleValue}
          onSubmit={handleRoleUpdate}
        />
      )}
    </div>
  )
}
