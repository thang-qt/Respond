// WebSocket event types from the server.
export type WSEventType =
  | "turn.new"
  | "debate.joined"
  | "debate.ended"
  | "debate.seat_open"
  | "debate.replacement"
  | "debate.draw_proposed"
  | "debate.draw_responded"
  | "debate.extension_update"
  | "debate.event"
  | "timer.sync"
  | "notification.new"

export interface WSEvent<T = unknown> {
  type: WSEventType
  data: T
}

export type WSEventHandler = (event: WSEvent) => void

const WS_BASE = (process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080")
  .replace(/^http/, "ws")
  .replace(/\/api\/v1$/, "")

/**
 * DebateSocket manages a single WebSocket connection to a debate room.
 * It handles connection, reconnection, authentication, and event dispatch.
 */
export class DebateSocket {
  private ws: WebSocket | null = null
  private url: string
  private token: string | null
  private handlers = new Set<WSEventHandler>()
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null
  private reconnectDelay = 1000
  private maxReconnectDelay = 30000
  private intentionallyClosed = false

  constructor(debateId: string, token: string | null) {
    this.url = `${WS_BASE}/ws/debates/${debateId}`
    this.token = token
  }

  /** Register an event handler. Returns an unsubscribe function. */
  on(handler: WSEventHandler): () => void {
    this.handlers.add(handler)
    return () => {
      this.handlers.delete(handler)
    }
  }

  /** Open the WebSocket connection. */
  connect() {
    this.intentionallyClosed = false
    this.doConnect()
  }

  /** Close the connection and stop reconnecting. */
  close() {
    this.intentionallyClosed = true
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer)
      this.reconnectTimer = null
    }
    if (this.ws) {
      this.ws.close(1000, "client closed")
      this.ws = null
    }
  }

  /** Update the auth token (e.g. after a refresh). */
  updateToken(token: string | null) {
    this.token = token
  }

  private doConnect() {
    if (this.intentionallyClosed) return

    try {
      this.ws = new WebSocket(this.url)
    } catch {
      this.scheduleReconnect()
      return
    }

    this.ws.onopen = () => {
      this.reconnectDelay = 1000 // reset backoff on success
      // Send auth message if we have a token.
      if (this.token && this.ws) {
        this.ws.send(JSON.stringify({ type: "auth", token: this.token }))
      }
    }

    this.ws.onmessage = (ev) => {
      try {
        const event = JSON.parse(ev.data) as WSEvent
        for (const handler of this.handlers) {
          handler(event)
        }
      } catch {
        // Ignore malformed messages.
      }
    }

    this.ws.onclose = () => {
      this.ws = null
      this.scheduleReconnect()
    }

    this.ws.onerror = () => {
      // onerror is always followed by onclose, which handles reconnect.
    }
  }

  private scheduleReconnect() {
    if (this.intentionallyClosed) return

    this.reconnectTimer = setTimeout(() => {
      this.reconnectTimer = null
      this.doConnect()
    }, this.reconnectDelay)

    // Exponential backoff with cap.
    this.reconnectDelay = Math.min(this.reconnectDelay * 2, this.maxReconnectDelay)
  }
}
