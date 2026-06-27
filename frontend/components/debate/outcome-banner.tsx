"use client"

import Link from "next/link"
import type { ReactNode } from "react"
import { useTranslations } from "next-intl"
import type { DebateDetail } from "@/lib/debates"
import { resolveUsername, resolveDisplayName } from "@/lib/debates"

interface OutcomeBannerProps {
  debate: DebateDetail
}

export function OutcomeBanner({ debate }: OutcomeBannerProps) {
  const t = useTranslations("debate.outcome")

  if (debate.status === "expired") {
    return (
      <div className="w-full px-5 py-3 border rounded-lg text-center text-sm font-medium font-sans bg-[var(--bg-surface-alt)] border-[var(--border-default)] text-[var(--text-secondary)]">
        {t("expired")}
      </div>
    )
  }

  if (debate.status !== "finished" || !debate.outcome) return null

  const sideAUsername = resolveUsername(debate.side_a)
  const sideBUsername = resolveUsername(debate.side_b)
  const sideADisplay = resolveDisplayName(debate.side_a, t("sideA"))
  const sideBDisplay = resolveDisplayName(debate.side_b, t("sideB"))
  const winnerDisplay = debate.winner_side === "a" ? sideADisplay : debate.winner_side === "b" ? sideBDisplay : null
  const winnerUsername = debate.winner_side === "a" ? sideAUsername : debate.winner_side === "b" ? sideBUsername : null

  const sideClass =
    debate.winner_side === "a"
      ? "bg-[var(--side-a-bg)] border-[var(--side-a-border)] text-[var(--side-a)]"
      : debate.winner_side === "b"
        ? "bg-[var(--side-b-bg)] border-[var(--side-b-border)] text-[var(--side-b)]"
        : "bg-[var(--bg-surface-alt)] border-[var(--border-default)] text-[var(--text-secondary)]"

  const outcomeClass: Record<string, string> = {
    concession: sideClass,
    walkover: sideClass,
    draw: "bg-[var(--warning-light)] border-[var(--warning)] text-[var(--warning)]",
    turn_limit: "bg-[var(--warning-light)] border-[var(--warning)] text-[var(--warning)]",
  }

  const sideADelta = debate.side_a_rating_delta
  const sideBDelta = debate.side_b_rating_delta
  const hasRatingChanges = typeof sideADelta === "number" || typeof sideBDelta === "number"

  let outcomeMessage: ReactNode
  if (debate.outcome === "draw") {
    outcomeMessage = t("draw")
  } else if (debate.outcome === "turn_limit") {
    outcomeMessage = t("turnLimit")
  } else if (debate.outcome === "concession") {
    outcomeMessage = winnerDisplay
      ? (
        <>
          {winnerUsername ? (
            <Link href={`/profile/${encodeURIComponent(winnerUsername)}`} className="hover:underline underline-offset-2">
              {winnerDisplay}
            </Link>
          ) : winnerDisplay}{t("winsByConcession")}
        </>
      )
      : t("endedByConcession")
  } else {
    outcomeMessage = winnerDisplay
      ? (
        <>
          {winnerUsername ? (
            <Link href={`/profile/${encodeURIComponent(winnerUsername)}`} className="hover:underline underline-offset-2">
              {winnerDisplay}
            </Link>
          ) : winnerDisplay}{t("winsByWalkover")}
        </>
      )
      : t("endedByWalkover")
  }

  return (
    <div className={`w-full px-5 py-3 border rounded-lg text-center font-sans ${outcomeClass[debate.outcome]}`}>
      <div className="text-sm font-medium">{outcomeMessage}</div>
      {hasRatingChanges && (
        <div className="mt-1 text-xs opacity-90">
          {t("rating", { sideA: formatDelta(sideADelta), sideB: formatDelta(sideBDelta) })}
        </div>
      )}
    </div>
  )
}

function formatDelta(value?: number | null) {
  if (typeof value !== "number") return "—"
  return value > 0 ? `+${value}` : `${value}`
}
