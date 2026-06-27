import { defaultLocale, type AppLocale } from "@/i18n/config"

export const localeCookieName = "NEXT_LOCALE"

export function setLocaleCookie(locale: AppLocale) {
  document.cookie = `${localeCookieName}=${encodeURIComponent(locale)}; Path=/; Max-Age=31536000; SameSite=Lax`
}

export function getLocaleCookie(): AppLocale | null {
  if (typeof document === "undefined") return null

  const cookie = document.cookie
    .split(";")
    .map((part) => part.trim())
    .find((part) => part.startsWith(`${localeCookieName}=`))

  if (!cookie) return null

  const value = decodeURIComponent(cookie.slice(localeCookieName.length + 1))
  return value === "en" || value === "vi" ? value : null
}

export function getEffectiveClientLocale(): AppLocale {
  return getLocaleCookie() ?? defaultLocale
}
