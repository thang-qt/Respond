import { type Dispatch, type MutableRefObject, type SetStateAction, useCallback } from "react"
import type { DebateDetail, DebateEvent, DebateTurn } from "@/lib/debates"
import { fetchDebate } from "@/lib/debates-api"
import { normalizeDebateDetail } from "@/lib/debate-detail"
import { useDebateSocket } from "@/hooks/use-debate-socket"
import type { WSEvent } from "@/lib/websocket"

type UseDebateRealtimeArgs = {
  debate: DebateDetail | null
  debateId: string
  debateRef: MutableRefObject<DebateDetail | null>
  isNearPageBottom: () => boolean
  setDebate: Dispatch<SetStateAction<DebateDetail | null>>
  setPendingScrollTurn: Dispatch<SetStateAction<number | null>>
  setShowNewTurnIndicator: Dispatch<SetStateAction<boolean>>
}

export function useDebateRealtime({
  debate,
  debateId,
  debateRef,
  isNearPageBottom,
  setDebate,
  setPendingScrollTurn,
  setShowNewTurnIndicator,
}: UseDebateRealtimeArgs) {
  const refreshDebate = useCallback(() => {
    void fetchDebate(debateId)
      .then((res) => setDebate(normalizeDebateDetail(res.data)))
      .catch(() => undefined)
  }, [debateId, setDebate])

  const handleWSEvent = useCallback((event: WSEvent) => {
    switch (event.type) {
      case "turn.new": {
        const turn = event.data as DebateTurn
        if (turn.is_system) break
        const previous = debateRef.current
        if (!previous) break
        if (previous.turns.some((existingTurn) => existingTurn.id === turn.id)) break

        const shouldAutoScroll = isNearPageBottom()

        setDebate((prev) => {
          if (!prev) return prev
          if (prev.turns.some((existingTurn) => existingTurn.id === turn.id)) return prev
          const nextSide = turn.side === "a" ? "b" : "a"
          return {
            ...prev,
            turns: [...prev.turns, turn],
            timeline: [...(prev.timeline ?? []), { type: "turn", created_at: turn.created_at, turn }],
            turn_count: prev.turn_count + 1,
            current_turn_side: nextSide,
          }
        })

        if (shouldAutoScroll) {
          setPendingScrollTurn(turn.turn_number)
          setShowNewTurnIndicator(false)
        } else {
          setShowNewTurnIndicator(true)
        }
        break
      }
      case "debate.joined": {
        const data = event.data as { side: string; anonymous_id: string; turn_deadline: string | null }
        setDebate((prev) => {
          if (!prev) return prev
          return {
            ...prev,
            status: "active",
            started_at: prev.started_at ?? new Date().toISOString(),
            turn_deadline: data.turn_deadline,
            side_b: {
              anonymous_id: data.anonymous_id,
              revealed: false,
              user: null,
            },
          }
        })
        break
      }
      case "debate.ended": {
        const data = event.data as { outcome: string | null; winner_side: string | null; ended_at: string | null }
        setDebate((prev) => {
          if (!prev) return prev
          return {
            ...prev,
            status: "finished",
            outcome: data.outcome as DebateDetail["outcome"],
            winner_side: data.winner_side as DebateDetail["winner_side"],
            ended_at: data.ended_at,
          }
        })
        refreshDebate()
        break
      }
      case "debate.seat_open": {
        const data = event.data as { side: string }
        setDebate((prev) => {
          if (!prev) return prev
          return {
            ...prev,
            status: "waiting_replacement",
            open_side: data.side as "a" | "b",
          }
        })
        break
      }
      case "debate.replacement": {
        const data = event.data as { side: string; anonymous_id: string; turn_deadline: string | null }
        setDebate((prev) => {
          if (!prev) return prev
          const sideKey = data.side === "a" ? "side_a" : "side_b"
          return {
            ...prev,
            status: "active",
            open_side: null,
            turn_deadline: data.turn_deadline,
            [sideKey]: {
              anonymous_id: data.anonymous_id,
              revealed: false,
              user: null,
            },
          }
        })
        break
      }
      case "debate.draw_proposed": {
        const data = event.data as { proposed_by: string }
        setDebate((prev) => {
          if (!prev) return prev
          return {
            ...prev,
            draw_proposed_by: data.proposed_by as "a" | "b",
          }
        })
        break
      }
      case "debate.draw_responded": {
        const data = event.data as { accepted: boolean }
        if (!data.accepted) {
          setDebate((prev) => {
            if (!prev) return prev
            return {
              ...prev,
              draw_proposed_by: null,
            }
          })
        }
        break
      }
      case "debate.extension_update": {
        const data = event.data as { status: string; turn_limit?: number; outcome?: string | null; winner_side?: string | null }
        setDebate((prev) => {
          if (!prev) return prev
          if (data.status === "active" && data.turn_limit) {
            return {
              ...prev,
              status: "active",
              turn_limit: data.turn_limit,
              extension_deadline: null,
              extension_a_accepted: null,
              extension_b_accepted: null,
            }
          }
          if (data.status === "finished") {
            refreshDebate()
            return {
              ...prev,
              status: "finished",
              outcome: (data.outcome ?? null) as DebateDetail["outcome"],
              winner_side: (data.winner_side ?? null) as DebateDetail["winner_side"],
              extension_deadline: null,
              extension_a_accepted: null,
              extension_b_accepted: null,
            }
          }
          return prev
        })
        break
      }
      case "debate.event": {
        const debateEvent = event.data as DebateEvent
        setDebate((prev) => {
          if (!prev) return prev
          const existing = prev.timeline ?? []
          if (existing.some((entry) => entry.type === "event" && entry.event?.id === debateEvent.id)) {
            return prev
          }
          return {
            ...prev,
            timeline: [...existing, { type: "event", created_at: debateEvent.created_at, event: debateEvent }],
          }
        })
        break
      }
      case "timer.sync": {
        const data = event.data as { current_turn_side: string; turn_deadline: string | null }
        setDebate((prev) => {
          if (!prev) return prev
          return {
            ...prev,
            current_turn_side: data.current_turn_side as "a" | "b",
            turn_deadline: data.turn_deadline,
          }
        })
        break
      }
    }
  }, [debateRef, isNearPageBottom, refreshDebate, setDebate, setPendingScrollTurn, setShowNewTurnIndicator])

  const wsEnabled = Boolean(debate && debate.status !== "finished" && debate.status !== "expired")
  useDebateSocket(debateId, handleWSEvent, wsEnabled)
}
