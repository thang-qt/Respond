import Link from "next/link"
import type { HiddenContentItem, ReportDetail, ReportItem, ReportResolution, UserCapability, UserEnforcementAction, UserEnforcementActionType } from "@/lib/moderation"

export type StatusFilter = "open" | "dismissed" | "actioned" | "all"
export type EnforcementFilter = "active" | "expired" | "revoked" | "all"
export type RoleValue = "user" | "moderator" | "admin"

export const ENFORCEMENT_ACTION_TYPES = ["warning", "restriction", "suspension", "ban"] as const
export const CAPABILITY_TYPES = ["create_debate", "comment", "vote", "follow", "report"] as const satisfies readonly UserCapability[]

export function capabilityTranslationKey(capability: UserCapability) {
  return `capability${capability.split("_").map((part) => part.charAt(0).toUpperCase() + part.slice(1)).join("")}`
}

export function moderationStatusKey(value: string) {
  return `status${value.charAt(0).toUpperCase() + value.slice(1)}`
}

export function moderationResolutionKey(value: string) {
  return `resolution${value.charAt(0).toUpperCase() + value.slice(1)}`
}

export function enforcementKey(value: string) {
  return `enforcement${value.charAt(0).toUpperCase() + value.slice(1)}`
}

export function reportTargetHref(report: { target_type: "debate" | "turn" | "comment"; target_id: string; debate_slug?: string; debate_id?: string; turn_number?: number }) {
  const debateSegment = report.debate_slug || report.debate_id
  if (!debateSegment) return null
  const base = `/debate/${debateSegment}`
  if (report.target_type === "debate") return base
  if (report.target_type === "turn") return typeof report.turn_number === "number" ? `${base}?turn=${report.turn_number}` : base
  return `${base}#comment-${report.target_id}`
}

export function KeyValue({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <div className="text-[12px] text-[var(--text-muted)] uppercase tracking-wide">{label}</div>
      <div className="text-[14px] text-[var(--text-primary)] mt-1">{value}</div>
    </div>
  )
}

export function ReportsQueueSection({
  items,
  loading,
  meta,
  page,
  selectedID,
  statusFilter,
  t,
  onPageChange,
  onSelect,
  onStatusFilterChange,
}: {
  items: ReportItem[]
  loading: boolean
  meta: { page: number; total_pages: number; total: number }
  page: number
  selectedID: string | null
  statusFilter: StatusFilter
  t: (key: string, values?: Record<string, unknown>) => string
  onPageChange: (page: number | ((value: number) => number)) => void
  onSelect: (id: string) => void
  onStatusFilterChange: (value: StatusFilter) => void
}) {
  return (
    <>
      <div className="flex items-center justify-between gap-3 mb-5">
        <h1 className="text-[22px] font-semibold font-sans text-[var(--text-primary)]">{t("queueTitle")}</h1>
        <select
          value={statusFilter}
          onChange={(event) => onStatusFilterChange(event.target.value as StatusFilter)}
          className="h-9 rounded-md border border-[var(--border-default)] bg-[var(--bg-surface)] px-3 text-[13px] text-[var(--text-primary)]"
        >
          {(["open", "dismissed", "actioned", "all"] as const).map((value) => (
            <option key={value} value={value}>{t(moderationStatusKey(value))}</option>
          ))}
        </select>
      </div>

      <section className="rounded-lg border border-[var(--border-default)] bg-[var(--bg-surface)]">
        {loading ? (
          <div className="p-4 text-[13px] text-[var(--text-secondary)] font-sans">{t("loadingReports")}</div>
        ) : items.length === 0 ? (
          <div className="p-4 text-[13px] text-[var(--text-secondary)] font-sans">{t("noReports")}</div>
        ) : (
          <>
            <div className="divide-y divide-[var(--border-subtle)]">
              {items.map((item) => (
                <button
                  key={item.id}
                  type="button"
                  onClick={() => onSelect(item.id)}
                  className={`w-full text-left p-4 transition-colors ${selectedID === item.id ? "bg-[var(--bg-surface-alt)]" : "hover:bg-[var(--bg-surface-alt)]/50"}`}
                >
                  <div className="flex items-center justify-between gap-2">
                    <span className="text-[12px] font-medium uppercase tracking-wide text-[var(--text-secondary)]">{item.target_type}</span>
                    <span className="text-[11px] text-[var(--text-muted)]">{t(moderationStatusKey(item.status))}</span>
                  </div>
                  <div className="mt-1 text-[13px] text-[var(--text-primary)] font-medium">{item.reason}</div>
                  {item.resolution && <div className="mt-1 text-[11px] text-[var(--text-secondary)]">{t("action")}: {t(moderationResolutionKey(item.resolution))}</div>}
                  <div className="mt-1 text-[11px] text-[var(--text-secondary)]">{t("reportedUserLabel", { username: item.target_author?.username ?? t("unknown") })}</div>
                  <div className="mt-1 text-[11px] text-[var(--text-muted)]">{new Date(item.created_at).toLocaleString()}</div>
                  {item.trusted_report && <div className="mt-2 inline-flex items-center rounded-full bg-[var(--warning-light)] text-[var(--warning)] text-[10px] px-2 py-0.5 font-medium">{t("trustedReport")}</div>}
                </button>
              ))}
            </div>
            <div className="flex items-center justify-between gap-2 border-t border-[var(--border-subtle)] p-3">
              <div className="text-[12px] text-[var(--text-muted)]">{t("pageInfo", { page: meta.page, totalPages: Math.max(meta.total_pages, 1), total: meta.total })}</div>
              <div className="flex items-center gap-2">
                <button type="button" onClick={() => onPageChange((value) => Math.max(value - 1, 1))} className="h-8 px-3 rounded-md border border-[var(--border-default)] text-[12px] text-[var(--text-secondary)] disabled:opacity-50" disabled={loading || page <= 1}>{t("previous")}</button>
                <button type="button" onClick={() => onPageChange((value) => value + 1)} className="h-8 px-3 rounded-md border border-[var(--border-default)] text-[12px] text-[var(--text-secondary)] disabled:opacity-50" disabled={loading || page >= meta.total_pages}>{t("next")}</button>
              </div>
            </div>
          </>
        )}
      </section>
    </>
  )
}

