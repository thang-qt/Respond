import { useEffect, useRef, useCallback } from "react"
import { getAccessToken } from "@/lib/auth-token"
import type { WSEvent, WSEventHandler } from "@/lib/websocket"

const WS_BASE = (process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080")
  .replace(/^http/, "ws")
  .replace(/\/api\/v1$/, "")

/**
 * useNotificationSocket connects to /ws/notifications and calls
 * onEvent whenever a notification.new event arrives. Handles
 * reconnection with exponential backoff.
 *
 * Only connects when enabled is true (i.e. user is authenticated).
 */
export function useNotificationSocket(
  onEvent: WSEventHandler,
  enabled: boolean
) {
  const wsRef = useRef<WebSocket | null>(null)
  const handlerRef = useRef<WSEventHandler>(onEvent)
  const reconnectTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const reconnectDelayRef = useRef(1000)
  const closedRef = useRef(false)

  useEffect(() => {
    handlerRef.current = onEvent
  }, [onEvent])

  const connect = useCallback(() => {
    if (closedRef.current) return

    const token = getAccessToken()
    if (!token) return // can't connect without auth

    const url = `${WS_BASE}/ws/notifications`
    let ws: WebSocket
    try {
      ws = new WebSocket(url)
    } catch {
      scheduleReconnect()
      return
    }

    wsRef.current = ws

    ws.onopen = () => {
      reconnectDelayRef.current = 1000
      ws.send(JSON.stringify({ type: "auth", token }))
    }

    ws.onmessage = (ev) => {
      try {
        const event = JSON.parse(ev.data) as WSEvent
        handlerRef.current(event)
      } catch {
        // ignore
      }
    }

    ws.onclose = () => {
      wsRef.current = null
      scheduleReconnect()
    }

    ws.onerror = () => {
      // onclose will fire after this
    }

    function scheduleReconnect() {
      if (closedRef.current) return
      reconnectTimerRef.current = setTimeout(() => {
        reconnectTimerRef.current = null
        connect()
      }, reconnectDelayRef.current)
      reconnectDelayRef.current = Math.min(reconnectDelayRef.current * 2, 30000)
    }
  }, [])

  useEffect(() => {
    if (!enabled) return

    closedRef.current = false
    connect()

    return () => {
      closedRef.current = true
      if (reconnectTimerRef.current) {
        clearTimeout(reconnectTimerRef.current)
        reconnectTimerRef.current = null
      }
      if (wsRef.current) {
        wsRef.current.close(1000, "client closed")
        wsRef.current = null
      }
    }
  }, [enabled, connect])
}
