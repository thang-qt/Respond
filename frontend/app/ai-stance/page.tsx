import type { Metadata } from "next"
import { getTranslations } from "next-intl/server"
import { siteConfig } from "@/lib/config"

export async function generateMetadata(): Promise<Metadata> {
  const t = await getTranslations("aiStance")
  return {
    title: `Our Stance on AI — ${siteConfig.name}`,
    description: t("metaDescription"),
  }
}

export default async function AIStancePage() {
  const t = await getTranslations("aiStance")
  const encourageItems = t.raw("sections.encourage.items") as string[]
  const enforcementSteps = t.raw("sections.handle.steps") as string[]

  return (
    <div className="min-h-screen bg-[var(--bg-primary)]">
      <div className="max-w-[820px] mx-auto px-3 sm:px-6 py-10 sm:py-16">
        <div className="max-w-[600px]">
          <h1 className="text-[28px] sm:text-[32px] font-semibold font-sans text-[var(--text-primary)] mb-3 leading-tight">
            {t("title")}
          </h1>
          <p className="text-[15px] font-sans text-[var(--text-muted)] mb-10 leading-relaxed">{t("subtitle")}</p>

          <div className="space-y-10">
            <div>
              <h2 className="text-[17px] font-semibold font-sans text-[var(--text-primary)] mb-3">{t("sections.reality.heading")}</h2>
              {(t.raw("sections.reality.body") as string[]).map((paragraph, i) => (
                <p key={i} className="text-[15px] font-sans text-[var(--text-secondary)] leading-relaxed">{paragraph}</p>
              ))}
            </div>

            <div>
              <h2 className="text-[17px] font-semibold font-sans text-[var(--text-primary)] mb-3">{t("sections.encourage.heading")}</h2>
              <ul className="space-y-2">
                {encourageItems.map((item) => (
                  <li key={item} className="text-[15px] font-sans text-[var(--text-secondary)] leading-relaxed flex gap-2">
                    <span className="text-[var(--text-muted)] mt-1 flex-shrink-0">–</span>
                    {item}
                  </li>
                ))}
              </ul>
            </div>

            <div>
              <h2 className="text-[17px] font-semibold font-sans text-[var(--text-primary)] mb-3">{t("sections.purpose.heading")}</h2>
              {(t.raw("sections.purpose.body") as string[]).map((paragraph, i) => (
                <p key={i} className={`text-[15px] font-sans text-[var(--text-secondary)] leading-relaxed ${i > 0 ? "mt-3" : ""}`}>{paragraph}</p>
              ))}
            </div>

            <div>
              <h2 className="text-[17px] font-semibold font-sans text-[var(--text-primary)] mb-3">{t("sections.handle.heading")}</h2>
              <p className="text-[15px] font-sans text-[var(--text-secondary)] leading-relaxed mb-4">{t("sections.handle.intro")}</p>
              <div className="space-y-4">
                <div>
                  <p className="text-[15px] font-semibold font-sans text-[var(--text-primary)] mb-1">{t("sections.handle.transparencyTitle")}</p>
                  <p className="text-[15px] font-sans text-[var(--text-secondary)] leading-relaxed">{t("sections.handle.transparencyBody")}</p>
                </div>
                <div>
                  <p className="text-[15px] font-semibold font-sans text-[var(--text-primary)] mb-1">{t("sections.handle.accountabilityTitle")}</p>
                  <p className="text-[15px] font-sans text-[var(--text-secondary)] leading-relaxed">{t("sections.handle.accountabilityBody")}</p>
                </div>
                <div>
                  <p className="text-[15px] font-semibold font-sans text-[var(--text-primary)] mb-1">{t("sections.handle.enforcementTitle")}</p>
                  <p className="text-[15px] font-sans text-[var(--text-secondary)] leading-relaxed mb-2">{t("sections.handle.enforcementBody")}</p>
                  <ol className="space-y-1">
                    {enforcementSteps.map((step, i) => (
                      <li key={step} className="text-[15px] font-sans text-[var(--text-secondary)] leading-relaxed flex gap-2.5">
                        <span className="text-[var(--text-muted)] font-semibold flex-shrink-0 tabular-nums">{i + 1}.</span>
                        {step}
                      </li>
                    ))}
                  </ol>
                  <p className="text-[14px] font-sans text-[var(--text-muted)] leading-relaxed mt-3">{t("sections.handle.note")}</p>
                </div>
              </div>
            </div>

            {(["line", "matters"] as const).map((key) => (
              <div key={key}>
                <h2 className="text-[17px] font-semibold font-sans text-[var(--text-primary)] mb-3">{t(`sections.${key}.heading`)}</h2>
                {(t.raw(`sections.${key}.body`) as string[]).map((paragraph, i) => (
                  <p key={i} className={`text-[15px] font-sans text-[var(--text-secondary)] leading-relaxed ${i > 0 ? "mt-3" : ""}`}>{paragraph}</p>
                ))}
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  )
}
