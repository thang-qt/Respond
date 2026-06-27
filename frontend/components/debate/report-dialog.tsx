"use client"

import { useMemo, useState } from "react"
import { useTranslations } from "next-intl"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import type { ReportReason, ReportTargetType } from "@/lib/moderation"

interface ReportDialogProps {
  open: boolean
  targetType: ReportTargetType | null
  onOpenChange: (open: boolean) => void
  onSubmit: (payload: { reason: ReportReason; details?: string }) => Promise<void>
}

const reasonOptions: ReportReason[] = ["hate", "harassment", "spam", "off_topic", "illegal", "other"]

export function ReportDialog({ open, targetType, onOpenChange, onSubmit }: ReportDialogProps) {
  const t = useTranslations("reportDialog")
  const tCommon = useTranslations("common")
  const [reason, setReason] = useState<ReportReason>("harassment")
  const [details, setDetails] = useState("")
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  const title = useMemo(() => {
    if (targetType === "debate") return t("title.debate")
    if (targetType === "turn") return t("title.turn")
    if (targetType === "comment") return t("title.comment")
    return t("title.content")
  }, [targetType, t])

  const handleSubmit = async () => {
    if (submitting) return
    const trimmed = details.trim()
    if (trimmed.length > 500) {
      setError(t("errors.detailsTooLong"))
      return
    }

    setSubmitting(true)
    setError(null)
    try {
      await onSubmit({ reason, details: trimmed.length > 0 ? trimmed : undefined })
      setDetails("")
      setReason("harassment")
      onOpenChange(false)
    } catch (err) {
      setError(err instanceof Error ? err.message : t("errors.generic"))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[460px] bg-[var(--bg-surface)] border border-[var(--border-default)]">
        <DialogHeader>
          <DialogTitle className="text-[var(--text-primary)] font-sans">{title}</DialogTitle>
          <DialogDescription className="text-[var(--text-secondary)] font-sans">
            {t("description")}
          </DialogDescription>
        </DialogHeader>

        <div className="flex flex-col gap-3 py-2">
          <label className="flex flex-col gap-1 text-[12px] font-medium text-[var(--text-secondary)] font-sans">
            {t("reason")}
            <select
              value={reason}
              onChange={(event) => setReason(event.target.value as ReportReason)}
              className="h-9 rounded-md border border-[var(--border-default)] bg-[var(--bg-surface)] px-3 text-[13px] text-[var(--text-primary)] outline-none focus:ring-2 focus:ring-[var(--border-strong)]"
            >
              {reasonOptions.map((option) => (
                <option key={option} value={option}>
                  {t(`reasons.${option}`)}
                </option>
              ))}
            </select>
          </label>

          <label className="flex flex-col gap-1 text-[12px] font-medium text-[var(--text-secondary)] font-sans">
            {t("details")}
            <textarea
              value={details}
              onChange={(event) => setDetails(event.target.value)}
              rows={4}
              maxLength={500}
              placeholder={t("detailsPlaceholder")}
              className="rounded-md border border-[var(--border-default)] bg-[var(--bg-surface)] px-3 py-2 text-[13px] text-[var(--text-primary)] outline-none focus:ring-2 focus:ring-[var(--border-strong)] resize-y"
            />
          </label>
          <div className="text-[11px] text-[var(--text-muted)] font-sans text-right">{details.length}/500</div>
          {error && <div className="text-[12px] text-[var(--error)] font-sans">{error}</div>}
        </div>

        <DialogFooter>
          <button
            type="button"
            onClick={() => onOpenChange(false)}
            className="h-9 px-4 rounded-md border border-[var(--border-default)] text-[var(--text-secondary)] text-[13px] font-medium font-sans hover:bg-[var(--bg-surface-alt)]"
            disabled={submitting}
          >
            {tCommon("cancel")}
          </button>
          <button
            type="button"
            onClick={handleSubmit}
            className="h-9 px-4 rounded-md bg-[var(--text-primary)] text-[var(--bg-primary)] text-[13px] font-medium font-sans hover:opacity-90 disabled:opacity-60"
            disabled={submitting}
          >
            {submitting ? t("submitting") : t("submit")}
          </button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