export function ReportDetailSection({ detail, detailLoading, resolutionNote, selectedID, t, updating, onResolve, onResolutionNoteChange }: { detail: ReportDetail | null; detailLoading: boolean; resolutionNote: string; selectedID: string | null; t: (key: string, values?: Record<string, unknown>) => string; updating: boolean; onResolve: (resolution: ReportResolution) => void; onResolutionNoteChange: (value: string) => void }) {
  return (
    <section className="rounded-lg border border-[var(--border-default)] bg-[var(--bg-surface)] p-5">
      {!selectedID ? <div className="text-[13px] text-[var(--text-secondary)]">{t("selectToReview")}</div> : detailLoading ? <div className="text-[13px] text-[var(--text-secondary)]">{t("loadingDetail")}</div> : detail ? (
        <div className="flex flex-col gap-4">
          <div><div className="text-[12px] text-[var(--text-muted)] uppercase tracking-wide">{t("reportId")}</div><div className="text-[14px] text-[var(--text-primary)] font-medium mt-1">{detail.id}</div></div>
          <div className="grid gap-3 sm:grid-cols-2">
            <KeyValue label={t("targetType")} value={detail.target_type} />
            <KeyValue label={t("reason")} value={detail.reason} />
            <KeyValue label={t("status")} value={t(moderationStatusKey(detail.status))} />
            <KeyValue label={t("action")} value={detail.resolution ? t(moderationResolutionKey(detail.resolution)) : "-"} />
            <KeyValue label={t("trusted")} value={detail.trusted_report ? t("yes") : t("no")} />
            <KeyValue label={t("reporter")} value={detail.reporter?.username ?? t("unknown")} />
            <KeyValue label={t("reportedUser")} value={detail.target_author?.username ?? t("unknown")} />
          </div>
          {reportTargetHref(detail) && <div><Link href={reportTargetHref(detail) ?? "#"} className="inline-flex h-8 items-center rounded-md border border-[var(--border-default)] px-3 text-[12px] text-[var(--text-secondary)] hover:bg-[var(--bg-surface-alt)]">{t("openTarget")}</Link></div>}
          {detail.details && <div><div className="text-[12px] text-[var(--text-muted)] uppercase tracking-wide mb-1">{t("reporterDetails")}</div><div className="text-[13px] text-[var(--text-secondary)] whitespace-pre-wrap">{detail.details}</div></div>}
          <div><div className="text-[12px] text-[var(--text-muted)] uppercase tracking-wide mb-1">{t("targetContent")}</div><div className="rounded-md border border-[var(--border-default)] bg-[var(--bg-surface-alt)] p-3 text-[13px] text-[var(--text-primary)] whitespace-pre-wrap">{detail.target.content}</div></div>
          {detail.status === "open" && <><textarea value={resolutionNote} onChange={(event) => onResolutionNoteChange(event.target.value)} rows={3} maxLength={500} placeholder={t("noteRequired")} className="w-full rounded-md border border-[var(--border-default)] bg-[var(--bg-surface)] px-3 py-2 text-[13px] text-[var(--text-primary)]" /><div className="text-[11px] text-[var(--text-muted)] text-right">{resolutionNote.length}/500</div><div className="flex flex-wrap gap-2"><button type="button" onClick={() => onResolve("dismiss")} className="h-9 px-3 rounded-md border border-[var(--border-default)] text-[13px] text-[var(--text-secondary)] hover:bg-[var(--bg-surface-alt)]" disabled={updating}>{t("dismiss")}</button><button type="button" onClick={() => onResolve("hide")} className="h-9 px-3 rounded-md bg-[var(--error)] text-white text-[13px] hover:opacity-90" disabled={updating}>{t("hideContent")}</button><button type="button" onClick={() => onResolve("restore")} className="h-9 px-3 rounded-md bg-[var(--text-primary)] text-[var(--bg-primary)] text-[13px] hover:opacity-90" disabled={updating}>{t("restoreContent")}</button></div></>}
          {detail.status !== "open" && detail.resolution_note && <div className="text-[12px] text-[var(--text-secondary)]">{t("resolutionNote", { note: detail.resolution_note })}</div>}
        </div>
      ) : <div className="text-[13px] text-[var(--text-secondary)]">{t("failedLoadReport")}</div>}
    </section>
  )
}

