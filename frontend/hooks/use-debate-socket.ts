import { useEffect, useRef, useCallback } from "react"
import { getAccessToken } from "@/lib/auth-token"
import { DebateSocket, type WSEvent, type WSEventHandler } from "@/lib/websocket"

/**
 * useDebateSocket connects to a debate's WebSocket room and
 * dispatches events to the provided handler. Handles connect,
 * disconnect, and reconnection automatically.
 *
 * @param debateId - The debate UUID or slug to subscribe to.
 * @param onEvent - Called for every incoming WebSocket event.
 * @param enabled - Set to false to skip connecting (e.g. for finished debates).
 */
export function useDebateSocket(
  debateId: string | null,
  onEvent: WSEventHandler,
  enabled = true
) {
  const socketRef = useRef<DebateSocket | null>(null)
  const handlerRef = useRef<WSEventHandler>(onEvent)

  // Keep the handler ref current without reconnecting.
  useEffect(() => {
    handlerRef.current = onEvent
  }, [onEvent])

  useEffect(() => {
    if (!debateId || !enabled) return

    const token = getAccessToken()
    const socket = new DebateSocket(debateId, token)
    socketRef.current = socket

    socket.on((event: WSEvent) => {
      handlerRef.current(event)
    })

    socket.connect()

    return () => {
      socket.close()
      socketRef.current = null
    }
  }, [debateId, enabled])

  // Provide a way to update the auth token after refresh.
  const updateToken = useCallback((token: string | null) => {
    socketRef.current?.updateToken(token)
  }, [])

  return { updateToken }
}
