"use client"

import Link from "next/link"
import type { DebateDetail, DebateEvent, DebateTurn } from "@/lib/debates"
import { TurnNavigator } from "@/components/debate/turn-navigator"
import { TurnCard } from "@/components/debate/turn-card"
import { TimelineEventCard } from "@/components/debate/timeline-event-card"

type TimelineItem =
  | { kind: "turn"; id: string; turn: DebateTurn }
  | { kind: "event"; id: string; event: DebateEvent }

type Props = {
  argumentTurns: DebateTurn[]
  collapsedTurns: Set<number>
  debate: DebateDetail
  displayNumberMap: Map<number, number>
  onCollapseAll: () => void
  onExpandAll: () => void
  onReportTurn: (turnID: string) => void
  onToggleTurn: (turnNumber: number) => void
  orderedTurnNumbers: number[]
  sideADisplay: string
  sideAUsername: string | null
  sideBDisplay: string
  sideBUsername: string | null
  timelineItems: TimelineItem[]
}

function participantLabel(username: string | null, fallback: string) {
  if (!username) return fallback
  return (
    <Link
      href={`/profile/${encodeURIComponent(username)}`}
      onClick={(event) => event.stopPropagation()}
      className="hover:underline underline-offset-2"
    >
      {username}
    </Link>
  )
}

export function DebateTimelinePanel({
  argumentTurns,
  collapsedTurns,
  debate,
  displayNumberMap,
  onCollapseAll,
  onExpandAll,
  onReportTurn,
  onToggleTurn,
  orderedTurnNumbers,
  sideADisplay,
  sideAUsername,
  sideBDisplay,
  sideBUsername,
  timelineItems,
}: Props) {
  if (debate.turns.length === 0) return null

  return (
    <div className="flex flex-col gap-3">
      <TurnNavigator
        turns={argumentTurns}
        collapsedTurns={collapsedTurns}
        onExpandAll={onExpandAll}
        onCollapseAll={onCollapseAll}
        showDiscussionLink={debate.status === "finished"}
        sideALabel={sideADisplay}
        sideBLabel={sideBDisplay}
        displayNumberMap={displayNumberMap}
      />

      {timelineItems.map((item) => {
        if (item.kind === "turn") {
          const turn = item.turn
          return (
            <TurnCard
              key={item.id}
              turn={turn}
              displayName={turn.side === "a"
                ? participantLabel(sideAUsername, sideADisplay)
                : participantLabel(sideBUsername, sideBDisplay)}
              isCollapsed={collapsedTurns.has(turn.turn_number)}
              onToggle={() => onToggleTurn(turn.turn_number)}
              displayNumber={displayNumberMap.get(turn.turn_number)}
              orderedTurnNumbers={orderedTurnNumbers}
              canReport={!(debate.viewer?.is_participant && Boolean(debate.viewer.side) && debate.viewer.side === turn.side)}
              onReport={onReportTurn}
            />
          )
        }

        return <TimelineEventCard key={item.id} event={item.event} debate={debate} />
      })}
    </div>
  )
}
