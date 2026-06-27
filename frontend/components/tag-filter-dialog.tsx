"use client"

import { CheckCircle, MagnifyingGlass } from "@phosphor-icons/react"
import { useEffect, useMemo, useState } from "react"
import { useTranslations } from "next-intl"
import type { Tag } from "@/lib/debates"
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog"

interface TagFilterDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  tags: Tag[]
  selectedSlugs: string[]
  onApply: (slugs: string[]) => void
  title?: string
}

export default function TagFilterDialog({
  open,
  onOpenChange,
  tags,
  selectedSlugs,
  onApply,
  title,
}: TagFilterDialogProps) {
  const t = useTranslations("tagFilter")
  const tCommon = useTranslations("common")
  const tTags = useTranslations("tags")
  const [query, setQuery] = useState("")
  const [draftSlugs, setDraftSlugs] = useState<string[]>([])

  useEffect(() => {
    if (!open) return
    setDraftSlugs(selectedSlugs)
    setQuery("")
  }, [open, selectedSlugs])

  const filteredTags = useMemo(() => {
    const q = query.trim().toLowerCase()
    const sorted = [...tags].sort((a, b) => a.name.localeCompare(b.name))
    if (!q) return sorted
    return sorted.filter((tag) => tag.name.toLowerCase().includes(q) || tag.slug.toLowerCase().includes(q))
  }, [query, tags])

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[640px] bg-[var(--bg-surface)] border border-[var(--border-default)]">
        <DialogHeader>
          <DialogTitle className="text-[var(--text-primary)] font-sans">{title ?? t("title")}</DialogTitle>
        </DialogHeader>

        <div className="space-y-3">
          <div className="relative">
            <MagnifyingGlass
              size={16}
              className="absolute left-3 top-1/2 -translate-y-1/2 text-[var(--text-muted)]"
              aria-hidden
            />
            <input
              value={query}
              onChange={(event) => setQuery(event.target.value)}
              placeholder={tTags("searchPlaceholder")}
              className="w-full h-10 rounded-lg border border-[var(--border-default)] bg-[var(--bg-primary)] pl-9 pr-3 text-[14px] text-[var(--text-primary)] outline-none focus:border-[var(--border-strong)] font-sans"
            />
          </div>

          <div className="max-h-[360px] overflow-y-auto rounded-md border border-[var(--border-subtle)] bg-[var(--bg-primary)]">
            {filteredTags.length > 0 ? (
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-2 p-2">
                {filteredTags.map((tag) => {
                  const selected = draftSlugs.includes(tag.slug)
                  return (
                    <button
                      key={tag.id}
                      type="button"
                      onClick={() => {
                        setDraftSlugs((prev) =>
                          prev.includes(tag.slug) ? prev.filter((slug) => slug !== tag.slug) : [...prev, tag.slug]
                        )
                      }}
                      className={`rounded-lg border px-3 py-2 text-left transition-colors ${
                        selected
                          ? "border-[var(--text-primary)] bg-[var(--bg-surface)]"
                          : "border-[var(--border-default)] bg-[var(--bg-surface)] hover:border-[var(--border-strong)]"
                      }`}
                    >
                      <div className="flex items-center gap-2">
                        <div className="min-w-0 flex-1">
                          <div className="text-[13px] font-semibold text-[var(--text-primary)] font-sans">{tag.name}</div>
                          <div className="text-[11px] text-[var(--text-muted)] font-sans">/{tag.slug}</div>
                        </div>
                        {selected && <CheckCircle size={14} weight="fill" className="text-[var(--text-primary)]" />}
                      </div>
                    </button>
                  )
                })}
              </div>
            ) : (
              <div className="px-3 py-8 text-center text-[12px] text-[var(--text-muted)]">{tTags("noMatch")}</div>
            )}
          </div>

          <div className="flex items-center justify-between gap-2">
            <button
              type="button"
              onClick={() => setDraftSlugs([])}
              className="h-9 rounded-md border border-[var(--border-default)] px-3 text-[12px] font-medium text-[var(--text-secondary)] hover:border-[var(--border-strong)]"
            >
              {tCommon("clearSelection")}
            </button>
            <button
              type="button"
              onClick={() => {
                onApply(draftSlugs)
                onOpenChange(false)
              }}
              className="h-9 rounded-md bg-[var(--text-primary)] px-3 text-[12px] font-medium text-[var(--bg-primary)] hover:opacity-90"
            >
              {tCommon("applyFilters")}
            </button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}
