"use client"

import { useEffect, useMemo, useRef, useState } from "react"
import { useTranslations } from "next-intl"
import { Button } from "@/components/ui/button"
import { Textarea } from "@/components/ui/textarea"
import type { DebateSide, DebateTurn } from "@/lib/debates"
import { ArrowLeft, ArrowRight, ArrowsOut, Robot, X } from "@phosphor-icons/react"
import { formatDate } from "@/lib/utils"

const minTurnLength = 100
const maxTurnLength = 5000
const maxAINoteLength = 300
const LS_SHOW_PREVIOUS_KEY = "respond.zen_show_previous"



interface TurnSubmitPayload {
  content: string
  ai_assisted?: boolean
  ai_note?: string
}

interface TurnInputProps {
  side: DebateSide
  turnNumber: number
  anonymousId: string
  isSubmitting: boolean
  onSubmit: (payload: TurnSubmitPayload) => Promise<void>
  error: string | null
  /** All non-system argument turns so far — user can browse any of them */
  allTurns?: DebateTurn[]
}

export function TurnInput({
  side,
  turnNumber,
  anonymousId,
  isSubmitting,
  onSubmit,
  error,
  allTurns = [],
}: TurnInputProps) {
  const t = useTranslations("turnInput")
  const tCommon = useTranslations("common")

  const aiChips = useMemo(() => [
    { label: t("aiChips.brainstorm.label"), snippet: t("aiChips.brainstorm.snippet") },
    { label: t("aiChips.outline.label"), snippet: t("aiChips.outline.snippet") },
    { label: t("aiChips.rewrite.label"), snippet: t("aiChips.rewrite.snippet") },
    { label: t("aiChips.grammar.label"), snippet: t("aiChips.grammar.snippet") },
    { label: t("aiChips.sources.label"), snippet: t("aiChips.sources.snippet") },
  ], [t])

  const [content, setContent] = useState("")
  const [aiAssisted, setAIAssisted] = useState(false)
  const [aiNote, setAINote] = useState("")
  const [showConfirm, setShowConfirm] = useState(false)
  const [zenMode, setZenMode] = useState(false)

  // Persisted: whether the previous-turn panel is visible
  const [showPrevious, setShowPreviousState] = useState<boolean>(() => {
    if (typeof window === "undefined") return false
    return window.localStorage.getItem(LS_SHOW_PREVIOUS_KEY) === "true"
  })

  // Which turn the user is currently reading in the panel
  const [selectedTurnIdx, setSelectedTurnIdx] = useState<number>(() =>
    Math.max(0, allTurns.length - 1)
  )

  const zenTextareaRef = useRef<HTMLTextAreaElement>(null)

  const charCount = useMemo(() => content.trim().length, [content])
  const aiNoteCount = useMemo(() => aiNote.trim().length, [aiNote])
  const isValid = charCount >= minTurnLength && charCount <= maxTurnLength
  const aiNoteValid = aiNoteCount <= maxAINoteLength

  const isA = side === "a"
  const borderColor = isA ? "border-[var(--side-a-border)]" : "border-[var(--side-b-border)]"
  const bgColor = isA ? "bg-[var(--side-a-bg)]" : "bg-[var(--side-b-bg)]"
  const badgeBg = isA ? "bg-[var(--side-a)]" : "bg-[var(--side-b)]"
  const nameColor = isA ? "text-[var(--side-a)]" : "text-[var(--side-b)]"

  // Keep selectedTurnIdx in-bounds if allTurns changes
  useEffect(() => {
    if (allTurns.length === 0) return
    setSelectedTurnIdx((prev) => Math.min(prev, allTurns.length - 1))
  }, [allTurns.length])

  function setShowPrevious(next: boolean) {
    setShowPreviousState(next)
    try {
      window.localStorage.setItem(LS_SHOW_PREVIOUS_KEY, String(next))
    } catch { /* ignore */ }
  }

  // Lock body scroll in zen mode
  useEffect(() => {
    if (zenMode) {
      document.body.style.overflow = "hidden"
      const t = setTimeout(() => zenTextareaRef.current?.focus(), 80)
      return () => { clearTimeout(t); document.body.style.overflow = "" }
    }
    document.body.style.overflow = ""
  }, [zenMode])

  // Escape closes zen mode
  useEffect(() => {
    if (!zenMode) return
    const handler = (e: KeyboardEvent) => { if (e.key === "Escape") setZenMode(false) }
    window.addEventListener("keydown", handler)
    return () => window.removeEventListener("keydown", handler)
  }, [zenMode])

  async function handleSubmit() {
    if (!isValid || !aiNoteValid || isSubmitting) return
    const payload: TurnSubmitPayload = { content: content.trim(), ai_assisted: aiAssisted }
    if (aiAssisted && aiNote.trim().length > 0) payload.ai_note = aiNote.trim()
    await onSubmit(payload)
    setContent("")
    setAIAssisted(false)
    setAINote("")
    setShowConfirm(false)
    setZenMode(false)
  }

  function toggleChip(snippet: string) {
    const current = aiNote.trim()
    if (current.includes(snippet)) {
      setAINote(current.replace(snippet, "").replace(/\s{2,}/g, " ").trim())
      return
    }
    setAINote(current ? `${current} ${snippet}` : snippet)
  }

  const hasTurns = allTurns.length > 0
  const selectedTurn = hasTurns ? allTurns[Math.min(selectedTurnIdx, allTurns.length - 1)] : null
  const canGoPrev = selectedTurnIdx > 0
  const canGoNext = selectedTurnIdx < allTurns.length - 1

  // ─── Shared sub-pieces ───────────────────────────────────────────────────

  const charCounterNode = (
    <span className={`text-[11px] font-sans tabular-nums shrink-0 ${charCount > maxTurnLength ? "text-[var(--error)]" : "text-[var(--text-muted)]"}`}>
      {charCount.toLocaleString()}/{maxTurnLength.toLocaleString()}
    </span>
  )

  /**
   * AI-expanded section: chips + optional note textarea.
   * Shown *above* the submit row when aiAssisted is true.
   */
  const aiExpandedSection = aiAssisted && (
    <div className="space-y-2 pt-1">
      <div className="flex flex-wrap gap-1.5">
        {aiChips.map((chip) => {
          const active = aiNote.includes(chip.snippet)
          return (
            <button
              key={chip.label}
              type="button"
              onClick={() => toggleChip(chip.snippet)}
              disabled={isSubmitting}
              className={`px-2 py-0.5 rounded text-[11px] border transition-colors ${active
                ? "bg-[var(--text-primary)] text-[var(--bg-primary)] border-[var(--text-primary)]"
                : "bg-transparent text-[var(--text-secondary)] border-[var(--border-default)] hover:bg-[var(--bg-surface-alt)]"
                }`}
            >
              {chip.label}
            </button>
          )
        })}
      </div>
      <div className="relative">
        <Textarea
          placeholder={t("aiNotePlaceholder")}
          value={aiNote}
          onChange={(e) => setAINote(e.target.value)}
          className="min-h-[64px] bg-[var(--bg-primary)] border-[var(--border-default)] pr-12"
          disabled={isSubmitting}
        />
        <span className={`absolute bottom-2 right-3 text-[10px] tabular-nums pointer-events-none ${aiNoteValid ? "text-[var(--text-muted)]" : "text-[var(--error)]"}`}>
          {aiNoteCount}/{maxAINoteLength}
        </span>
      </div>
    </div>
  )

  /**
   * The AI toggle + submit button on one row.
   * idPrefix keeps checkbox ids unique between normal and zen instances.
   */
  const actionRow = (idPrefix: string) => (
    showConfirm ? (
      <div className="flex items-center justify-between gap-3 bg-[var(--bg-primary)] border border-[var(--border-default)] rounded-lg px-4 py-3">
        <p className="text-[13px] text-[var(--text-secondary)] font-sans">
          {t("confirmSubmitText")}
        </p>
        <div className="flex gap-2 shrink-0">
          <Button variant="outline" size="sm" onClick={() => setShowConfirm(false)} disabled={isSubmitting}>
            {tCommon("cancel")}
          </Button>
          <Button size="sm" onClick={handleSubmit} disabled={isSubmitting || !aiNoteValid}>
            {isSubmitting ? tCommon("submitting") : tCommon("confirm")}
          </Button>
        </div>
      </div>
    ) : (
      <div className="flex items-center justify-between gap-3">
        {/* AI toggle — left side of action row */}
        <label
          htmlFor={`${idPrefix}-ai`}
          className={`flex items-center gap-1.5 cursor-pointer select-none group ${isSubmitting ? "opacity-50 pointer-events-none" : ""}`}
        >
          <input
            id={`${idPrefix}-ai`}
            type="checkbox"
            checked={aiAssisted}
            onChange={(e) => setAIAssisted(e.target.checked)}
            disabled={isSubmitting}
            className="sr-only peer"
          />
          {/* custom checkbox look */}
          <span
            className={`w-4 h-4 rounded border flex items-center justify-center transition-colors shrink-0 ${aiAssisted
              ? "bg-[var(--text-primary)] border-[var(--text-primary)]"
              : "border-[var(--border-strong)] bg-transparent group-hover:border-[var(--text-secondary)]"
              }`}
          >
            {aiAssisted && (
              <svg width="9" height="7" viewBox="0 0 9 7" fill="none" className="text-[var(--bg-primary)]">
                <path d="M1 3.5L3.5 6L8 1" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
              </svg>
            )}
          </span>
          <span className={`text-[12px] font-sans flex items-center gap-1 ${aiAssisted ? "text-[var(--text-primary)]" : "text-[var(--text-muted)] group-hover:text-[var(--text-secondary)]"} transition-colors`}>
            <Robot size={12} className="shrink-0" />
            {t("aiAssisted")}
          </span>
        </label>

        {/* Submit button — right */}
        <Button
          onClick={() => setShowConfirm(true)}
          disabled={!isValid || !aiNoteValid || isSubmitting}
        >
          {t("submitTurn")}
        </Button>
      </div>
    )
  )

  // ─── Previous-turn panel ─────────────────────────────────────────────────

  const PreviousTurnPanel = () => {
    if (!hasTurns || !selectedTurn) return null
    const turnIsA = selectedTurn.side === "a"
    return (
      <div className="flex flex-col h-full">
        <div
          className="flex items-center justify-between gap-2 px-5 py-3 border-b shrink-0"
          style={{ borderColor: "var(--border-subtle)" }}
        >
          <div className="flex items-center gap-2 min-w-0">
            <span className={`w-5 h-5 rounded-full text-white text-[9px] font-bold flex items-center justify-center font-mono shrink-0 ${turnIsA ? "bg-[var(--side-a)]" : "bg-[var(--side-b)]"}`}>
              {turnIsA ? "A" : "B"}
            </span>
            <span className={`text-[12px] font-semibold font-sans truncate ${turnIsA ? "text-[var(--side-a)]" : "text-[var(--side-b)]"}`}>
              {t("turnTitle", { number: selectedTurnIdx + 1 })}
            </span>
            <span className="text-[var(--border-strong)] text-[11px]">&middot;</span>
            <span className="text-[11px] text-[var(--text-muted)] font-sans truncate">
              {formatDate(selectedTurn.created_at)}
            </span>
          </div>
          <div className="flex items-center gap-0.5 shrink-0">
            <span className="text-[11px] text-[var(--text-muted)] font-sans tabular-nums mr-1">
              {selectedTurnIdx + 1}/{allTurns.length}
            </span>
            <button
              type="button"
              onClick={() => setSelectedTurnIdx((i) => Math.max(0, i - 1))}
              disabled={!canGoPrev}
              title={t("earlierTurn")}
              className={`w-6 h-6 flex items-center justify-center rounded transition-colors ${canGoPrev ? "text-[var(--text-muted)] hover:text-[var(--text-secondary)] hover:bg-black/[0.04]" : "text-[var(--border-default)] cursor-default"}`}
            >
              <ArrowLeft size={13} />
            </button>
            <button
              type="button"
              onClick={() => setSelectedTurnIdx((i) => Math.min(allTurns.length - 1, i + 1))}
              disabled={!canGoNext}
              title={t("laterTurn")}
              className={`w-6 h-6 flex items-center justify-center rounded transition-colors ${canGoNext ? "text-[var(--text-muted)] hover:text-[var(--text-secondary)] hover:bg-black/[0.04]" : "text-[var(--border-default)] cursor-default"}`}
            >
              <ArrowRight size={13} />
            </button>
          </div>
        </div>
        <div className="flex-1 overflow-y-auto px-5 py-5">
          <p className="text-[14px] leading-[1.85] text-[var(--text-secondary)] font-sans whitespace-pre-wrap">
            {selectedTurn.content}
          </p>
        </div>
      </div>
    )
  }

  // ─── Zen Mode overlay ─────────────────────────────────────────────────────

  const zenOverlay = zenMode && (
    <div className="fixed inset-0 z-[100] flex flex-col" style={{ background: "var(--bg-primary)" }}>
      {/* Top bar */}
      <div
        className="flex items-center justify-between gap-3 px-4 sm:px-6 py-3 border-b shrink-0"
        style={{ borderColor: "var(--border-subtle)" }}
      >
        <div className="flex items-center gap-2 shrink-0">
          <span className={`w-6 h-6 rounded-full ${badgeBg} text-white text-[10px] font-bold flex items-center justify-center font-mono shrink-0`}>
            {isA ? "A" : "B"}
          </span>
          <span className={`text-[13px] font-semibold font-sans ${nameColor}`}>{anonymousId}</span>
          <span className="text-[var(--border-strong)] text-[11px] hidden sm:inline">&middot;</span>
          <span className="text-[var(--text-muted)] text-[11px] font-sans hidden sm:inline">{t("turnTitle", { number: turnNumber })}</span>
        </div>

        <div className="flex items-center gap-2">
          {hasTurns && (
            <button
              type="button"
              onClick={() => {
                if (!showPrevious) setSelectedTurnIdx(allTurns.length - 1)
                setShowPrevious(!showPrevious)
              }}
              className={`h-7 px-3 rounded-full text-[11px] font-medium font-sans border transition-colors whitespace-nowrap ${showPrevious
                ? "bg-[var(--text-primary)] text-[var(--bg-primary)] border-[var(--text-primary)]"
                : "bg-transparent text-[var(--text-secondary)] border-[var(--border-default)] hover:bg-[var(--bg-surface-alt)]"
                }`}
            >
              {showPrevious ? t("hideTurns") : t("showTurns")}
            </button>
          )}
          {charCounterNode}
          <button
            type="button"
            onClick={() => setZenMode(false)}
            title={t("exitZenMode")}
            className="w-7 h-7 flex items-center justify-center rounded-md text-[var(--text-muted)] hover:text-[var(--text-primary)] hover:bg-[var(--bg-surface-alt)] transition-colors"
          >
            <X size={16} />
          </button>
        </div>
      </div>

      {/* Body */}
      <div className="flex flex-1 overflow-hidden">
        {/* Turns panel — desktop left column */}
        {showPrevious && hasTurns && (
          <div
            className="hidden sm:flex flex-col w-[42%] max-w-[500px] border-r shrink-0 overflow-hidden"
            style={{ borderColor: "var(--border-subtle)" }}
          >
            <PreviousTurnPanel />
          </div>
        )}

        {/* Writing column */}
        <div className="flex flex-col flex-1 min-w-0 overflow-hidden">
          <div className="flex-1 overflow-y-auto px-4 sm:px-8 md:px-14 py-6">
            <textarea
              ref={zenTextareaRef}
              placeholder={t("textareaPlaceholder")}
              value={content}
              onChange={(e) => setContent(e.target.value)}
              disabled={isSubmitting}
              className="w-full h-full min-h-[300px] resize-none bg-transparent border-0 shadow-none ring-0 outline-none text-[var(--text-primary)] text-[15px] sm:text-[16px] leading-[1.85] placeholder:text-[var(--text-muted)] font-sans focus:outline-none focus:ring-0"
            />
          </div>

          {/* Bottom toolbar */}
          <div
            className="shrink-0 border-t px-4 sm:px-8 md:px-14 py-4 space-y-3"
            style={{ borderColor: "var(--border-subtle)" }}
          >
            {aiExpandedSection}
            {error && (
              <div className="rounded-lg border border-[var(--error)] bg-[var(--error-light)] px-3 py-2 text-[12px] text-[var(--error)]">
                {error}
              </div>
            )}
            {actionRow("zen")}
          </div>
        </div>
      </div>

      {/* Mobile: turns panel as bottom drawer */}
      {showPrevious && hasTurns && (
        <div
          className="sm:hidden border-t shrink-0 flex flex-col"
          style={{ borderColor: "var(--border-subtle)", maxHeight: "40vh" }}
        >
          <PreviousTurnPanel />
        </div>
      )}
    </div>
  )

  // ─── Normal (inline) view ─────────────────────────────────────────────────

  return (
    <>
      {zenOverlay}

      <div className={`rounded-lg border ${borderColor} ${bgColor} mt-3`}>
        {/* Header */}
        <div className="flex items-center gap-2 sm:gap-3 px-4 sm:px-5 py-3">
          <span className={`w-6 h-6 rounded-full ${badgeBg} text-white text-[10px] font-bold flex items-center justify-center font-mono shrink-0`}>
            {isA ? "A" : "B"}
          </span>
          <div className="flex items-center gap-1.5 flex-1 min-w-0">
            <span className={`text-[13px] font-semibold font-sans ${nameColor}`}>{anonymousId}</span>
            <span className="text-[var(--border-strong)] text-[11px]">&middot;</span>
            <span className="text-[var(--text-muted)] text-[11px] font-sans">{t("turnTitle", { number: turnNumber })}</span>
          </div>
          <div className="flex items-center gap-2">
            {charCounterNode}
            <button
              type="button"
              onClick={() => {
                setSelectedTurnIdx(Math.max(0, allTurns.length - 1))
                setZenMode(true)
              }}
              title={t("zenModeTitle")}
              className="w-7 h-7 flex items-center justify-center rounded-md text-[var(--text-muted)] hover:text-[var(--text-primary)] hover:bg-black/[0.04] transition-colors"
            >
              <ArrowsOut size={15} />
            </button>
          </div>
        </div>

        {/* Input area */}
        <div className="px-4 sm:px-5 pb-4 sm:pb-5 pt-0 space-y-3">
          <Textarea
            placeholder={t("textareaPlaceholder")}
            value={content}
            onChange={(e) => setContent(e.target.value)}
            className="min-h-[200px] bg-[var(--bg-primary)] border-[var(--border-default)]"
            disabled={isSubmitting}
          />

          <p className="text-[11px] text-[var(--text-muted)]">
            {minTurnLength}–{maxTurnLength.toLocaleString()} characters. Once submitted, this turn cannot be edited.
          </p>

          {aiExpandedSection}

          {error && (
            <div className="rounded-lg border border-[var(--error)] bg-[var(--error-light)] px-3 py-2 text-[12px] text-[var(--error)]">
              {error}
            </div>
          )}

          {actionRow("turn")}
        </div>
      </div>
    </>
  )
}
