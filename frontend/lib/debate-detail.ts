import type { DebateDetail } from "@/lib/debates"

export function normalizeDebateDetail(debate: DebateDetail): DebateDetail {
  const participantHistory = debate.participant_history
  return {
    ...debate,
    turns: Array.isArray(debate.turns) ? debate.turns : [],
    timeline: Array.isArray(debate.timeline) ? debate.timeline : [],
    participant_history: {
      side_a: Array.isArray(participantHistory?.side_a) ? participantHistory.side_a : [],
      side_b: Array.isArray(participantHistory?.side_b) ? participantHistory.side_b : [],
    },
  }
}
