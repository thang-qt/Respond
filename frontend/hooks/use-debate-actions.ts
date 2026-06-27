"use client"

import { useCallback, useState } from "react"
import type { AppRouterInstance } from "next/dist/shared/lib/app-router-context.shared-runtime"
import { toast } from "sonner"
import { ApiError } from "@/lib/api"
import type { DebateDetail } from "@/lib/debates"
import {
  fetchDebate,
  followDebate,
  inviteToDebate,
  joinDebate,
  replaceDebate,
  respondChallenge,
  respondExtension,
  revealDebateIdentity,
  submitTurn,
  toggleDebateVote,
  unfollowDebate,
} from "@/lib/debates-api"
import { normalizeDebateDetail } from "@/lib/debate-detail"
import { EMAIL_VERIFICATION_REQUIRED_MESSAGE, isEmailNotVerifiedError } from "@/lib/verification"

type Options = {
  debate: DebateDetail | null
  debateId: string
  requireVerified: (setError?: (message: string | null) => void) => boolean
  router: AppRouterInstance
  setDebate: React.Dispatch<React.SetStateAction<DebateDetail | null>>
}

export function useDebateActions({ debate, debateId, requireVerified, router, setDebate }: Options) {
  const [isJoining, setIsJoining] = useState(false)
  const [isRespondingChallenge, setIsRespondingChallenge] = useState(false)
  const [joinError, setJoinError] = useState<string | null>(null)
  const [inviteModalOpen, setInviteModalOpen] = useState(false)
  const [invitingUsername, setInvitingUsername] = useState<string | null>(null)
  const [inviteActionError, setInviteActionError] = useState<string | null>(null)
  const [invitedUsernames, setInvitedUsernames] = useState<Set<string>>(new Set())
  const [isReplacing, setIsReplacing] = useState(false)
  const [replaceError, setReplaceError] = useState<string | null>(null)
  const [isSubmittingTurn, setIsSubmittingTurn] = useState(false)
  const [turnError, setTurnError] = useState<string | null>(null)
  const [isSubmittingReveal, setIsSubmittingReveal] = useState(false)
  const [revealError, setRevealError] = useState<string | null>(null)
  const [isSubmittingExtension, setIsSubmittingExtension] = useState(false)
  const [extensionError, setExtensionError] = useState<string | null>(null)

  const refreshDebate = useCallback(async (id = debate?.id ?? debateId) => {
    const res = await fetchDebate(id)
    setDebate(normalizeDebateDetail(res.data))
  }, [debate?.id, debateId, setDebate])

  const handleToggleDebateVote = useCallback(async () => {
    if (!debate) return
    if (!requireVerified()) return
    try {
      const res = await toggleDebateVote(debate.id)
      setDebate((prev) => {
        if (!prev) return prev
        const nextViewer = prev.viewer ?? { is_participant: false, side: null, has_upvoted: false, is_following: false, reveal_choice: null }
        return {
          ...prev,
          upvote_count: res.data.upvote_count,
          viewer: { ...nextViewer, has_upvoted: res.data.voted },
        }
      })
    } catch (err) {
      if (isEmailNotVerifiedError(err)) {
        toast.error(EMAIL_VERIFICATION_REQUIRED_MESSAGE)
      } else if (err instanceof Error) {
        toast.error(err.message)
      } else {
        toast.error("Failed to vote.")
      }
    }
  }, [debate, requireVerified, setDebate])

  const handleToggleFollow = useCallback(async () => {
    if (!debate) return
    if (!requireVerified()) return
    try {
      const isFollowing = Boolean(debate.viewer?.is_following)
      if (isFollowing) {
        await unfollowDebate(debate.id)
      } else {
        await followDebate(debate.id)
      }
      setDebate((prev) => {
        if (!prev) return prev
        const nextViewer = prev.viewer ?? { is_participant: false, side: null, has_upvoted: false, is_following: false, reveal_choice: null }
        return {
          ...prev,
          viewer: { ...nextViewer, is_following: !isFollowing },
        }
      })
    } catch (err) {
      if (isEmailNotVerifiedError(err)) {
        toast.error(EMAIL_VERIFICATION_REQUIRED_MESSAGE)
      } else if (err instanceof Error) {
        toast.error(err.message)
      } else {
        toast.error("Failed to update follow state.")
      }
    }
  }, [debate, requireVerified, setDebate])

  const handleRevealChoice = useCallback(async (reveal: boolean) => {
    if (!debate) return
    if (!requireVerified(setRevealError)) return
    setIsSubmittingReveal(true)
    setRevealError(null)
    try {
      const res = await revealDebateIdentity(debate.id, reveal)
      setDebate((prev) => {
        if (!prev) return prev
        const nextViewer = prev.viewer
          ? { ...prev.viewer, reveal_choice: reveal }
          : { is_participant: true, side: res.data.side, has_upvoted: false, is_following: false, reveal_choice: reveal }
        if (res.data.side === "a") {
          return { ...prev, side_a: { ...prev.side_a, revealed: reveal, user: reveal ? res.data.user : null }, viewer: nextViewer }
        }
        return { ...prev, side_b: { ...prev.side_b, revealed: reveal, user: reveal ? res.data.user : null }, viewer: nextViewer }
      })
    } catch (err) {
      if (isEmailNotVerifiedError(err)) {
        setRevealError(EMAIL_VERIFICATION_REQUIRED_MESSAGE)
      } else if (err instanceof ApiError && err.code === "REVEAL_ALREADY_CHOSEN") {
        try { await refreshDebate(debateId) } catch (_) {}
        setRevealError("You already made a reveal choice.")
      } else if (err instanceof Error) {
        setRevealError(err.message)
      } else {
        setRevealError("Something went wrong.")
      }
    } finally {
      setIsSubmittingReveal(false)
    }
  }, [debate, debateId, refreshDebate, requireVerified, setDebate])

  const handleExtensionResponse = useCallback(async (accept: boolean) => {
    if (!debate) return
    if (!requireVerified(setExtensionError)) return
    setIsSubmittingExtension(true)
    setExtensionError(null)
    try {
      await respondExtension(debate.id, accept)
      await refreshDebate(debate.id)
    } catch (err) {
      setExtensionError(err instanceof ApiError ? err.message : "Something went wrong. Try again.")
    } finally {
      setIsSubmittingExtension(false)
    }
  }, [debate, refreshDebate, requireVerified])

  const handleJoinDebate = useCallback(async () => {
    if (!debate) return
    if (!requireVerified(setJoinError)) return
    setIsJoining(true)
    setJoinError(null)
    try {
      await joinDebate(debate.id)
      await refreshDebate(debate.id)
    } catch (err) {
      setJoinError(err instanceof ApiError ? err.message : "Something went wrong. Try again.")
    } finally {
      setIsJoining(false)
    }
  }, [debate, refreshDebate, requireVerified])

  const handleRespondChallenge = useCallback(async (accept: boolean) => {
    if (!debate) return
    if (!requireVerified(setJoinError)) return
    setIsRespondingChallenge(true)
    setJoinError(null)
    try {
      const res = await respondChallenge(debate.id, accept)
      if (accept && res.data.accepted) {
        router.push(`/debate/${res.data.debate_id || debate.id}`)
        return
      }
      router.push("/challenges")
    } catch (err) {
      setJoinError(err instanceof ApiError ? err.message : "Could not respond to challenge.")
    } finally {
      setIsRespondingChallenge(false)
    }
  }, [debate, requireVerified, router])

  const handleOpenInviteModal = useCallback(() => {
    if (!requireVerified(setInviteActionError)) return
    setInviteModalOpen(true)
  }, [requireVerified])

  const handleInviteUser = useCallback(async (username: string) => {
    if (!debate) return
    if (!requireVerified(setInviteActionError)) return

    const normalized = username.trim().replace(/^@+/, "")
    if (!normalized) return

    setInvitingUsername(normalized)
    setInviteActionError(null)
    try {
      await inviteToDebate(debate.id, normalized)
      setInvitedUsernames((prev) => {
        const next = new Set(prev)
        next.add(normalized.toLowerCase())
        return next
      })
    } catch (err) {
      setInviteActionError(err instanceof ApiError ? err.message : "Could not send invite.")
    } finally {
      setInvitingUsername(null)
    }
  }, [debate, requireVerified])

  const handleReplaceDebate = useCallback(async () => {
    if (!debate) return
    if (!requireVerified(setReplaceError)) return
    setIsReplacing(true)
    setReplaceError(null)
    try {
      await replaceDebate(debate.id)
      await refreshDebate(debate.id)
    } catch (err) {
      setReplaceError(err instanceof ApiError ? err.message : "Something went wrong. Try again.")
    } finally {
      setIsReplacing(false)
    }
  }, [debate, refreshDebate, requireVerified])

  const handleSubmitTurn = useCallback(async (payload: { content: string; ai_assisted?: boolean; ai_note?: string }) => {
    if (!debate) return
    if (!requireVerified(setTurnError)) return
    setIsSubmittingTurn(true)
    setTurnError(null)
    try {
      await submitTurn(debate.id, payload)
      await refreshDebate(debate.id)
    } catch (err) {
      setTurnError(err instanceof ApiError ? err.message : "Something went wrong. Try again.")
    } finally {
      setIsSubmittingTurn(false)
    }
  }, [debate, refreshDebate, requireVerified])

  const handleStartRechallenge = useCallback((mode: "same_side" | "new_topic") => {
    if (!debate || debate.status !== "finished" || !debate.viewer?.is_participant || !debate.viewer.side) return
    const next = new URLSearchParams()
    next.set("source_debate", debate.slug || debate.id)
    next.set("prefill", mode === "same_side" ? "1" : "0")
    router.push(`/create?${next.toString()}`)
  }, [debate, router])

  return {
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
  }
}
