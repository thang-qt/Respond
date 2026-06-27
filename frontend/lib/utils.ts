import { clsx, type ClassValue } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

export function formatDate(dateStr: string): string {
  const date = new Date(dateStr)
  return date.toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  })
}

export function formatTimeAgo(dateStr: string): string {
  const now = new Date()
  const date = new Date(dateStr)
  const diffMs = now.getTime() - date.getTime()
  const isFuture = diffMs < 0
  const absDiffMs = Math.abs(diffMs)
  const diffMins = Math.floor(absDiffMs / 60000)
  const diffHours = Math.floor(diffMins / 60)
  const diffDays = Math.floor(diffHours / 24)

  if (diffMins < 60) return isFuture ? `in ${diffMins}m` : `${diffMins}m ago`
  if (diffHours < 24) return isFuture ? `in ${diffHours}h` : `${diffHours}h ago`
  if (diffDays < 7) return isFuture ? `in ${diffDays}d` : `${diffDays}d ago`
  return date.toLocaleDateString("en-US", { month: "short", day: "numeric" })
}
