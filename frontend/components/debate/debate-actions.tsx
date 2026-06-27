"use client"

import { useCallback, useState } from "react"
import { useTranslations } from "next-intl"
import { MoreHorizontal, Flag, HandshakeIcon, LogOut } from "lucide-react"
import { ApiError } from "@/lib/api"
import { Button } from "@/components/ui/button"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import {
  concedeDebate,
  resignDebate,
  proposeDraw,
  respondDraw,
  fetchDebate,
} from "@/lib/debates-api"
import type { DebateDetail, DebateSide } from "@/lib/debates"
import { useAuth } from "@/hooks/use-auth"
import { EMAIL_VERIFICATION_REQUIRED_MESSAGE } from "@/lib/verification"

interface DebateActionsProps {
  debate: DebateDetail
  onDebateUpdate: (debate: DebateDetail) => void
}

export function DebateActions({ debate, onDebateUpdate }: DebateActionsProps) {
  const { user } = useAuth()
  const t = useTranslations("debate.actionsMenu")
  const tCommon = useTranslations("debate.common")
  const [confirmAction, setConfirmAction] = useState<"concede" | "resign" | null>(null)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const viewerSide = debate.viewer?.side as DebateSide | null

  const reload = useCallback(async () => {
    const res = await fetchDebate(debate.id)
    onDebateUpdate(res.data)
  }, [debate.id, onDebateUpdate])

  const handleConcede = useCallback(async () => {
    if (!user?.email_verified) {
      setError(EMAIL_VERIFICATION_REQUIRED_MESSAGE)
      return
    }
    setIsLoading(true)
    setError(null)
    try {
      await concedeDebate(debate.id)
      await reload()
      setConfirmAction(null)
    } catch (err) {
      setError(err instanceof ApiError ? err.message : tCommon("somethingWrong"))
    } finally {
      setIsLoading(false)
    }
  }, [debate.id, reload, user?.email_verified])

  const handleResign = useCallback(async () => {
    if (!user?.email_verified) {
      setError(EMAIL_VERIFICATION_REQUIRED_MESSAGE)
      return
    }
    setIsLoading(true)
    setError(null)
    try {
      await resignDebate(debate.id)
      await reload()
      setConfirmAction(null)
    } catch (err) {
      setError(err instanceof ApiError ? err.message : tCommon("somethingWrong"))
    } finally {
      setIsLoading(false)
    }
  }, [debate.id, reload, user?.email_verified])

  const handleProposeDraw = useCallback(async () => {
    if (!user?.email_verified) {
      setError(EMAIL_VERIFICATION_REQUIRED_MESSAGE)
      return
    }
    setIsLoading(true)
    setError(null)
    try {
      await proposeDraw(debate.id)
      await reload()
    } catch (err) {
      setError(err instanceof ApiError ? err.message : tCommon("somethingWrong"))
    } finally {
      setIsLoading(false)
    }
  }, [debate.id, reload, user?.email_verified])

  const handleRespondDraw = useCallback(async (accept: boolean) => {
    if (!user?.email_verified) {
      setError(EMAIL_VERIFICATION_REQUIRED_MESSAGE)
      return
    }
    setIsLoading(true)
    setError(null)
    try {
      await respondDraw(debate.id, accept)
      await reload()
    } catch (err) {
      setError(err instanceof ApiError ? err.message : tCommon("somethingWrong"))
    } finally {
      setIsLoading(false)
    }
  }, [debate.id, reload, user?.email_verified])

  if (!debate.viewer?.is_participant || debate.status !== "active") {
    return null
  }

  const drawProposedByOpponent = debate.draw_proposed_by && debate.draw_proposed_by !== viewerSide
  const drawProposedByViewer = debate.draw_proposed_by === viewerSide

  return (
    <div className="mt-4 flex flex-col gap-3">
      {/* Draw proposal banner — opponent proposed */}
      {drawProposedByOpponent && (
        <div className="py-3 px-4 bg-[var(--warning-light)] border border-[var(--warning)] rounded-lg flex items-center justify-between gap-3">
          <p className="text-[13px] text-[var(--text-secondary)] font-sans">
            {t("drawProposed")}
          </p>
          <div className="flex gap-2 shrink-0">
            <Button
              variant="outline"
              size="sm"
              onClick={() => handleRespondDraw(false)}
              disabled={isLoading}
            >
              {t("decline")}
            </Button>
            <Button
              size="sm"
              onClick={() => handleRespondDraw(true)}
              disabled={isLoading}
            >
              {isLoading ? t("loading") : t("acceptDraw")}
            </Button>
          </div>
        </div>
      )}

      {/* Draw pending banner — viewer proposed */}
      {drawProposedByViewer && (
        <div className="py-3 px-4 bg-[var(--bg-surface-alt)] border border-[var(--border-default)] rounded-lg">
          <p className="text-[13px] text-[var(--text-muted)] font-sans">
            {t("pendingDraw")}
          </p>
        </div>
      )}

      {/* Confirmation inline bar */}
      {confirmAction && (
        <div className="flex items-center justify-between gap-3 bg-[var(--error-light)] border border-[var(--error)] rounded-lg px-4 py-3">
          <p className="text-[13px] text-[var(--text-secondary)] font-sans">
            {confirmAction === "concede"
              ? t("confirmConcede")
              : t("confirmResign")}
          </p>
          <div className="flex gap-2 shrink-0">
            <Button
              variant="outline"
              size="sm"
              onClick={() => { setConfirmAction(null); setError(null) }}
              disabled={isLoading}
            >
              {t("cancel")}
            </Button>
            <Button
              variant="destructive"
              size="sm"
              onClick={confirmAction === "concede" ? handleConcede : handleResign}
              disabled={isLoading}
            >
              {isLoading ? t("loading") : confirmAction === "concede" ? t("concede") : t("resign")}
            </Button>
          </div>
        </div>
      )}

      {error && (
        <div className="rounded-lg border border-[var(--error)] bg-[var(--error-light)] px-3 py-2 text-[12px] text-[var(--error)]">
          {error}
        </div>
      )}

      {/* Actions dropdown — only show when not in confirm mode */}
      {!confirmAction && (
        <div className="flex justify-end">
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="outline" size="sm" className="gap-1.5">
                <MoreHorizontal className="size-4" />
                {t("actions")}
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="w-48">
              {!debate.draw_proposed_by && (
                <DropdownMenuItem onClick={handleProposeDraw} disabled={isLoading}>
                  <HandshakeIcon className="size-4" />
                  {t("proposeDraw")}
                </DropdownMenuItem>
              )}
              <DropdownMenuItem onClick={() => setConfirmAction("concede")}>
                <Flag className="size-4" />
                {t("concede")}
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem
                variant="destructive"
                onClick={() => setConfirmAction("resign")}
              >
                <LogOut className="size-4" />
                {t("resign")}
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      )}
    </div>
  )
}
