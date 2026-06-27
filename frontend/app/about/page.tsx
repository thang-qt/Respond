import type { Metadata } from "next"
import { getTranslations } from "next-intl/server"
import { siteConfig } from "@/lib/config"

export async function generateMetadata(): Promise<Metadata> {
  const t = await getTranslations("about")
  return {
    title: `About — ${siteConfig.name}`,
    description: t("metaDescription"),
  }
}

export default async function AboutPage() {
  const t = await getTranslations("about")
  const paragraphs = t.raw("paragraphs") as string[]

  return (
    <div className="min-h-screen bg-[var(--bg-primary)]">
      <div className="max-w-[820px] mx-auto px-3 sm:px-6 py-10 sm:py-16">
        <div className="max-w-[600px]">
          <h1 className="text-[28px] sm:text-[32px] font-semibold font-sans text-[var(--text-primary)] mb-8 leading-tight">
            {t("title")}
          </h1>

          <div className="space-y-5 text-[15px] font-sans text-[var(--text-secondary)] leading-relaxed">
            {paragraphs.map((paragraph, index) => (
              <p key={index}>{paragraph}</p>
            ))}
          </div>
        </div>
      </div>
    </div>
  )
}
