"use client"

import { useEffect, useState } from "react"
import Link from "next/link"
import { usePathname } from "next/navigation"
import { useTranslations } from "next-intl"
import BrandLogo from "@/components/brand-logo"
import { siteConfig } from "@/lib/config"
import { useAuth } from "@/hooks/use-auth"
import { useNotifications } from "@/hooks/use-notifications"
import NotificationsDropdown from "@/components/notifications-dropdown"
import { NavItem, ThemeMenuItem, ThemeToggle } from "@/components/app-shell-nav"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import {
  Bell,
  CaretDown,
  CaretDoubleLeft,
  Compass,
  DoorOpen,
  Gear,
  House,
  MagnifyingGlass,
  Plus,
  SignOut,
  SquaresFour,
  ShieldCheck,
  User,
  UserCircle,
  X,
} from "@phosphor-icons/react"
import { MessageSquareQuote } from "lucide-react"

export function MobileSidebar({
  onClose,
  redirectTo,
  loginUrl,
  status,
  username,
  userRole,
  pendingChallengesCount,
  onLogout,
}: {
  onClose: () => void
  redirectTo: string
  loginUrl: string
  status: "loading" | "authenticated" | "unauthenticated"
  username: string | null
  userRole: "user" | "moderator" | "admin" | null
  pendingChallengesCount: number
  onLogout: () => void
}) {
  const pathname = usePathname()
  const isActive = (path: string) => pathname === path
  const isExploreDebatesRoute = pathname.startsWith("/explore/debates")
  const isExploreUsersRoute = pathname.startsWith("/explore/users")
  const isExploreRoute = pathname.startsWith("/explore")
  const isSearchRoute = pathname === "/search"
  const isTagRoute = pathname.startsWith("/tags") || pathname.startsWith("/categories")
  const isChallengeRoute = pathname === "/challenges"
  const isAuthed = status === "authenticated"
  const canModerate = userRole === "moderator" || userRole === "admin"
  const { logout } = useAuth()
  const { unreadCount } = useNotifications()
  const t = useTranslations("shell")
  const tCommon = useTranslations("common")
  const [exploreOpen, setExploreOpen] = useState(isExploreRoute)

  useEffect(() => {
    if (isExploreRoute) setExploreOpen(true)
  }, [isExploreRoute])

  return (
    <aside
      className="absolute top-0 left-0 h-full w-[240px] bg-[var(--bg-primary)] border-r border-[var(--border-subtle)] flex flex-col animate-in slide-in-from-left-2 fade-in duration-200 motion-reduce:animate-none"
      role="dialog"
      aria-modal="true"
    >
      <div className="h-14 px-3 flex items-center justify-between shrink-0">
        <Link href="/" className="flex items-center gap-2 text-[var(--text-primary)] truncate" onClick={onClose}>
          <BrandLogo size={20} />
          <span className="text-[15px] font-semibold font-sans">{siteConfig.name}</span>
        </Link>
        <button
          type="button"
          onClick={onClose}
          className="w-8 h-8 rounded-md flex items-center justify-center text-[var(--text-secondary)] hover:bg-[var(--border-subtle)] hover:text-[var(--text-primary)] transition-colors"
          aria-label={t("closeMenu")}
        >
          <X size={18} />
        </button>
      </div>

      <nav className="flex-1 overflow-y-auto overflow-x-hidden px-2 flex flex-col justify-center">
        <div className="flex flex-col gap-0.5">
          <NavItem href="/" icon={<House size={18} />} label={t("nav.home")} active={isActive("/")} collapsed={false} onClick={onClose} />
          <button
            type="button"
            onClick={() => setExploreOpen((prev) => !prev)}
            className={`flex w-full items-center gap-2.5 px-2.5 py-2 rounded-md transition-colors font-sans text-[13px] font-medium ${
              isExploreRoute
                ? "bg-[var(--bg-surface)] shadow-[0px_0px_0px_0.75px_var(--border-default)_inset] text-[var(--text-primary)]"
                : "text-[var(--text-secondary)] hover:bg-[var(--border-subtle)] hover:text-[var(--text-primary)]"
            }`}
          >
            <Compass size={18} className="shrink-0" />
            <span className="flex-1 truncate text-left">{t("nav.explore")}</span>
            <CaretDown size={14} className={`transition-transform ${exploreOpen ? "rotate-180" : "rotate-0"}`} />
          </button>
          {exploreOpen && (
            <div className="ml-4 mt-0.5 flex flex-col gap-0.5 border-l border-[var(--border-subtle)] pl-1.5">
              <NavItem href="/explore/debates/hot" icon={<MessageSquareQuote size={16} />} label={t("nav.debates")} active={isExploreDebatesRoute} collapsed={false} onClick={onClose} />
              <NavItem href="/explore/users" icon={<UserCircle size={16} />} label={t("nav.users")} active={isExploreUsersRoute} collapsed={false} onClick={onClose} />
            </div>
          )}
          <NavItem href="/search" icon={<MagnifyingGlass size={18} />} label={t("nav.search")} active={isSearchRoute} collapsed={false} onClick={onClose} />
          <NavItem href="/tags" icon={<SquaresFour size={18} />} label={t("nav.tags")} active={isTagRoute} collapsed={false} onClick={onClose} />
          {isAuthed && (
            <NavItem
              href="/challenges"
              icon={<DoorOpen size={18} />}
              label={t("nav.challenges")}
              active={isChallengeRoute}
              collapsed={false}
              badge={pendingChallengesCount > 0 ? (pendingChallengesCount > 99 ? "99+" : `${pendingChallengesCount}`) : undefined}
              badgeColor="bg-[var(--text-primary)] text-[var(--bg-primary)]"
              onClick={onClose}
            />
          )}
          {canModerate && (
            <NavItem href="/admin/moderation" icon={<ShieldCheck size={18} />} label={t("nav.moderation")} active={pathname.startsWith("/admin/moderation")} collapsed={false} onClick={onClose} />
          )}

        </div>
      </nav>

      <div className="shrink-0 px-2 py-2 flex flex-col gap-1">
        {isAuthed && (
          <Link
            href="/create"
            onClick={onClose}
            className="flex items-center gap-2.5 px-2.5 py-2 rounded-md transition-colors font-sans text-[13px] font-medium bg-[var(--text-primary)] text-[var(--bg-primary)] hover:opacity-90"
          >
            <Plus size={18} className="shrink-0" />
            <span className="truncate">{t("createDebate")}</span>
          </Link>
        )}
        {!isAuthed && <ThemeToggle onClick={onClose} />}
        {isAuthed ? (
          <div className="flex flex-col gap-1">
            <Link
              href="/notifications"
              onClick={onClose}
              className="relative flex items-center gap-2.5 px-2.5 py-2 rounded-md transition-colors font-sans text-[13px] font-medium text-[var(--text-secondary)] hover:bg-[var(--border-subtle)] hover:text-[var(--text-primary)]"
            >
              <Bell size={18} className="shrink-0" />
              <span className="truncate">{t("nav.notifications")}</span>
              {unreadCount > 0 && (
                <span className="ml-auto min-w-4 h-4 px-1 rounded-full bg-[var(--text-primary)] text-[10px] text-[var(--bg-primary)] font-semibold flex items-center justify-center">
                  {unreadCount > 9 ? "9+" : unreadCount}
                </span>
              )}
            </Link>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <button
                  type="button"
                  className="flex items-center gap-2.5 px-2.5 py-2 rounded-md transition-colors text-[var(--text-secondary)] hover:bg-[var(--border-subtle)] hover:text-[var(--text-primary)]"
                >
                  <span className="w-6 h-6 rounded-full bg-[var(--bg-surface-alt)] text-[var(--text-primary)] text-[11px] font-semibold flex items-center justify-center shrink-0">
                    {username ? username[0]?.toUpperCase() : "U"}
                  </span>
                  <span className="text-[13px] font-medium font-sans truncate">{username ?? tCommon("profile")}</span>
                </button>
              </DropdownMenuTrigger>
              <DropdownMenuContent
                align="start"
                className="w-56 bg-[var(--bg-surface)] border-[var(--border-default)] text-[var(--text-primary)] shadow-[0px_10px_30px_rgba(15,12,10,0.12)] p-2"
              >
                <DropdownMenuLabel className="px-2 py-2 font-sans text-[12px] text-[var(--text-secondary)]">
                  <div className="flex items-center gap-2">
                    <span className="w-8 h-8 rounded-full bg-[var(--bg-surface-alt)] text-[var(--text-primary)] text-[12px] font-semibold flex items-center justify-center shrink-0">
                      {username ? username[0]?.toUpperCase() : "U"}
                    </span>
                    <div className="flex flex-col leading-tight">
                      <span className="text-[13px] font-semibold text-[var(--text-primary)]">
                        {username ?? tCommon("profile")}
                      </span>
                      <span className="text-[11px] text-[var(--text-muted)]">{t("signedIn")}</span>
                    </div>
                  </div>
                </DropdownMenuLabel>
                <DropdownMenuSeparator className="bg-[var(--border-subtle)]" />
                <DropdownMenuItem
                  asChild
                  className="font-sans text-[13px] text-[var(--text-secondary)] focus:bg-[var(--bg-surface-alt)] focus:text-[var(--text-primary)]"
                >
                  <Link href={username ? `/profile/${encodeURIComponent(username)}` : "/profile"} onClick={onClose} className="flex items-center gap-2">
                    <UserCircle size={16} />
                    {tCommon("viewProfile")}
                  </Link>
                </DropdownMenuItem>
                <DropdownMenuItem
                  asChild
                  className="font-sans text-[13px] text-[var(--text-secondary)] focus:bg-[var(--bg-surface-alt)] focus:text-[var(--text-primary)]"
                >
                  <Link href="/settings" onClick={onClose} className="flex items-center gap-2">
                    <Gear size={16} />
                    {tCommon("settings")}
                  </Link>
                </DropdownMenuItem>
                <ThemeMenuItem onSelect={onClose} />
                <DropdownMenuItem
                  onClick={() => {
                    logout()
                    onClose()
                    onLogout()
                  }}
                  className="font-sans text-[13px] text-[var(--error)] focus:bg-[var(--error-light)] focus:text-[var(--error)]"
                >
                  <SignOut size={16} />
                  {tCommon("signOut")}
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        ) : (
          <Link
            href={`/auth/login?redirect=${encodeURIComponent(redirectTo)}`}
            onClick={onClose}
            className="flex items-center gap-2.5 px-2.5 py-2 rounded-md transition-colors text-[var(--text-secondary)] hover:bg-[var(--border-subtle)] hover:text-[var(--text-primary)]"
          >
            <User size={18} className="shrink-0" />
            <span className="text-[13px] font-medium font-sans truncate">{tCommon("signIn")}</span>
          </Link>
        )}
      </div>
    </aside>
  )
}

