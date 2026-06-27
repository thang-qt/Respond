"use client"

import { useCallback, useEffect, useRef, useState } from "react"
import { useParams, useRouter, useSearchParams } from "next/navigation"
import { useTranslations } from "next-intl"
import { toast } from "sonner"
import { ApiError } from "@/lib/api"
import type { DebateDetail } from "@/lib/debates"
import { fetchDebate } from "@/lib/debates-api"
import { createReport, moderateAdminContent } from "@/lib/moderation-api"
import { normalizeDebateDetail } from "@/lib/debate-detail"
import { useAuthRedirect } from "@/hooks/use-auth-redirect"
import { useAuth } from "@/hooks/use-auth"
import { useDebateRealtime } from "@/hooks/use-debate-realtime"
import { useDebateComments } from "@/hooks/use-debate-comments"
import { useDebateScroll } from "@/hooks/use-debate-scroll"
import { useDebateActions } from "@/hooks/use-debate-actions"
import { useTimer } from "@/hooks/use-timer"
import { EMAIL_VERIFICATION_REQUIRED_MESSAGE } from "@/lib/verification"
import { NotFoundState } from "@/components/not-found-state"
import type { ReportTargetType } from "@/lib/moderation"
import { DebatePageView } from "@/components/debate/page/debate-page-view"

