"use client"

import Link from "next/link"
import { CalendarDots, ChartBar, FloppyDisk, PencilSimple, Prohibit, Trophy, UserCircle } from "@phosphor-icons/react"
import { MoreHorizontal } from "lucide-react"
import type { DebateFeedItem } from "@/lib/debates"
import type { ChallengeLobbyEntry } from "@/lib/debates"
import type { UserProfile } from "@/lib/users"
import type { UserCapability, UserEnforcementAction, UserEnforcementActionType } from "@/lib/moderation"
import DebateCard from "@/components/debate-card"
import LobbyEntryCard from "@/components/lobby-entry-card"
import { formatTimeAgo } from "@/lib/utils"
import { Button } from "@/components/ui/button"
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@/components/ui/dropdown-menu"

export const enforcementActions: Array<Exclude<UserEnforcementActionType, "revoke">> = ["warning", "restriction", "suspension", "ban"]
export const enforcementCapabilities: UserCapability[] = ["create_debate", "comment", "vote", "follow", "report", "invite"]
export const enforcementFilters = ["active", "expired", "revoked", "all"] as const

type Props = {
  bioDraft: string
  bioError: string | null
  blockActionError: string | null
  blockedByMe: boolean
  canModerateProfile: boolean
  debates: DebateFeedItem[]
  editingBio: boolean
  enforcementAction: Exclude<UserEnforcementActionType, "revoke">
  enforcementCapabilitiesSelected: UserCapability[]
  enforcementError: string | null
  enforcementExpiresAt: string
  enforcementFilter: "active" | "expired" | "revoked" | "all"
  enforcementItems: UserEnforcementAction[]
  enforcementLoading: boolean
  enforcementNote: string
  enforcementSubmitting: boolean
  handleCreateEnforcementAction: () => Promise<void>
  handleLoadMore: () => Promise<void>
  handleRevokeEnforcementAction: (action: UserEnforcementAction) => Promise<void>
  handleSaveBio: () => Promise<void>
  handleToggleBlock: () => Promise<void>
  handleToggleUpvote: (debateId: string) => Promise<void>
  isBlocking: boolean
  isOwnProfile: boolean
  loadEnforcementHistory: () => Promise<void>
  loadingMore: boolean
  lobbyEntry: ChallengeLobbyEntry | null
  page: number
  profile: UserProfile
  savingBio: boolean
  setBioDraft: (value: string) => void
  setBioError: (value: string | null) => void
  setEditingBio: (value: boolean) => void
  setEnforcementAction: (value: Exclude<UserEnforcementActionType, "revoke">) => void
  setEnforcementCapabilitiesSelected: (value: UserCapability[] | ((current: UserCapability[]) => UserCapability[])) => void
  setEnforcementExpiresAt: (value: string) => void
  setEnforcementFilter: (value: "active" | "expired" | "revoked" | "all") => void
  setEnforcementNote: (value: string) => void
  setShowModerationPanel: (value: boolean | ((current: boolean) => boolean)) => void
  showModerationPanel: boolean
  status: "loading" | "authenticated" | "unauthenticated"
  t: any
  totalPages: number
  userEmailVerified?: boolean
  visibleDebatesTotal: number
}

