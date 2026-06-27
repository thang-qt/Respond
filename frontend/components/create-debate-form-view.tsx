"use client"

import { ArrowLeft, MagnifyingGlass } from "@phosphor-icons/react"
import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import type { Tag } from "@/lib/debates"

export const turnLimitOptions = [10, 20, 30, 40]
export const minOpeningTurn = 100
export const maxOpeningTurn = 5000
export const maxOpeningTurnAINote = 300
export const aiChipKeys = ["brainstorm", "outline", "rewrite", "grammar", "sources"] as const
export const timeModeKeys = ["marathon", "standard", "rapid", "blitz"] as const

type Props = {
  addTag: (tagID: string) => void
  challengeUsername: string
  context: string
  contextCount: number
  filteredTags: Tag[]
  handleNext: () => void
  handleSubmit: (event: React.FormEvent<HTMLFormElement>) => void
  isChallengeMode: boolean
  isDisabled: boolean
  isRechallengeMode: boolean
  isSubmitting: boolean
  limitReached: boolean
  loadingTags: boolean
  openingTurn: string
  openingTurnAIAssisted: boolean
  openingTurnAINote: string
  openingTurnAINoteCount: number
  openingTurnCount: number
  prefillFromSource: boolean
  selectedTagIDs: string[]
  selectedTags: Tag[]
  selectTriggerClass: string
  setContext: (value: string) => void
  setOpeningTurn: (value: string) => void
  setOpeningTurnAIAssisted: (value: boolean) => void
  setOpeningTurnAINote: (value: string) => void
  setStep: (value: 1 | 2) => void
  setSubmitError: (value: string | null) => void
  setTagDropdownOpen: (value: boolean) => void
  setTagQuery: (value: string) => void
  setTimeMode: (value: "marathon" | "standard" | "rapid" | "blitz") => void
  setTopic: (value: string) => void
  setTurnLimit: (value: number) => void
  step: 1 | 2
  submitError: string | null
  tagDropdownOpen: boolean
  tagQuery: string
  t: any
  tDebate: any
  timeMode: "marathon" | "standard" | "rapid" | "blitz"
  toggleAINoteChip: (snippet: string) => void
  toggleTag: (tagID: string) => void
  topic: string
  topicCount: number
  turnLimit: number
  userEmailVerified?: boolean
}

