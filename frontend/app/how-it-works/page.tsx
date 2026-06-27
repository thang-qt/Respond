import type { Metadata } from "next"
import { getTranslations } from "next-intl/server"
import { siteConfig } from "@/lib/config"

export async function generateMetadata(): Promise<Metadata> {
  const t = await getTranslations("howItWorks")
  return {
    title: `How It Works — ${siteConfig.name}`,
    description: t("metaDescription"),
  }
}

const stepKeys = ["start", "opponent", "turns", "anonymous", "conclusion", "discuss", "rating"] as const
const conclusionItems = ["concede", "draw", "resign", "walkover"] as const

export default async function HowItWorksPage() {
  const t = await getTranslations("howItWorks")

  return (
    <div className="min-h-screen bg-[var(--bg-primary)]">
      <div className="max-w-[820px] mx-auto px-3 sm:px-6 py-10 sm:py-16">
        <div className="max-w-[600px]">
          <h1 className="text-[28px] sm:text-[32px] font-semibold font-sans text-[var(--text-primary)] mb-10 leading-tight">
            {t("title")}
          </h1>

          <div className="space-y-8">
            {stepKeys.map((key, index) => (
              <div key={key} className="flex gap-5">
                <div className="flex-shrink-0 w-6 h-6 mt-0.5 rounded-full bg-[var(--bg-surface)] border border-[var(--border-default)] flex items-center justify-center">
                  <span className="text-[11px] font-semibold font-sans text-[var(--text-muted)]">{index + 1}</span>
                </div>
                <div>
                  <h2 className="text-[16px] font-semibold font-sans text-[var(--text-primary)] mb-2">
                    {t(`steps.${key}.heading`)}
                  </h2>
                  {key === "conclusion" ? (
                    <ul className="space-y-1.5 mt-1">
                      {conclusionItems.map((item) => (
                        <li key={item} className="text-[15px] font-sans text-[var(--text-secondary)] leading-relaxed">
                          <span className="font-medium text-[var(--text-primary)]">{t(`steps.conclusion.items.${item}.label`)}</span>
                          {" — "}
                          {t(`steps.conclusion.items.${item}.description`)}
                        </li>
                      ))}
                    </ul>
                  ) : (
                    <p className="text-[15px] font-sans text-[var(--text-secondary)] leading-relaxed">
                      {t(`steps.${key}.body`)}
                    </p>
                  )}
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  )
}
