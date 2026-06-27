"use client"

import Link from "next/link"
import { useState } from "react"
import { useTranslations } from "next-intl"
import { ArrowUp, Flag } from "@phosphor-icons/react"
import { MoreHorizontal } from "lucide-react"
import type { DebateComment } from "@/lib/debates"
import { formatDate } from "@/lib/utils"
import { useAuthRedirect } from "@/hooks/use-auth-redirect"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"

interface CommentCardProps {
  comment: DebateComment
  isLocked: boolean
  onPostReply: (parentId: string, content: string) => Promise<void>
  onToggleUpvote: (commentId: string) => Promise<void>
  onEditComment: (commentId: string, content: string) => Promise<void>
  onDeleteComment: (commentId: string) => Promise<void>
  onReportComment: (commentId: string) => void
  canModerateComments?: boolean
  onModerateComment?: (commentId: string, action: "hide" | "restore") => void
  highlightCommentId: string | null
  readOnly?: boolean
}

export function CommentCard({
  comment,
  isLocked,
  onPostReply,
  onToggleUpvote,
  onEditComment,
  onDeleteComment,
  onReportComment,
  canModerateComments,
  onModerateComment,
  highlightCommentId,
  readOnly,
}: CommentCardProps) {
  const t = useTranslations("commentCard")
  const tCommon = useTranslations("common")
  const [showReplies, setShowReplies] = useState(true)
  const [showReplyInput, setShowReplyInput] = useState(false)
  const [replyContent, setReplyContent] = useState("")
  const [isReplying, setIsReplying] = useState(false)
  const [replyError, setReplyError] = useState<string | null>(null)
  const [isVoting, setIsVoting] = useState(false)
  const [isEditing, setIsEditing] = useState(false)
  const [editContent, setEditContent] = useState(comment.content)
  const [isSaving, setIsSaving] = useState(false)
  const [isDeleting, setIsDeleting] = useState(false)
  const [editError, setEditError] = useState<string | null>(null)
  const { requireAuth, status } = useAuthRedirect()
  const replies = comment.replies ?? []
  const isReply = Boolean(comment.parent_id)
  const canReply = status === "authenticated" && !readOnly && !isLocked && !isReply && !comment.hidden
  const canUpvote = status === "authenticated" && !readOnly && !comment.hidden
  const upvoteActive = comment.viewer_has_upvoted
  const isDeleted = comment.content === "[deleted]"
  const editWindowMs = 5 * 60 * 1000
  const canEdit = comment.is_author && !readOnly && !isDeleted && !isLocked && !comment.hidden &&
    Date.now() - new Date(comment.created_at).getTime() <= editWindowMs
  const canDelete = comment.is_author && !readOnly && !isDeleted && !isLocked && !comment.hidden

  const handleToggleUpvote = async () => {
    if (!canUpvote) {
      requireAuth()
      return
    }
    if (isVoting) return
    setIsVoting(true)
    try {
      await onToggleUpvote(comment.id)
    } finally {
      setIsVoting(false)
    }
  }

  const handleReplySubmit = async () => {
    if (!canReply || isReplying) return
    const trimmed = replyContent.trim()
    if (!trimmed) return
    setIsReplying(true)
    setReplyError(null)
    try {
      await onPostReply(comment.id, trimmed)
      setReplyContent("")
      setShowReplyInput(false)
    } catch (err) {
      const message = err instanceof Error ? err.message : tCommon("tryAgain")
      setReplyError(message)
    } finally {
      setIsReplying(false)
    }
  }

  const handleEditSubmit = async () => {
    if (!canEdit || isSaving) return
    const trimmed = editContent.trim()
    if (!trimmed) return
    setIsSaving(true)
    setEditError(null)
    try {
      await onEditComment(comment.id, trimmed)
      setIsEditing(false)
    } catch (err) {
      const message = err instanceof Error ? err.message : tCommon("tryAgain")
      setEditError(message)
    } finally {
      setIsSaving(false)
    }
  }

  const handleDelete = async () => {
    if (!canDelete || isDeleting) return
    if (!window.confirm(t("deleteConfirmation"))) return
    setIsDeleting(true)
    try {
      await onDeleteComment(comment.id)
    } finally {
      setIsDeleting(false)
    }
  }

  const isHighlighted = highlightCommentId === comment.id
  const username = comment.user.username || null
  const authorDisplayName = username || comment.debater_anonymous_id || t("anonymousDebater")

  return (
    <div id={`comment-${comment.id}`} className="flex flex-col">
      <div
        className={`px-4 sm:px-5 py-4 flex flex-col gap-2 ${
          comment.is_debater
            ? "bg-[var(--warning-light)] border-l-2 border-l-[var(--warning)]"
            : ""
        } ${isHighlighted ? "comment-highlight" : ""}`}
      >
        <div className="flex items-center gap-2">
          <span className="text-[var(--text-primary)] text-sm font-semibold font-sans">
            {username ? (
              <Link href={`/profile/${encodeURIComponent(username)}`} className="hover:underline underline-offset-2">
                {authorDisplayName}
              </Link>
            ) : (
              authorDisplayName
            )}
          </span>
          {comment.is_reflection && comment.is_debater && (
            <span className="px-1.5 py-0.5 text-[10px] font-medium rounded-full bg-[var(--warning-light)] text-[var(--warning)] border border-[var(--warning)]">
              {t("postMatchReflection")}
            </span>
          )}
          <span className="text-[var(--text-muted)] text-xs font-sans ml-auto">
            {formatDate(comment.created_at)}
          </span>
        </div>
        {isEditing ? (
          <div className="flex flex-col gap-2">
            <textarea
              value={editContent}
              onChange={(event) => setEditContent(event.target.value)}
              className="w-full min-h-[70px] resize-none bg-[var(--bg-surface-alt)] rounded border border-[var(--border-subtle)] px-3 py-2 text-sm text-[var(--text-primary)] font-sans placeholder:text-[var(--text-muted)] focus:outline-none focus:ring-2 focus:ring-[var(--border-strong)]"
              disabled={isSaving}
            />
            <div className="flex items-center justify-end gap-2">
              <button
                className="text-[11px] font-medium text-[var(--text-secondary)] hover:text-[var(--text-primary)] transition-colors"
                onClick={() => {
                  setIsEditing(false)
                  setEditContent(comment.content)
                  setEditError(null)
                }}
                disabled={isSaving}
              >
                {tCommon("cancel")}
              </button>
              <button
                className="h-7 px-3 bg-[var(--text-primary)] text-[var(--bg-primary)] text-[11px] font-medium rounded-full font-sans hover:opacity-90 transition-colors disabled:opacity-60"
                onClick={handleEditSubmit}
                disabled={isSaving || editContent.trim().length === 0}
              >
                {isSaving ? t("saving") : tCommon("save")}
              </button>
            </div>
            {editError && (
              <div className="text-xs text-[var(--error)] font-sans">
                {editError}
              </div>
            )}
          </div>
        ) : comment.hidden ? (
          canModerateComments || comment.is_author ? (
            <div className="flex flex-col gap-2">
              <div className="rounded-md border border-[var(--warning)] bg-[var(--warning-light)] px-3 py-2 text-[12px] text-[var(--warning)] font-sans">
                {canModerateComments
                  ? t("hiddenModOnly")
                  : t("hiddenAuthorAndModOnly")}
              </div>
              <p className="text-[var(--text-secondary)] text-sm leading-relaxed font-sans">
                {comment.content}
              </p>
            </div>
          ) : (
            <div className="rounded-md border border-[var(--border-default)] bg-[var(--bg-surface)] px-3 py-2 text-[13px] text-[var(--text-secondary)] font-sans">
              {t("commentDeletedText")}
            </div>
          )
        ) : (
          <p className="text-[var(--text-secondary)] text-sm leading-relaxed font-sans">
            {comment.content}
          </p>
        )}
        <div className="flex items-center gap-3 mt-1">
          <button
            className={`flex items-center gap-1 transition-colors ${
              upvoteActive
                ? "text-[var(--warning)]"
                : "text-[var(--text-secondary)] hover:text-[var(--text-primary)]"
            }`}
            onClick={handleToggleUpvote}
            aria-label={t("ariaUpvote")}
            disabled={isVoting}
          >
            <ArrowUp size={14} />
            <span className="text-xs font-medium">{comment.upvote_count}</span>
          </button>
          {canReply && (
            <button
              onClick={() => setShowReplyInput((prev) => !prev)}
              className="text-xs text-[var(--text-secondary)] hover:text-[var(--text-primary)] font-medium transition-colors"
            >
              {showReplyInput ? tCommon("cancel") : t("reply")}
            </button>
          )}
          {replies.length > 0 && (
            <button
              onClick={() => setShowReplies(!showReplies)}
              className="text-xs text-[var(--text-secondary)] hover:text-[var(--text-primary)] font-medium transition-colors"
            >
              {showReplies ? t("hide") : t("show")} {t("repliesCount", { count: replies.length })}
            </button>
          )}
          {canEdit && !isEditing && (
            <button
              onClick={() => setIsEditing(true)}
              className="text-xs text-[var(--text-secondary)] hover:text-[var(--text-primary)] font-medium transition-colors"
            >
              {t("edit")}
            </button>
          )}
          {canDelete && (
            <button
              onClick={handleDelete}
              className="text-xs text-[var(--text-secondary)] hover:text-[var(--error)] font-medium transition-colors"
            >
              {isDeleting ? t("deleting") : t("delete")}
            </button>
          )}
          {!readOnly && !comment.hidden && !comment.is_author && status === "authenticated" && (
            <button
              onClick={() => onReportComment(comment.id)}
              className="text-xs text-[var(--text-secondary)] hover:text-[var(--error)] font-medium transition-colors inline-flex items-center gap-1"
            >
              <Flag size={12} />
              {t("report")}
            </button>
          )}
          {canModerateComments && (
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <button
                  className="text-xs text-[var(--text-secondary)] hover:text-[var(--text-primary)] font-medium transition-colors inline-flex items-center gap-1"
                  aria-label={t("ariaModerate")}
                >
                  <MoreHorizontal size={12} />
                  {t("moderate")}
                </button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end" className="w-44">
                <DropdownMenuItem onClick={() => onModerateComment?.(comment.id, "hide")} disabled={comment.hidden}>
                  {t("hideComment")}
                </DropdownMenuItem>
                <DropdownMenuItem onClick={() => onModerateComment?.(comment.id, "restore")} disabled={!comment.hidden}>
                  {t("restoreComment")}
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          )}
        </div>
        {showReplyInput && canReply && (
          <div className="mt-2 flex flex-col gap-2">
            <textarea
              value={replyContent}
              onChange={(event) => setReplyContent(event.target.value)}
              placeholder={t("replyPlaceholder")}
              className="w-full min-h-[70px] resize-none bg-[var(--bg-surface-alt)] rounded border border-[var(--border-subtle)] px-3 py-2 text-sm text-[var(--text-primary)] font-sans placeholder:text-[var(--text-muted)] focus:outline-none focus:ring-2 focus:ring-[var(--border-strong)]"
              disabled={isReplying}
            />
            <div className="flex justify-end">
              <button
                className="h-7 px-3 bg-[var(--text-primary)] text-[var(--bg-primary)] text-[11px] font-medium rounded-full font-sans hover:opacity-90 transition-colors disabled:opacity-60"
                onClick={handleReplySubmit}
                disabled={isReplying || replyContent.trim().length === 0}
              >
                {isReplying ? t("posting") : t("reply")}
              </button>
            </div>
            {replyError && (
              <div className="text-xs text-[var(--error)] font-sans">
                {replyError}
              </div>
            )}
          </div>
        )}
      </div>

      {showReplies && replies.length > 0 && (
        <div className="ml-6 sm:ml-8 border-l border-[var(--border-subtle)]">
          {replies.map((reply) => (
            <CommentCard
              key={reply.id}
              comment={reply}
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
      )}
    </div>
  )
}