export function DesktopSidebar({
  collapsed,
  setCollapsed,
  redirectTo,
  loginUrl,
  status,
  username,
  userRole,
  pendingChallengesCount,
  onLogout,
}: {
  collapsed: boolean
  setCollapsed: (v: boolean) => void
  redirectTo: string
  loginUrl: string
  status: "loading" | "authenticated" | "unauthenticated"
  username: string | null
  userRole: "user" | "moderator" | "admin" | null
  pendingChallengesCount: number
  onLogout: () => void
}) {
  const pathname = usePathname()
  const isActive = (path: string) => pathname === path
  const isExploreDebatesRoute = pathname.startsWith("/explore/debates")
  const isExploreUsersRoute = pathname.startsWith("/explore/users")
  const isExploreRoute = pathname.startsWith("/explore")
  const isSearchRoute = pathname === "/search"
  const isTagRoute = pathname.startsWith("/tags") || pathname.startsWith("/categories")
  const isChallengeRoute = pathname === "/challenges"
  const isAuthed = status === "authenticated"
  const canModerate = userRole === "moderator" || userRole === "admin"
  const { logout } = useAuth()
  const t = useTranslations("shell")
  const tCommon = useTranslations("common")
  const [exploreOpen, setExploreOpen] = useState(isExploreRoute)

  useEffect(() => {
    if (isExploreRoute) setExploreOpen(true)
  }, [isExploreRoute])

  return (
    <aside
      className={`hidden lg:flex fixed top-0 left-0 h-full z-40 bg-[var(--bg-primary)] border-r border-[var(--border-subtle)] flex-col transition-all duration-200 ease-in-out ${collapsed ? "w-[60px]" : "w-[240px]"
        }`}
    >
      {/* Header: Logo + collapse toggle */}
      <div className="h-14 px-3 flex items-center justify-between shrink-0">
        {collapsed ? (
          <button
            onClick={() => setCollapsed(false)}
            className="w-8 h-8 rounded-md flex items-center justify-center text-[var(--text-secondary)] hover:bg-[var(--border-subtle)] hover:text-[var(--text-primary)] transition-colors mx-auto"
            title={t("expandSidebar")}
          >
            <BrandLogo size={18} />
          </button>
        ) : (
          <>
            <Link href="/" className="flex items-center gap-2 text-[var(--text-primary)] truncate">
              <BrandLogo size={20} />
              <span className="text-[15px] font-semibold font-sans">{siteConfig.name}</span>
            </Link>
            <button
              onClick={() => setCollapsed(true)}
              className="w-8 h-8 rounded-md flex items-center justify-center text-[var(--text-secondary)] hover:bg-[var(--border-subtle)] hover:text-[var(--text-primary)] transition-colors shrink-0"
              title={t("collapseSidebar")}
            >
              <CaretDoubleLeft size={18} />
            </button>
          </>
        )}
      </div>

      {/* Navigation */}
      <nav className="flex-1 overflow-y-auto overflow-x-hidden px-2 flex flex-col justify-center">
        <div className="flex flex-col gap-0.5">
          <NavItem href="/" icon={<House size={18} />} label={t("nav.home")} active={isActive("/")} collapsed={collapsed} />
          {collapsed ? (
            <NavItem href="/explore/debates/hot" icon={<Compass size={18} />} label={t("nav.explore")} active={isExploreRoute} collapsed={collapsed} />
          ) : (
            <>
              <button
                type="button"
                onClick={() => setExploreOpen((prev) => !prev)}
                className={`flex w-full items-center gap-2.5 px-2.5 py-2 rounded-md transition-colors font-sans text-[13px] font-medium ${
                  isExploreRoute
                    ? "bg-[var(--bg-surface)] shadow-[0px_0px_0px_0.75px_var(--border-default)_inset] text-[var(--text-primary)]"
                    : "text-[var(--text-secondary)] hover:bg-[var(--border-subtle)] hover:text-[var(--text-primary)]"
                }`}
              >
                <Compass size={18} className="shrink-0" />
                <span className="flex-1 truncate text-left">{t("nav.explore")}</span>
                <CaretDown size={14} className={`transition-transform ${exploreOpen ? "rotate-180" : "rotate-0"}`} />
              </button>
              {exploreOpen && (
                <div className="ml-4 mt-0.5 flex flex-col gap-0.5 border-l border-[var(--border-subtle)] pl-1.5">
                  <NavItem href="/explore/debates/hot" icon={<MessageSquareQuote size={16} />} label={t("nav.debates")} active={isExploreDebatesRoute} collapsed={false} />
                  <NavItem href="/explore/users" icon={<UserCircle size={16} />} label={t("nav.users")} active={isExploreUsersRoute} collapsed={false} />
                </div>
              )}
            </>
          )}
          <NavItem href="/search" icon={<MagnifyingGlass size={18} />} label={t("nav.search")} active={isSearchRoute} collapsed={collapsed} />
          <NavItem href="/tags" icon={<SquaresFour size={18} />} label={t("nav.tags")} active={isTagRoute} collapsed={collapsed} />
          {isAuthed && (
            <NavItem
              href="/challenges"
              icon={<DoorOpen size={18} />}
              label={t("nav.challenges")}
              active={isChallengeRoute}
              collapsed={collapsed}
              badge={pendingChallengesCount > 0 ? (pendingChallengesCount > 99 ? "99+" : `${pendingChallengesCount}`) : undefined}
              badgeColor="bg-[var(--text-primary)] text-[var(--bg-primary)]"
            />
          )}
          {canModerate && (
            <NavItem href="/admin/moderation" icon={<ShieldCheck size={18} />} label={t("nav.moderation")} active={pathname.startsWith("/admin/moderation")} collapsed={collapsed} />
          )}

        </div>
      </nav>

      {/* Footer: Create + Theme + Sign in */}
      <div className="shrink-0 px-2 py-2 flex flex-col gap-1">
        {isAuthed && (
          <Link
            href="/create"
            className={`flex items-center gap-2.5 px-2.5 py-2 rounded-md transition-colors font-sans text-[13px] font-medium bg-[var(--text-primary)] text-[var(--bg-primary)] hover:opacity-90 ${collapsed ? "justify-center" : ""
              }`}
            title={collapsed ? t("createDebate") : undefined}
          >
            <Plus size={18} className="shrink-0" />
            {!collapsed && <span className="truncate">{t("createDebate")}</span>}
          </Link>
        )}
        {!isAuthed && <ThemeToggle collapsed={collapsed} />}
        {isAuthed ? (
          <div className="flex flex-col gap-1">
            <NotificationsDropdown collapsed={collapsed} />
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <button
                  type="button"
                  className={`flex items-center gap-2.5 px-2.5 py-2 rounded-md transition-colors text-[var(--text-secondary)] hover:bg-[var(--border-subtle)] hover:text-[var(--text-primary)] ${collapsed ? "justify-center" : ""
                    }`}
                  title={collapsed ? tCommon("profile") : undefined}
                >
                  <span className="w-6 h-6 rounded-full bg-[var(--bg-surface-alt)] text-[var(--text-primary)] text-[11px] font-semibold flex items-center justify-center shrink-0">
                    {username ? username[0]?.toUpperCase() : "U"}
                  </span>
                  {!collapsed && <span className="text-[13px] font-medium font-sans truncate">{username ?? tCommon("profile")}</span>}
                </button>
              </DropdownMenuTrigger>
              <DropdownMenuContent
                align="start"
                className="w-56 bg-[var(--bg-surface)] border-[var(--border-default)] text-[var(--text-primary)] shadow-[0px_10px_30px_rgba(15,12,10,0.12)] p-2"
              >
                <DropdownMenuLabel className="px-2 py-2 font-sans text-[12px] text-[var(--text-secondary)]">
                  <div className="flex items-center gap-2">
                    <span className="w-8 h-8 rounded-full bg-[var(--bg-surface-alt)] text-[var(--text-primary)] text-[12px] font-semibold flex items-center justify-center shrink-0">
                      {username ? username[0]?.toUpperCase() : "U"}
                    </span>
                    <div className="flex flex-col leading-tight">
                      <span className="text-[13px] font-semibold text-[var(--text-primary)]">
                        {username ?? tCommon("profile")}
                      </span>
                      <span className="text-[11px] text-[var(--text-muted)]">{t("signedIn")}</span>
                    </div>
                  </div>
                </DropdownMenuLabel>
                <DropdownMenuSeparator className="bg-[var(--border-subtle)]" />
                <DropdownMenuItem
                  asChild
                  className="font-sans text-[13px] text-[var(--text-secondary)] focus:bg-[var(--bg-surface-alt)] focus:text-[var(--text-primary)]"
                >
                  <Link href={username ? `/profile/${encodeURIComponent(username)}` : "/profile"} className="flex items-center gap-2">
                    <UserCircle size={16} />
                    {tCommon("viewProfile")}
                  </Link>
                </DropdownMenuItem>
                <DropdownMenuItem
                  asChild
                  className="font-sans text-[13px] text-[var(--text-secondary)] focus:bg-[var(--bg-surface-alt)] focus:text-[var(--text-primary)]"
                >
                  <Link href="/settings" className="flex items-center gap-2">
                    <Gear size={16} />
                    {tCommon("settings")}
                  </Link>
                </DropdownMenuItem>
                <ThemeMenuItem />
                <DropdownMenuItem
                  onClick={() => {
                    logout()
                    onLogout()
                  }}
                  className="font-sans text-[13px] text-[var(--error)] focus:bg-[var(--error-light)] focus:text-[var(--error)]"
                >
                  <SignOut size={16} />
                  {tCommon("signOut")}
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        ) : (
          <Link
            href={`/auth/login?redirect=${encodeURIComponent(redirectTo)}`}
            className={`flex items-center gap-2.5 px-2.5 py-2 rounded-md transition-colors text-[var(--text-secondary)] hover:bg-[var(--border-subtle)] hover:text-[var(--text-primary)] ${collapsed ? "justify-center" : ""
              }`}
            title={collapsed ? tCommon("signIn") : undefined}
          >
            <User size={18} className="shrink-0" />
            {!collapsed && <span className="text-[13px] font-medium font-sans truncate">{tCommon("signIn")}</span>}
          </Link>
        )}
      </div>
    </aside>
  )
}
