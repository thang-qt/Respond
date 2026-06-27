import { useCallback, useEffect, useRef, useState } from "react"

/**
 * useTimer returns a live countdown string (e.g. "2d 5h 12m")
 * that ticks every second. It accepts a deadline ISO string and
 * optionally syncs from timer.sync WebSocket events.
 */
export function useTimer(deadline: string | null) {
  const [remaining, setRemaining] = useState<string | null>(null)
  const deadlineRef = useRef<string | null>(deadline)

  // Allow external sync (from WS timer.sync).
  const syncDeadline = useCallback((newDeadline: string | null) => {
    deadlineRef.current = newDeadline
  }, [])

  // Keep ref in sync with prop changes.
  useEffect(() => {
    deadlineRef.current = deadline
  }, [deadline])

  useEffect(() => {
    function tick() {
      const dl = deadlineRef.current
      if (!dl) {
        setRemaining(null)
        return
      }

      const diff = new Date(dl).getTime() - Date.now()
      if (diff <= 0) {
        setRemaining("Expired")
        return
      }

      setRemaining(formatDuration(diff))
    }

    tick()
    const id = setInterval(tick, 1000)
    return () => clearInterval(id)
  }, [deadline]) // restart interval when deadline prop changes

  return { remaining, syncDeadline }
}

function formatDuration(ms: number): string {
  const seconds = Math.floor(ms / 1000)
  const minutes = Math.floor(seconds / 60)
  const hours = Math.floor(minutes / 60)
  const days = Math.floor(hours / 24)

  if (days > 0) {
    const h = hours % 24
    return `${days}d ${h}h`
  }
  if (hours > 0) {
    const m = minutes % 60
    return `${hours}h ${m}m`
  }
  if (minutes > 0) {
    const s = seconds % 60
    return `${minutes}m ${s}s`
  }
  return `${seconds}s`
}
