"use client"

import Link from "next/link"
import type { Dispatch, SetStateAction } from "react"
import type { DebateComment, DebateDetail, DebateEvent, DebateTurn } from "@/lib/debates"
import { resolveDisplayName, resolveUsername } from "@/lib/debates"
import type { ReportTargetType } from "@/lib/moderation"
import { DebateHeader } from "@/components/debate/debate-header"
import { TurnInput } from "@/components/debate/turn-input"
import { DebateActions } from "@/components/debate/debate-actions"
import { DiscussionSection } from "@/components/debate/discussion-section"
import { RevealPrompt } from "@/components/debate/reveal-prompt"
import { DebateTimelinePanel } from "@/components/debate/page/debate-timeline-panel"
import { DebatePageOverlays } from "@/components/debate/page/debate-page-overlays"

type Props = { [key: string]: any } & { debate: DebateDetail; comments: DebateComment[]; setDebate: Dispatch<SetStateAction<DebateDetail | null>> }

function hasViewerReflection(comments: DebateComment[]): boolean {
  return comments.some((comment) => {
    if (comment.is_reflection && comment.is_author) return true
    if (comment.replies && comment.replies.length > 0) return hasViewerReflection(comment.replies)
    return false
  })
}

export function DebatePageView(props: Props) {
  const {
    collapsedTurns, commentSort, commentTotal, comments, debate, extensionError, handleDeleteComment, handleEditComment, handleExtensionResponse, handleInviteUser, handleJoinDebate, handleModerateTarget, handleOpenInviteModal, handlePostComment, handlePostReply, handleReplaceDebate, handleRespondChallenge, handleRevealChoice, handleStartRechallenge, handleSubmitReport, handleSubmitTurn, handleToggleCommentVote, handleToggleDebateVote, handleToggleFollow, highlightCommentId, inviteActionError, inviteModalOpen, invitedUsernames, invitingUsername, collapseAll, expandAll, isJoining, isReplacing, isRespondingChallenge, isSubmittingExtension, isSubmittingReveal, isSubmittingTurn, joinError, openReportDialog, pendingScrollTurn, replaceError, revealError, reportDialogOpen, reportTargetType, setCollapsedTurns, setCommentSort, setInviteActionError, setInviteModalOpen, setPendingScrollTurn, setReportDialogOpen, setReportTargetID, setReportTargetType, setShowNewTurnIndicator, setDebate, showNewTurnIndicator, t, tCommon, timerRemaining, toggleTurn, turnError, user
  } = props
  const sideAUsername = resolveUsername(debate.side_a)
  const sideBUsername = resolveUsername(debate.side_b)
  const sideADisplay = resolveDisplayName(debate.side_a, t("common.sideA"))
  const sideBDisplay = resolveDisplayName(debate.side_b, t("common.sideB"))
  const argumentTurns = debate.turns.filter((turn) => !turn.is_system)

  const timelineItems: Array<
    | { kind: "turn"; id: string; turn: DebateTurn }
    | { kind: "event"; id: string; event: DebateEvent }
  > = (() => {
    if (debate.timeline && debate.timeline.length > 0) {
      return debate.timeline.reduce<Array<{ kind: "turn"; id: string; turn: DebateTurn } | { kind: "event"; id: string; event: DebateEvent }>>((acc, item) => {
        if (item.type === "turn" && item.turn && !item.turn.is_system) {
          acc.push({ kind: "turn", id: item.turn.id, turn: item.turn })
        } else if (item.type === "event" && item.event) {
          acc.push({ kind: "event", id: item.event.id, event: item.event })
        }
        return acc
      }, [])
    }

    return debate.turns.map((turn) => {
      if (turn.is_system) {
        return {
          kind: "event" as const,
          id: `legacy-${turn.id}`,
          event: {
            id: turn.id,
            event_type: "legacy_system_turn",
            side: turn.side,
            payload_json: { content: turn.content },
            created_at: turn.created_at,
          } as DebateEvent,
        }
      }
      return { kind: "turn" as const, id: turn.id, turn }
    })
  })()

  // Compute display numbers for argument turns (1, 2, 3, ...) skipping system turns,
  // and an ordered list of turn_numbers for prev/next navigation.
  const displayNumberMap = new Map<number, number>()
  const orderedTurnNumbers = argumentTurns.map((t) => t.turn_number)
  let argCount = 0
  for (const turn of argumentTurns) {
    argCount++
    displayNumberMap.set(turn.turn_number, argCount)
  }
  const showRevealPrompt = Boolean(
    debate.status === "finished" &&
    debate.viewer?.is_participant &&
    debate.viewer.reveal_choice === null
  )
  const isDebateHidden = debate.hidden
  const canModerate = user?.role === "moderator" || user?.role === "admin"
  const canViewHiddenDebate = canModerate || Boolean(debate.viewer?.is_participant)
  const isHiddenReadOnly = isDebateHidden && canViewHiddenDebate && !canModerate
  const realTurnCount = argumentTurns.length
  const turnProgress = debate.turn_limit > 0 ? (realTurnCount / debate.turn_limit) * 100 : 0
  const latestTurnNumber = argumentTurns.length > 0 ? argumentTurns[argumentTurns.length - 1].turn_number : null
  const isChallengeWaiting = debate.status === "waiting" && debate.is_challenge
  return (
    <div className="min-h-screen bg-[var(--bg-primary)]">
      <DebateHeader
        debate={debate}
        turnProgress={turnProgress}
        timerRemaining={timerRemaining}
        onToggleUpvote={handleToggleDebateVote}
        onToggleFollow={handleToggleFollow}
        onReportDebate={() => openReportDialog("debate", debate.id)}
        canModerateDebate={canModerate}
        onModerateDebate={(action) => {
          void handleModerateTarget("debate", debate.id, action)
        }}
        isReadOnly={isHiddenReadOnly}
      />

      <div className="max-w-[820px] mx-auto px-3 sm:px-6 py-4 sm:py-6">
        {isDebateHidden && !canViewHiddenDebate ? (
          <div className="rounded-lg border border-[var(--border-default)] bg-[var(--bg-surface)] p-5 text-center">
            <div className="text-[var(--text-primary)] text-sm font-medium font-sans">{t("flagged.title")}</div>
            <div className="mt-1 text-[var(--text-secondary)] text-xs font-sans">{t("flagged.description")}</div>
          </div>
        ) : (
          <>
            {isDebateHidden && canModerate && (
              <div className="mb-4 rounded-lg border border-[var(--warning)] bg-[var(--warning-light)] px-4 py-3">
                <div className="text-[var(--warning)] text-sm font-medium font-sans">{t("moderation.hiddenTitle")}</div>
                <div className="mt-1 text-[var(--text-secondary)] text-xs font-sans">{t("moderation.adminViewDesc")}</div>
              </div>
            )}
            {isHiddenReadOnly && (
              <div className="mb-4 rounded-lg border border-[var(--warning)] bg-[var(--warning-light)] px-4 py-3">
                <div className="text-[var(--warning)] text-sm font-medium font-sans">{t("moderation.hiddenTitle")}</div>
                <div className="mt-1 text-[var(--text-secondary)] text-xs font-sans">{t("moderation.participantViewDesc")}</div>
              </div>
            )}
            {showRevealPrompt && (
              <RevealPrompt
                anonymousId={debate.viewer?.side === "a" ? sideADisplay : sideBDisplay}
                isSubmitting={isSubmittingReveal}
                error={revealError}
                onReveal={() => handleRevealChoice(true)}
                onStayAnonymous={() => handleRevealChoice(false)}
              />
            )}
            <DebateTimelinePanel
              argumentTurns={argumentTurns}
              collapsedTurns={collapsedTurns}
              debate={debate}
              displayNumberMap={displayNumberMap}
              onCollapseAll={collapseAll}
              onExpandAll={expandAll}
              onReportTurn={(turnID) => openReportDialog("turn", turnID)}
              onToggleTurn={toggleTurn}
              orderedTurnNumbers={orderedTurnNumbers}
              sideADisplay={sideADisplay}
              sideAUsername={sideAUsername}
              sideBDisplay={sideBDisplay}
              sideBUsername={sideBUsername}
              timelineItems={timelineItems}
            />

            {/* Join button for waiting debates — visible to non-participants */}
            {debate.status === "waiting" && !debate.viewer?.is_participant && (
              <div className="mt-6 py-6 px-5 bg-[var(--bg-surface)] border border-[var(--border-default)] rounded-lg text-center">
                <div className="text-[var(--text-secondary)] text-sm font-sans mb-1">
                  {isChallengeWaiting ? t("join.challengedToThis") : t("join.waitingChallenger")}
                </div>
                <div className="text-[var(--text-muted)] text-xs font-sans mb-4">
                  {isChallengeWaiting
                    ? t("join.challengeActionDesc")
                    : t("join.joinActionDesc")}
                </div>
                {joinError && (
                  <div className="mb-3 rounded-lg border border-[var(--error)] bg-[var(--error-light)] px-3 py-2 text-[12px] text-[var(--error)]">
                    {joinError}
                  </div>
                )}
                {isChallengeWaiting ? (
                  <div className="flex items-center justify-center gap-3">
                    <button
                      className="h-10 px-6 py-[6px] bg-[var(--text-primary)] shadow-[0px_0px_0px_2.5px_rgba(255,255,255,0.08)_inset] overflow-hidden rounded-full text-[var(--bg-primary)] text-sm font-medium font-sans hover:opacity-90 transition-colors disabled:opacity-50"
                      onClick={() => void handleRespondChallenge(true)}
                      disabled={isRespondingChallenge}
                    >
                      {isRespondingChallenge ? t("join.working") : t("join.acceptChallenge")}
                    </button>
                    <button
                      className="h-10 px-6 py-[6px] bg-transparent border border-[var(--border-default)] rounded-full text-[var(--text-secondary)] text-sm font-medium font-sans hover:bg-[var(--bg-hover)] transition-colors disabled:opacity-50"
                      onClick={() => void handleRespondChallenge(false)}
                      disabled={isRespondingChallenge}
                    >
                      {isRespondingChallenge ? t("join.working") : tCommon("decline")}
                    </button>
                  </div>
                ) : (
                  <button
                    className="h-10 px-6 py-[6px] bg-[var(--text-primary)] shadow-[0px_0px_0px_2.5px_rgba(255,255,255,0.08)_inset] overflow-hidden rounded-full text-[var(--bg-primary)] text-sm font-medium font-sans hover:opacity-90 transition-colors disabled:opacity-50"
                    onClick={handleJoinDebate}
                    disabled={isJoining}
                  >
                    {isJoining ? t("join.joining") : t("join.joinAsSideB")}
                  </button>
                )}
              </div>
            )}

            {/* Replacement seat banner — visible when a side resigned and seat is open */}
            {debate.status === "waiting_replacement" && !debate.viewer?.is_participant && (
              <div className="mt-6 py-6 px-5 bg-[var(--bg-surface)] border border-[var(--warning)]/30 rounded-lg text-center">
                <div className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full bg-[var(--warning)]/10 text-[var(--warning)] text-xs font-medium font-sans mb-2">
                  {t("replacement.openSeat")}
                </div>
                <div className="text-[var(--text-secondary)] text-sm font-sans mb-1">
                  {t("replacement.resignedSideOpen", { side: debate.open_side?.toUpperCase() ?? debate.current_turn_side.toUpperCase() })}
                </div>
                <div className="text-[var(--text-muted)] text-xs font-sans mb-4">
                  {t("replacement.readAndTake")}
                </div>
                {replaceError && (
                  <div className="mb-3 rounded-lg border border-[var(--error)] bg-[var(--error-light)] px-3 py-2 text-[12px] text-[var(--error)]">
                    {replaceError}
                  </div>
                )}
                <button
                  className="h-10 px-6 py-[6px] bg-[var(--text-primary)] shadow-[0px_0px_0px_2.5px_rgba(255,255,255,0.08)_inset] overflow-hidden rounded-full text-[var(--bg-primary)] text-sm font-medium font-sans hover:opacity-90 transition-colors disabled:opacity-50"
                  onClick={handleReplaceDebate}
                  disabled={isReplacing}
                >
                  {isReplacing ? t("join.joining") : t("replacement.takeSeatAsSide", { side: debate.open_side?.toUpperCase() ?? debate.current_turn_side.toUpperCase() })}
                </button>
              </div>
            )}

            {/* Remaining debater sees waiting-for-replacement message */}
            {debate.status === "waiting_replacement" && debate.viewer?.is_participant && (
              <div className="mt-6 py-4 px-5 bg-[var(--bg-surface)] border border-[var(--warning)]/30 rounded-lg text-center">
                <div className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full bg-[var(--warning)]/10 text-[var(--warning)] text-xs font-medium font-sans mb-2">
                  {t("replacement.openSeat")}
                </div>
                <div className="text-[var(--text-secondary)] text-sm font-sans mb-1">
                  {t("replacement.opponentResigned")}
                </div>
                <div className="text-[var(--text-muted)] text-xs font-sans">
                  {t("replacement.resumeWhenJoined")}
                </div>
              </div>
            )}

            {/* Extension vote — participants see accept/decline buttons */}
            {debate.status === "pending_extension" && debate.viewer?.is_participant && (() => {
              const viewerSide = debate.viewer?.side
              const viewerAccepted = viewerSide === "a"
                ? debate.extension_a_accepted
                : debate.extension_b_accepted
              const opponentAccepted = viewerSide === "a"
                ? debate.extension_b_accepted
                : debate.extension_a_accepted

              if (viewerAccepted === true) {
                return (
                  <div className="mt-6 py-4 px-5 bg-[var(--bg-surface)] border border-[var(--accent-blue)]/30 rounded-lg text-center">
                    <div className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full bg-[var(--accent-blue)]/10 text-[var(--accent-blue)] text-xs font-medium font-sans mb-2">
                      Extension Vote
                    </div>
                    <div className="text-[var(--text-secondary)] text-sm font-sans mb-1">
                      {t("extension.waitingOpponent")}
                    </div>
                    {opponentAccepted === null && (
                      <div className="text-[var(--text-muted)] text-xs font-sans">
                        {t("extension.waitingOpponentDesc")}
                      </div>
                    )}
                  </div>
                )
              }

              return (
                <div className="mt-6 py-6 px-5 bg-[var(--bg-surface)] border border-[var(--accent-blue)]/30 rounded-lg text-center">
                  <div className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full bg-[var(--accent-blue)]/10 text-[var(--accent-blue)] text-xs font-medium font-sans mb-2">
                    Extension Vote
                  </div>
                  <div className="text-[var(--text-secondary)] text-sm font-sans mb-1">
                    {t("extension.limitReached", { limit: debate.turn_limit })}
                  </div>
                  {opponentAccepted === true && (
                    <div className="text-[var(--text-muted)] text-xs font-sans mb-3">
                      {t("extension.opponentAgreed")}
                    </div>
                  )}
                  {opponentAccepted === null && (
                    <div className="text-[var(--text-muted)] text-xs font-sans mb-3">
                      {t("extension.bothMustAgree")}
                    </div>
                  )}
                  {extensionError && (
                    <div className="mb-3 rounded-lg border border-[var(--error)] bg-[var(--error-light)] px-3 py-2 text-[12px] text-[var(--error)]">
                      {extensionError}
                    </div>
                  )}
                  <div className="flex items-center justify-center gap-3">
                    <button
                      className="h-10 px-6 py-[6px] bg-[var(--text-primary)] shadow-[0px_0px_0px_2.5px_rgba(255,255,255,0.08)_inset] overflow-hidden rounded-full text-[var(--bg-primary)] text-sm font-medium font-sans hover:opacity-90 transition-colors disabled:opacity-50"
                      onClick={() => handleExtensionResponse(true)}
                      disabled={isSubmittingExtension}
                    >
                      {isSubmittingExtension ? t("extension.submitting") : t("extension.acceptExtension")}
                    </button>
                    <button
                      className="h-10 px-6 py-[6px] bg-transparent border border-[var(--border-default)] rounded-full text-[var(--text-secondary)] text-sm font-medium font-sans hover:bg-[var(--bg-hover)] transition-colors disabled:opacity-50"
                      onClick={() => handleExtensionResponse(false)}
                      disabled={isSubmittingExtension}
                    >
                      {isSubmittingExtension ? t("extension.submitting") : t("extension.declineDraw")}
                    </button>
                  </div>
                </div>
              )
            })()}

            {/* Extension vote — non-participants see a status banner */}
            {debate.status === "pending_extension" && !debate.viewer?.is_participant && (
              <div className="mt-6 py-4 px-5 bg-[var(--bg-surface)] border border-[var(--accent-blue)]/30 rounded-lg text-center">
                <div className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full bg-[var(--accent-blue)]/10 text-[var(--accent-blue)] text-xs font-medium font-sans mb-2">
                  Extension Vote
                </div>
                <div className="text-[var(--text-secondary)] text-sm font-sans mb-1">
                  {t("extension.voting")}
                </div>
                <div className="text-[var(--text-muted)] text-xs font-sans">
                  {t("extension.votingDesc")}
                </div>
              </div>
            )}

            {/* Creator waiting message */}
            {debate.status === "waiting" && debate.viewer?.is_participant && (
              <div className="mt-6 py-4 px-5 bg-[var(--bg-surface)] border border-[var(--border-default)] rounded-lg text-center">
                <div className="text-[var(--text-secondary)] text-sm font-sans mb-1">
                  {isChallengeWaiting
                    ? (debate.invited_username
                      ? t("creatorWaiting.waitingForInvited", { username: debate.invited_username })
                      : t("creatorWaiting.waitingForOpponent"))
                    : t("creatorWaiting.waitingForChallenger")}
                </div>
                <div className="text-[var(--text-muted)] text-xs font-sans">
                  {isChallengeWaiting
                    ? t("creatorWaiting.challengePrivate")
                    : t("creatorWaiting.openingLive")}
                </div>
                {!isChallengeWaiting && (
                  <div className="mt-4 border-t border-[var(--border-subtle)] pt-4">
                    <div className="text-[var(--text-secondary)] text-xs font-sans mb-3">
                      {t("creatorWaiting.wantSpecific")}
                    </div>
                    <button
                      type="button"
                      onClick={handleOpenInviteModal}
                      className="h-9 px-3 rounded-md text-[12px] font-medium bg-[var(--text-primary)] text-[var(--bg-primary)]"
                    >
                      {t("creatorWaiting.browseUsers")}
                    </button>
                  </div>
                )}
              </div>
            )}

            {/* Turn input — shown when it's the viewer's turn */}
            {!isHiddenReadOnly && debate.status === "active" && debate.viewer?.is_participant && debate.viewer.side === debate.current_turn_side && (
              <TurnInput
                side={debate.current_turn_side}
                turnNumber={debate.turn_count + 1}
                anonymousId={
                  debate.current_turn_side === "a"
                    ? (debate.side_a.anonymous_id ?? t("common.anonymousFallbackA"))
                    : (debate.side_b.anonymous_id ?? t("common.anonymousFallbackB"))
                }
                isSubmitting={isSubmittingTurn}
                onSubmit={handleSubmitTurn}
                error={turnError}
                allTurns={argumentTurns}
              />
            )}

            {/* Waiting for opponent — shown when it's NOT the viewer's turn */}
            {debate.status === "active" && (
              !debate.viewer?.is_participant || debate.viewer.side !== debate.current_turn_side
            ) && (
                <div className="mt-6 py-4 px-5 bg-[var(--bg-surface)] border border-[var(--border-default)] rounded-lg text-center">
                  <div className="text-[var(--text-secondary)] text-sm font-sans mb-1">
                    {t.rich("active.waitingForOpponent", {
                      opponent: () => (
                        debate.current_turn_side === "a"
                          ? (sideAUsername
                            ? (
                              <Link
                                href={`/profile/${encodeURIComponent(sideAUsername)}`}
                                className="hover:underline underline-offset-2"
                              >
                                {sideAUsername}
                              </Link>
                            )
                            : sideADisplay)
                          : (sideBUsername
                            ? (
                              <Link
                                href={`/profile/${encodeURIComponent(sideBUsername)}`}
                                className="hover:underline underline-offset-2"
                              >
                                {sideBUsername}
                              </Link>
                            )
                            : sideBDisplay)
                      )
                    })}
                  </div>
                  <div className="text-[var(--text-muted)] text-xs font-sans">
                    {t("active.spectatorCount", { count: debate.spectator_count })}
                  </div>
                </div>
              )}

            {/* Debate actions — concede, resign, draw for participants in active debates */}
            {!isHiddenReadOnly && <DebateActions debate={debate} onDebateUpdate={setDebate} />}

            {debate.status === "finished" && debate.viewer?.is_participant && (
              <div className="mt-4 rounded-lg border border-[var(--border-default)] bg-[var(--bg-surface)] p-4">
                <div className="text-[13px] font-medium text-[var(--text-primary)] font-sans">{t("rechallenge.title")}</div>
                <p className="mt-1 text-[12px] text-[var(--text-secondary)] font-sans">
                  {t("rechallenge.description")}
                </p>
                <div className="mt-3 grid gap-2 sm:grid-cols-2">
                  <button
                    type="button"
                    onClick={() => handleStartRechallenge("same_side")}
                    className="h-9 rounded-md bg-[var(--text-primary)] px-3 text-[12px] font-medium text-[var(--bg-primary)]"
                  >
                    {t("rechallenge.buttonPrefill")}
                  </button>
                  <button
                    type="button"
                    onClick={() => handleStartRechallenge("new_topic")}
                    className="h-9 rounded-md border border-[var(--border-default)] bg-[var(--bg-primary)] px-3 text-[12px] font-medium text-[var(--text-primary)]"
                  >
                    {t("rechallenge.buttonNewTopic")}
                  </button>
                </div>
                <p className="mt-2 text-[11px] text-[var(--text-muted)] font-sans">
                  {t("rechallenge.note")}
                </p>
              </div>
            )}

            {debate.status === "finished" && (
              <DiscussionSection
                comments={comments}
                totalCount={commentTotal}
                sort={commentSort}
                onSortChange={setCommentSort}
                onPostComment={handlePostComment}
                onPostReply={handlePostReply}
                onToggleUpvote={handleToggleCommentVote}
                onEditComment={handleEditComment}
                onDeleteComment={handleDeleteComment}
                onReportComment={(commentID) => openReportDialog("comment", commentID)}
                canModerateComments={canModerate}
                onModerateComment={(commentID, action) => {
                  void handleModerateTarget("comment", commentID, action)
                }}
                highlightCommentId={highlightCommentId}
                isLocked={
                  Boolean(debate.ended_at) &&
                  Date.now() - new Date(debate.ended_at ?? 0).getTime() > 7 * 24 * 60 * 60 * 1000
                }
                canPostReflection={Boolean(debate.viewer?.is_participant)}
                hasPostedReflection={hasViewerReflection(comments)}
                readOnly={isHiddenReadOnly}
              />
            )}
          </>
        )}
      </div>

      <DebatePageOverlays
        actionError={inviteActionError}
        canShowNewTurnIndicator={(!isDebateHidden || canViewHiddenDebate) && !isHiddenReadOnly && showNewTurnIndicator}
        currentUsername={user?.username}
        invitedUsernames={invitedUsernames}
        invitingUsername={invitingUsername}
        inviteOpen={inviteModalOpen}
        latestTurnNumber={latestTurnNumber}
        onInviteActionErrorChange={setInviteActionError}
        onInviteOpenChange={setInviteModalOpen}
        onInviteUser={(username) => void handleInviteUser(username)}
        onPendingScrollTurnChange={setPendingScrollTurn}
        onReportOpenChange={setReportDialogOpen}
        onReportTargetIdChange={setReportTargetID}
        onReportTargetTypeChange={setReportTargetType}
        onSetShowNewTurnIndicator={setShowNewTurnIndicator}
        onSubmitReport={handleSubmitReport}
        reportOpen={reportDialogOpen}
        reportTargetType={reportTargetType}
        t={t}
      />
    </div>
  )
}
