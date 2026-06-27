"use client"

import { ChatCenteredText, Info } from "@phosphor-icons/react"
import { useTranslations } from "next-intl"
import type { DebateTurn } from "@/lib/debates"

interface TurnNavigatorProps {
  turns: DebateTurn[]
  collapsedTurns: Set<number>
  onExpandAll: () => void
  onCollapseAll: () => void
  showDiscussionLink: boolean
  sideALabel: string
  sideBLabel: string
  displayNumberMap: Map<number, number>
}

export function TurnNavigator({
  turns,
  collapsedTurns,
  onExpandAll,
  onCollapseAll,
  showDiscussionLink,
  sideALabel,
  sideBLabel,
  displayNumberMap,
}: TurnNavigatorProps) {
  const t = useTranslations("debate.turnNavigator")
  const scrollToTurn = (turnNumber: number) => {
    const el = document.getElementById(`turn-${turnNumber}`)
    if (el) el.scrollIntoView({ behavior: "smooth", block: "start" })
  }

  const scrollToDiscussion = () => {
    const el = document.getElementById("discussion-section")
    if (el) el.scrollIntoView({ behavior: "smooth", block: "start" })
  }

  const regularTurns = turns.filter((t) => !t.is_system)
  const allCollapsed = collapsedTurns.size === regularTurns.length
  const allExpanded = collapsedTurns.size === 0

  return (
    <div className="flex items-center gap-2 flex-wrap">
      <div className="flex items-center gap-1 flex-wrap flex-1">
        {turns.map((turn) => {
          if (turn.is_system) {
            return (
              <button
                key={turn.id}
                onClick={() => scrollToTurn(turn.turn_number)}
                className="w-7 h-7 rounded-md text-[11px] flex items-center justify-center transition-colors bg-[var(--bg-surface-alt)] text-[var(--text-muted)] hover:bg-[var(--border-default)]"
                title={turn.content}
              >
                <Info size={12} />
              </button>
            )
          }
          const isA = turn.side === "a"
          const dn = displayNumberMap.get(turn.turn_number) ?? turn.turn_number
          return (
            <button
              key={turn.id}
              onClick={() => scrollToTurn(turn.turn_number)}
              className={`w-7 h-7 rounded-md text-[11px] font-bold font-mono flex items-center justify-center transition-colors ${
                isA
                  ? "bg-[var(--side-a-bg)] text-[var(--side-a)] hover:bg-[var(--side-a-border)]"
                  : "bg-[var(--side-b-bg)] text-[var(--side-b)] hover:bg-[var(--side-b-border)]"
              } ${collapsedTurns.has(turn.turn_number) ? "opacity-40" : ""}`}
              title={t("jumpToTurn", { turn: dn, side: isA ? sideALabel : sideBLabel })}
            >
              {dn}
            </button>
          )
        })}

        {showDiscussionLink && (
          <button
            onClick={scrollToDiscussion}
            className="h-7 px-2 rounded-md text-[11px] font-medium font-sans flex items-center gap-1 transition-colors bg-[var(--bg-surface-alt)] text-[var(--text-secondary)] hover:bg-[var(--border-default)]"
            title={t("jumpToDiscussion")}
          >
            <ChatCenteredText size={12} />
            {t("discussion")}
          </button>
        )}
      </div>

      <button
        onClick={allExpanded ? onCollapseAll : onExpandAll}
        className="text-[11px] text-[var(--text-muted)] hover:text-[var(--text-secondary)] font-medium font-sans transition-colors whitespace-nowrap"
      >
        {allExpanded ? t("collapseAll") : t("expandAll")}
      </button>
    </div>
  )
}
