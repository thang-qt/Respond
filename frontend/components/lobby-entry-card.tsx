"use client"

import Link from "next/link"
import { useTranslations } from "next-intl"
import type { ChallengeLobbyEntry } from "@/lib/debates"

interface Props {
  entry: ChallengeLobbyEntry
  /** If true, shows an "Edit" link instead of the Challenge CTA (own profile context). */
  isOwn?: boolean
  /**
   * If true, hides the @username + rating row — use when those details
   * are already displayed above the card (e.g. on the profile page).
   */
  compact?: boolean
  /** If false, hides the right-side action button. */
  showAction?: boolean
  /** Optional embedded title shown at the top of the card. */
  label?: string
}

export default function LobbyEntryCard({ entry, isOwn = false, compact = false, showAction = true, label }: Props) {
  const t = useTranslations("lobbyEntryCard")
  const record = t("record", { wins: entry.wins, losses: entry.losses, draws: entry.draws })

  return (
    <div className="bg-[var(--bg-surface)] border border-[var(--border-default)] rounded-lg p-4 hover:border-[var(--border-strong)] transition-colors">
      {label && (
        <div className="mb-2 text-[12px] font-semibold text-[var(--text-primary)] font-sans">{label}</div>
      )}
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0 flex-1">
          {!compact && (
            <div className="flex items-center gap-2 flex-wrap mb-1.5">
              <Link
                href={`/profile/${entry.username}`}
                className="text-[14px] font-semibold text-[var(--text-primary)] hover:underline font-sans"
              >
                @{entry.username}
              </Link>
              <span className="text-[11px] text-[var(--text-muted)] font-sans">
                {t("ratingRecord", { rating: entry.rating, record })}
              </span>
            </div>
          )}

          {entry.bio_note ? (
            <p className="text-[13px] text-[var(--text-secondary)] font-sans leading-relaxed">
              {entry.bio_note}
            </p>
          ) : (
            <p className="text-[13px] text-[var(--text-muted)] font-sans italic">
              {t("openAnyTopic")}
            </p>
          )}

          {entry.tags.length > 0 && (
            <div className="mt-2 flex flex-wrap gap-1">
              {entry.tags.map((tag) => (
                <span
                  key={tag.id}
                  className="px-2 py-0.5 rounded-full text-[11px] font-medium border border-[var(--border-default)] text-[var(--text-secondary)] bg-[var(--bg-surface-alt)]"
                >
                  {tag.name}
                </span>
              ))}
            </div>
          )}
        </div>

        {showAction && (
          <div className="shrink-0">
            {isOwn ? (
              <Link
                href="/challenges"
                className="px-3 py-1.5 rounded-md text-[12px] font-medium border border-[var(--border-default)] text-[var(--text-secondary)] hover:bg-[var(--bg-surface-alt)] hover:text-[var(--text-primary)] transition-colors"
              >
                {t("edit")}
              </Link>
            ) : (
              <Link
                href={`/create?challenge=${entry.username}`}
                className="px-3 py-1.5 rounded-md text-[12px] font-medium bg-[var(--text-primary)] text-[var(--bg-primary)] hover:opacity-90 transition-opacity"
              >
                {t("challenge")}
              </Link>
            )}
          </div>
        )}
      </div>
    </div>
  )
}
