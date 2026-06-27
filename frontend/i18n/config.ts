export const locales = ["en", "vi"] as const

export type AppLocale = (typeof locales)[number]

export const defaultLocale: AppLocale = "en"

export function isAppLocale(value: string): value is AppLocale {
  return locales.includes(value as AppLocale)
}

export function normalizeLocale(value: string | null | undefined): AppLocale | null {
  if (!value) return null

  const normalized = value.trim().toLowerCase()
  if (!normalized) return null

  if (isAppLocale(normalized)) return normalized

  const base = normalized.split("-")[0]
  return isAppLocale(base) ? base : null
}
