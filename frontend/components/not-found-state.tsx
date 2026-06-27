"use client"

import Link from "next/link"
import { ArrowLeft } from "@phosphor-icons/react"

type NotFoundAction = {
  href: string
  label: string
  primary?: boolean
}

interface NotFoundStateProps {
  code?: string
  title: string
  description: string
  actions?: NotFoundAction[]
}

export function NotFoundState({ code, title, description, actions = [] }: NotFoundStateProps) {
  return (
    <div className="min-h-[calc(100vh-56px)] bg-[var(--bg-primary)] flex items-center justify-center">
      <div className="w-full max-w-[640px] px-4 sm:px-6 py-12 sm:py-16 text-center">
        {code ? (
          <div className="text-[72px] sm:text-[92px] font-semibold text-[var(--text-primary)] font-sans tracking-[-0.02em]">
            {code}
          </div>
        ) : null}
        <h1 className="mt-2 text-[26px] sm:text-[30px] font-semibold text-[var(--text-primary)] font-sans">{title}</h1>
        <p className="mt-3 text-[13px] text-[var(--text-muted)] font-sans">{description}</p>

        {actions.length > 0 ? (
          <div className="mt-8 flex flex-wrap items-center justify-center gap-2">
            {actions.map((action, index) => (
              <Link
                key={`${action.href}:${action.label}`}
                href={action.href}
                className={
                  action.primary || (!actions.some((item) => item.primary) && index === 0)
                    ? "inline-flex items-center gap-2 rounded-md bg-[var(--text-primary)] px-4 py-2 text-[12px] font-semibold text-[var(--bg-primary)] font-sans hover:opacity-90"
                    : "inline-flex items-center gap-2 rounded-md border border-[var(--border-default)] px-4 py-2 text-[12px] font-semibold text-[var(--text-primary)] font-sans hover:bg-[var(--bg-surface-alt)]"
                }
              >
                {index === 0 ? <ArrowLeft size={14} /> : null}
                {action.label}
              </Link>
            ))}
          </div>
        ) : null}
      </div>
    </div>
  )
}
