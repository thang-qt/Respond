"use client"

import { useTranslations } from "next-intl"

interface RevealPromptProps {
  anonymousId: string
  isSubmitting: boolean
  error: string | null
  onReveal: () => Promise<void>
  onStayAnonymous: () => Promise<void>
}

export function RevealPrompt({
  anonymousId,
  isSubmitting,
  error,
  onReveal,
  onStayAnonymous,
}: RevealPromptProps) {
  const t = useTranslations("debate.reveal")
  const tCommon = useTranslations("common")

  return (
    <div className="mt-6 mb-4 p-4 sm:p-5 bg-[var(--bg-surface)] border border-[var(--border-default)] rounded-lg">
      <div className="flex flex-col gap-2">
        <div className="text-[var(--text-primary)] text-sm font-semibold font-sans">
          {t("title")}
        </div>
        <div className="text-[var(--text-muted)] text-xs font-sans">
          {t.rich("body", {
            anonymousId: () => <span className="font-semibold text-[var(--text-secondary)]">{anonymousId}</span>,
          })}
        </div>
        {error && (
          <div className="mt-1 text-xs text-[var(--error)] font-sans">
            {error}
          </div>
        )}
        <div className="mt-2 flex flex-wrap items-center gap-2">
          <button
            className="h-8 px-4 rounded-full bg-[var(--text-primary)] text-[var(--bg-primary)] text-xs font-medium font-sans hover:opacity-90 transition-colors disabled:opacity-60"
            onClick={onReveal}
            disabled={isSubmitting}
          >
            {isSubmitting ? tCommon("saving") : t("reveal")}
          </button>
          <button
            className="h-8 px-4 rounded-full border border-[var(--border-default)] text-[var(--text-secondary)] text-xs font-medium font-sans hover:text-[var(--text-primary)] hover:border-[var(--border-strong)] transition-colors disabled:opacity-60"
            onClick={onStayAnonymous}
            disabled={isSubmitting}
          >
            {t("stayAnonymous")}
          </button>
        </div>
      </div>
    </div>
  )
}
