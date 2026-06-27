"use client"

import { ReportDialog } from "@/components/debate/report-dialog"
import { InviteDialog } from "@/components/debate/invite-dialog"
import type { ReportTargetType } from "@/lib/moderation"

type Props = {
  actionError: string | null
  canShowNewTurnIndicator: boolean
  currentUsername?: string
  invitedUsernames: Set<string>
  invitingUsername: string | null
  inviteOpen: boolean
  latestTurnNumber: number | null
  onInviteActionErrorChange: (value: string | null) => void
  onInviteOpenChange: (value: boolean) => void
  onInviteUser: (username: string) => void
  onPendingScrollTurnChange: (value: number | null) => void
  onReportOpenChange: (value: boolean) => void
  onReportTargetIdChange: (value: string | null) => void
  onReportTargetTypeChange: (value: ReportTargetType | null) => void
  onSetShowNewTurnIndicator: (value: boolean) => void
  onSubmitReport: (payload: { reason: "hate" | "harassment" | "spam" | "off_topic" | "illegal" | "other"; details?: string }) => Promise<void>
  reportOpen: boolean
  reportTargetType: ReportTargetType | null
  t: any
}

export function DebatePageOverlays({
  actionError,
  canShowNewTurnIndicator,
  currentUsername,
  invitedUsernames,
  invitingUsername,
  inviteOpen,
  latestTurnNumber,
  onInviteActionErrorChange,
  onInviteOpenChange,
  onInviteUser,
  onPendingScrollTurnChange,
  onReportOpenChange,
  onReportTargetIdChange,
  onReportTargetTypeChange,
  onSetShowNewTurnIndicator,
  onSubmitReport,
  reportOpen,
  reportTargetType,
  t,
}: Props) {
  return (
    <>
      {canShowNewTurnIndicator && latestTurnNumber != null && (
        <button
          type="button"
          onClick={() => {
            onSetShowNewTurnIndicator(false)
            onPendingScrollTurnChange(latestTurnNumber)
          }}
          className="fixed bottom-5 right-5 z-40 rounded-full bg-[var(--text-primary)] px-4 py-2 text-[12px] font-medium text-[var(--bg-primary)] shadow-[0px_10px_30px_rgba(15,12,10,0.2)] hover:opacity-90"
        >
          {t("active.newTurn")}
        </button>
      )}

      <InviteDialog
        open={inviteOpen}
        currentUsername={currentUsername}
        actionError={actionError}
        invitedUsernames={invitedUsernames}
        invitingUsername={invitingUsername}
        onActionErrorChange={onInviteActionErrorChange}
        onInviteUser={onInviteUser}
        onOpenChange={onInviteOpenChange}
      />

      <ReportDialog
        open={reportOpen}
        targetType={reportTargetType}
        onOpenChange={(open) => {
          onReportOpenChange(open)
          if (!open) {
            onReportTargetTypeChange(null)
            onReportTargetIdChange(null)
          }
        }}
        onSubmit={onSubmitReport}
      />
    </>
  )
}