export function HiddenContentSection({ hiddenItems, hiddenLoading, t, updating, onRestore }: { hiddenItems: HiddenContentItem[]; hiddenLoading: boolean; t: (key: string, values?: Record<string, unknown>) => string; updating: boolean; onRestore: (item: HiddenContentItem) => void }) {
  return <section className="mt-6 rounded-lg border border-[var(--border-default)] bg-[var(--bg-surface)] p-5"><h2 className="text-[16px] font-semibold text-[var(--text-primary)]">{t("hiddenQueueTitle")}</h2>{hiddenLoading ? <div className="mt-3 text-[13px] text-[var(--text-secondary)]">{t("loadingHidden")}</div> : hiddenItems.length === 0 ? <div className="mt-3 text-[13px] text-[var(--text-secondary)]">{t("noHidden")}</div> : <div className="mt-3 divide-y divide-[var(--border-subtle)] border border-[var(--border-default)] rounded-md">{hiddenItems.map((item) => { const href = reportTargetHref(item); return <div key={`${item.target_type}:${item.target_id}`} className="p-3"><div className="flex items-center justify-between gap-3"><div className="text-[12px] text-[var(--text-secondary)] uppercase tracking-wide">{item.target_type}{item.target_author?.username ? ` • ${item.target_author.username}` : ""}</div><button type="button" onClick={() => onRestore(item)} className="h-8 px-3 rounded-md bg-[var(--text-primary)] text-[var(--bg-primary)] text-[12px] hover:opacity-90" disabled={updating}>{t("restore")}</button></div><div className="mt-2 text-[13px] text-[var(--text-primary)] whitespace-pre-wrap">{item.content}</div>{href && <Link href={href} className="mt-2 inline-flex text-[12px] text-[var(--text-secondary)] underline">{t("openInDebate")}</Link>}</div>})}</div>}</section>
}

