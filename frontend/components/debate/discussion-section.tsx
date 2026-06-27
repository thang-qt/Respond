"use client"

import { useState } from "react"
import Link from "next/link"
import { useTranslations } from "next-intl"
import type { DebateComment } from "@/lib/debates"
import { CommentCard } from "@/components/debate/comment-card"
import { useAuthRedirect } from "@/hooks/use-auth-redirect"
import { useAuth } from "@/hooks/use-auth"
import { EMAIL_VERIFICATION_REQUIRED_MESSAGE } from "@/lib/verification"

interface DiscussionSectionProps {
  comments: DebateComment[]
  totalCount: number
  sort: "newest" | "top"
  onSortChange: (value: "newest" | "top") => void
  onPostComment: (content: string, isReflection: boolean) => Promise<void>
  onPostReply: (parentId: string, content: string) => Promise<void>
  onToggleUpvote: (commentId: string) => Promise<void>
  onEditComment: (commentId: string, content: string) => Promise<void>
  onDeleteComment: (commentId: string) => Promise<void>
  onReportComment: (commentId: string) => void
  canModerateComments?: boolean
  onModerateComment?: (commentId: string, action: "hide" | "restore") => void
  highlightCommentId: string | null
  isLocked: boolean
  canPostReflection: boolean
  hasPostedReflection: boolean
  readOnly?: boolean
}

