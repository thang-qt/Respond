import type { Metadata } from "next"
import { getTranslations } from "next-intl/server"
import { siteConfig } from "@/lib/config"

export async function generateMetadata(): Promise<Metadata> {
  const t = await getTranslations("philosophy")
  return {
    title: `Philosophy — ${siteConfig.name}`,
    description: t("metaDescription"),
  }
}

const sectionKeys = ["structure", "anonymity", "slow", "fairness", "endings", "disagreement", "opponent"] as const

export default async function PhilosophyPage() {
  const t = await getTranslations("philosophy")

  return (
    <div className="min-h-screen bg-[var(--bg-primary)]">
      <div className="max-w-[820px] mx-auto px-3 sm:px-6 py-10 sm:py-16">
        <div className="max-w-[600px]">
          <h1 className="text-[28px] sm:text-[32px] font-semibold font-sans text-[var(--text-primary)] mb-3 leading-tight">
            {t("title")}
          </h1>
          <p className="text-[15px] font-sans text-[var(--text-muted)] mb-10 leading-relaxed">
            {t("subtitle")}
          </p>

          <div className="space-y-10">
            {sectionKeys.map((key) => {
              const paragraphs = t.raw(`sections.${key}.body`) as string[]
              return (
                <div key={key}>
                  <h2 className="text-[17px] font-semibold font-sans text-[var(--text-primary)] mb-3">
                    {t(`sections.${key}.heading`)}
                  </h2>
                  <div className="space-y-3">
                    {paragraphs.map((paragraph, i) => (
                      <p key={i} className="text-[15px] font-sans text-[var(--text-secondary)] leading-relaxed">
                        {paragraph}
                      </p>
                    ))}
                  </div>
                </div>
              )
            })}
          </div>
        </div>
      </div>
    </div>
  )
}
