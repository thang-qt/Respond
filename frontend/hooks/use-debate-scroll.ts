"use client"

import { useCallback, useEffect } from "react"
import type { ReadonlyURLSearchParams } from "next/navigation"
import type { AppRouterInstance } from "next/dist/shared/lib/app-router-context.shared-runtime"
import type { DebateComment, DebateDetail } from "@/lib/debates"

type UseDebateScrollOptions = {
  comments: DebateComment[]
  debate: DebateDetail | null
  debateId: string
  highlightCommentId: string | null
  isNearPageBottom: () => boolean
  pendingScrollTurn: number | null
  router: AppRouterInstance
  searchParams: ReadonlyURLSearchParams
  setHighlightCommentId: (id: string | null) => void
  setPendingScrollTurn: (turn: number | null) => void
  setShowNewTurnIndicator: (show: boolean) => void
  showNewTurnIndicator: boolean
}

export function useDebateScroll({
  comments,
  debate,
  debateId,
  highlightCommentId,
  isNearPageBottom,
  pendingScrollTurn,
  router,
  searchParams,
  setHighlightCommentId,
  setPendingScrollTurn,
  setShowNewTurnIndicator,
  showNewTurnIndicator,
}: UseDebateScrollOptions) {
  const scrollToTurn = useCallback((turnNumber: number) => {
    const target = document.getElementById(`turn-${turnNumber}`)
    if (!target) return false
    target.scrollIntoView({ behavior: "smooth", block: "start" })
    return true
  }, [])

  useEffect(() => {
    if (!debate?.slug) return
    if (debateId === debate.slug) return

    const query = searchParams.toString()
    const hash = window.location.hash
    const canonical = `/debate/${debate.slug}${query ? `?${query}` : ""}${hash}`
    router.replace(canonical, { scroll: false })
  }, [debate?.slug, debateId, router, searchParams])

  useEffect(() => {
    if (!debate) return

    const turnParam = searchParams.get("turn")
    if (!turnParam) return

    let targetTurn: number | null = null
    if (turnParam === "latest") {
      const latest = [...debate.turns].reverse().find((turn) => !turn.is_system)
      targetTurn = latest?.turn_number ?? null
    } else {
      const parsed = Number.parseInt(turnParam, 10)
      targetTurn = Number.isFinite(parsed) ? parsed : null
    }

    if (targetTurn == null) return

    let attempts = 0
    let timer: number | null = null
    const maxAttempts = 12
    const turnToScroll = targetTurn

    const tryScroll = () => {
      attempts += 1
      const ok = scrollToTurn(turnToScroll)
      if (ok) {
        const next = new URLSearchParams(searchParams.toString())
        next.delete("turn")
        const hash = window.location.hash
        const base = `/debate/${debate.slug || debateId}`
        router.replace(`${base}${next.toString() ? `?${next.toString()}` : ""}${hash}`, { scroll: false })
        return
      }

      if (attempts < maxAttempts) {
        timer = window.setTimeout(tryScroll, 100)
      }
    }

    timer = window.setTimeout(tryScroll, 60)

    return () => {
      if (timer != null) window.clearTimeout(timer)
    }
  }, [debate, debateId, router, scrollToTurn, searchParams])

  useEffect(() => {
    if (pendingScrollTurn == null) return
    const timer = window.setTimeout(() => {
      scrollToTurn(pendingScrollTurn)
      setPendingScrollTurn(null)
    }, 60)
    return () => window.clearTimeout(timer)
  }, [pendingScrollTurn, scrollToTurn, setPendingScrollTurn, debate?.turns?.length])

  useEffect(() => {
    if (!showNewTurnIndicator) return

    const onScroll = () => {
      if (isNearPageBottom()) {
        setShowNewTurnIndicator(false)
      }
    }

    window.addEventListener("scroll", onScroll, { passive: true })
    return () => {
      window.removeEventListener("scroll", onScroll)
    }
  }, [showNewTurnIndicator, isNearPageBottom, setShowNewTurnIndicator])

  useEffect(() => {
    if (!highlightCommentId) return
    const target = document.getElementById(`comment-${highlightCommentId}`)
    if (target) {
      target.scrollIntoView({ behavior: "smooth", block: "center" })
    }
    const timer = window.setTimeout(() => {
      setHighlightCommentId(null)
    }, 2000)
    return () => {
      window.clearTimeout(timer)
    }
  }, [highlightCommentId, comments, setHighlightCommentId])

  return { scrollToTurn }
}
