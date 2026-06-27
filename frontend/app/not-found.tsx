"use client"

import { useTranslations } from "next-intl"
import { NotFoundState } from "@/components/not-found-state"

export default function NotFound() {
  const t = useTranslations("notFound")
  const tCommon = useTranslations("common")

  return (
    <NotFoundState
      code="404"
      title={t("title")}
      description={t("description")}
      actions={[
        { href: "/", label: tCommon("backToHome"), primary: true },
        { href: "/tags", label: t("browseTags") },
        { href: "/?tab=live", label: t("seeLiveDebates") },
      ]}
    />
  )
}
