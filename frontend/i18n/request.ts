import { cookies, headers } from "next/headers"
import { getRequestConfig } from "next-intl/server"
import { defaultLocale, normalizeLocale } from "./config"
import { loadMessages } from "./messages"

function localeFromAcceptLanguage(value: string | null): string | null {
  if (!value) return null

  const parts = value
    .split(",")
    .map((part) => part.split(";")[0]?.trim())
    .filter(Boolean)

  for (const part of parts) {
    const locale = normalizeLocale(part)
    if (locale) return locale
  }

  return null
}

export default getRequestConfig(async () => {
  const cookieLocale = normalizeLocale(cookies().get("NEXT_LOCALE")?.value)
  const headerLocale = normalizeLocale(localeFromAcceptLanguage(headers().get("accept-language")))
  const locale = cookieLocale ?? headerLocale ?? defaultLocale

  return {
    locale,
    messages: await loadMessages(locale),
  }
})