export function EnforcementSection({ enforcementAction, enforcementExpiresAt, enforcementFilter, enforcementItems, enforcementLoading, enforcementNote, enforcementUserID, selectedCapabilities, t, updating, onActionChange, onCapabilitiesChange, onCreate, onExpiresAtChange, onFilterChange, onLoadHistory, onNoteChange, onRevoke, onUserIDChange }: { enforcementAction: Exclude<UserEnforcementActionType, "revoke">; enforcementExpiresAt: string; enforcementFilter: EnforcementFilter; enforcementItems: UserEnforcementAction[]; enforcementLoading: boolean; enforcementNote: string; enforcementUserID: string; selectedCapabilities: UserCapability[]; t: (key: string, values?: Record<string, unknown>) => string; updating: boolean; onActionChange: (value: Exclude<UserEnforcementActionType, "revoke">) => void; onCapabilitiesChange: (value: UserCapability[] | ((current: UserCapability[]) => UserCapability[])) => void; onCreate: () => void; onExpiresAtChange: (value: string) => void; onFilterChange: (value: EnforcementFilter) => void; onLoadHistory: () => void; onNoteChange: (value: string) => void; onRevoke: (action: UserEnforcementAction) => void; onUserIDChange: (value: string) => void }) {
  return <section className="mt-6 rounded-lg border border-[var(--border-default)] bg-[var(--bg-surface)] p-5"><h2 className="text-[16px] font-semibold text-[var(--text-primary)]">{t("enforcementTitle")}</h2><p className="text-[12px] text-[var(--text-secondary)] mt-1">{t("enforcementDescription")}</p><div className="mt-3 grid gap-3"><input value={enforcementUserID} onChange={(event) => onUserIDChange(event.target.value)} placeholder={t("targetUserId")} className="h-9 rounded-md border border-[var(--border-default)] px-3 text-[13px] bg-[var(--bg-surface)] text-[var(--text-primary)]" /><div className="grid gap-2 sm:grid-cols-3"><select value={enforcementAction} onChange={(event) => onActionChange(event.target.value as Exclude<UserEnforcementActionType, "revoke">)} className="h-9 rounded-md border border-[var(--border-default)] px-3 text-[13px] bg-[var(--bg-surface)] text-[var(--text-primary)]">{ENFORCEMENT_ACTION_TYPES.map((action) => <option key={action} value={action}>{t(enforcementKey(action))}</option>)}</select>{(enforcementAction === "restriction" || enforcementAction === "suspension") && <input type="datetime-local" value={enforcementExpiresAt} onChange={(event) => onExpiresAtChange(event.target.value)} className="h-9 rounded-md border border-[var(--border-default)] px-3 text-[13px] bg-[var(--bg-surface)] text-[var(--text-primary)]" />}<button type="button" onClick={onLoadHistory} className="h-9 px-3 rounded-md border border-[var(--border-default)] text-[13px] text-[var(--text-secondary)] hover:bg-[var(--bg-surface-alt)]" disabled={updating || enforcementLoading}>{t("loadHistory")}</button></div>{enforcementAction === "restriction" && <div className="flex flex-wrap gap-2">{CAPABILITY_TYPES.map((capability) => { const selected = selectedCapabilities.includes(capability); return <button key={capability} type="button" onClick={() => onCapabilitiesChange((current) => current.includes(capability) ? current.filter((item) => item !== capability) : [...current, capability])} className={`h-8 px-3 rounded-md border text-[12px] ${selected ? "border-[var(--text-primary)] bg-[var(--bg-surface-alt)] text-[var(--text-primary)]" : "border-[var(--border-default)] text-[var(--text-secondary)]"}`}>{t(capabilityTranslationKey(capability))}</button>})}</div>}<textarea value={enforcementNote} onChange={(event) => onNoteChange(event.target.value)} rows={3} maxLength={500} placeholder={t("noteRequiredField")} className="w-full rounded-md border border-[var(--border-default)] bg-[var(--bg-surface)] px-3 py-2 text-[13px] text-[var(--text-primary)]" /><div className="text-[11px] text-[var(--text-muted)] text-right">{enforcementNote.length}/500</div><div><button type="button" onClick={onCreate} className="h-9 px-3 rounded-md bg-[var(--text-primary)] text-[var(--bg-primary)] text-[13px] hover:opacity-90" disabled={updating}>{t("createAction")}</button></div></div><div className="mt-5"><div className="flex items-center justify-between gap-2"><h3 className="text-[14px] font-medium text-[var(--text-primary)]">{t("enforcementHistory")}</h3><select value={enforcementFilter} onChange={(event) => onFilterChange(event.target.value as EnforcementFilter)} className="h-8 rounded-md border border-[var(--border-default)] px-2 text-[12px] bg-[var(--bg-surface)] text-[var(--text-primary)]"><option value="active">{t("statusFilterActive")}</option><option value="expired">{t("statusFilterExpired")}</option><option value="revoked">{t("statusFilterRevoked")}</option><option value="all">{t("statusFilterAll")}</option></select></div>{!enforcementUserID.trim() ? <div className="mt-3 text-[12px] text-[var(--text-secondary)]">{t("enterUserId")}</div> : enforcementLoading ? <div className="mt-3 text-[12px] text-[var(--text-secondary)]">{t("loadingActions")}</div> : enforcementItems.length === 0 ? <div className="mt-3 text-[12px] text-[var(--text-secondary)]">{t("noActionsInFilter")}</div> : <div className="mt-3 divide-y divide-[var(--border-subtle)] border border-[var(--border-default)] rounded-md">{enforcementItems.map((item) => <div key={item.id} className="p-3"><div className="flex items-center justify-between gap-2"><div className="text-[12px] text-[var(--text-primary)] font-medium">{t(enforcementKey(item.action))}{item.status ? ` • ${t(`statusFilter${item.status.charAt(0).toUpperCase() + item.status.slice(1)}`)}` : ""}</div>{(item.action === "restriction" || item.action === "suspension") && item.status === "active" && <button type="button" onClick={() => onRevoke(item)} className="h-7 px-2 rounded-md border border-[var(--border-default)] text-[11px] text-[var(--text-secondary)] hover:bg-[var(--bg-surface-alt)]" disabled={updating}>{t("revoke")}</button>}</div><div className="mt-1 text-[12px] text-[var(--text-secondary)]">{t("by")} {item.created_by?.username ?? item.actor_user_id}{item.expires_at ? ` ${t("expiresAt", { date: new Date(item.expires_at).toLocaleString() })}` : ""}</div>{item.capabilities.length > 0 && <div className="mt-1 text-[12px] text-[var(--text-secondary)]">{t("capabilities")} {item.capabilities.join(", ")}</div>}<div className="mt-1 text-[12px] text-[var(--text-secondary)]">{t("noteLabel")} {item.note}</div></div>)}</div>}</div></section>
}

