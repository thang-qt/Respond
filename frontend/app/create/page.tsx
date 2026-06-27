"use client"

import { useEffect, useMemo, useState } from "react"
import { useRouter, useSearchParams } from "next/navigation"
import { useTranslations } from "next-intl"
import { ApiError } from "@/lib/api"
import { createChallengeDebate, createDebate, createRechallengeDebate, fetchDebate, fetchTags } from "@/lib/debates-api"
import type { Tag } from "@/lib/debates"
import { useAuthRedirect } from "@/hooks/use-auth-redirect"
import { useAuth } from "@/hooks/use-auth"
import { EMAIL_VERIFICATION_REQUIRED_MESSAGE } from "@/lib/verification"
import { CreateDebateFormView, maxOpeningTurn, maxOpeningTurnAINote, minOpeningTurn } from "@/components/create-debate-form-view"

export default function CreateDebatePage() {
  const router = useRouter()
  const searchParams = useSearchParams()
  const { status, loginUrl } = useAuthRedirect()
  const { user } = useAuth()
  const t = useTranslations("create")
  const tDebate = useTranslations("debate")
  const [tags, setTags] = useState<Tag[]>([])
  const [loadingTags, setLoadingTags] = useState(true)
  const [submitError, setSubmitError] = useState<string | null>(null)

  const selectTriggerClass =
    "h-8 rounded-md px-3 text-[13px] bg-[var(--bg-surface)] border border-[var(--border-default)] shadow-xs hover:bg-[var(--bg-surface-alt)]"
  const [limitReached, setLimitReached] = useState(false)
  const [isSubmitting, setIsSubmitting] = useState(false)

  const [step, setStep] = useState<1 | 2>(1)
  const [topic, setTopic] = useState("")
  const [selectedTagIDs, setSelectedTagIDs] = useState<string[]>([])
  const [tagQuery, setTagQuery] = useState("")
  const [tagDropdownOpen, setTagDropdownOpen] = useState(false)
  const [timeMode, setTimeMode] = useState<"marathon" | "standard" | "rapid" | "blitz">("standard")
  const [turnLimit, setTurnLimit] = useState<number>(20)
  const [context, setContext] = useState("")
  const [openingTurn, setOpeningTurn] = useState("")
  const [openingTurnAIAssisted, setOpeningTurnAIAssisted] = useState(false)
  const [openingTurnAINote, setOpeningTurnAINote] = useState("")

  const challengeUsername = useMemo(() => {
    const raw = searchParams.get("challenge")
    return raw ? decodeURIComponent(raw).trim() : ""
  }, [searchParams])
  const sourceDebateID = useMemo(() => {
    const raw = searchParams.get("source_debate")
    return raw ? decodeURIComponent(raw).trim() : ""
  }, [searchParams])
  const prefillFromSource = useMemo(
    () => sourceDebateID.length > 0 && searchParams.get("prefill") !== "0",
    [searchParams, sourceDebateID]
  )

  const isChallengeMode = challengeUsername.length > 0
  const isRechallengeMode = sourceDebateID.length > 0
  const [sourcePrefillApplied, setSourcePrefillApplied] = useState(false)

  useEffect(() => {
    setSourcePrefillApplied(false)
  }, [sourceDebateID])

  useEffect(() => {
    if (status === "unauthenticated") {
      router.push(loginUrl)
    }
  }, [status, loginUrl, router])

  useEffect(() => {
    let active = true
    async function loadTags() {
      try {
        const res = await fetchTags()
        if (!active) return
        setTags(res.data)
      } catch (_) {
        if (!active) return
        setTags([])
      } finally {
        if (active) setLoadingTags(false)
      }
    }

    loadTags()
    return () => {
      active = false
    }
  }, [])

  useEffect(() => {
    if (status !== "authenticated") return
    if (!prefillFromSource || sourcePrefillApplied) return

    let active = true

    void fetchDebate(sourceDebateID)
      .then((res) => {
        if (!active) return
        setTopic(res.data.topic)
        setSelectedTagIDs(res.data.tags.map((tag) => tag.id))
        setTimeMode(res.data.time_mode)
        setTurnLimit(res.data.turn_limit)
        setContext(res.data.context ?? "")
        setSourcePrefillApplied(true)
      })
      .catch(() => {
        if (!active) return
        setSourcePrefillApplied(true)
        setSubmitError(t("errors.preload"))
      })

    return () => {
      active = false
    }
  }, [prefillFromSource, sourceDebateID, sourcePrefillApplied, status])

  const topicCount = useMemo(() => topic.trim().length, [topic])
  const contextCount = useMemo(() => context.trim().length, [context])
  const openingTurnCount = useMemo(() => openingTurn.trim().length, [openingTurn])
  const openingTurnAINoteCount = useMemo(() => openingTurnAINote.trim().length, [openingTurnAINote])
  const selectedTags = useMemo(() => {
    const selectedSet = new Set(selectedTagIDs)
    return tags.filter((tag) => selectedSet.has(tag.id))
  }, [selectedTagIDs, tags])
  const filteredTags = useMemo(() => {
    const query = tagQuery.trim().toLowerCase()
    const selectedSet = new Set(selectedTagIDs)
    const available = tags.filter((tag) => !selectedSet.has(tag.id))
    if (!query) {
      return available.slice(0, 10)
    }
    return available
      .filter((tag) => {
        const name = tag.name.toLowerCase()
        const slug = tag.slug.toLowerCase()
        return name.includes(query) || slug.includes(query)
      })
      .slice(0, 10)
  }, [selectedTagIDs, tagQuery, tags])

  if (status !== "authenticated") {
    return <div className="min-h-screen bg-[var(--bg-primary)]" />
  }

  function handleNext() {
    setSubmitError(null)

    if (topicCount < 10 || topicCount > 200) {
      setSubmitError(t("errors.topic"))
      return
    }
    if (selectedTagIDs.length < 1 || selectedTagIDs.length > 3) {
      setSubmitError(t("errors.tags"))
      return
    }

    setStep(2)
  }

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setSubmitError(null)

    if (!user?.email_verified) {
      setSubmitError(EMAIL_VERIFICATION_REQUIRED_MESSAGE)
      return
    }

    if (openingTurnCount < minOpeningTurn || openingTurnCount > maxOpeningTurn) {
      setSubmitError(t("errors.opening", { min: minOpeningTurn, max: maxOpeningTurn.toLocaleString() }))
      return
    }

    if (openingTurnAIAssisted && openingTurnAINoteCount > maxOpeningTurnAINote) {
      setSubmitError(t("errors.aiNote", { max: maxOpeningTurnAINote }))
      return
    }

    setIsSubmitting(true)
    setLimitReached(false)
    try {
      const payload = {
        topic: topic.trim(),
        tag_ids: selectedTagIDs,
        time_mode: timeMode,
        turn_limit: turnLimit,
        context: context.trim() ? context.trim() : undefined,
        opening_turn: openingTurn.trim(),
        opening_turn_ai_assisted: openingTurnAIAssisted,
        opening_turn_ai_note:
          openingTurnAIAssisted && openingTurnAINote.trim().length > 0
            ? openingTurnAINote.trim()
            : undefined,
      }

      const res = isRechallengeMode
        ? await createRechallengeDebate(sourceDebateID, payload)
        : isChallengeMode
          ? await createChallengeDebate({ invited_username: challengeUsername, ...payload })
          : await createDebate(payload)
      router.push(`/debate/${res.data.slug || res.data.id}`)
    } catch (err) {
      if (err instanceof ApiError) {
        if (err.code === "DEBATE_LIMIT_REACHED") {
          setLimitReached(true)
          setSubmitError(t("errors.limit"))
        } else if (err.code === "USER_NOT_FOUND") {
          setSubmitError(t("errors.userNotFound"))
        } else if (err.code === "DEBATE_NOT_PARTICIPANT") {
          setSubmitError(t("errors.notParticipant"))
        } else if (err.code === "DEBATE_NOT_FINISHED") {
          setSubmitError(t("errors.notFinished"))
        } else {
          setSubmitError(err.message)
        }
      } else {
        setSubmitError(t("errors.generic"))
      }
    } finally {
      setIsSubmitting(false)
    }
  }

  const isDisabled = isSubmitting || loadingTags || limitReached

  function toggleTag(tagID: string) {
    setSelectedTagIDs((current) => {
      if (current.includes(tagID)) {
        return current.filter((id) => id !== tagID)
      }
      if (current.length >= 3) {
        return current
      }
      return [...current, tagID]
    })
  }

  function addTag(tagID: string) {
    if (selectedTagIDs.includes(tagID) || selectedTagIDs.length >= 3) return
    setSelectedTagIDs((current) => [...current, tagID])
    setTagQuery("")
    setTagDropdownOpen(false)
  }

  function toggleAINoteChip(snippet: string) {
    const current = openingTurnAINote.trim()
    if (current.includes(snippet)) {
      setOpeningTurnAINote(current.replace(snippet, "").replace(/\s{2,}/g, " ").trim())
      return
    }
    setOpeningTurnAINote(current ? `${current} ${snippet}` : snippet)
  }

  return (
    <CreateDebateFormView
      addTag={addTag}
      challengeUsername={challengeUsername}
      context={context}
      contextCount={contextCount}
      filteredTags={filteredTags}
      handleNext={handleNext}
      handleSubmit={handleSubmit}
      isChallengeMode={isChallengeMode}
      isDisabled={isDisabled}
      isRechallengeMode={isRechallengeMode}
      isSubmitting={isSubmitting}
      limitReached={limitReached}
      loadingTags={loadingTags}
      openingTurn={openingTurn}
      openingTurnAIAssisted={openingTurnAIAssisted}
      openingTurnAINote={openingTurnAINote}
      openingTurnAINoteCount={openingTurnAINoteCount}
      openingTurnCount={openingTurnCount}
      prefillFromSource={prefillFromSource}
      selectedTagIDs={selectedTagIDs}
      selectedTags={selectedTags}
      selectTriggerClass={selectTriggerClass}
      setContext={setContext}
      setOpeningTurn={setOpeningTurn}
      setOpeningTurnAIAssisted={setOpeningTurnAIAssisted}
      setOpeningTurnAINote={setOpeningTurnAINote}
      setStep={setStep}
      setSubmitError={setSubmitError}
      setTagDropdownOpen={setTagDropdownOpen}
      setTagQuery={setTagQuery}
      setTimeMode={setTimeMode}
      setTopic={setTopic}
      setTurnLimit={setTurnLimit}
      step={step}
      submitError={submitError}
      tagDropdownOpen={tagDropdownOpen}
      tagQuery={tagQuery}
      t={t}
      tDebate={tDebate}
      timeMode={timeMode}
      toggleAINoteChip={toggleAINoteChip}
      toggleTag={toggleTag}
      topic={topic}
      topicCount={topicCount}
      turnLimit={turnLimit}
      userEmailVerified={user?.email_verified}
    />
  )
}
