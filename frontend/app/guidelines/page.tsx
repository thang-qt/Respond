import type { Metadata } from "next"
import { getTranslations } from "next-intl/server"
import { siteConfig } from "@/lib/config"

export async function generateMetadata(): Promise<Metadata> {
  const t = await getTranslations("guidelines")
  return {
    title: `Guidelines — ${siteConfig.name}`,
    description: t("metaDescription"),
  }
}

const allowedKeys = ["opinions", "disagreement", "provocative", "unpopular", "criticism", "sarcasm", "ai"] as const
const notAllowedKeys = ["attacks", "hate", "threats", "sexual", "doxxing", "spam", "illegal"] as const
const exampleKeys = ["logic", "idiot", "policy", "subhuman", "evidence", "threat"] as const
const allowedExamples = new Set<(typeof exampleKeys)[number]>(["logic", "policy", "evidence"])
const enforcementSteps = ["warning", "restriction", "suspension", "ban"] as const

export default async function GuidelinesPage() {
  const t = await getTranslations("guidelines")

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
            <div>
              <h2 className="text-[17px] font-semibold font-sans text-[var(--text-primary)] mb-4">{t("allowedTitle")}</h2>
              <ul className="space-y-2">
                {allowedKeys.map((key) => (
                  <li key={key} className="flex gap-2.5 text-[15px] font-sans text-[var(--text-secondary)] leading-relaxed">
                    <span className="text-[var(--text-muted)] flex-shrink-0 mt-0.5">–</span>
                    {t(`allowed.${key}`)}
                  </li>
                ))}
              </ul>
            </div>

            <div>
              <h2 className="text-[17px] font-semibold font-sans text-[var(--text-primary)] mb-4">{t("notAllowedTitle")}</h2>
              <ul className="space-y-2">
                {notAllowedKeys.map((key) => (
                  <li key={key} className="flex gap-2.5 text-[15px] font-sans text-[var(--text-secondary)] leading-relaxed">
                    <span className="text-[var(--text-muted)] flex-shrink-0 mt-0.5">–</span>
                    {t(`notAllowed.${key}`)}
                  </li>
                ))}
              </ul>
            </div>

            <div>
              <h2 className="text-[17px] font-semibold font-sans text-[var(--text-primary)] mb-4">{t("lineTitle")}</h2>
              <div className="space-y-3">
                {exampleKeys.map((key) => {
                  const allowed = allowedExamples.has(key)
                  return (
                    <div key={key} className="flex items-start gap-3 py-2.5 border-b border-[var(--border-subtle)] last:border-0">
                      <span className={`flex-shrink-0 text-[11px] font-semibold font-sans uppercase tracking-wide mt-0.5 ${allowed ? "text-[var(--text-muted)]" : "text-[var(--error)]"}`}>
                        {allowed ? t("ok") : t("no")}
                      </span>
                      <div>
                        <p className="text-[14px] font-sans text-[var(--text-primary)] leading-snug">{t(`examples.${key}.text`)}</p>
                        <p className="text-[13px] font-sans text-[var(--text-muted)] mt-0.5">{t(`examples.${key}.note`)}</p>
                      </div>
                    </div>
                  )
                })}
              </div>
            </div>

            <div>
              <h2 className="text-[17px] font-semibold font-sans text-[var(--text-primary)] mb-3">{t("enforcementTitle")}</h2>
              <div className="space-y-3 text-[15px] font-sans text-[var(--text-secondary)] leading-relaxed">
                <p>{t("enforcement.p1")}</p>
                <p>{t("enforcement.p2")}</p>
                <p>{t("enforcement.p3")}</p>
                <ol className="space-y-1 ml-1">
                  {enforcementSteps.map((step, i) => (
                    <li key={step} className="flex gap-2.5 text-[15px] font-sans text-[var(--text-secondary)]">
                      <span className="text-[var(--text-muted)] font-semibold flex-shrink-0 tabular-nums">{i + 1}.</span>
                      {t(`enforcement.steps.${step}`)}
                    </li>
                  ))}
                </ol>
                <p>{t("enforcement.p4")}</p>
              </div>
            </div>

            <div>
              <h2 className="text-[17px] font-semibold font-sans text-[var(--text-primary)] mb-3">{t("reportingTitle")}</h2>
              <p className="text-[15px] font-sans text-[var(--text-secondary)] leading-relaxed">{t("reporting")}</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
