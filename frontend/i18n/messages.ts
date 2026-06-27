import type { AppLocale } from "./config"

type Messages = Record<string, unknown>

async function loadDomainMessages(locale: AppLocale): Promise<Messages[]> {
  switch (locale) {
    case "vi":
      return Promise.all([
        import("../messages/vi/core.json").then((module) => module.default),
        import("../messages/vi/auth.json").then((module) => module.default),
        import("../messages/vi/debates.json").then((module) => module.default),
        import("../messages/vi/discovery.json").then((module) => module.default),
        import("../messages/vi/content.json").then((module) => module.default),
        import("../messages/vi/account.json").then((module) => module.default),
      ])
    case "en":
    default:
      return Promise.all([
        import("../messages/en/core.json").then((module) => module.default),
        import("../messages/en/auth.json").then((module) => module.default),
        import("../messages/en/debates.json").then((module) => module.default),
        import("../messages/en/discovery.json").then((module) => module.default),
        import("../messages/en/content.json").then((module) => module.default),
        import("../messages/en/account.json").then((module) => module.default),
      ])
  }
}

export async function loadMessages(locale: AppLocale): Promise<Messages> {
  const domains = await loadDomainMessages(locale)
  return Object.assign({}, ...domains)
}