export default function DebatePage() {
  const params = useParams()
  const router = useRouter()
  const searchParams = useSearchParams()
  const debateId = params.id as string
  const t = useTranslations("debatePage")
  const tCommon = useTranslations("common")
  const [debate, setDebate] = useState<DebateDetail | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<ApiError | null>(null)
  const [collapsedTurns, setCollapsedTurns] = useState<Set<number>>(new Set())
  const [showNewTurnIndicator, setShowNewTurnIndicator] = useState(false)
  const [reportDialogOpen, setReportDialogOpen] = useState(false)
  const [reportTargetType, setReportTargetType] = useState<ReportTargetType | null>(null)
  const [reportTargetID, setReportTargetID] = useState<string | null>(null)
  const [pendingScrollTurn, setPendingScrollTurn] = useState<number | null>(null)
  const debateRef = useRef<DebateDetail | null>(null)
  const { status: authStatus, requireAuth } = useAuthRedirect()
  const { user } = useAuth()

  useEffect(() => {
    debateRef.current = debate
  }, [debate])

  const isNearPageBottom = useCallback(() => {
    const threshold = 180
    return window.innerHeight + window.scrollY >= document.documentElement.scrollHeight - threshold
  }, [])

  useDebateRealtime({
    debate,
    debateId,
    debateRef,
    isNearPageBottom,
    setDebate,
    setPendingScrollTurn,
    setShowNewTurnIndicator,
  })

  // Live countdown timer synced via WS timer.sync events.
  const { remaining: timerRemaining } = useTimer(debate?.turn_deadline ?? null)

  const requireVerified = useCallback((setError?: (message: string | null) => void) => {
    if (authStatus !== "authenticated") {
      requireAuth()
      return false
    }
    if (!user?.email_verified) {
      if (setError) {
        setError(EMAIL_VERIFICATION_REQUIRED_MESSAGE)
      } else {
        toast.error(EMAIL_VERIFICATION_REQUIRED_MESSAGE)
      }
      return false
    }
    return true
  }, [authStatus, requireAuth, user?.email_verified])

  const {
    comments,
    commentSort,
    commentTotal,
    handleDeleteComment,
    handleEditComment,
    handlePostComment,
    handlePostReply,
    handleToggleCommentVote,
    highlightCommentId,
    loadComments,
    setCommentSort,
    setCommentTotal,
    setComments,
    setHighlightCommentId,
  } = useDebateComments({ debateId, requireVerified })

  useEffect(() => {
    const activeFlag = { current: true }
    setLoading(true)
    setError(null)

    fetchDebate(debateId)
      .then((res) => {
        if (!activeFlag.current) return
        setDebate(normalizeDebateDetail(res.data))
        setCollapsedTurns(new Set())

        if (res.data.status === "finished") {
          loadComments(activeFlag)
        } else {
          setComments([])
          setCommentTotal(0)
        }
      })
      .catch((err: ApiError) => {
        if (!activeFlag.current) return
        setError(err)
        setDebate(null)
      })
      .finally(() => {
        if (!activeFlag.current) return
        setLoading(false)
      })

    return () => {
      activeFlag.current = false
    }
  }, [debateId, loadComments, authStatus])

  const openReportDialog = useCallback((targetType: ReportTargetType, targetID: string) => {
    if (authStatus !== "authenticated") {
      requireAuth()
      return
    }

    if (targetType === "debate" && debate?.viewer?.is_participant) {
      toast.error("You cannot report your own content.")
      return
    }

    if (targetType === "turn" && debate?.viewer?.is_participant && debate.viewer.side) {
      const targetTurn = debate.turns.find((turn) => turn.id === targetID)
      if (targetTurn && targetTurn.side === debate.viewer.side) {
        toast.error("You cannot report your own content.")
        return
      }
    }

    if (targetType === "comment") {
      const flattenComments = (items: typeof comments): typeof comments =>
        items.flatMap((item) => [item, ...(item.replies ?? [])])
      const targetComment = flattenComments(comments).find((comment) => comment.id === targetID)
      if (targetComment?.is_author) {
        toast.error("You cannot report your own content.")
        return
      }
    }

    setReportTargetType(targetType)
    setReportTargetID(targetID)
    setReportDialogOpen(true)
  }, [authStatus, comments, debate, requireAuth])

  const handleSubmitReport = useCallback(async ({ reason, details }: { reason: "hate" | "harassment" | "spam" | "off_topic" | "illegal" | "other"; details?: string }) => {
    if (!reportTargetID || !reportTargetType) {
      throw new Error("Select a target to report.")
    }

    await createReport({
      target_type: reportTargetType,
      target_id: reportTargetID,
      reason,
      details,
    })

    toast.success("Report submitted. Thank you.")
  }, [reportTargetID, reportTargetType])

  const handleModerateTarget = useCallback(async (targetType: ReportTargetType, targetID: string, action: "hide" | "restore") => {
    if (!(user?.role === "moderator" || user?.role === "admin")) {
      return
    }

    const note = window.prompt(`Moderator note (required to ${action} ${targetType}):`)?.trim()
    if (!note) {
      toast.error("A moderator note is required.")
      return
    }
    if (note.length > 500) {
      toast.error("Moderator note must be at most 500 characters.")
      return
    }

    try {
      await moderateAdminContent(targetType, targetID, action, note)
      toast.success(action === "hide" ? "Content hidden." : "Content restored.")

      if (targetType === "comment") {
        const activeFlag = { current: true }
        await loadComments(activeFlag)
        return
      }

      const refreshed = await fetchDebate(debateId)
      setDebate(normalizeDebateDetail(refreshed.data))
    } catch (err) {
      if (err instanceof Error) {
        toast.error(err.message)
      } else {
        toast.error("Failed to apply moderation action.")
      }
    }
  }, [debateId, loadComments, user?.role])

  const toggleTurn = useCallback((turnNumber: number) => {
    setCollapsedTurns((prev) => {
      const next = new Set(prev)
      if (next.has(turnNumber)) {
        next.delete(turnNumber)
      } else {
        next.add(turnNumber)
      }
      return next
    })
  }, [])

  const expandAll = useCallback(() => setCollapsedTurns(new Set()), [])
  const collapseAll = useCallback(() => {
    if (!debate) return
    setCollapsedTurns(new Set(debate.turns.filter((t) => !t.is_system).map((turn) => turn.turn_number)))
  }, [debate])

  const {
    extensionError,
    handleExtensionResponse,
    handleInviteUser,
    handleJoinDebate,
    handleOpenInviteModal,
    handleReplaceDebate,
    handleRespondChallenge,
    handleRevealChoice,
    handleStartRechallenge,
    handleSubmitTurn,
    handleToggleDebateVote,
    handleToggleFollow,
    inviteActionError,
    inviteModalOpen,
    invitedUsernames,
    invitingUsername,
    isJoining,
    isReplacing,
    isRespondingChallenge,
    isSubmittingExtension,
    isSubmittingReveal,
    isSubmittingTurn,
    joinError,
    replaceError,
    revealError,
    setInviteActionError,
    setInviteModalOpen,
    setInvitedUsernames,
    setInvitingUsername,
    turnError,
  } = useDebateActions({ debate, debateId, requireVerified, router, setDebate })

  if (loading) {
    return <div className="min-h-screen bg-[var(--bg-primary)]" />
  }

  if (error?.status === 403 && error.code === "DEBATE_HIDDEN_BY_BLOCK") {
    return (
      <NotFoundState
        title={t("errors.hiddenByBlockTitle")}
        description={t("errors.hiddenByBlockDesc")}
        actions={[
          { href: "/settings", label: t("errors.manageBlocked"), primary: true },
          { href: "/explore", label: t("errors.exploreDebates") },
        ]}
      />
    )
  }

  if (error?.status === 404 || !debate) {
    return (
      <NotFoundState
        title={t("errors.notFoundTitle")}
        description={t("errors.notFoundDesc")}
        actions={[
          { href: "/", label: tCommon("backToHome"), primary: true },
          { href: "/explore", label: t("errors.exploreDebates") },
        ]}
      />
    )
  }

  return (
    <DebatePageView
      canModerate={user?.role === "moderator" || user?.role === "admin"}
      collapsedTurns={collapsedTurns}
      commentSort={commentSort}
      commentTotal={commentTotal}
      comments={comments}
      debate={debate}
      extensionError={extensionError}
      handleDeleteComment={handleDeleteComment}
      handleEditComment={handleEditComment}
      handleExtensionResponse={handleExtensionResponse}
      handleInviteUser={handleInviteUser}
      handleJoinDebate={handleJoinDebate}
      handleModerateTarget={handleModerateTarget}
      handleOpenInviteModal={handleOpenInviteModal}
      handlePostComment={handlePostComment}
      handlePostReply={handlePostReply}
      handleReplaceDebate={handleReplaceDebate}
      handleRespondChallenge={handleRespondChallenge}
      handleRevealChoice={handleRevealChoice}
      handleStartRechallenge={handleStartRechallenge}
      handleSubmitReport={handleSubmitReport}
      handleSubmitTurn={handleSubmitTurn}
      handleToggleCommentVote={handleToggleCommentVote}
      handleToggleDebateVote={handleToggleDebateVote}
      handleToggleFollow={handleToggleFollow}
      highlightCommentId={highlightCommentId}
      collapseAll={collapseAll}
      expandAll={expandAll}
      inviteActionError={inviteActionError}
      inviteModalOpen={inviteModalOpen}
      invitedUsernames={invitedUsernames}
      invitingUsername={invitingUsername}
      isJoining={isJoining}
      isReplacing={isReplacing}
      isRespondingChallenge={isRespondingChallenge}
      isSubmittingExtension={isSubmittingExtension}
      isSubmittingReveal={isSubmittingReveal}
      isSubmittingTurn={isSubmittingTurn}
      joinError={joinError}
      openReportDialog={openReportDialog}
      replaceError={replaceError}
      revealError={revealError}
      reportDialogOpen={reportDialogOpen}
      reportTargetType={reportTargetType}
      setCommentSort={setCommentSort}
      setInviteActionError={setInviteActionError}
      setInviteModalOpen={setInviteModalOpen}
      setPendingScrollTurn={setPendingScrollTurn}
      setReportDialogOpen={setReportDialogOpen}
      setReportTargetID={setReportTargetID}
      setReportTargetType={setReportTargetType}
      setShowNewTurnIndicator={setShowNewTurnIndicator}
      setDebate={setDebate}
      showNewTurnIndicator={showNewTurnIndicator}
      t={t}
      tCommon={tCommon}
      timerRemaining={timerRemaining}
      toggleTurn={toggleTurn}
      turnError={turnError}
      user={user}
    />
  )
}
