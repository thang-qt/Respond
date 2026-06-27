import { Info } from "@phosphor-icons/react"
import { useTranslations } from "next-intl"
import type { DebateDetail, DebateEvent } from "@/lib/debates"

function sidePublicName(side: "a" | "b", debate: DebateDetail): string | null {
  const info = side === "a" ? debate.side_a : debate.side_b
  return info.user?.username ?? null
}

function eventSideLabel(side: DebateEvent["side"]): string | null {
  return side ? side.toUpperCase() : null
}

function winnerSideKey(winnerSide: string | null): "a" | "b" | null {
  const normalized = winnerSide?.toLowerCase()
  return normalized === "a" || normalized === "b" ? normalized : null
}

function formatEventContent(event: DebateEvent, debate: DebateDetail, t: ReturnType<typeof useTranslations<"debatePage">>): string {
  const sideLabel = eventSideLabel(event.side)
  const payload = event.payload_json ?? {}
  const winnerSide = typeof payload.winner_side === "string" ? payload.winner_side.toUpperCase() : null
  const turnLimit = typeof payload.turn_limit === "number" ? payload.turn_limit : null
  const actorAnonymousId = typeof payload.anonymous_id === "string" ? payload.anonymous_id : null

  switch (event.event_type) {
    case "seat_opened":
      if (actorAnonymousId) {
        return t("events.seatOpenedActor", { actor: actorAnonymousId, side: sideLabel ?? "?" })
      }
      return t("events.seatOpened", { side: sideLabel ?? "?" })
    case "replacement_joined":
      if (event.side && (event.side === "a" || event.side === "b")) {
        const currentAnon = event.side === "a" ? debate.side_a.anonymous_id : debate.side_b.anonymous_id
        const revealedName = sidePublicName(event.side, debate)
        if (revealedName && actorAnonymousId && currentAnon === actorAnonymousId) {
          return t("events.replacementJoinedRevealed", { name: revealedName, actor: actorAnonymousId, side: sideLabel ?? "?" })
        }
      }
      if (actorAnonymousId) {
        return t("events.replacementJoinedActor", { actor: actorAnonymousId, side: sideLabel ?? "?" })
      }
      return t("events.replacementJoined", { side: sideLabel ?? "?" })
    case "conceded": {
      const conceder = event.side ? (sidePublicName(event.side, debate) ?? `Side ${sideLabel ?? "?"}`) : `Side ${sideLabel ?? "?"}`
      const winnerKey = winnerSideKey(winnerSide)
      if (winnerKey) {
        const winner = sidePublicName(winnerKey, debate) ?? `${t("common.unknownSide")} ${winnerSide}`
        return t("events.concededWithWinner", { conceder, winner })
      }
      return t("events.conceded", { conceder })
    }
    case "draw_proposed":
      return t("events.drawProposed", { side: sideLabel ?? "?" })
    case "draw_declined":
      return t("events.drawDeclined", { side: sideLabel ?? "?" })
    case "draw_accepted":
      return t("events.drawAccepted", { side: sideLabel ?? "?" })
    case "extension_proposed":
      return turnLimit != null
        ? t("events.extensionProposedWithLimit", { limit: turnLimit })
        : t("events.extensionProposed")
    case "extension_accepted":
      return t("events.extensionAccepted", { side: sideLabel ?? "?" })
    case "extension_declined":
      return t("events.extensionDeclined", { side: sideLabel ?? "?" })
    case "walkover":
      return winnerSide
        ? t("events.walkoverWithWinner", { side: sideLabel ?? "?", winner: winnerSide })
        : t("events.walkover")
    case "replacement_expired":
      return winnerSide
        ? t("events.replacementExpiredWithWinner", { winner: winnerSide })
        : t("events.replacementExpired")
    case "extension_expired":
      return t("events.extensionExpired")
    case "legacy_system_turn":
      return typeof payload.content === "string" ? payload.content : t("events.systemUpdate")
    default:
      return t("events.systemUpdate")
  }
}

export function TimelineEventCard({ event, debate }: { event: DebateEvent; debate: DebateDetail }) {
  const t = useTranslations("debatePage")
  return (
    <div className="flex items-center gap-3 py-2 px-4">
      <div className="flex-1 h-px bg-[var(--border-default)]" />
      <div className="flex items-center gap-1.5 text-[var(--text-muted)]">
        <Info size={14} className="shrink-0" />
        <span className="text-[12px] font-medium font-sans">{formatEventContent(event, debate, t)}</span>
      </div>
      <div className="flex-1 h-px bg-[var(--border-default)]" />
    </div>
  )
}
