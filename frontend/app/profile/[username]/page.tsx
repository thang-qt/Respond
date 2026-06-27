"use client"

import { useCallback, useEffect, useState } from "react"
import Link from "next/link"
import { useParams } from "next/navigation"
import { useTranslations } from "next-intl"
import { Prohibit } from "@phosphor-icons/react"
import { ApiError } from "@/lib/api"
import type { DebateFeedItem } from "@/lib/debates"
import type { UserProfile } from "@/lib/users"
import { blockUser, fetchMyBlockedUsers, fetchUserDebates, fetchUserProfile, unblockUser, updateMyProfile } from "@/lib/users-api"
import type { UserCapability, UserEnforcementAction, UserEnforcementActionType } from "@/lib/moderation"
import { UserProfileView, enforcementActions, enforcementCapabilities, enforcementFilters } from "@/components/profile/user-profile-view"
import { createAdminUserEnforcementAction, listAdminUserEnforcementActions, revokeAdminUserEnforcementAction } from "@/lib/moderation-api"
import { toggleDebateVote, getUserLobbyEntry } from "@/lib/debates-api"
import type { ChallengeLobbyEntry } from "@/lib/debates"
import { useAuth } from "@/hooks/use-auth"
import { NotFoundState } from "@/components/not-found-state"
import { Button } from "@/components/ui/button"

const PER_PAGE = 20

