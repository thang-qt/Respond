"use client"

import Link from "next/link"
import { useTranslations } from "next-intl"
import { useTheme } from "@/components/theme-provider"
import { DropdownMenuItem } from "@/components/ui/dropdown-menu"
import { Moon, Sun } from "@phosphor-icons/react"
import { Monitor } from "lucide-react"

export function NavItem({
  href,
  icon,
  label,
  active,
  collapsed,
  badge,
  badgeColor,
  highlight,
  onClick,
}: {
  href: string
  icon: React.ReactNode
  label: string
  active: boolean
  collapsed: boolean
  badge?: string
  badgeColor?: string
  highlight?: boolean
  onClick?: () => void
}) {
  const base = `flex items-center gap-2.5 px-2.5 py-2 rounded-md transition-colors font-sans text-[13px] font-medium ${collapsed ? "justify-center" : ""
    }`

  const stateClass = active
    ? "bg-[var(--bg-surface)] shadow-[0px_0px_0px_0.75px_var(--border-default)_inset] text-[var(--text-primary)]"
    : highlight
      ? "bg-[var(--text-primary)] text-[var(--bg-primary)] hover:opacity-90"
      : "text-[var(--text-secondary)] hover:bg-[var(--border-subtle)] hover:text-[var(--text-primary)]"

  return (
    <Link
      href={href}
      onClick={onClick}
      className={`${base} ${stateClass}`}
      title={collapsed ? label : undefined}
    >
      <span className="shrink-0">{icon}</span>
      {!collapsed && <span className="flex-1 truncate">{label}</span>}
      {!collapsed && badge && (
        <span className={`text-[10px] text-white font-bold px-1.5 py-0.5 rounded-full ${badgeColor}`}>
          {badge}
        </span>
      )}
    </Link>
  )
}

export function ThemeToggle({ collapsed, onClick }: { collapsed?: boolean; onClick?: () => void }) {
  const { theme, mounted, toggleTheme } = useTheme()
  const t = useTranslations("shell.theme")
  const currentLabel = theme === "system" ? t("followSystem") : theme === "dark" ? t("darkMode") : t("lightMode")
  const nextLabel = theme === "light" ? t("switchToDark") : theme === "dark" ? t("switchToSystem") : t("switchToLight")

  return (
    <button
      type="button"
      onClick={() => {
        toggleTheme()
        onClick?.()
      }}
      className={`flex items-center gap-2.5 px-2.5 py-2 rounded-md transition-colors font-sans text-[13px] font-medium text-[var(--text-secondary)] hover:bg-[var(--border-subtle)] hover:text-[var(--text-primary)] ${collapsed ? "justify-center" : ""
        }`}
      title={collapsed ? `${currentLabel} • ${nextLabel}` : undefined}
      aria-label={t("toggleLabel", { current: currentLabel, next: nextLabel })}
    >
      {mounted ? theme === "system" ? <Monitor size={18} /> : theme === "dark" ? <Moon size={18} /> : <Sun size={18} /> : <span className="inline-block w-[18px] h-[18px]" />}
      {!collapsed && <span className="truncate">{mounted ? currentLabel : "\u00A0"}</span>}
    </button>
  )
}

export function ThemeMenuItem({ onSelect }: { onSelect?: () => void }) {
  const { theme, toggleTheme } = useTheme()
  const t = useTranslations("shell.theme")
  const label = theme === "system" ? t("followSystem") : theme === "dark" ? t("darkMode") : t("lightMode")

  return (
    <DropdownMenuItem
      onClick={() => {
        toggleTheme()
        onSelect?.()
      }}
      className="font-sans text-[13px] text-[var(--text-secondary)] focus:bg-[var(--bg-surface-alt)] focus:text-[var(--text-primary)]"
    >
      {theme === "system" ? <Monitor size={16} /> : theme === "dark" ? <Moon size={16} /> : <Sun size={16} />}
      {label}
    </DropdownMenuItem>
  )
}
