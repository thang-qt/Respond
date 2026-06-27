"use client"

import { useCallback, useState } from "react"
import Link from "next/link"
import { ArrowUp, ChatCenteredText, Eye, BookmarkSimple, Flag } from "@phosphor-icons/react"
import { toast } from "sonner"
import { useTranslations } from "next-intl"
import type { DebateFeedItem, DebateOutcome, DebateSide, DebateStatus } from "@/lib/debates"
import { resolveDisplayName } from "@/lib/debates"
import { formatTimeAgo } from "@/lib/utils"
import { useAuthRedirect } from "@/hooks/use-auth-redirect"
import { useAuth } from "@/hooks/use-auth"
import { EMAIL_VERIFICATION_REQUIRED_MESSAGE } from "@/lib/verification"
import { ReportDialog } from "@/components/debate/report-dialog"
import { createReport } from "@/lib/moderation-api"

function StatusBadge({ status }: { status: DebateStatus }) {
  const t = useTranslations("debateCard.status")
  const styles: Record<DebateStatus, string> = {
    waiting: "bg-[var(--warning-light)] text-[var(--warning)] border-[var(--warning)]",
    active: "bg-[var(--success-light)] text-[var(--success)] border-[var(--success)]",
    pending_extension: "bg-[var(--warning-light)] text-[var(--warning)] border-[var(--warning)]",
    waiting_replacement: "bg-[var(--warning-light)] text-[var(--warning)] border-[var(--warning)]",
    finished: "bg-[var(--bg-surface-alt)] text-[var(--text-secondary)] border-[var(--border-default)]",
    expired: "bg-[var(--bg-surface-alt)] text-[var(--text-secondary)] border-[var(--border-default)]",
  }

  return (
    <span
      className={`px-2 py-0.5 text-[11px] font-medium rounded-full border ${styles[status]}`}
    >
      {t(status)}
    </span>
  )
}

function OutcomeBadge({
  outcome,
  winner_side,
  sideADisplay,
  sideBDisplay,
}: {
  outcome: DebateOutcome
  winner_side: DebateSide | null
  sideADisplay: string
  sideBDisplay: string
}) {
  if (!outcome) return null

  const t = useTranslations("debateCard.outcome")
  const winnerDisplay = winner_side === "a" ? sideADisplay : winner_side === "b" ? sideBDisplay : null

  const labels: Record<string, string> = {
    concession: winnerDisplay ? t("concession", { winner: winnerDisplay }) : t("concessionNoWinner"),
    draw: t("draw"),
    turn_limit: t("turnLimit"),
    walkover: winnerDisplay ? t("walkover", { winner: winnerDisplay }) : t("walkoverNoWinner"),
  }

  const label = labels[outcome]
  if (!label) return null

  return (
    <span className="px-2 py-0.5 text-[11px] font-medium rounded-full border border-[var(--border-default)] bg-[var(--bg-surface)] text-[var(--text-secondary)]">
      {label}
    </span>
  )
}

