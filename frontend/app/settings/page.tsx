"use client"

import { useCallback, useEffect, useMemo, useState } from "react"
import { useRouter } from "next/navigation"
import { useLocale, useTranslations } from "next-intl"
import { setLocaleCookie } from "@/i18n/client-locale"
import type { AppLocale } from "@/i18n/config"
import { ApiError } from "@/lib/api"
import {
  createMyInvite,
  fetchMyInvites,
  fetchMyNotificationSettings,
  type InviteRecord,
  revokeMyInvite,
  resendVerificationEmail,
  updateMyEmail,
  updateMyNotificationSettings,
  updateMyPassword,
  updateMyProfileSettings,
  type NotificationSettings,
} from "@/lib/settings-api"
import { fetchMyBlockedUsers, unblockUser } from "@/lib/users-api"
import type { BlockedUser } from "@/lib/users"
import { useAuth } from "@/hooks/use-auth"
import {
  BlockedUsersSection,
  DEFAULT_NOTIFICATION_SETTINGS,
  EmailSettingsSection,
  InvitesSettingsSection,
  LanguageSettingsSection,
  NotificationSettingsSection,
  ProfileSettingsSection,
  ResourceLinksSection,
  SecuritySettingsSection,
  SettingsHeader,
} from "@/components/settings/settings-sections"