export function UserProfileView({
bioDraft,
  bioError,
  blockActionError,
  blockedByMe,
  canModerateProfile,
  debates,
  editingBio,
  enforcementAction,
  enforcementCapabilitiesSelected,
  enforcementError,
  enforcementExpiresAt,
  enforcementFilter,
  enforcementItems,
  enforcementLoading,
  enforcementNote,
  enforcementSubmitting,
  handleCreateEnforcementAction,
  handleLoadMore,
  handleRevokeEnforcementAction,
  handleSaveBio,
  handleToggleBlock,
  handleToggleUpvote,
  isBlocking,
  isOwnProfile,
  loadEnforcementHistory,
  loadingMore,
  lobbyEntry,
  page,
  profile,
  savingBio,
  setBioDraft,
  setBioError,
  setEditingBio,
  setEnforcementAction,
  setEnforcementCapabilitiesSelected,
  setEnforcementExpiresAt,
  setEnforcementFilter,
  setEnforcementNote,
  setShowModerationPanel,
  showModerationPanel,
  status,
  t,
  totalPages,
  userEmailVerified,
  visibleDebatesTotal
}: Props) {
  return (
    <div className="min-h-screen bg-[var(--bg-primary)]">
      <div className="max-w-[820px] mx-auto px-3 sm:px-6 py-5 sm:py-6">
        <div className="bg-[var(--bg-surface)] border border-[var(--border-default)] rounded-lg p-5 sm:p-6">
          <div className="flex items-start justify-between gap-3">
            <div className="flex-1 min-w-0">
              <div className="flex items-center gap-2">
                <UserCircle size={18} className="text-[var(--text-secondary)]" />
                <h1 className="text-[18px] sm:text-[20px] font-semibold text-[var(--text-primary)] font-sans">
                  @{profile.username}
                </h1>
              </div>
              {isOwnProfile && editingBio ? (
                <div className="mt-2">
                  <textarea
                    value={bioDraft}
                    onChange={(event) => setBioDraft(event.target.value)}
                    maxLength={200}
                    rows={3}
                    className="w-full rounded-md border border-[var(--border-default)] bg-[var(--bg-surface)] px-3 py-2 text-[14px] text-[var(--text-secondary)] font-sans"
                    placeholder={t("bio.placeholder")}
                  />
                  <div className="mt-1 flex items-center justify-between gap-2">
                    <span className="text-[11px] text-[var(--text-muted)] font-sans">{bioDraft.length}/200</span>
                    <div className="flex items-center gap-2">
                      <button
                        type="button"
                        onClick={() => {
                          setEditingBio(false)
                          setBioDraft(profile.bio ?? "")
                          setBioError(null)
                        }}
                        className="px-2.5 py-1 text-[12px] font-medium rounded-md border border-[var(--border-default)] text-[var(--text-secondary)] hover:bg-[var(--bg-surface-alt)]"
                      >
                        {t("bio.cancel")}
                      </button>
                      <button
                        type="button"
                        onClick={() => void handleSaveBio()}
                        disabled={savingBio}
                        className="px-2.5 py-1 text-[12px] font-medium rounded-md bg-[var(--text-primary)] text-[var(--bg-primary)] hover:opacity-90 disabled:opacity-60 inline-flex items-center gap-1.5"
                      >
                        <FloppyDisk size={13} />
                        {savingBio ? t("bio.saving") : t("bio.save")}
                      </button>
                    </div>
                  </div>
                  {bioError && <div className="mt-1 text-[11px] text-[var(--error)] font-sans">{bioError}</div>}
                </div>
              ) : (
                <>
                  {profile.bio ? (
                    <p className="mt-2 text-[14px] text-[var(--text-secondary)] font-sans leading-relaxed whitespace-pre-wrap">{profile.bio}</p>
                  ) : (
                    <p className="mt-2 text-[13px] text-[var(--text-muted)] font-sans">{t("bio.empty")}</p>
                  )}
                  {isOwnProfile && (
                    <button
                      type="button"
                      onClick={() => {
                        setEditingBio(true)
                        setBioDraft(profile.bio ?? "")
                        setBioError(null)
                      }}
                      className="mt-2 inline-flex items-center gap-1.5 text-[12px] font-medium text-[var(--text-secondary)] hover:text-[var(--text-primary)]"
                    >
                      <PencilSimple size={13} />
                      {t("bio.edit")}
                    </button>
                  )}
                </>
              )}
            </div>
            {status === "authenticated" && !isOwnProfile && (
              <div className="shrink-0 flex items-start gap-2">
                {userEmailVerified && !blockedByMe && !isBlocking ? (
                  <Button asChild size="sm">
                    <Link href={`/create?challenge=${encodeURIComponent(profile.username)}`}>
                      {t("actions.challenge")}
                    </Link>
                  </Button>
                ) : (
                  <Button size="sm" disabled title={blockedByMe ? t("actions.unblockToChallenge") : t("actions.verifyToChallenge")}>
                    {t("actions.challenge")}
                  </Button>
                )}
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Button
                      type="button"
                      variant="ghost"
                      size="icon"
                      aria-label={t("actions.options")}
                      disabled={isBlocking}
                    >
                      <MoreHorizontal className="size-4" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end" className="w-44">
                    <DropdownMenuItem
                      variant={blockedByMe ? "default" : "destructive"}
                      onClick={() => void handleToggleBlock()}
                      disabled={isBlocking}
                    >
                      <Prohibit className="size-4" />
                      {isBlocking ? t("blocked.updating") : blockedByMe ? t("blocked.unblock") : t("actions.block")}
                    </DropdownMenuItem>
                    {canModerateProfile && (
                      <DropdownMenuItem
                        onClick={() => setShowModerationPanel((current) => !current)}
                      >
                        <PencilSimple className="size-4" />
                        {showModerationPanel ? t("actions.hideModeration") : t("actions.moderate")}
                      </DropdownMenuItem>
                    )}
                  </DropdownMenuContent>
                </DropdownMenu>
                {blockActionError && <div className="mt-1 text-[11px] text-[var(--error)] font-sans">{blockActionError}</div>}
              </div>
            )}
          </div>

          {lobbyEntry && (
            <div className="mt-4">
              <LobbyEntryCard entry={lobbyEntry} isOwn={isOwnProfile} compact showAction={false} label={t("lobbyLabel")} />
            </div>
          )}

          <div className="mt-4 grid grid-cols-2 sm:grid-cols-4 gap-2">
            <div className="rounded-md border border-[var(--border-default)] bg-[var(--bg-surface-alt)] px-3 py-2">
              <div className="flex items-center gap-1.5 text-[11px] text-[var(--text-muted)] font-medium font-sans uppercase tracking-wide">
                <ChartBar size={13} />
                {t("stats.rating")}
              </div>
              <div className="mt-1 text-[16px] font-semibold text-[var(--text-primary)] font-sans">{profile.rating}</div>
            </div>

            <div className="rounded-md border border-[var(--border-default)] bg-[var(--bg-surface-alt)] px-3 py-2">
              <div className="flex items-center gap-1.5 text-[11px] text-[var(--text-muted)] font-medium font-sans uppercase tracking-wide">
                <Trophy size={13} />
                {t("stats.record")}
              </div>
              <div className="mt-1 text-[13px] font-medium text-[var(--text-primary)] font-sans">
                {t("stats.recordValue", { wins: profile.wins, losses: profile.losses, draws: profile.draws })}
              </div>
            </div>

            <div className="rounded-md border border-[var(--border-default)] bg-[var(--bg-surface-alt)] px-3 py-2">
              <div className="text-[11px] text-[var(--text-muted)] font-medium font-sans uppercase tracking-wide">{t("stats.debates")}</div>
              <div className="mt-1 text-[16px] font-semibold text-[var(--text-primary)] font-sans">{profile.debates_count}</div>
            </div>

            <div className="rounded-md border border-[var(--border-default)] bg-[var(--bg-surface-alt)] px-3 py-2">
              <div className="flex items-center gap-1.5 text-[11px] text-[var(--text-muted)] font-medium font-sans uppercase tracking-wide">
                <CalendarDots size={13} />
                {t("stats.joined")}
              </div>
              <div className="mt-1 text-[13px] font-medium text-[var(--text-primary)] font-sans">{formatTimeAgo(profile.created_at)}</div>
            </div>
          </div>

          {canModerateProfile && showModerationPanel && (
            <div className="mt-4 rounded-md border border-[var(--border-default)] bg-[var(--bg-surface-alt)] p-4">
              <h3 className="text-[14px] font-semibold text-[var(--text-primary)] font-sans">{t("moderation.title")}</h3>
              <p className="mt-1 text-[12px] text-[var(--text-secondary)] font-sans">{t("moderation.description")}</p>

              <div className="mt-3 grid gap-2 sm:grid-cols-3">
                <select
                  value={enforcementAction}
                  onChange={(event) => {
                    setEnforcementAction(event.target.value as Exclude<UserEnforcementActionType, "revoke">)
                    if (event.target.value !== "restriction") {
                      setEnforcementCapabilitiesSelected([])
                    }
                  }}
                  className="h-9 rounded-md border border-[var(--border-default)] bg-[var(--bg-surface)] px-3 text-[12px] text-[var(--text-primary)]"
                >
                  {enforcementActions.map((item) => (
                    <option key={item} value={item}>{t(`moderation.actions.${item}`)}</option>
                  ))}
                </select>
                {(enforcementAction === "restriction" || enforcementAction === "suspension") && (
                  <input
                    type="datetime-local"
                    value={enforcementExpiresAt}
                    onChange={(event) => setEnforcementExpiresAt(event.target.value)}
                    className="h-9 rounded-md border border-[var(--border-default)] bg-[var(--bg-surface)] px-3 text-[12px] text-[var(--text-primary)]"
                  />
                )}
                <button
                  type="button"
                  onClick={() => void loadEnforcementHistory()}
                  disabled={enforcementLoading}
                  className="h-9 rounded-md border border-[var(--border-default)] px-3 text-[12px] text-[var(--text-secondary)] hover:bg-[var(--bg-surface)]"
                >
                  {enforcementLoading ? t("moderation.loading") : t("moderation.refresh")}
                </button>
              </div>

              {enforcementAction === "restriction" && (
                <div className="mt-2 flex flex-wrap gap-2">
                  {enforcementCapabilities.map((capability) => {
                    const selected = enforcementCapabilitiesSelected.includes(capability)
                    return (
                      <button
                        key={capability}
                        type="button"
                        onClick={() => {
                          setEnforcementCapabilitiesSelected((current) =>
                            current.includes(capability)
                              ? current.filter((value) => value !== capability)
                              : [...current, capability]
                          )
                        }}
                        className={`h-7 rounded-md border px-2 text-[11px] ${selected
                          ? "border-[var(--text-primary)] text-[var(--text-primary)]"
                          : "border-[var(--border-default)] text-[var(--text-secondary)]"
                          }`}
                      >
                        {t(`moderation.capability.${capability}`)}
                      </button>
                    )
                  })}
                </div>
              )}

              <textarea
                value={enforcementNote}
                onChange={(event) => setEnforcementNote(event.target.value)}
                maxLength={500}
                rows={3}
                placeholder={t("moderation.notePlaceholder")}
                className="mt-2 w-full rounded-md border border-[var(--border-default)] bg-[var(--bg-surface)] px-3 py-2 text-[12px] text-[var(--text-primary)]"
              />
              <div className="mt-1 text-right text-[11px] text-[var(--text-muted)] font-sans">{enforcementNote.length}/500</div>

              <div className="mt-2 flex items-center gap-2">
                <Button
                  type="button"
                  onClick={() => void handleCreateEnforcementAction()}
                  disabled={enforcementSubmitting}
                  size="sm"
                >
                  {enforcementSubmitting ? t("moderation.saving") : t("moderation.create")}
                </Button>
                {enforcementError && <span className="text-[11px] text-[var(--error)] font-sans">{enforcementError}</span>}
              </div>

              <div className="mt-4">
                <div className="flex items-center justify-between gap-2">
                  <h4 className="text-[12px] font-medium text-[var(--text-primary)] font-sans">{t("moderation.history")}</h4>
                  <select
                    value={enforcementFilter}
                    onChange={(event) => setEnforcementFilter(event.target.value as "active" | "expired" | "revoked" | "all")}
                    className="h-7 rounded-md border border-[var(--border-default)] bg-[var(--bg-surface)] px-2 text-[11px] text-[var(--text-primary)]"
                  >
                    {enforcementFilters.map((filter) => (
                      <option key={filter} value={filter}>{t(`moderation.filters.${filter}`)}</option>
                    ))}
                  </select>
                </div>

                <div className="mt-2 space-y-2">
                  {enforcementItems.length === 0 ? (
                    <div className="text-[12px] text-[var(--text-secondary)] font-sans">{t("moderation.empty")}</div>
                  ) : (
                    enforcementItems.map((item) => (
                      <div key={item.id} className="rounded-md border border-[var(--border-default)] bg-[var(--bg-surface)] p-2">
                        <div className="flex items-center justify-between gap-2">
                          <div className="text-[12px] font-medium text-[var(--text-primary)] font-sans">
                            {item.action}{item.status ? ` • ${item.status}` : ""}
                          </div>
                          {(item.action === "restriction" || item.action === "suspension") && item.status === "active" && (
                            <button
                              type="button"
                              onClick={() => void handleRevokeEnforcementAction(item)}
                              disabled={enforcementSubmitting}
                              className="rounded-md border border-[var(--border-default)] px-2 py-1 text-[11px] text-[var(--text-secondary)] hover:bg-[var(--bg-surface-alt)]"
                            >
                              {t("moderation.revoke")}
                            </button>
                          )}
                        </div>
                        <div className="mt-1 text-[11px] text-[var(--text-secondary)] font-sans">
                          {item.created_by?.username ? t("moderation.byUser", { username: item.created_by.username }) : t("moderation.byModerator")}
                          {item.expires_at ? ` • ${t("moderation.until", { date: new Date(item.expires_at).toLocaleString() })}` : ""}
                        </div>
                        {item.capabilities.length > 0 && (
                          <div className="mt-1 text-[11px] text-[var(--text-secondary)] font-sans">{t("moderation.capabilities", { capabilities: item.capabilities.join(", ") })}</div>
                        )}
                        <div className="mt-1 text-[11px] text-[var(--text-secondary)] font-sans">{t("moderation.note", { note: item.note })}</div>
                      </div>
                    ))
                  )}
                </div>
              </div>
            </div>
          )}
        </div>

        <div className="mt-5">
          <div className="mb-3">
            <h2 className="text-[14px] font-semibold text-[var(--text-primary)] font-sans">{t("recent.title")}</h2>
            {!isOwnProfile && profile.debates_count !== visibleDebatesTotal && (
              <p className="mt-1 text-[12px] text-[var(--text-muted)] font-sans">
                {t("recent.showing", { visible: visibleDebatesTotal, total: profile.debates_count })}
              </p>
            )}
          </div>

          {debates.length === 0 ? (
            <div className="bg-[var(--bg-surface)] border border-[var(--border-default)] rounded-lg p-5 text-[13px] text-[var(--text-secondary)] font-sans">
              {t("recent.empty")}
            </div>
          ) : (
            <div className="flex flex-col gap-2">
              {debates.map((debate) => (
                <DebateCard
                  key={debate.id}
                  debate={debate}
                  onToggleUpvote={handleToggleUpvote}
                  showParticipantBadge={false}
                />
              ))}

              {page < totalPages && (
                <div className="pt-2 flex justify-center">
                  <button
                    type="button"
                    onClick={() => void handleLoadMore()}
                    disabled={loadingMore}
                    className="px-4 py-2 rounded-md text-[12px] font-medium font-sans text-[var(--text-secondary)] border border-[var(--border-default)] hover:bg-[var(--bg-surface-alt)] hover:text-[var(--text-primary)] disabled:opacity-60"
                  >
                    {loadingMore ? t("recent.loading") : t("recent.loadMore")}
                  </button>
                </div>
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
