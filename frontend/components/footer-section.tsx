"use client"

import Link from "next/link"
import { GithubLogo, XLogo } from "@phosphor-icons/react"
import { useTranslations } from "next-intl"
import BrandLogo from "@/components/brand-logo"
import { siteConfig } from "@/lib/config"

export default function FooterSection() {
  const tSite = useTranslations("site")
  const t = useTranslations("footer")

  return (
    <footer className="w-full border-t border-[var(--border-subtle)] mt-12">
      <div className="max-w-[820px] mx-auto px-4 sm:px-6 py-8 sm:py-10">
        <div className="flex flex-col sm:flex-row gap-8 sm:gap-12">
          <div className="flex flex-col gap-2 sm:min-w-[140px]">
            <Link href="/" className="flex items-center gap-2 text-[var(--text-primary)]">
              <BrandLogo size={18} />
              <span className="text-[15px] font-semibold font-sans">{siteConfig.name}</span>
            </Link>
            <p className="text-[var(--text-muted)] text-[13px] font-sans leading-snug">
              {tSite("tagline")}
            </p>
          </div>

          <div className="flex flex-wrap gap-8 sm:gap-12 flex-1">
            <div className="flex flex-col gap-2.5 min-w-[100px]">
              <span className="text-[var(--text-muted)] text-[11px] font-semibold font-sans uppercase tracking-wider">{t("platform")}</span>
              <Link href="/" className="text-[var(--text-secondary)] text-[13px] font-sans hover:text-[var(--text-primary)] transition-colors">{t("browseDebates")}</Link>
              <Link href="/create" className="text-[var(--text-secondary)] text-[13px] font-sans hover:text-[var(--text-primary)] transition-colors">{t("startDebate")}</Link>
              <Link href="/how-it-works" className="text-[var(--text-secondary)] text-[13px] font-sans hover:text-[var(--text-primary)] transition-colors">{t("howItWorks")}</Link>
            </div>
            <div className="flex flex-col gap-2.5 min-w-[100px]">
              <span className="text-[var(--text-muted)] text-[11px] font-semibold font-sans uppercase tracking-wider">{t("community")}</span>
              <Link href="/guidelines" className="text-[var(--text-secondary)] text-[13px] font-sans hover:text-[var(--text-primary)] transition-colors">{t("guidelines")}</Link>
              <Link href="/ai-stance" className="text-[var(--text-secondary)] text-[13px] font-sans hover:text-[var(--text-primary)] transition-colors">{t("aiStance")}</Link>
              <Link href="/fairness" className="text-[var(--text-secondary)] text-[13px] font-sans hover:text-[var(--text-primary)] transition-colors">{t("fairness")}</Link>
            </div>
            <div className="flex flex-col gap-2.5 min-w-[100px]">
              <span className="text-[var(--text-muted)] text-[11px] font-semibold font-sans uppercase tracking-wider">{t("company")}</span>
              <Link href="/about" className="text-[var(--text-secondary)] text-[13px] font-sans hover:text-[var(--text-primary)] transition-colors">{t("about")}</Link>
              <Link href="/philosophy" className="text-[var(--text-secondary)] text-[13px] font-sans hover:text-[var(--text-primary)] transition-colors">{t("philosophy")}</Link>
              <Link href="/transparency" className="text-[var(--text-secondary)] text-[13px] font-sans hover:text-[var(--text-primary)] transition-colors">{t("transparency")}</Link>
            </div>
          </div>
        </div>

        <div className="flex flex-col-reverse sm:flex-row items-start sm:items-center justify-between gap-4 mt-8 pt-6 border-t border-[var(--border-subtle)]">
          <span className="text-[var(--text-muted)] text-[12px] font-sans">
            {siteConfig.domain}
          </span>
          <div className="flex items-center gap-4">
            <a href="#" className="text-[var(--text-muted)] hover:text-[var(--text-secondary)] transition-colors">
              <XLogo size={16} weight="fill" />
            </a>
            <a href="#" className="text-[var(--text-muted)] hover:text-[var(--text-secondary)] transition-colors">
              <GithubLogo size={16} weight="fill" />
            </a>
          </div>
        </div>
      </div>
    </footer>
  )
}