export default function SettingsPage() {
  const router = useRouter()
  const currentLocale = useLocale() as AppLocale
  const { user, status, refresh } = useAuth()
  const t = useTranslations("settings")

  const [bio, setBio] = useState("")
  const [defaultReveal, setDefaultReveal] = useState(false)
  const [profileSaving, setProfileSaving] = useState(false)
  const [profileMessage, setProfileMessage] = useState<string | null>(null)
  const [profileError, setProfileError] = useState<string | null>(null)
  const [locale, setLocale] = useState<AppLocale>(currentLocale)
  const [localeSaving, setLocaleSaving] = useState(false)
  const [localeMessage, setLocaleMessage] = useState<string | null>(null)
  const [localeError, setLocaleError] = useState<string | null>(null)

  const [notificationSettings, setNotificationSettings] = useState<NotificationSettings>(DEFAULT_NOTIFICATION_SETTINGS)
  const [notificationsLoading, setNotificationsLoading] = useState(true)
  const [notificationsSaving, setNotificationsSaving] = useState(false)
  const [notificationMessage, setNotificationMessage] = useState<string | null>(null)
  const [notificationError, setNotificationError] = useState<string | null>(null)

  const [currentPassword, setCurrentPassword] = useState("")
  const [newPassword, setNewPassword] = useState("")
  const [confirmNewPassword, setConfirmNewPassword] = useState("")
  const [passwordSaving, setPasswordSaving] = useState(false)
  const [passwordMessage, setPasswordMessage] = useState<string | null>(null)
  const [passwordError, setPasswordError] = useState<string | null>(null)

  const [newEmail, setNewEmail] = useState("")
  const [emailPassword, setEmailPassword] = useState("")
  const [emailSaving, setEmailSaving] = useState(false)
  const [emailMessage, setEmailMessage] = useState<string | null>(null)
  const [emailError, setEmailError] = useState<string | null>(null)
  const [resendSaving, setResendSaving] = useState(false)
  const [resendMessage, setResendMessage] = useState<string | null>(null)
  const [resendError, setResendError] = useState<string | null>(null)

  const [blockedUsers, setBlockedUsers] = useState<BlockedUser[]>([])
  const [blockedUsersLoading, setBlockedUsersLoading] = useState(true)
  const [blockedUsersError, setBlockedUsersError] = useState<string | null>(null)
  const [unblockingUsername, setUnblockingUsername] = useState<string | null>(null)

  const [inviteEmail, setInviteEmail] = useState("")
  const [inviteSaving, setInviteSaving] = useState(false)
  const [inviteMessage, setInviteMessage] = useState<string | null>(null)
  const [inviteError, setInviteError] = useState<string | null>(null)
  const [invites, setInvites] = useState<InviteRecord[]>([])
  const [invitesLoading, setInvitesLoading] = useState(true)
  const [invitesError, setInvitesError] = useState<string | null>(null)
  const [revokingInviteID, setRevokingInviteID] = useState<string | null>(null)

  useEffect(() => {
    if (status === "unauthenticated") {
      router.replace("/auth/login?redirect=/settings")
    }
  }, [router, status])

  useEffect(() => {
    if (!user) return
    setBio(user.bio ?? "")
    setDefaultReveal(Boolean(user.default_reveal))
    setLocale(user.locale ?? currentLocale)
    setNewEmail("")
  }, [currentLocale, user])

  const loadNotificationSettings = useCallback(async () => {
    if (status !== "authenticated") return
    setNotificationsLoading(true)
    setNotificationError(null)
    try {
      const res = await fetchMyNotificationSettings()
      setNotificationSettings(res.data)
    } catch (err) {
      setNotificationError(err instanceof Error ? err.message : t("notifications.loadError"))
    } finally {
      setNotificationsLoading(false)
    }
  }, [status])

  useEffect(() => {
    void loadNotificationSettings()
  }, [loadNotificationSettings])

  const loadBlockedUsers = useCallback(async () => {
    if (status !== "authenticated") return
    setBlockedUsersLoading(true)
    setBlockedUsersError(null)
    try {
      const res = await fetchMyBlockedUsers()
      setBlockedUsers(res.data ?? [])
    } catch (err) {
      setBlockedUsersError(err instanceof Error ? err.message : t("blocked.loadError"))
    } finally {
      setBlockedUsersLoading(false)
    }
  }, [status])

  useEffect(() => {
    void loadBlockedUsers()
  }, [loadBlockedUsers])

  const loadInvites = useCallback(async () => {
    if (status !== "authenticated") return
    setInvitesLoading(true)
    setInvitesError(null)
    try {
      const res = await fetchMyInvites({ status: "all", per_page: 20 })
      setInvites(res.data ?? [])
    } catch (err) {
      setInvitesError(err instanceof Error ? err.message : t("invites.loadError"))
    } finally {
      setInvitesLoading(false)
    }
  }, [status])

  useEffect(() => {
    void loadInvites()
  }, [loadInvites])

  const bioCount = useMemo(() => bio.length, [bio])

  const handleSaveProfile = useCallback(async () => {
    setProfileSaving(true)
    setProfileError(null)
    setProfileMessage(null)
    try {
      await updateMyProfileSettings({
        bio,
        default_reveal: defaultReveal,
      })
      await refresh()
      setProfileMessage(t("saved"))
    } catch (err) {
      setProfileError(err instanceof Error ? err.message : t("profile.saveError"))
    } finally {
      setProfileSaving(false)
    }
  }, [bio, defaultReveal, refresh])

  const handleSaveLocale = useCallback(async () => {
    setLocaleSaving(true)
    setLocaleError(null)
    setLocaleMessage(null)
    try {
      await updateMyProfileSettings({ locale })
      setLocaleCookie(locale)
      await refresh()
      router.refresh()
      setLocaleMessage(t("saved"))
    } catch (err) {
      setLocaleError(err instanceof Error ? err.message : t("language.saveError"))
    } finally {
      setLocaleSaving(false)
    }
  }, [locale, refresh, router])

  const handleSaveNotifications = useCallback(async () => {
    setNotificationsSaving(true)
    setNotificationError(null)
    setNotificationMessage(null)
    try {
      const res = await updateMyNotificationSettings(notificationSettings)
      setNotificationSettings(res.data)
      setNotificationMessage(t("saved"))
    } catch (err) {
      setNotificationError(err instanceof Error ? err.message : t("notifications.saveError"))
    } finally {
      setNotificationsSaving(false)
    }
  }, [notificationSettings])

  const handleChangePassword = useCallback(async () => {
    if (newPassword.length < 8) {
      setPasswordError(t("security.tooShort"))
      return
    }
    if (newPassword !== confirmNewPassword) {
      setPasswordError(t("security.mismatch"))
      return
    }
    setPasswordSaving(true)
    setPasswordError(null)
    setPasswordMessage(null)
    try {
      await updateMyPassword({
        current_password: currentPassword,
        new_password: newPassword,
      })
      setCurrentPassword("")
      setNewPassword("")
      setConfirmNewPassword("")
      setPasswordMessage(t("security.changed"))
    } catch (err) {
      setPasswordError(err instanceof Error ? err.message : t("security.changeError"))
    } finally {
      setPasswordSaving(false)
    }
  }, [confirmNewPassword, currentPassword, newPassword])

  const handleChangeEmail = useCallback(async () => {
    setEmailSaving(true)
    setEmailError(null)
    setEmailMessage(null)
    try {
      await updateMyEmail({ email: newEmail, password: emailPassword })
      setNewEmail("")
      setEmailPassword("")
      await refresh()
      setEmailMessage(t("email.updated"))
    } catch (err) {
      if (err instanceof ApiError && err.code === "AUTH_EMAIL_TAKEN") {
        setEmailError(t("email.taken"))
      } else {
        setEmailError(err instanceof Error ? err.message : t("email.changeError"))
      }
    } finally {
      setEmailSaving(false)
    }
  }, [newEmail, emailPassword, refresh])

  const handleResendVerification = useCallback(async () => {
    setResendSaving(true)
    setResendMessage(null)
    setResendError(null)
    try {
      await resendVerificationEmail()
      setResendMessage(t("email.verificationSent"))
    } catch (err) {
      if (err instanceof ApiError && err.code === "AUTH_ALREADY_VERIFIED") {
        await refresh()
        setResendMessage(t("email.alreadyVerified"))
      } else {
        setResendError(err instanceof Error ? err.message : t("email.resendError"))
      }
    } finally {
      setResendSaving(false)
    }
  }, [refresh])

  const handleUnblock = useCallback(async (username: string) => {
    setUnblockingUsername(username)
    setBlockedUsersError(null)
    try {
      await unblockUser(username)
      setBlockedUsers((prev) => prev.filter((entry) => entry.username.toLowerCase() !== username.toLowerCase()))
    } catch (err) {
      setBlockedUsersError(err instanceof Error ? err.message : t("blocked.unblockError"))
    } finally {
      setUnblockingUsername(null)
    }
  }, [])

  const handleCreateInvite = useCallback(async () => {
    setInviteSaving(true)
    setInviteError(null)
    setInviteMessage(null)
    try {
      await createMyInvite({ email: inviteEmail.trim() })
      setInviteEmail("")
      setInviteMessage(t("invites.sent"))
      await loadInvites()
    } catch (err) {
      setInviteError(err instanceof Error ? err.message : t("invites.sendError"))
    } finally {
      setInviteSaving(false)
    }
  }, [inviteEmail, loadInvites])

  const handleRevokeInvite = useCallback(async (inviteID: string) => {
    setRevokingInviteID(inviteID)
    setInvitesError(null)
    try {
      await revokeMyInvite(inviteID)
      await loadInvites()
    } catch (err) {
      setInvitesError(err instanceof Error ? err.message : t("invites.revokeError"))
    } finally {
      setRevokingInviteID(null)
    }
  }, [loadInvites])

  if (status !== "authenticated") {
    return (
      <div className="min-h-screen bg-[var(--bg-primary)]">
        <div className="max-w-[820px] mx-auto px-3 sm:px-6 py-8">
          <div className="text-[var(--text-muted)] text-[13px] font-sans">{t("loading")}</div>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-[var(--bg-primary)]">
      <div className="max-w-[820px] mx-auto px-3 sm:px-6 py-5 sm:py-6 space-y-5">
        <SettingsHeader t={t as any} />
        <ProfileSettingsSection bio={bio} bioCount={bioCount} defaultReveal={defaultReveal} error={profileError} message={profileMessage} saving={profileSaving} t={t as any} onBioChange={setBio} onDefaultRevealChange={setDefaultReveal} onSave={() => void handleSaveProfile()} />
        <LanguageSettingsSection currentUserLocale={user?.locale as AppLocale | undefined} error={localeError} locale={locale} message={localeMessage} saving={localeSaving} t={t as any} onLocaleChange={setLocale} onSave={() => void handleSaveLocale()} />
        <NotificationSettingsSection error={notificationError} loading={notificationsLoading} message={notificationMessage} saving={notificationsSaving} settings={notificationSettings} t={t as any} onSave={() => void handleSaveNotifications()} onSettingsChange={setNotificationSettings} />
        <SecuritySettingsSection confirmNewPassword={confirmNewPassword} currentPassword={currentPassword} error={passwordError} message={passwordMessage} newPassword={newPassword} saving={passwordSaving} t={t as any} onConfirmNewPasswordChange={setConfirmNewPassword} onCurrentPasswordChange={setCurrentPassword} onNewPasswordChange={setNewPassword} onSave={() => void handleChangePassword()} />
        <BlockedUsersSection blockedUsers={blockedUsers} error={blockedUsersError} loading={blockedUsersLoading} t={t as any} unblockingUsername={unblockingUsername} onUnblock={(username) => void handleUnblock(username)} />
        <EmailSettingsSection email={user?.email} emailError={emailError} emailMessage={emailMessage} emailPassword={emailPassword} emailSaving={emailSaving} emailVerified={user?.email_verified} newEmail={newEmail} resendError={resendError} resendMessage={resendMessage} resendSaving={resendSaving} t={t as any} onEmailPasswordChange={setEmailPassword} onNewEmailChange={setNewEmail} onResend={() => void handleResendVerification()} onSave={() => void handleChangeEmail()} />
        <InvitesSettingsSection inviteEmail={inviteEmail} inviteError={inviteError} inviteMessage={inviteMessage} inviteSaving={inviteSaving} invites={invites} invitesError={invitesError} invitesLoading={invitesLoading} revokingInviteID={revokingInviteID} t={t as any} onCreate={() => void handleCreateInvite()} onInviteEmailChange={setInviteEmail} onRevoke={(inviteID) => void handleRevokeInvite(inviteID)} />
        <ResourceLinksSection t={t as any} />
      </div>
    </div>
  )
}