export default function UserProfilePage() {
  const params = useParams<{ username: string }>()
  const username = decodeURIComponent(params.username)
  const { user, status } = useAuth()
  const t = useTranslations("profile")

  const [profile, setProfile] = useState<UserProfile | null>(null)
  const [debates, setDebates] = useState<DebateFeedItem[]>([])
  const [loading, setLoading] = useState(true)
  const [loadingMore, setLoadingMore] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [userNotFound, setUserNotFound] = useState(false)
  const [profileBlocked, setProfileBlocked] = useState(false)
  const [blockedByMe, setBlockedByMe] = useState(false)
  const [isBlocking, setIsBlocking] = useState(false)
  const [blockActionError, setBlockActionError] = useState<string | null>(null)
  const [page, setPage] = useState(1)
  const [totalPages, setTotalPages] = useState(1)
  const [visibleDebatesTotal, setVisibleDebatesTotal] = useState(0)
  const [editingBio, setEditingBio] = useState(false)
  const [bioDraft, setBioDraft] = useState("")
  const [savingBio, setSavingBio] = useState(false)
  const [bioError, setBioError] = useState<string | null>(null)
  const [showModerationPanel, setShowModerationPanel] = useState(false)
  const [enforcementAction, setEnforcementAction] = useState<Exclude<UserEnforcementActionType, "revoke">>("warning")
  const [enforcementCapabilitiesSelected, setEnforcementCapabilitiesSelected] = useState<UserCapability[]>([])
  const [enforcementExpiresAt, setEnforcementExpiresAt] = useState("")
  const [enforcementNote, setEnforcementNote] = useState("")
  const [enforcementItems, setEnforcementItems] = useState<UserEnforcementAction[]>([])
  const [enforcementFilter, setEnforcementFilter] = useState<"active" | "expired" | "revoked" | "all">("active")
  const [enforcementLoading, setEnforcementLoading] = useState(false)
  const [enforcementSubmitting, setEnforcementSubmitting] = useState(false)
  const [enforcementError, setEnforcementError] = useState<string | null>(null)

  // Lobby entry state
  const [lobbyEntry, setLobbyEntry] = useState<ChallengeLobbyEntry | null>(null)

  const loadProfileAndDebates = useCallback(async () => {
    setLoading(true)
    setError(null)
    setUserNotFound(false)
    setProfileBlocked(false)
    setBlockActionError(null)

    try {
      const [profileRes, debatesRes] = await Promise.all([
        fetchUserProfile(username),
        fetchUserDebates(username, { page: 1, perPage: PER_PAGE }),
      ])

      setProfile(profileRes.data)
      setBioDraft(profileRes.data.bio ?? "")
      setDebates(debatesRes.data ?? [])
      setPage(debatesRes.meta?.page ?? 1)
      setTotalPages(debatesRes.meta?.total_pages ?? 1)
      setVisibleDebatesTotal(debatesRes.meta?.total ?? 0)

      // Load lobby entry (silently; 404 means none)
      getUserLobbyEntry(username)
        .then((res) => setLobbyEntry(res.data))
        .catch(() => setLobbyEntry(null))

      if (status === "authenticated") {
        try {
          const blockedRes = await fetchMyBlockedUsers()
          const blocked = blockedRes.data.some((entry) => entry.username.toLowerCase() === username.toLowerCase())
          setBlockedByMe(blocked)
        } catch {
          setBlockedByMe(false)
        }
      } else {
        setBlockedByMe(false)
      }
    } catch (err) {
      if (err instanceof ApiError && err.code === "USER_NOT_FOUND") {
        setUserNotFound(true)
        setError(null)
      } else if (err instanceof ApiError && err.code === "USER_HIDDEN_BY_BLOCK") {
        setProfileBlocked(true)
        setError(null)
        if (status === "authenticated") {
          try {
            const blockedRes = await fetchMyBlockedUsers()
            const blocked = blockedRes.data.some((entry) => entry.username.toLowerCase() === username.toLowerCase())
            setBlockedByMe(blocked)
          } catch {
            setBlockedByMe(false)
          }
        } else {
          setBlockedByMe(false)
        }
      } else if (err instanceof Error) {
        setError(err.message)
        setBlockedByMe(false)
      } else {
        setError(t("errors.load"))
        setBlockedByMe(false)
      }
      setProfile(null)
      setDebates([])
      setPage(1)
      setTotalPages(1)
      setVisibleDebatesTotal(0)
    } finally {
      setLoading(false)
    }
  }, [status, username])

  const handleToggleBlock = useCallback(async () => {
    if (status !== "authenticated") return

    setIsBlocking(true)
    setBlockActionError(null)
    try {
      if (blockedByMe) {
        await unblockUser(username)
        setBlockedByMe(false)
        await loadProfileAndDebates()
      } else {
        await blockUser(username)
        setBlockedByMe(true)
        setProfileBlocked(true)
        setProfile(null)
        setDebates([])
        setVisibleDebatesTotal(0)
      }
    } catch (err) {
      setBlockActionError(err instanceof Error ? err.message : t("errors.block"))
    } finally {
      setIsBlocking(false)
    }
  }, [blockedByMe, loadProfileAndDebates, status, username])

  useEffect(() => {
    if (status === "loading") return
    void loadProfileAndDebates()
  }, [loadProfileAndDebates, status])

  const handleLoadMore = useCallback(async () => {
    if (loadingMore) return
    if (page >= totalPages) return

    setLoadingMore(true)
    try {
      const nextPage = page + 1
      const res = await fetchUserDebates(username, { page: nextPage, perPage: PER_PAGE })
      setDebates((prev) => [...prev, ...(res.data ?? [])])
      setPage(res.meta?.page ?? nextPage)
      setTotalPages(res.meta?.total_pages ?? totalPages)
      setVisibleDebatesTotal(res.meta?.total ?? visibleDebatesTotal)
    } catch (err) {
      setError(err instanceof Error ? err.message : t("errors.loadMore"))
    } finally {
      setLoadingMore(false)
    }
  }, [loadingMore, page, totalPages, username, visibleDebatesTotal])

  const handleToggleUpvote = useCallback(async (debateId: string) => {
    const res = await toggleDebateVote(debateId)
    setDebates((prev) =>
      prev.map((debate) =>
        debate.id === res.data.debate_id
          ? { ...debate, upvote_count: res.data.upvote_count, viewer_has_upvoted: res.data.voted }
          : debate
      )
    )
  }, [])

  const isOwnProfile = Boolean(user && profile && user.username.toLowerCase() === profile.username.toLowerCase())
  const canModerateProfile = Boolean(user && profile?.id && !isOwnProfile && (user.role === "moderator" || user.role === "admin"))

  const loadEnforcementHistory = useCallback(async () => {
    if (!canModerateProfile || !profile?.id) return
    setEnforcementLoading(true)
    setEnforcementError(null)
    try {
      const res = await listAdminUserEnforcementActions({
        user_id: profile.id,
        status: enforcementFilter,
        page: 1,
        per_page: 50,
      })
      setEnforcementItems(res.data)
    } catch (err) {
      setEnforcementError(err instanceof Error ? err.message : t("errors.enforcementHistory"))
    } finally {
      setEnforcementLoading(false)
    }
  }, [canModerateProfile, enforcementFilter, profile])

  useEffect(() => {
    if (!showModerationPanel) return
    void loadEnforcementHistory()
  }, [showModerationPanel, loadEnforcementHistory])

  useEffect(() => {
    if (!showModerationPanel || !canModerateProfile) return
    void loadEnforcementHistory()
  }, [enforcementFilter, showModerationPanel, canModerateProfile, loadEnforcementHistory])

  const handleCreateEnforcementAction = useCallback(async () => {
    if (!profile?.id || !canModerateProfile || enforcementSubmitting) return
    const trimmedNote = enforcementNote.trim()
    if (!trimmedNote) {
      setEnforcementError(t("errors.moderatorNoteRequired"))
      return
    }
    if (trimmedNote.length > 500) {
      setEnforcementError(t("errors.moderatorNoteLength"))
      return
    }
    if (enforcementAction === "restriction" && enforcementCapabilitiesSelected.length === 0) {
      setEnforcementError(t("errors.capabilityRequired"))
      return
    }

    const expiresAt = enforcementExpiresAt ? new Date(enforcementExpiresAt) : null
    if (expiresAt && Number.isNaN(expiresAt.getTime())) {
      setEnforcementError(t("errors.invalidExpiration"))
      return
    }

    setEnforcementSubmitting(true)
    setEnforcementError(null)
    try {
      await createAdminUserEnforcementAction(profile.id, {
        action: enforcementAction,
        capabilities: enforcementAction === "restriction" ? enforcementCapabilitiesSelected : undefined,
        expires_at: (enforcementAction === "restriction" || enforcementAction === "suspension") && expiresAt ? expiresAt.toISOString() : undefined,
        note: trimmedNote,
      })
      setEnforcementNote("")
      await loadEnforcementHistory()
    } catch (err) {
      setEnforcementError(err instanceof Error ? err.message : t("errors.createEnforcement"))
    } finally {
      setEnforcementSubmitting(false)
    }
  }, [profile, canModerateProfile, enforcementSubmitting, enforcementNote, enforcementAction, enforcementCapabilitiesSelected, enforcementExpiresAt, loadEnforcementHistory])

  const handleRevokeEnforcementAction = useCallback(async (action: UserEnforcementAction) => {
    if (!profile?.id || !canModerateProfile || enforcementSubmitting) return
    const note = window.prompt(t("moderation.prompt"))?.trim()
    if (!note) {
      setEnforcementError(t("errors.moderatorNoteRevoke"))
      return
    }
    if (note.length > 500) {
      setEnforcementError(t("errors.moderatorNoteLength"))
      return
    }
    setEnforcementSubmitting(true)
    setEnforcementError(null)
    try {
      await revokeAdminUserEnforcementAction(profile.id, action.id, note)
      await loadEnforcementHistory()
    } catch (err) {
      setEnforcementError(err instanceof Error ? err.message : t("errors.revokeEnforcement"))
    } finally {
      setEnforcementSubmitting(false)
    }
  }, [profile, canModerateProfile, enforcementSubmitting, loadEnforcementHistory])

  const handleSaveBio = useCallback(async () => {
    const nextBio = bioDraft
    if (nextBio.length > 200) {
      setBioError(t("errors.bioLength"))
      return
    }

    setSavingBio(true)
    setBioError(null)
    try {
      const res = await updateMyProfile({ bio: nextBio })
      setProfile((prev) => (prev ? { ...prev, bio: res.data.bio } : prev))
      setBioDraft(res.data.bio)
      setEditingBio(false)
    } catch (err) {
      setBioError(err instanceof Error ? err.message : t("errors.bioUpdate"))
    } finally {
      setSavingBio(false)
    }
  }, [bioDraft])

  if (loading) {
    return (
      <div className="min-h-screen bg-[var(--bg-primary)]">
        <div className="max-w-[820px] mx-auto px-3 sm:px-6 py-8">
          <div className="text-[var(--text-muted)] text-[13px] font-sans">{t("loading")}</div>
        </div>
      </div>
    )
  }

  if (error || !profile) {
    if (userNotFound) {
      return (
        <NotFoundState
          title={t("notFound.title")}
          description={t("notFound.description", { username })}
          actions={[
            { href: "/", label: t("notFound.home"), primary: true },
            { href: "/search", label: t("notFound.search") },
          ]}
        />
      )
    }
    if (profileBlocked) {
      return (
        <div className="min-h-screen bg-[var(--bg-primary)]">
          <div className="max-w-[820px] mx-auto px-3 sm:px-6 py-12">
            <div className="bg-[var(--bg-surface)] border border-[var(--border-default)] rounded-lg p-5">
              <div className="inline-flex items-center gap-2 text-[15px] font-semibold text-[var(--text-primary)] font-sans">
                <Prohibit size={16} />
                {t("blocked.title")}
              </div>
              <div className="mt-2 text-[13px] text-[var(--text-secondary)] font-sans">
                {t("blocked.description")}
              </div>
              <div className="mt-4 flex flex-wrap items-center gap-2">
                {blockedByMe && (
                  <Button
                    onClick={() => void handleToggleBlock()}
                    disabled={isBlocking}
                    size="sm"
                  >
                    {isBlocking ? t("blocked.updating") : t("blocked.unblock")}
                  </Button>
                )}
                <Button asChild variant="outline" size="sm">
                  <Link href="/settings">{t("blocked.manage")}</Link>
                </Button>
              </div>
              {blockActionError && <div className="mt-2 text-[12px] text-[var(--error)] font-sans">{blockActionError}</div>}
            </div>
          </div>
        </div>
      )
    }
    return (
      <div className="min-h-screen bg-[var(--bg-primary)]">
        <div className="max-w-[820px] mx-auto px-3 sm:px-6 py-12">
          <div className="bg-[var(--bg-surface)] border border-[var(--border-default)] rounded-lg p-5">
            <div className="text-[15px] font-semibold text-[var(--text-primary)] font-sans mb-1">{t("loadErrorTitle")}</div>
            <div className="text-[13px] text-[var(--text-secondary)] font-sans">{error ?? t("errors.unknown")}</div>
          </div>
        </div>
      </div>
    )
  }

  return (
    <UserProfileView
      bioDraft={bioDraft}
      bioError={bioError}
      blockActionError={blockActionError}
      blockedByMe={blockedByMe}
      canModerateProfile={canModerateProfile}
      debates={debates}
      editingBio={editingBio}
      enforcementAction={enforcementAction}
      enforcementCapabilitiesSelected={enforcementCapabilitiesSelected}
      enforcementError={enforcementError}
      enforcementExpiresAt={enforcementExpiresAt}
      enforcementFilter={enforcementFilter}
      enforcementItems={enforcementItems}
      enforcementLoading={enforcementLoading}
      enforcementNote={enforcementNote}
      enforcementSubmitting={enforcementSubmitting}
      handleCreateEnforcementAction={handleCreateEnforcementAction}
      handleLoadMore={handleLoadMore}
      handleRevokeEnforcementAction={handleRevokeEnforcementAction}
      handleSaveBio={handleSaveBio}
      handleToggleBlock={handleToggleBlock}
      handleToggleUpvote={handleToggleUpvote}
      isBlocking={isBlocking}
      isOwnProfile={isOwnProfile}
      loadEnforcementHistory={loadEnforcementHistory}
      loadingMore={loadingMore}
      lobbyEntry={lobbyEntry}
      page={page}
      profile={profile}
      savingBio={savingBio}
      setBioDraft={setBioDraft}
      setBioError={setBioError}
      setEditingBio={setEditingBio}
      setEnforcementAction={setEnforcementAction}
      setEnforcementCapabilitiesSelected={setEnforcementCapabilitiesSelected}
      setEnforcementExpiresAt={setEnforcementExpiresAt}
      setEnforcementFilter={setEnforcementFilter}
      setEnforcementNote={setEnforcementNote}
      setShowModerationPanel={setShowModerationPanel}
      showModerationPanel={showModerationPanel}
      status={status}
      t={t}
      totalPages={totalPages}
      userEmailVerified={user?.email_verified}
      visibleDebatesTotal={visibleDebatesTotal}
    />
  )
}
