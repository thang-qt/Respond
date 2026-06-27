"use client"

import { useCallback, useEffect, useState } from "react"
import Link from "next/link"
import { useTranslations } from "next-intl"
import { fetchTags } from "@/lib/debates-api"
import {
    fetchLobbyEntries,
    getMyLobbyEntry,
    upsertMyLobbyEntry,
    deleteMyLobbyEntry,
} from "@/lib/debates-api"
import type { Tag, ChallengeLobbyEntry } from "@/lib/debates"
import { ApiError } from "@/lib/api"
import { useAuth } from "@/hooks/use-auth"
import { Button } from "@/components/ui/button"
import { formatTimeAgo } from "@/lib/utils"
import LobbyEntryCard from "@/components/lobby-entry-card"

interface LobbyTabProps {
    mode?: "all" | "browse" | "my-entry"
    defaultMyEntryOpen?: boolean
    showSectionHeaders?: boolean
    showMyEntrySection?: boolean
    filterTagSlugs?: string[]
    showMyEntryToggle?: boolean
}

export default function LobbyTab({
    mode = "all",
    defaultMyEntryOpen = false,
    showSectionHeaders = true,
    showMyEntrySection,
    filterTagSlugs,
    showMyEntryToggle = true,
}: LobbyTabProps) {
    const { user } = useAuth()
    const t = useTranslations("lobbyTab")
    // ── My Entry ──────────────────────────────────────────────────────────────
    const [myEntry, setMyEntry] = useState<ChallengeLobbyEntry | null>(null)
    const [myEntryLoading, setMyEntryLoading] = useState(true)
    const [bioNote, setBioNote] = useState("")
    const [selectedTagIds, setSelectedTagIds] = useState<string[]>([])
    const [allTags, setAllTags] = useState<Tag[]>([])
    const [saveLoading, setSaveLoading] = useState(false)
    const [deleteLoading, setDeleteLoading] = useState(false)
    const [myError, setMyError] = useState<string | null>(null)
    const [mySuccess, setMySuccess] = useState<string | null>(null)
    const [showMyEntryEditor, setShowMyEntryEditor] = useState(defaultMyEntryOpen)

    useEffect(() => {
        if (defaultMyEntryOpen) {
            setShowMyEntryEditor(true)
        }
    }, [defaultMyEntryOpen])

    // ── Browse ────────────────────────────────────────────────────────────────
    const [entries, setEntries] = useState<ChallengeLobbyEntry[]>([])
    const [browseLoading, setBrowseLoading] = useState(true)
    const [browseLoadingMore, setBrowseLoadingMore] = useState(false)
    const [page, setPage] = useState(1)
    const [totalPages, setTotalPages] = useState(1)
    const [internalFilterTagSlugs] = useState<string[]>([])
    const [browseError, setBrowseError] = useState<string | null>(null)

    const effectiveFilterTagSlugs = filterTagSlugs ?? internalFilterTagSlugs

    const perPage = 20

    // Load all system tags for selects
    useEffect(() => {
        fetchTags()
            .then((res) => setAllTags(res.data ?? []))
            .catch(() => { })
    }, [])

    // Load own entry
    useEffect(() => {
        if (!user) {
            setMyEntryLoading(false)
            return
        }
        setMyEntryLoading(true)
        getMyLobbyEntry()
            .then((res) => {
                setMyEntry(res.data)
                setBioNote(res.data.bio_note)
                setSelectedTagIds(res.data.tags.map((t) => t.id))
            })
            .catch((err) => {
                // 404 means no entry yet — that's fine
                if (!(err instanceof ApiError) || err.status !== 404) {
                    setMyError(t("errors.loadMy"))
                }
            })
            .finally(() => setMyEntryLoading(false))
    }, [user])

    // Load browse list
    const loadBrowse = useCallback(async (nextPage: number, append: boolean) => {
        if (append) setBrowseLoadingMore(true)
        else setBrowseLoading(true)
        setBrowseError(null)
        try {
            const res = await fetchLobbyEntries({
                tagSlugs: effectiveFilterTagSlugs.length > 0 ? effectiveFilterTagSlugs : undefined,
                page: nextPage,
                perPage,
            })
            setEntries((prev) => (append ? [...prev, ...(res.data ?? [])] : (res.data ?? [])))
            setPage(res.meta?.page ?? nextPage)
            setTotalPages(res.meta?.total_pages ?? 1)
        } catch {
            setBrowseError(t("errors.loadBrowse"))
        } finally {
            setBrowseLoading(false)
            setBrowseLoadingMore(false)
        }
    }, [effectiveFilterTagSlugs])

    useEffect(() => {
        void loadBrowse(1, false)
    }, [loadBrowse])

    async function handleSave() {
        if (!user?.email_verified) {
            setMyError(t("errors.verify"))
            return
        }
        setSaveLoading(true)
        setMyError(null)
        setMySuccess(null)
        try {
            const res = await upsertMyLobbyEntry(bioNote, selectedTagIds)
            setMyEntry(res.data)
            setMySuccess(t("saved"))
            // Refresh browse list so own entry is visible
            void loadBrowse(1, false)
        } catch (err) {
            setMyError(err instanceof ApiError ? err.message : t("errors.save"))
        } finally {
            setSaveLoading(false)
        }
    }

    async function handleDelete() {
        setDeleteLoading(true)
        setMyError(null)
        setMySuccess(null)
        try {
            await deleteMyLobbyEntry()
            setMyEntry(null)
            setBioNote("")
            setSelectedTagIds([])
            setMySuccess(t("removed"))
            void loadBrowse(1, false)
        } catch (err) {
            setMyError(err instanceof ApiError ? err.message : t("errors.remove"))
        } finally {
            setDeleteLoading(false)
        }
    }

    function toggleTagId(id: string) {
        setSelectedTagIds((prev) =>
            prev.includes(id) ? prev.filter((t) => t !== id) : prev.length >= 15 ? prev : [...prev, id]
        )
    }

    const bioNoteLength = [...bioNote].length

    const shouldShowMyEntrySection = showMyEntrySection ?? (mode === "all" || mode === "my-entry")
    const showBrowseSection = mode === "all" || mode === "browse"
    const isEditorVisible = showMyEntryToggle ? showMyEntryEditor : true

    return (
        <div className="space-y-6">
            {/* ── My Entry ── */}
            {shouldShowMyEntrySection && user ? (
                <section className="rounded-lg border border-[var(--border-default)] bg-[var(--bg-surface)] p-3">
                    <div className="flex flex-wrap items-center justify-between gap-2">
                        <div>
                            {showSectionHeaders && (
                                <>
                                    <h2 className="text-[13px] font-semibold text-[var(--text-primary)] font-sans">
                                        {myEntry ? t("yourTitle") : t("postTitle")}
                                    </h2>
                                    <p className="text-[11px] text-[var(--text-muted)] font-sans">
                                        {t("description")}
                                    </p>
                                </>
                            )}
                            {!showSectionHeaders && (
                                <div className="text-[12px] font-medium text-[var(--text-secondary)] font-sans">{t("standing")}</div>
                            )}
                            {myEntry && (
                                <p className={`${showSectionHeaders ? "mt-1" : ""} text-[11px] text-[var(--text-muted)] font-sans`}>
                                    {t("lastUpdated", { time: formatTimeAgo(myEntry.updated_at) })}
                                </p>
                            )}
                        </div>
                        {showMyEntryToggle && (
                            <button
                                type="button"
                                onClick={() => setShowMyEntryEditor((current) => !current)}
                                className="px-3 py-1.5 rounded-md text-[12px] font-medium border border-[var(--border-default)] text-[var(--text-secondary)] hover:bg-[var(--bg-surface-alt)] hover:text-[var(--text-primary)]"
                            >
                                {showMyEntryEditor ? t("hide") : myEntry ? t("edit") : t("open")}
                            </button>
                        )}
                    </div>

                    {isEditorVisible && (
                        <div className="mt-3 border-t border-[var(--border-default)] pt-3 space-y-4">
                            {myEntryLoading ? (
                                <div className="text-[13px] text-[var(--text-muted)]">{t("loading")}</div>
                            ) : (
                                <>
                                    <div>
                                        <label className="block text-[12px] font-medium text-[var(--text-secondary)] mb-1">
                                            {t("prompt")} <span className="text-[var(--text-muted)]">{t("optional")}</span>
                                        </label>
                                        <textarea
                                            value={bioNote}
                                            onChange={(e) => setBioNote(e.target.value)}
                                            maxLength={300}
                                            rows={3}
                                            placeholder={t("placeholder")}
                                            className="w-full rounded-md border border-[var(--border-default)] bg-[var(--bg-primary)] px-3 py-2 text-[13px] text-[var(--text-primary)] placeholder:text-[var(--text-muted)] resize-none focus:outline-none focus:border-[var(--text-secondary)]"
                                        />
                                        <div className={`text-right text-[11px] mt-0.5 ${bioNoteLength > 280 ? "text-[var(--warning)]" : "text-[var(--text-muted)]"}`}>
                                            {bioNoteLength}/300
                                        </div>
                                    </div>

                                    <div>
                                        <div className="flex items-center justify-between mb-2">
                                            <label className="text-[12px] font-medium text-[var(--text-secondary)]">
                                                {t("topics")} <span className="text-[var(--text-muted)]">{t("upTo")}</span>
                                            </label>
                                            {selectedTagIds.length > 0 && (
                                                <button
                                                    type="button"
                                                    onClick={() => setSelectedTagIds([])}
                                                    className="text-[11px] text-[var(--text-muted)] hover:text-[var(--text-secondary)]"
                                                >
                                                    {t("clearAll")}
                                                </button>
                                            )}
                                        </div>
                                        <div className="flex flex-wrap gap-1.5">
                                            {allTags.map((tag) => {
                                                const selected = selectedTagIds.includes(tag.id)
                                                return (
                                                    <button
                                                        key={tag.id}
                                                        type="button"
                                                        onClick={() => toggleTagId(tag.id)}
                                                        disabled={!selected && selectedTagIds.length >= 15}
                                                        className={`px-2.5 py-1 rounded-full text-[11px] font-medium border transition-colors ${selected
                                                            ? "bg-[var(--text-primary)] text-[var(--bg-primary)] border-[var(--text-primary)]"
                                                            : "bg-transparent text-[var(--text-secondary)] border-[var(--border-default)] hover:border-[var(--text-secondary)] disabled:opacity-40"
                                                            }`}
                                                    >
                                                        {tag.name}
                                                    </button>
                                                )
                                            })}
                                        </div>
                                    </div>

                                    {myError && (
                                        <div className="rounded-md border border-[var(--error)] bg-[var(--error-light)] px-3 py-2 text-[12px] text-[var(--error)]">
                                            {myError}
                                        </div>
                                    )}
                                    {mySuccess && (
                                        <div className="rounded-md border border-[var(--success)] bg-[var(--success-light)] px-3 py-2 text-[12px] text-[var(--success)]">
                                            {mySuccess}
                                        </div>
                                    )}

                                    <div className="flex items-center gap-2">
                                        <Button size="sm" onClick={handleSave} disabled={saveLoading}>
                                            {saveLoading ? t("saving") : myEntry ? t("update") : t("publish")}
                                        </Button>
                                        {myEntry && (
                                            <Button size="sm" variant="outline" onClick={handleDelete} disabled={deleteLoading}>
                                                {deleteLoading ? t("removing") : t("remove")}
                                            </Button>
                                        )}
                                    </div>
                                </>
                            )}
                        </div>
                    )}
                </section>
            ) : shouldShowMyEntrySection ? (
                <div className="bg-[var(--bg-surface)] border border-[var(--border-default)] rounded-lg p-4 text-[13px] text-[var(--text-secondary)] font-sans">
                    <Link href="/auth/login?redirect=%2Fchallenges" className="underline text-[var(--text-primary)]">
                        {t("signIn")}
                    </Link>{" "}
                    {t("signInSuffix")}
                </div>
            ) : null}

            {/* ── Browse ── */}
            {showBrowseSection && (
            <section>
                {browseError && (
                    <div className="mb-4 rounded-md border border-[var(--error)] bg-[var(--error-light)] px-3 py-2 text-[12px] text-[var(--error)]">
                        {browseError}
                    </div>
                )}

                {browseLoading ? (
                    <div className="text-[13px] text-[var(--text-muted)] font-sans">{t("loadingChallengers")}</div>
                ) : entries.length === 0 ? (
                    <div className="bg-[var(--bg-surface)] border border-[var(--border-default)] rounded-lg p-4 text-[13px] text-[var(--text-secondary)] font-sans">
                        {effectiveFilterTagSlugs.length > 0
                            ? t("emptyFiltered")
                            : t("empty")}
                    </div>
                ) : (
                    <div className="space-y-2">
                        {entries.map((entry) => (
                            <LobbyEntryCard key={entry.username} entry={entry} />
                        ))}
                        {page < totalPages && (
                            <div className="pt-2 flex justify-center">
                                <button
                                    type="button"
                                    onClick={() => void loadBrowse(page + 1, true)}
                                    disabled={browseLoadingMore}
                                    className="px-4 py-2 rounded-md text-[12px] font-medium text-[var(--text-secondary)] border border-[var(--border-default)] hover:bg-[var(--bg-surface-alt)] hover:text-[var(--text-primary)] disabled:opacity-60"
                                >
                                    {browseLoadingMore ? t("loading") : t("loadMore")}
                                </button>
                            </div>
                        )}
                    </div>
                )}
            </section>
            )}
        </div>
    )
}
