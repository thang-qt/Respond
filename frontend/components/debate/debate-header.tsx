"use client"

import Link from "next/link"
import { useTranslations } from "next-intl"
import { ArrowLeft, ArrowUp, ChatCenteredText, Eye, BookmarkSimple, Flag } from "@phosphor-icons/react"
import { MoreHorizontal } from "lucide-react"
import { toast } from "sonner"
import type { DebateDetail, DebateStatus } from "@/lib/debates"
import { resolveUsername, resolveDisplayName } from "@/lib/debates"
import { OutcomeBanner } from "@/components/debate/outcome-banner"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { useAuthRedirect } from "@/hooks/use-auth-redirect"
import { useAuth } from "@/hooks/use-auth"
import { EMAIL_VERIFICATION_REQUIRED_MESSAGE } from "@/lib/verification"

interface DebateHeaderProps {
  debate: DebateDetail
  turnProgress: number
  timerRemaining: string | null
  onToggleUpvote: () => Promise<void>
  onToggleFollow: () => Promise<void>
  onReportDebate?: () => void
  canModerateDebate?: boolean
  onModerateDebate?: (action: "hide" | "restore") => void
  isReadOnly?: boolean
}

export function DebateHeader({ debate, turnProgress, timerRemaining, onToggleUpvote, onToggleFollow, onReportDebate, canModerateDebate, onModerateDebate, isReadOnly }: DebateHeaderProps) {
  const { requireAuth } = useAuthRedirect()
  const t = useTranslations("debate.header")
  const tDebate = useTranslations("debate")
  const { user } = useAuth()
  const hasUpvoted = Boolean(debate.viewer?.has_upvoted)
  const isFollowing = Boolean(debate.viewer?.is_following)
  const sideAUsername = resolveUsername(debate.side_a)
  const sideBUsername = resolveUsername(debate.side_b)
  const statusClasses: Record<DebateStatus, string> = {
    waiting: "bg-[var(--warning-light)] text-[var(--warning)] border-[var(--warning)]",
    active: "bg-[var(--success-light)] text-[var(--success)] border-[var(--success)]",
    pending_extension: "bg-[var(--warning-light)] text-[var(--warning)] border-[var(--warning)]",
    waiting_replacement: "bg-[var(--warning-light)] text-[var(--warning)] border-[var(--warning)]",
    finished: "bg-[var(--bg-surface-alt)] text-[var(--text-secondary)] border-[var(--border-default)]",
    expired: "bg-[var(--bg-surface-alt)] text-[var(--text-secondary)] border-[var(--border-default)]",
  }

  return (
    <div className="border-b border-[var(--border-subtle)]">
      <div className="max-w-[820px] mx-auto px-3 sm:px-6 pt-4 sm:pt-6 pb-5 flex flex-col gap-3 sm:gap-4">
        <Link
          href="/"
          className="inline-flex items-center gap-1.5 text-[var(--text-secondary)] text-sm font-sans hover:text-[var(--text-primary)] transition-colors w-fit"
        >
          <ArrowLeft size={16} />
          {t("back")}
        </Link>

        <div className="flex items-center gap-2 flex-wrap">
          {debate.tags.map((tag) => (
            <span
              key={tag.id}
              className="px-2 py-0.5 text-[11px] font-medium rounded-full bg-[var(--bg-surface-alt)] text-[var(--text-secondary)] border border-[var(--border-default)]"
            >
              {tag.name}
            </span>
          ))}
          <span className="px-2 py-0.5 text-[11px] font-medium rounded-full bg-[var(--bg-surface-alt)] text-[var(--text-secondary)] border border-[var(--border-default)]">
            {tDebate(`timeMode.${debate.time_mode}`)}
          </span>
          <span className={`px-2 py-0.5 text-[11px] font-medium rounded-full border ${statusClasses[debate.status]}`}>
            {tDebate(`status.${debate.status}`)}
          </span>
          {debate.viewer?.is_participant && (
            <span className="px-2 py-0.5 text-[11px] font-medium rounded-full bg-[var(--success-light)] text-[var(--success)] border border-[var(--success)]">
              {t("participant")}
            </span>
          )}
        </div>

        <h1 className="text-[var(--text-primary)] text-xl sm:text-2xl md:text-[28px] font-semibold leading-tight font-sans tracking-tight">
          {debate.topic}
        </h1>

        {debate.context && (
          <div className="px-4 py-3 bg-[var(--bg-surface)] border border-[var(--border-default)] rounded-lg">
            <div className="text-[var(--text-muted)] text-[11px] font-semibold font-sans uppercase tracking-wider mb-1">
              {t("contextTitle")}
            </div>
            <p className="text-[var(--text-secondary)] text-sm leading-relaxed font-sans whitespace-pre-line">
              {debate.context}
            </p>
          </div>
        )}

        <div className="flex items-center gap-4 sm:gap-6">
          <div className="flex items-center gap-2">
            <div className="w-6 h-6 rounded-full bg-[var(--side-a)] text-white text-[10px] font-bold flex items-center justify-center font-mono">A</div>
            {sideAUsername ? (
              <Link
                href={`/profile/${encodeURIComponent(sideAUsername)}`}
                className="text-[var(--side-a)] text-sm font-semibold font-sans hover:underline underline-offset-2"
              >
                {sideAUsername}
              </Link>
            ) : (
              <span className="text-[var(--side-a)] text-sm font-semibold font-sans">
                {resolveDisplayName(debate.side_a, tDebate("common.sideA"))}
              </span>
            )}
            {debate.viewer?.side === "a" && (
              <span className="text-[12px] text-[var(--text-muted)] font-sans">{tDebate("common.you")}</span>
            )}
          </div>
          <span className="text-[var(--border-strong)] text-sm font-sans">vs</span>
          {debate.side_b.anonymous_id ? (
            <div className="flex items-center gap-2">
              <div className="w-6 h-6 rounded-full bg-[var(--side-b)] text-white text-[10px] font-bold flex items-center justify-center font-mono">B</div>
              {sideBUsername ? (
                <Link
                  href={`/profile/${encodeURIComponent(sideBUsername)}`}
                  className="text-[var(--side-b)] text-sm font-semibold font-sans hover:underline underline-offset-2"
                >
                  {sideBUsername}
                </Link>
              ) : (
                <span className="text-[var(--side-b)] text-sm font-semibold font-sans">
                  {resolveDisplayName(debate.side_b, tDebate("common.sideB"))}
                </span>
              )}
              {debate.viewer?.side === "b" && (
                <span className="text-[12px] text-[var(--text-muted)] font-sans">{tDebate("common.you")}</span>
              )}
            </div>
          ) : (
            <span className="text-[var(--warning)] text-sm italic font-sans">{t("waitingChallenger")}</span>
          )}
        </div>

        <div className="flex items-center gap-3">
          <div className="flex-1 h-1.5 bg-[var(--border-subtle)] rounded-full overflow-hidden">
            <div
              className="h-full bg-[var(--border-strong)] rounded-full transition-all duration-300"
              style={{ width: `${turnProgress}%` }}
            />
          </div>
          <span
            className="text-[var(--text-secondary)] text-xs font-semibold font-sans whitespace-nowrap"
            title={t("limit", { limit: debate.turn_limit })}
          >
            {t("turn", { turn: debate.turns.filter((turn) => !turn.is_system).length })}
          </span>
          {timerRemaining && (debate.status === "active" || debate.status === "pending_extension") && (
            <span
              className={`text-xs font-semibold font-sans whitespace-nowrap ${timerRemaining === "Expired" ? "text-[var(--error)]" : "text-[var(--text-muted)]"
                }`}
              title={t("timeRemaining")}
            >
              ⏱ {timerRemaining}
            </span>
          )}
        </div>

        <div className="flex items-center gap-5 text-[13px] text-[var(--text-secondary)] font-sans">
          {!isReadOnly && (
            <button
              className={`flex items-center gap-1.5 transition-colors ${hasUpvoted ? "text-[var(--warning)]" : "hover:text-[var(--text-primary)]"
                }`}
              onClick={() => {
                if (!requireAuth()) return
                if (!user?.email_verified) {
                  toast.error(EMAIL_VERIFICATION_REQUIRED_MESSAGE)
                  return
                }
                onToggleUpvote()
              }}
              aria-label={t("upvote")}
            >
              <ArrowUp size={15} />
              <span className="font-medium">{debate.upvote_count}</span>
            </button>
          )}
          <div className={`flex items-center gap-1.5 ${isReadOnly ? "" : "hover:text-[var(--text-primary)] transition-colors"}`}>
            <ChatCenteredText size={15} />
            <span className="font-medium">{debate.comment_count}</span>
          </div>
          {!isReadOnly && !debate.viewer?.is_participant && (
            <button
              className="flex items-center gap-1.5 hover:text-[var(--text-primary)] transition-colors"
              onClick={() => {
                if (!requireAuth()) return
                onReportDebate?.()
              }}
              aria-label={t("reportDebate")}
            >
              <Flag size={15} />
              <span className="font-medium">{t("report")}</span>
            </button>
          )}
          {canModerateDebate && (
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <button
                  className="flex items-center gap-1.5 hover:text-[var(--text-primary)] transition-colors"
                  aria-label={t("moderateDebate")}
                >
                  <MoreHorizontal size={14} />
                  <span className="font-medium">{t("moderate")}</span>
                </button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end" className="w-44">
                <DropdownMenuItem onClick={() => onModerateDebate?.("hide")} disabled={debate.hidden}>
                  {t("hideDebate")}
                </DropdownMenuItem>
                <DropdownMenuItem onClick={() => onModerateDebate?.("restore")} disabled={!debate.hidden}>
                  {t("restoreDebate")}
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          )}
          {!isReadOnly && debate.viewer && (
            <button
              className={`flex items-center gap-1.5 transition-colors ${isFollowing ? "text-[var(--text-primary)]" : "hover:text-[var(--text-primary)]"
                }`}
              onClick={() => {
                if (!requireAuth()) return
                if (!user?.email_verified) {
                  toast.error(EMAIL_VERIFICATION_REQUIRED_MESSAGE)
                  return
                }
                onToggleFollow()
              }}
              aria-label={isFollowing ? t("unfollowDebate") : t("followDebate")}
            >
              <BookmarkSimple size={15} weight={isFollowing ? "fill" : "regular"} />
              <span className="font-medium">{isFollowing ? t("following") : t("follow")}</span>
            </button>
          )}
          {debate.status === "active" && debate.spectator_count > 0 && (
            <div className="flex items-center gap-1.5 text-[var(--success)]">
              <Eye size={15} />
              <span className="font-medium">{t("watching", { count: debate.spectator_count })}</span>
            </div>
          )}
        </div>

        <OutcomeBanner debate={debate} />
      </div>
    </div>
  )
}
