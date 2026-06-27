"use client"

import { useState, type ReactNode } from "react"
import { useTranslations } from "next-intl"
import { CaretDown, CaretLeft, CaretRight, Flag, Info, Robot } from "@phosphor-icons/react"
import type { DebateTurn } from "@/lib/debates"
import { formatDate } from "@/lib/utils"

interface TurnCardProps {
  turn: DebateTurn
  displayName: ReactNode
  isCollapsed: boolean
  onToggle: () => void
  displayNumber?: number
  orderedTurnNumbers: number[]
  onReport?: (turnId: string) => void
  canReport?: boolean
}

export function TurnCard({ turn, displayName, isCollapsed, onToggle, displayNumber, orderedTurnNumbers, onReport, canReport = true }: TurnCardProps) {
  const [showAINote, setShowAINote] = useState(false)
  const t = useTranslations("debate.turnCard")

  if (turn.is_system) {
    return <SystemEventCard turn={turn} />
  }

  const isA = turn.side === "a"

  const cardBg = isA ? "bg-[var(--side-a-bg)]" : "bg-[var(--side-b-bg)]"
  const borderColor = isA ? "border-[var(--side-a-border)]" : "border-[var(--side-b-border)]"
  const nameColor = isA ? "text-[var(--side-a)]" : "text-[var(--side-b)]"
  const badgeBg = isA ? "bg-[var(--side-a)]" : "bg-[var(--side-b)]"

  const dn = displayNumber ?? turn.turn_number
  const idx = orderedTurnNumbers.indexOf(turn.turn_number)
  const prevTurnNumber = idx > 0 ? orderedTurnNumbers[idx - 1] : undefined
  const nextTurnNumber = idx < orderedTurnNumbers.length - 1 ? orderedTurnNumbers[idx + 1] : undefined

  const scrollToTurn = (turnNumber: number) => {
    const el = document.getElementById(`turn-${turnNumber}`)
    if (el) el.scrollIntoView({ behavior: "smooth", block: "start" })
  }

  return (
    <div id={`turn-${turn.turn_number}`} className={`rounded-lg border ${borderColor} ${cardBg} transition-all duration-200`}>
      <div className="flex items-center gap-2 sm:gap-3 px-4 sm:px-5 py-3">
        <button
          onClick={onToggle}
          className="flex items-center gap-2 sm:gap-3 flex-1 min-w-0 text-left hover:bg-black/[0.02] -m-1 p-1 rounded transition-colors"
        >
          <span
            className={`w-6 h-6 rounded-full ${badgeBg} text-white text-[10px] font-bold flex items-center justify-center font-mono shrink-0`}
          >
            {isA ? "A" : "B"}
          </span>

          <div className="flex items-center gap-1.5 flex-wrap flex-1 min-w-0">
            <span className={`text-[13px] font-semibold font-sans ${nameColor}`}>
              {displayName}
            </span>
            <span className="text-[var(--border-strong)] text-[11px]">&middot;</span>
            <span className="text-[var(--text-muted)] text-[11px] font-sans">
              {t("turn", { number: dn })}
            </span>
            <span className="text-[var(--border-strong)] text-[11px] hidden sm:inline">&middot;</span>
            <span className="text-[var(--text-muted)] text-[11px] font-sans hidden sm:inline">
              {formatDate(turn.created_at)}
            </span>
          </div>
        </button>

        <div className="flex items-center gap-0.5 shrink-0">
          {turn.ai_assisted && (
            <button
              onClick={() => setShowAINote((current) => !current)}
              className="h-6 px-2 flex items-center gap-1 rounded border border-[var(--border-default)] text-[var(--text-secondary)] hover:text-[var(--text-primary)] hover:bg-black/[0.04] transition-colors"
              title={t("aiAssistedTitle")}
            >
              <Robot size={12} />
              <span className="text-[10px] font-medium font-sans hidden sm:inline">{t("aiAssisted")}</span>
            </button>
          )}
          {onReport && canReport && !turn.hidden && (
            <button
              onClick={() => onReport(turn.id)}
              className="w-6 h-6 flex items-center justify-center rounded text-[var(--text-muted)] hover:text-[var(--error)] hover:bg-black/[0.04] transition-colors"
              title={t("reportTurn")}
            >
              <Flag size={13} />
            </button>
          )}
          <button
            onClick={() => prevTurnNumber != null && scrollToTurn(prevTurnNumber)}
            className={`w-6 h-6 flex items-center justify-center rounded transition-colors ${
              prevTurnNumber != null
                ? "text-[var(--text-muted)] hover:text-[var(--text-secondary)] hover:bg-black/[0.04]"
                : "text-[var(--border-default)] cursor-default"
            }`}
            title={prevTurnNumber != null ? t("previous") : undefined}
            disabled={prevTurnNumber == null}
          >
            <CaretLeft size={14} />
          </button>
          <button
            onClick={() => nextTurnNumber != null && scrollToTurn(nextTurnNumber)}
            className={`w-6 h-6 flex items-center justify-center rounded transition-colors ${
              nextTurnNumber != null
                ? "text-[var(--text-muted)] hover:text-[var(--text-secondary)] hover:bg-black/[0.04]"
                : "text-[var(--border-default)] cursor-default"
            }`}
            title={nextTurnNumber != null ? t("next") : undefined}
            disabled={nextTurnNumber == null}
          >
            <CaretRight size={14} />
          </button>
          <button
            onClick={onToggle}
            className="w-6 h-6 flex items-center justify-center rounded text-[var(--text-muted)] hover:text-[var(--text-secondary)] hover:bg-black/[0.04] transition-colors"
            title={isCollapsed ? t("expand") : t("collapse")}
          >
            <CaretDown size={16} className={`transition-transform duration-200 ${isCollapsed ? "-rotate-90" : ""}`} />
          </button>
        </div>
      </div>

      {!isCollapsed && (
        <div className="px-4 sm:px-5 pb-4 sm:pb-5 pt-0">
          {turn.ai_assisted && showAINote && turn.ai_note && (
            <div className="mb-3 rounded-md border border-[var(--border-default)] bg-[var(--bg-primary)] px-3 py-2">
              <div className="flex items-center gap-1.5 text-[var(--text-secondary)] text-[11px] font-medium font-sans mb-1">
                <Robot size={12} />
                {t("aiNote")}
              </div>
              <p className="text-[12px] text-[var(--text-secondary)] font-sans whitespace-pre-wrap">
                {turn.ai_note}
              </p>
            </div>
          )}
          {turn.hidden ? (
            <div className="rounded-md border border-[var(--border-default)] bg-[var(--bg-surface)] px-3 py-2 text-[13px] text-[var(--text-secondary)] font-sans">
              {t("flagged")}
            </div>
          ) : (
            <div className="text-[var(--text-primary)] text-sm sm:text-[15px] leading-[1.75] font-sans whitespace-pre-wrap">
              {turn.content}
            </div>
          )}
          <div className="mt-2 text-[var(--text-muted)] text-[11px] font-sans sm:hidden">
            {formatDate(turn.created_at)}
          </div>
        </div>
      )}
    </div>
  )
}

function SystemEventCard({ turn }: { turn: DebateTurn }) {
  return (
    <div
      id={`turn-${turn.turn_number}`}
      className="flex items-center gap-3 py-2 px-4"
    >
      <div className="flex-1 h-px bg-[var(--border-default)]" />
      <div className="flex items-center gap-1.5 text-[var(--text-muted)]">
        <Info size={14} className="shrink-0" />
        <span className="text-[12px] font-medium font-sans">
          {turn.content}
        </span>
        <span className="text-[var(--border-strong)] text-[11px]">&middot;</span>
        <span className="text-[11px] font-sans">
          {formatDate(turn.created_at)}
        </span>
      </div>
      <div className="flex-1 h-px bg-[var(--border-default)]" />
    </div>
  )
}