export default function DebateCard({
  debate,
  onToggleUpvote,
  showParticipantBadge = true,
}: {
  debate: DebateFeedItem
  onToggleUpvote?: (debateId: string) => Promise<void>
  showParticipantBadge?: boolean
}) {
  const { requireAuth } = useAuthRedirect()
  const { status, user } = useAuth()
  const t = useTranslations("debateCard")
  const tCommon = useTranslations("common")
  const timeLabel = getDebateTimestamp(debate)
  const hasUpvoted = Boolean(debate.viewer_has_upvoted)
  const isFollowing = Boolean(debate.is_following)
  const isParticipant = Boolean(debate.viewer_is_participant)
  const sideADisplay = resolveDisplayName(debate.side_a, t("sideA"))
  const sideBDisplay = resolveDisplayName(debate.side_b, t("sideB"))
  const hasChallenger = Boolean(debate.side_b.user?.username || debate.side_b.anonymous_id)
  const [reportDialogOpen, setReportDialogOpen] = useState(false)

  const handleSubmitReport = useCallback(async ({
    reason,
    details,
  }: {
    reason: "hate" | "harassment" | "spam" | "off_topic" | "illegal" | "other"
    details?: string
  }) => {
    await createReport({
      target_type: "debate",
      target_id: debate.id,
      reason,
      details,
    })
  }, [debate.id])

  const debateHref = `/debate/${debate.slug || debate.id}`

  return (
    <>
      <Link href={debateHref}>
        <div className="group w-full bg-[var(--bg-surface)] shadow-[0px_0px_0px_0.75px_var(--border-default)_inset] hover:shadow-[0px_0px_0px_0.75px_var(--border-strong)_inset] overflow-hidden transition-all duration-200 cursor-pointer">
          <div className="px-5 py-4 sm:px-6 sm:py-5 flex flex-col gap-3">
          <div className="flex items-center justify-between gap-2 flex-wrap">
            <div className="flex items-center gap-2">
              {debate.tags.slice(0, 3).map((tag) => (
                <span
                  key={tag.id}
                  className="px-2 py-0.5 text-[11px] font-medium rounded-full bg-[var(--bg-surface-alt)] text-[var(--text-secondary)] border border-[var(--border-default)]"
                >
                  {tag.name}
                </span>
              ))}
              <span className="text-[11px] text-[var(--text-secondary)] font-medium">
                {t(`timeMode.${debate.time_mode}`)}
              </span>
            </div>
            <div className="flex items-center gap-2">
              {showParticipantBadge && isParticipant && (
                <span className="px-2 py-0.5 text-[11px] font-medium rounded-full bg-[var(--success-light)] text-[var(--success)] border border-[var(--success)]">
                  {t("youreInThis")}
                </span>
              )}
              <StatusBadge status={debate.status} />
            </div>
          </div>

          <h3 className="text-[var(--text-primary)] text-[15px] sm:text-base font-semibold leading-snug font-sans group-hover:opacity-80 transition-colors">
            {debate.topic}
          </h3>

          {debate.latest_turn?.content ? (
            <p className="text-[var(--text-secondary)] text-[13px] leading-relaxed font-sans line-clamp-3">
              {debate.latest_turn.content}
            </p>
          ) : (
            debate.context && (
              <p className="text-[var(--text-secondary)] text-[13px] leading-relaxed font-sans line-clamp-3">
                {debate.context}
              </p>
            )
          )}

          <div className="flex items-center gap-3 text-[12px] text-[var(--text-secondary)] font-sans">
            <span className="font-medium text-[var(--side-a)]">{sideADisplay}</span>
            {hasChallenger ? (
              <>
                <span className="text-[var(--border-default)]">{t("vs")}</span>
                <span className="font-medium text-[var(--side-b)]">{sideBDisplay}</span>
              </>
            ) : (
              <span className="text-[var(--warning)] italic">{t("waitingForChallenger")}</span>
            )}
            {debate.turn_count > 0 && (
              <span
                className="ml-auto text-[var(--text-muted)]"
                title={t("limit", { limit: debate.turn_limit })}
              >
                {t("turn", { count: debate.turn_count })}
              </span>
            )}
          </div>

          {debate.status === "finished" && (
            <OutcomeBadge
              outcome={debate.outcome}
              winner_side={debate.winner_side}
              sideADisplay={sideADisplay}
              sideBDisplay={sideBDisplay}
            />
          )}

            <div className="flex items-center gap-4 pt-1">
            <button
              className={`flex items-center gap-1.5 text-[12px] transition-colors ${
                hasUpvoted
                  ? "text-[var(--warning)]"
                  : "text-[var(--text-secondary)] hover:text-[var(--text-primary)]"
              }`}
              onClick={(event) => {
                event.preventDefault()
                event.stopPropagation()
                if (!requireAuth()) return
                if (!user?.email_verified) {
                  toast.error(EMAIL_VERIFICATION_REQUIRED_MESSAGE)
                  return
                }
                onToggleUpvote?.(debate.id)
              }}
              aria-label={t("upvote")}
            >
              <ArrowUp size={14} />
              <span>{debate.upvote_count}</span>
            </button>

            <button
              className="flex items-center gap-1.5 text-[12px] text-[var(--text-secondary)] hover:text-[var(--text-primary)] transition-colors"
              onClick={(event) => {
                event.preventDefault()
                event.stopPropagation()
                requireAuth()
              }}
              aria-label={t("comment")}
            >
              <ChatCenteredText size={14} />
              <span>{debate.comment_count}</span>
            </button>

            {!isParticipant && (
              <button
                className="flex items-center gap-1.5 text-[12px] text-[var(--text-secondary)] hover:text-[var(--text-primary)] transition-colors"
                onClick={(event) => {
                  event.preventDefault()
                  event.stopPropagation()
                  if (!requireAuth()) return
                  setReportDialogOpen(true)
                }}
                aria-label={tCommon("report")}
              >
                <Flag size={14} />
                <span>{tCommon("report")}</span>
              </button>
            )}

            {status === "authenticated" && isFollowing && (
              <div className="flex items-center gap-1.5 text-[12px] text-[var(--text-primary)]">
                <BookmarkSimple size={14} weight="fill" />
                <span>{tCommon("following")}</span>
              </div>
            )}

            {debate.status === "active" && debate.spectator_count > 0 && (
              <div className="flex items-center gap-1.5 text-[12px] text-[var(--success)]">
                <Eye size={14} />
                <span>{t("watching", { count: debate.spectator_count })}</span>
              </div>
            )}

            <span className="ml-auto text-[11px] text-[var(--text-muted)]">
              {formatTimeAgo(timeLabel)}
            </span>
            </div>
          </div>
        </div>
      </Link>

      <ReportDialog
        open={reportDialogOpen}
        targetType="debate"
        onOpenChange={setReportDialogOpen}
        onSubmit={async (payload) => {
          try {
            await handleSubmitReport(payload)
            toast.success(t("report.submitted"))
          } catch (error) {
            if (error instanceof Error && error.message.includes("already reported")) {
              toast.error(t("report.alreadyReported"))
              return
            }
            if (error instanceof Error && error.message.includes("rate limit")) {
              toast.error(t("report.rateLimit"))
              return
            }
            if (error instanceof Error && error.message.includes("own content")) {
              toast.error(t("report.ownContent"))
              return
            }
            toast.error(error instanceof Error ? error.message : t("report.failed"))
            throw error
          }
        }}
      />
    </>
  )
}

function getDebateTimestamp(debate: DebateFeedItem) {
  if (debate.status === "finished") {
    return debate.ended_at || debate.started_at || debate.created_at
  }
  if (debate.status === "active" || debate.status === "pending_extension") {
    return debate.started_at || debate.created_at
  }
  return debate.created_at
}