export function CreateDebateFormView({
addTag,
  challengeUsername,
  context,
  contextCount,
  filteredTags,
  handleNext,
  handleSubmit,
  isChallengeMode,
  isDisabled,
  isRechallengeMode,
  isSubmitting,
  limitReached,
  loadingTags,
  openingTurn,
  openingTurnAIAssisted,
  openingTurnAINote,
  openingTurnAINoteCount,
  openingTurnCount,
  prefillFromSource,
  selectedTagIDs,
  selectedTags,
  selectTriggerClass,
  setContext,
  setOpeningTurn,
  setOpeningTurnAIAssisted,
  setOpeningTurnAINote,
  setStep,
  setSubmitError,
  setTagDropdownOpen,
  setTagQuery,
  setTimeMode,
  setTopic,
  setTurnLimit,
  step,
  submitError,
  tagDropdownOpen,
  tagQuery,
  t,
  tDebate,
  timeMode,
  toggleAINoteChip,
  toggleTag,
  topic,
  topicCount,
  turnLimit,
  userEmailVerified
}: Props) {
  return (
    <div className="min-h-screen bg-[var(--bg-primary)]">
      <div className="max-w-[820px] mx-auto px-4 sm:px-6 py-8">
        <div className="mb-6">
          <h1 className="text-[26px] font-semibold text-[var(--text-primary)] font-sans">
            {isRechallengeMode
              ? t("title.rechallenge")
              : isChallengeMode
                ? t("title.challenge", { username: challengeUsername })
                : t("title.new")}
          </h1>
          <p className="mt-2 text-[14px] text-[var(--text-secondary)] font-sans">
            {isRechallengeMode
              ? t("subtitle.rechallenge")
              : isChallengeMode
              ? t("subtitle.challenge")
              : step === 1
              ? t("subtitle.setup")
              : t("subtitle.opening")}
          </p>
        </div>

        {isChallengeMode && !isRechallengeMode && (
          <div className="mb-6 rounded-lg border border-[var(--border-default)] bg-[var(--bg-surface)] px-4 py-3 text-[13px] text-[var(--text-secondary)]">
            {t("banners.challenge", { username: challengeUsername })}
          </div>
        )}
        {isRechallengeMode && prefillFromSource && (
          <div className="mb-6 rounded-lg border border-[var(--border-default)] bg-[var(--bg-surface)] px-4 py-3 text-[13px] text-[var(--text-secondary)]">
            {t("banners.prefill")}
          </div>
        )}
        {isRechallengeMode && !prefillFromSource && (
          <div className="mb-6 rounded-lg border border-[var(--border-default)] bg-[var(--bg-surface)] px-4 py-3 text-[13px] text-[var(--text-secondary)]">
            {t("banners.newTopic")}
          </div>
        )}

        {limitReached && (
          <div className="mb-6 rounded-lg border border-[var(--warning)] bg-[var(--warning-light)] px-4 py-3 text-[13px] text-[var(--warning)]">
            {t("banners.limit")}
          </div>
        )}
        {!userEmailVerified && (
          <div className="mb-6 rounded-lg border border-[var(--warning)] bg-[var(--warning-light)] px-4 py-3 text-[13px] text-[var(--warning)]">
            {t("banners.verify")}
          </div>
        )}

        {/* Step indicator */}
        <div className="flex items-center gap-2 mb-6">
          <div className={`flex items-center justify-center w-6 h-6 rounded-full text-[11px] font-bold font-mono ${
            step === 1
              ? "bg-[var(--side-a)] text-white"
              : "bg-[var(--border-default)] text-[var(--text-muted)]"
          }`}>
            1
          </div>
          <span className={`text-[12px] font-sans ${step === 1 ? "text-[var(--text-primary)] font-medium" : "text-[var(--text-muted)]"}`}>
            {t("steps.setup")}
          </span>
          <div className="w-8 h-px bg-[var(--border-default)]" />
          <div className={`flex items-center justify-center w-6 h-6 rounded-full text-[11px] font-bold font-mono ${
            step === 2
              ? "bg-[var(--side-a)] text-white"
              : "bg-[var(--border-default)] text-[var(--text-muted)]"
          }`}>
            2
          </div>
          <span className={`text-[12px] font-sans ${step === 2 ? "text-[var(--text-primary)] font-medium" : "text-[var(--text-muted)]"}`}>
            {t("steps.opening")}
          </span>
        </div>

        {/* Step 1: Debate setup */}
        {step === 1 && (
          <div className="space-y-6">
            <section className="bg-[var(--bg-surface)] border border-[var(--border-default)] rounded-xl p-5 sm:p-6 space-y-5">
              <div className="flex items-center justify-between">
                <h2 className="text-[15px] font-semibold text-[var(--text-primary)] font-sans">{t("setup.title")}</h2>
                <span className="text-[11px] text-[var(--text-muted)] font-sans">{t("setup.sideA")}</span>
              </div>

              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <Label htmlFor="topic" className="text-[12px] text-[var(--text-secondary)]">
                    {t("setup.topic")}
                  </Label>
                  <span className="text-[11px] text-[var(--text-muted)]">{topicCount}/200</span>
                </div>
                <Input
                  id="topic"
                  name="topic"
                  placeholder={t("setup.topicPlaceholder")}
                  value={topic}
                  onChange={(event) => setTopic(event.target.value)}
                  className="bg-[var(--bg-primary)] border-[var(--border-default)]"
                  required
                />
                <p className="text-[11px] text-[var(--text-muted)]">{t("setup.topicHelp")}</p>
              </div>

              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <Label className="text-[12px] text-[var(--text-secondary)]">{t("setup.tags")}</Label>
                  <span className="text-[11px] text-[var(--text-muted)]">{t("setup.selected", { count: selectedTagIDs.length })}</span>
                </div>
                <div className="space-y-2">
                  {selectedTags.length > 0 && (
                    <div className="flex flex-wrap gap-2">
                      {selectedTags.map((tag) => (
                        <button
                          key={tag.id}
                          type="button"
                          onClick={() => toggleTag(tag.id)}
                          className="px-2.5 py-1 text-[12px] rounded-full border bg-[var(--text-primary)] text-[var(--bg-primary)] border-[var(--text-primary)] font-sans"
                        >
                          {tag.name} x
                        </button>
                      ))}
                    </div>
                  )}

                  <div className="relative">
                    <div className="relative">
                      <MagnifyingGlass
                        size={14}
                        className="absolute left-3 top-1/2 -translate-y-1/2 text-[var(--text-muted)]"
                        aria-hidden
                      />
                      <Input
                        value={tagQuery}
                        onChange={(event) => {
                          setTagQuery(event.target.value)
                          setTagDropdownOpen(true)
                        }}
                        onFocus={() => setTagDropdownOpen(true)}
                        onBlur={() => window.setTimeout(() => setTagDropdownOpen(false), 120)}
                        placeholder={selectedTagIDs.length >= 3 ? t("setup.tagLimit") : t("setup.searchTags")}
                        className="bg-[var(--bg-primary)] border-[var(--border-default)] pl-9"
                        disabled={selectedTagIDs.length >= 3}
                      />
                    </div>
                    {tagDropdownOpen && selectedTagIDs.length < 3 && (
                      <div className="absolute z-30 mt-1 w-full overflow-hidden rounded-md border border-[var(--border-default)] bg-[var(--bg-surface)] shadow-lg">
                        {filteredTags.length > 0 ? (
                          <div className="max-h-56 overflow-y-auto py-1">
                            {filteredTags.map((tag) => (
                              <button
                                key={tag.id}
                                type="button"
                                onMouseDown={(event) => event.preventDefault()}
                                onClick={() => addTag(tag.id)}
                                className="w-full text-left px-3 py-2 text-[13px] text-[var(--text-primary)] hover:bg-[var(--bg-surface-alt)] font-sans"
                              >
                                {tag.name}
                                <span className="ml-2 text-[11px] text-[var(--text-muted)]">/{tag.slug}</span>
                              </button>
                            ))}
                          </div>
                        ) : (
                          <div className="px-3 py-2 text-[12px] text-[var(--text-muted)] font-sans">{t("setup.noTags")}</div>
                        )}
                      </div>
                    )}
                  </div>

                </div>
              </div>

              <div className="grid gap-4 sm:grid-cols-2">
                <div className="space-y-2">
                  <Label className="text-[12px] text-[var(--text-secondary)]">{t("setup.timeMode")}</Label>
                  <Select value={timeMode} onValueChange={(value) => setTimeMode(value as typeof timeMode)}>
                    <SelectTrigger className={selectTriggerClass}>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      {timeModeKeys.map((value) => (
                        <SelectItem key={value} value={value}>
                          {tDebate(`timeMode.${value}`)}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
                <div className="space-y-2">
                  <Label className="text-[12px] text-[var(--text-secondary)]">{t("setup.turnLimit")}</Label>
                  <Select value={String(turnLimit)} onValueChange={(value) => setTurnLimit(Number(value))}>
                    <SelectTrigger className={selectTriggerClass}>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      {turnLimitOptions.map((limit) => (
                        <SelectItem key={limit} value={String(limit)}>
                          {t("setup.turns", { count: limit })}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
              </div>

              <div className="space-y-2 border-t border-[var(--border-subtle)] pt-5">
                <div className="flex items-center justify-between">
                  <Label htmlFor="context" className="text-[12px] text-[var(--text-secondary)]">
                    {t("setup.context")}
                  </Label>
                  <span className="text-[11px] text-[var(--text-muted)]">{contextCount}/500</span>
                </div>
                <Textarea
                  id="context"
                  name="context"
                  placeholder={t("setup.contextPlaceholder")}
                  value={context}
                  onChange={(event) => setContext(event.target.value)}
                  className="min-h-[120px] bg-[var(--bg-primary)] border-[var(--border-default)]"
                />
                <p className="text-[11px] text-[var(--text-muted)]">
                  {t("setup.contextHelp")}
                </p>
              </div>
            </section>

            {submitError && (
              <div className="rounded-lg border border-[var(--error)] bg-[var(--error-light)] px-3 py-2 text-[12px] text-[var(--error)]">
                {submitError}
              </div>
            )}

            <section className="flex justify-end">
              <Button type="button" onClick={handleNext} disabled={loadingTags || limitReached}>
                {t("setup.next")}
              </Button>
            </section>
          </div>
        )}

        {/* Step 2: Opening argument */}
        {step === 2 && (
          <form onSubmit={handleSubmit} className="space-y-6">
            {/* Topic summary */}
            <div className="bg-[var(--bg-surface)] border border-[var(--border-default)] rounded-xl px-5 sm:px-6 py-4">
              <p className="text-[12px] text-[var(--text-muted)] font-sans mb-1">{t("opening.topic")}</p>
              <p className="text-[15px] text-[var(--text-primary)] font-semibold font-sans">{topic.trim()}</p>
            </div>

            {/* Opening argument input */}
            <section className="bg-[var(--bg-surface)] border border-[var(--border-default)] rounded-xl p-5 sm:p-6 space-y-4">
              <div className="flex items-center justify-between">
                <h2 className="text-[15px] font-semibold text-[var(--text-primary)] font-sans">{t("opening.title")}</h2>
                <span className={`text-[11px] font-sans ${
                  openingTurnCount < minOpeningTurn
                    ? "text-[var(--text-muted)]"
                    : openingTurnCount > maxOpeningTurn
                      ? "text-[var(--error)]"
                      : "text-[var(--text-muted)]"
                }`}>
                  {openingTurnCount.toLocaleString()}/{maxOpeningTurn.toLocaleString()}
                </span>
              </div>

              <Textarea
                id="opening-turn"
                name="opening_turn"
                placeholder={t("opening.placeholder")}
                value={openingTurn}
                onChange={(event) => setOpeningTurn(event.target.value)}
                className="min-h-[240px] bg-[var(--bg-primary)] border-[var(--border-default)]"
                autoFocus
              />

              <p className="text-[11px] text-[var(--text-muted)]">
                  {t("opening.help", { min: minOpeningTurn, max: maxOpeningTurn.toLocaleString() })}
              </p>

              <div className="rounded-md border border-[var(--border-default)] bg-[var(--bg-primary)] px-3 py-3 space-y-2">
                <div className="flex items-center gap-2">
                  <Checkbox
                    id="opening-turn-ai-assisted"
                    checked={openingTurnAIAssisted}
                    onCheckedChange={(checked) => setOpeningTurnAIAssisted(Boolean(checked))}
                    disabled={isSubmitting}
                  />
                  <Label
                    htmlFor="opening-turn-ai-assisted"
                    className="text-[12px] text-[var(--text-secondary)] font-normal cursor-pointer"
                  >
                    {t("opening.aiAssisted")}
                  </Label>
                </div>

                {openingTurnAIAssisted && (
                  <div className="space-y-2">
                    <div className="flex flex-wrap gap-2">
                      {aiChipKeys.map((key) => {
                        const snippet = t(`aiChips.${key}.snippet`)
                        const active = openingTurnAINote.includes(snippet)
                        return (
                          <button
                            key={key}
                            type="button"
                            onClick={() => toggleAINoteChip(snippet)}
                            className={`px-2 py-0.5 rounded text-[11px] border transition-colors ${
                              active
                                ? "bg-[var(--text-primary)] text-[var(--bg-primary)] border-[var(--text-primary)]"
                                : "bg-transparent text-[var(--text-secondary)] border-[var(--border-default)] hover:bg-[var(--bg-surface-alt)]"
                            }`}
                            disabled={isSubmitting}
                          >
                            {t(`aiChips.${key}.label`)}
                          </button>
                        )
                      })}
                    </div>

                    <Textarea
                      placeholder={t("opening.aiNotePlaceholder")}
                      value={openingTurnAINote}
                      onChange={(event) => setOpeningTurnAINote(event.target.value)}
                      className="min-h-[72px] bg-[var(--bg-surface)] border-[var(--border-default)]"
                      disabled={isSubmitting}
                    />
                    <div className={`text-[11px] text-right ${
                      openingTurnAINoteCount > maxOpeningTurnAINote ? "text-[var(--error)]" : "text-[var(--text-muted)]"
                    }`}>
                      {openingTurnAINoteCount}/{maxOpeningTurnAINote}
                    </div>
                  </div>
                )}
              </div>
            </section>

            {submitError && (
              <div className="rounded-lg border border-[var(--error)] bg-[var(--error-light)] px-3 py-2 text-[12px] text-[var(--error)]">
                {submitError}
              </div>
            )}

            <section className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
              <button
                type="button"
                onClick={() => { setStep(1); setSubmitError(null) }}
                className="flex items-center gap-1.5 text-[13px] text-[var(--text-secondary)] hover:text-[var(--text-primary)] transition-colors font-sans"
              >
                <ArrowLeft size={14} />
                {t("opening.back")}
              </button>
              <Button type="submit" disabled={isDisabled}>
                {isSubmitting ? t("opening.creating") : isChallengeMode || isRechallengeMode ? t("opening.sendChallenge") : t("opening.createDebate")}
              </Button>
            </section>
          </form>
        )}
      </div>
    </div>
  )
}