export function DiscussionSection({
  comments,
  totalCount,
  sort,
  onSortChange,
  onPostComment,
  onPostReply,
  onToggleUpvote,
  onEditComment,
  onDeleteComment,
  onReportComment,
  canModerateComments,
  onModerateComment,
  highlightCommentId,
  isLocked,
  canPostReflection,
  hasPostedReflection,
  readOnly,
}: DiscussionSectionProps) {
  const { requireAuth } = useAuthRedirect()
  const t = useTranslations("debate.discussion")
  const tCommon = useTranslations("debate.common")
  const { status, user } = useAuth()
  const [content, setContent] = useState("")
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [asReflection, setAsReflection] = useState(false)

  const handleSubmit = async () => {
    if (isSubmitting) return
    const trimmed = content.trim()
    if (!trimmed) return
    setIsSubmitting(true)
    setError(null)
    try {
      await onPostComment(trimmed, asReflection)
      setContent("")
      setAsReflection(false)
    } catch (err) {
      const message = err instanceof Error ? err.message : tCommon("somethingWrong")
      setError(message)
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <div id="discussion-section" className="mt-8">
      <div className="flex items-center gap-3 mb-4">
        <h2 className="text-[var(--text-primary)] text-base font-semibold font-sans">
          {t("title", { count: totalCount })}
        </h2>
        <div className="flex-1 h-px bg-[var(--border-subtle)]" />
        <div className="flex items-center gap-1">
          {(["newest", "top"] as const).map((option) => (
            <button
              key={option}
              onClick={() => onSortChange(option)}
              className={`px-2.5 py-1 text-[11px] font-medium rounded-md transition-colors font-sans capitalize ${
                sort === option
                  ? "bg-[var(--bg-surface)] shadow-[0px_0px_0px_0.75px_var(--border-default)_inset] text-[var(--text-primary)]"
                  : "text-[var(--text-muted)] hover:text-[var(--text-secondary)]"
              }`}
            >
              {t(`sort.${option}`)}
            </button>
          ))}
        </div>
      </div>

      {comments.length > 0 ? (
        <div className="flex flex-col divide-y divide-[var(--border-subtle)]">
          {comments.map((comment) => (
            <CommentCard
              key={comment.id}
              comment={comment}
              isLocked={isLocked}
              onPostReply={onPostReply}
              onToggleUpvote={onToggleUpvote}
              onEditComment={onEditComment}
              onDeleteComment={onDeleteComment}
              onReportComment={onReportComment}
              canModerateComments={canModerateComments}
              onModerateComment={onModerateComment}
              highlightCommentId={highlightCommentId}
              readOnly={readOnly}
            />
          ))}
        </div>
      ) : (
        <div className="py-12 text-center">
          <div className="text-[var(--text-secondary)] text-base font-sans mb-2">
            {t("emptyTitle")}
          </div>
          <div className="text-[var(--text-muted)] text-sm font-sans">
            {t("emptyDescription")}
          </div>
        </div>
      )}

      <div className="mt-6 p-4 bg-[var(--bg-surface)] border border-[var(--border-default)] rounded-lg">
        {isLocked ? (
          <div>
            <div className="text-[var(--text-primary)] text-sm font-semibold font-sans">
              {t("lockedTitle")}
            </div>
            <div className="text-[var(--text-muted)] text-xs font-sans mt-1">
              {t("lockedDescription")}
            </div>
          </div>
        ) : readOnly ? (
          <div>
            <div className="text-[var(--text-primary)] text-sm font-semibold font-sans">
              {t("readOnlyTitle")}
            </div>
            <div className="text-[var(--text-muted)] text-xs font-sans mt-1">
              {t("readOnlyDescription")}
            </div>
          </div>
        ) : status === "authenticated" && user?.email_verified ? (
          <>
            <textarea
              value={content}
              onChange={(event) => setContent(event.target.value)}
              placeholder={t("placeholder")}
              className="w-full min-h-[80px] resize-none bg-[var(--bg-surface-alt)] rounded border border-[var(--border-subtle)] px-3 py-2 text-sm text-[var(--text-primary)] font-sans placeholder:text-[var(--text-muted)] focus:outline-none focus:ring-2 focus:ring-[var(--border-strong)]"
              disabled={isSubmitting}
            />
            <div className="flex justify-end mt-3">
              <div className="flex items-center gap-3">
                {canPostReflection && (
                  <label className="flex items-center gap-2 text-xs text-[var(--text-muted)] font-sans">
                    <input
                      id="post-reflection"
                      type="checkbox"
                      checked={asReflection}
                      disabled={hasPostedReflection}
                      onChange={(event) => setAsReflection(event.target.checked)}
                    />
                    <span>{t("postAsReflection")}</span>
                    {hasPostedReflection && (
                      <span className="text-[10px] text-[var(--text-secondary)]">
                        {t("alreadyPosted")}
                      </span>
                    )}
                  </label>
                )}
                <button
                  className="h-8 px-4 bg-[var(--text-primary)] text-[var(--bg-primary)] text-xs font-medium rounded-full font-sans hover:opacity-90 transition-colors disabled:opacity-60"
                  onClick={handleSubmit}
                  disabled={isSubmitting || content.trim().length === 0}
                >
                  {isSubmitting ? t("posting") : t("post")}
                </button>
              </div>
            </div>
            {error && (
              <div className="mt-2 text-xs text-[var(--error)] font-sans">
                {error}
              </div>
            )}
          </>
        ) : status === "authenticated" ? (
          <div className="flex items-center justify-between gap-3">
            <div>
              <div className="text-[var(--text-primary)] text-sm font-semibold font-sans">
                {tCommon("emailVerificationRequired")}
              </div>
              <div className="text-[var(--text-muted)] text-xs font-sans">
                {EMAIL_VERIFICATION_REQUIRED_MESSAGE}
              </div>
            </div>
            <Link
              href="/settings"
              className="h-8 px-4 inline-flex items-center bg-[var(--text-primary)] text-[var(--bg-primary)] text-xs font-medium rounded-full font-sans hover:opacity-90 transition-colors"
            >
              {tCommon("openSettings")}
            </Link>
          </div>
        ) : (
          <div className="flex items-center justify-between gap-3">
            <div>
              <div className="text-[var(--text-primary)] text-sm font-semibold font-sans">
                {t("signInTitle")}
              </div>
              <div className="text-[var(--text-muted)] text-xs font-sans">
                {t("signInDescription")}
              </div>
            </div>
            <button
              className="h-8 px-4 bg-[var(--text-primary)] text-[var(--bg-primary)] text-xs font-medium rounded-full font-sans hover:opacity-90 transition-colors"
              onClick={() => requireAuth()}
            >
              {t("signIn")}
            </button>
          </div>
        )}
      </div>
    </div>
  )
}
