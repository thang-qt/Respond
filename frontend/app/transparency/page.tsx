import type { Metadata } from "next"
import { getTranslations } from "next-intl/server"
import { siteConfig } from "@/lib/config"

export async function generateMetadata(): Promise<Metadata> {
  const t = await getTranslations("transparency")
  return {
    title: `Transparency — ${siteConfig.name}`,
    description: t("metaDescription"),
  }
}

const ratingKeys = ["win", "loss", "draw", "resignation"] as const

export default async function TransparencyPage() {
  const t = await getTranslations("transparency")
  const collect = t.raw("collect") as string[]
  const dontCollect = t.raw("dontCollect") as string[]
  const moderation = t.raw("moderation") as string[]
  const openQuestions = t.raw("openQuestions") as string[]

  return (
    <div className="min-h-screen bg-[var(--bg-primary)]">
      <div className="max-w-[820px] mx-auto px-3 sm:px-6 py-10 sm:py-16">
        <div className="max-w-[600px]">
          <h1 className="text-[28px] sm:text-[32px] font-semibold font-sans text-[var(--text-primary)] mb-3 leading-tight">
            {t("title")}
          </h1>
          <p className="text-[15px] font-sans text-[var(--text-muted)] mb-10 leading-relaxed">{t("subtitle")}</p>

          <div className="space-y-10">
            <ListSection title={t("collectTitle")} items={collect} />
            <ListSection title={t("dontCollectTitle")} items={dontCollect} />

            <div>
              <h2 className="text-[17px] font-semibold font-sans text-[var(--text-primary)] mb-3">{t("moderationTitle")}</h2>
              <div className="space-y-3 text-[15px] font-sans text-[var(--text-secondary)] leading-relaxed">
                {moderation.map((paragraph, index) => <p key={index}>{paragraph}</p>)}
              </div>
            </div>

            <div>
              <h2 className="text-[17px] font-semibold font-sans text-[var(--text-primary)] mb-3">{t("ratingTitle")}</h2>
              <div className="space-y-3 text-[15px] font-sans text-[var(--text-secondary)] leading-relaxed">
                <p>{t("ratingIntro")}</p>
                <ul className="space-y-1.5">
                  {ratingKeys.map((key) => (
                    <li key={key} className="flex gap-2">
                      <span className="font-medium text-[var(--text-primary)] flex-shrink-0">{t(`ratingItems.${key}.label`)}</span>
                      <span>— {t(`ratingItems.${key}.note`)}</span>
                    </li>
                  ))}
                </ul>
                <p>{t("ratingOutro")}</p>
              </div>
            </div>

            <div>
              <h2 className="text-[17px] font-semibold font-sans text-[var(--text-primary)] mb-3">{t("openTitle")}</h2>
              <p className="text-[15px] font-sans text-[var(--text-secondary)] leading-relaxed mb-4">{t("openIntro")}</p>
              <ul className="space-y-2">
                {openQuestions.map((item) => (
                  <li key={item} className="flex gap-2.5 text-[15px] font-sans text-[var(--text-secondary)] leading-relaxed">
                    <span className="text-[var(--text-muted)] flex-shrink-0 mt-0.5">–</span>
                    {item}
                  </li>
                ))}
              </ul>
              <p className="text-[15px] font-sans text-[var(--text-secondary)] leading-relaxed mt-4">{t("openOutro")}</p>
            </div>

            <div>
              <h2 className="text-[17px] font-semibold font-sans text-[var(--text-primary)] mb-3">{t("feedbackTitle")}</h2>
              <p className="text-[15px] font-sans text-[var(--text-secondary)] leading-relaxed">
                {t("feedback")} {" "}
                <a href="mailto:feedback@respond.im" className="text-[var(--text-primary)] underline underline-offset-2 hover:opacity-80 transition-opacity">
                  feedback@respond.im
                </a>
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

function ListSection({ title, items }: { title: string; items: string[] }) {
  return (
    <div>
      <h2 className="text-[17px] font-semibold font-sans text-[var(--text-primary)] mb-4">{title}</h2>
      <ul className="space-y-2">
        {items.map((item) => (
          <li key={item} className="flex gap-2.5 text-[15px] font-sans text-[var(--text-secondary)] leading-relaxed">
            <span className="text-[var(--text-muted)] flex-shrink-0 mt-0.5">–</span>
            {item}
          </li>
        ))}
      </ul>
    </div>
  )
}
