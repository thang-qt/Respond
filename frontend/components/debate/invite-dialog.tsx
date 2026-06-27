import Link from "next/link"
import { useEffect, useState } from "react"
import { useTranslations } from "next-intl"
import { ApiError } from "@/lib/api"
import type { UserSearchProfile } from "@/lib/users"
import { fetchExploreUsers, searchUsers } from "@/lib/users-api"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"

type InviteDialogProps = {
  open: boolean
  currentUsername?: string | null
  actionError: string | null
  invitedUsernames: Set<string>
  invitingUsername: string | null
  onActionErrorChange: (message: string | null) => void
  onInviteUser: (username: string) => void
  onOpenChange: (open: boolean) => void
}

export function InviteDialog({
  open,
  currentUsername,
  actionError,
  invitedUsernames,
  invitingUsername,
  onActionErrorChange,
  onInviteUser,
  onOpenChange,
}: InviteDialogProps) {
  const t = useTranslations("debatePage")
  const [query, setQuery] = useState("")
  const [searchResults, setSearchResults] = useState<UserSearchProfile[]>([])
  const [searchLoading, setSearchLoading] = useState(false)
  const [searchError, setSearchError] = useState<string | null>(null)

  useEffect(() => {
    if (!open) {
      setQuery("")
      setSearchResults([])
      setSearchLoading(false)
      setSearchError(null)
      return
    }

    const trimmedQuery = query.trim()
    let active = true
    setSearchLoading(true)
    setSearchError(null)

    const timeoutID = window.setTimeout(() => {
      void (async () => {
        try {
          const res = trimmedQuery
            ? await searchUsers(trimmedQuery, { page: 1, perPage: 12 })
            : await fetchExploreUsers({ context: "invite", page: 1, perPage: 12 })
          if (!active) return
          const rows = Array.isArray(res.data) ? res.data : []
          const me = currentUsername?.toLowerCase() ?? ""
          setSearchResults(rows.filter((candidate) => candidate.username.toLowerCase() !== me))
        } catch (err) {
          if (!active) return
          setSearchError(err instanceof ApiError ? err.message : "Could not search users.")
          setSearchResults([])
        } finally {
          if (active) {
            setSearchLoading(false)
          }
        }
      })()
    }, 250)

    return () => {
      active = false
      window.clearTimeout(timeoutID)
    }
  }, [open, query, currentUsername])

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[620px] bg-[var(--bg-surface)] border border-[var(--border-default)]">
        <DialogHeader>
          <DialogTitle className="text-[var(--text-primary)] font-sans">{t("invite.title")}</DialogTitle>
          <DialogDescription className="text-[var(--text-secondary)] font-sans">
            {t("invite.description")}
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-3">
          <input
            type="text"
            value={query}
            onChange={(event) => {
              setQuery(event.target.value)
              if (actionError) onActionErrorChange(null)
            }}
            placeholder={t("invite.placeholder")}
            maxLength={50}
            className="h-10 w-full rounded-md border border-[var(--border-default)] bg-[var(--bg-primary)] px-3 text-[14px] text-[var(--text-primary)] placeholder:text-[var(--text-muted)] font-sans focus:outline-none focus:ring-2 focus:ring-[var(--text-primary)]/20"
          />

          {actionError && (
            <div className="rounded-md border border-[var(--error)] bg-[var(--error-light)] px-3 py-2 text-[12px] text-[var(--error)]">
              {actionError}
            </div>
          )}
          <div className="max-h-[360px] overflow-y-auto rounded-md border border-[var(--border-subtle)] bg-[var(--bg-primary)]">
            {searchLoading ? (
              <div className="px-4 py-8 text-center text-[12px] text-[var(--text-muted)] font-sans">
                {query.trim() ? t("invite.searching") : t("invite.loadingRecommended")}
              </div>
            ) : searchError ? (
              <div className="px-4 py-8 text-center text-[12px] text-[var(--error)] font-sans">{searchError}</div>
            ) : searchResults.length < 1 ? (
              <div className="px-4 py-8 text-center text-[12px] text-[var(--text-muted)] font-sans">
                {query.trim() ? t("invite.noMatches") : t("invite.noRecommendations")}
              </div>
            ) : (
              <div className="divide-y divide-[var(--border-subtle)]">
                {searchResults.map((candidate) => (
                  <InviteCandidateRow
                    key={candidate.username}
                    candidate={candidate}
                    invited={invitedUsernames.has(candidate.username.toLowerCase())}
                    inviting={invitingUsername === candidate.username}
                    disabled={invitingUsername !== null || invitedUsernames.has(candidate.username.toLowerCase())}
                    onInvite={() => onInviteUser(candidate.username)}
                  />
                ))}
              </div>
            )}
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}

function InviteCandidateRow({
  candidate,
  invited,
  inviting,
  disabled,
  onInvite,
}: {
  candidate: UserSearchProfile
  invited: boolean
  inviting: boolean
  disabled: boolean
  onInvite: () => void
}) {
  const t = useTranslations("debatePage")
  return (
    <div className="flex items-center justify-between gap-3 px-3 py-3">
      <div className="min-w-0">
        <Link
          href={`/profile/${encodeURIComponent(candidate.username)}`}
          className="text-[13px] font-semibold text-[var(--text-primary)] hover:underline underline-offset-2"
        >
          @{candidate.username}
        </Link>
        <div className="mt-0.5 text-[11px] text-[var(--text-secondary)] font-sans">
          {t("invite.stats", {
            rating: candidate.rating,
            wins: candidate.wins,
            losses: candidate.losses,
            draws: candidate.draws,
            debates: candidate.debates_count,
          })}
        </div>
        {candidate.bio ? (
          <div className="mt-1 text-[11px] text-[var(--text-muted)] font-sans">{candidate.bio}</div>
        ) : null}
        {candidate.shared_tags && candidate.shared_tags.length > 0 ? (
          <div className="mt-1.5 flex flex-wrap gap-1">
            {candidate.shared_tags.slice(0, 3).map((slug) => (
              <span
                key={`${candidate.username}-${slug}`}
                className="rounded-sm bg-[var(--bg-surface-alt)] px-1.5 py-0.5 text-[10px] uppercase tracking-[0.04em] text-[var(--text-secondary)]"
              >
                {slug}
              </span>
            ))}
          </div>
        ) : null}
      </div>

      <button
        type="button"
        onClick={onInvite}
        disabled={disabled}
        className="h-8 shrink-0 rounded-md bg-[var(--text-primary)] px-3 text-[11px] font-medium text-[var(--bg-primary)] disabled:opacity-60"
      >
        {inviting ? t("invite.sending") : invited ? t("invite.invited") : t("invite.invite")}
      </button>
    </div>
  )
}