export function RoleManagementSection({ roleUserID, roleValue, t, updating, onRoleUserIDChange, onRoleValueChange, onSubmit }: { roleUserID: string; roleValue: RoleValue; t: (key: string) => string; updating: boolean; onRoleUserIDChange: (value: string) => void; onRoleValueChange: (value: RoleValue) => void; onSubmit: () => void }) {
  return <section className="mt-6 rounded-lg border border-[var(--border-default)] bg-[var(--bg-surface)] p-5"><h2 className="text-[16px] font-semibold text-[var(--text-primary)]">{t("roleManagement")}</h2><p className="text-[12px] text-[var(--text-secondary)] mt-1">{t("roleDescription")}</p><div className="mt-3 flex flex-col sm:flex-row gap-2"><input value={roleUserID} onChange={(event) => onRoleUserIDChange(event.target.value)} placeholder={t("rolePlaceholder")} className="flex-1 h-9 rounded-md border border-[var(--border-default)] px-3 text-[13px] bg-[var(--bg-surface)] text-[var(--text-primary)]" /><select value={roleValue} onChange={(event) => onRoleValueChange(event.target.value as RoleValue)} className="h-9 rounded-md border border-[var(--border-default)] px-3 text-[13px] bg-[var(--bg-surface)] text-[var(--text-primary)]"><option value="user">{t("roleUser")}</option><option value="moderator">{t("roleModerator")}</option><option value="admin">{t("roleAdmin")}</option></select><button type="button" onClick={onSubmit} className="h-9 px-3 rounded-md bg-[var(--text-primary)] text-[var(--bg-primary)] text-[13px] hover:opacity-90" disabled={updating}>{t("updateRole")}</button></div></section>
}
