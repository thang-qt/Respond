export const CHALLENGES_REFRESH_EVENT = "challenges:refresh"

export function emitChallengesRefresh() {
  if (typeof window === "undefined") return
  window.dispatchEvent(new Event(CHALLENGES_REFRESH_EVENT))
}
