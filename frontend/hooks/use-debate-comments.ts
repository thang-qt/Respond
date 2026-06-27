"use client"

import { useCallback, useEffect, useState } from "react"
import { toast } from "sonner"
import type { DebateComment } from "@/lib/debates"
import {
  deleteDebateComment,
  fetchDebateComments,
  postDebateComment,
  toggleCommentVote,
  updateDebateComment,
} from "@/lib/debates-api"
import { EMAIL_VERIFICATION_REQUIRED_MESSAGE, isEmailNotVerifiedError } from "@/lib/verification"

type CommentSort = "newest" | "top"

type UseDebateCommentsOptions = {
  debateId: string
  requireVerified: () => boolean
}

function updateCommentVote(
  comments: DebateComment[],
  commentId: string,
  upvoteCount: number,
  voted: boolean
): DebateComment[] {
  return comments.map((comment) => {
    if (comment.id === commentId) {
      return { ...comment, upvote_count: upvoteCount, viewer_has_upvoted: voted }
    }
    if (comment.replies && comment.replies.length > 0) {
      return { ...comment, replies: updateCommentVote(comment.replies, commentId, upvoteCount, voted) }
    }
    return comment
  })
}

export function useDebateComments({ debateId, requireVerified }: UseDebateCommentsOptions) {
  const [comments, setComments] = useState<DebateComment[]>([])
  const [commentTotal, setCommentTotal] = useState(0)
  const [commentSort, setCommentSort] = useState<CommentSort>(() => {
    if (typeof window === "undefined") return "newest"
    const stored = window.localStorage.getItem("respond.comment_sort")
    return stored === "top" ? "top" : "newest"
  })
  const [highlightCommentId, setHighlightCommentId] = useState<string | null>(null)

  const loadComments = useCallback(
    async (activeFlag: { current: boolean }) => {
      try {
        const commentRes = await fetchDebateComments(debateId, { sort: commentSort })
        if (!activeFlag.current) return
        setComments(commentRes.data)
        const total = commentRes.data.reduce(
          (sum, comment) => sum + 1 + (comment.replies?.length ?? 0),
          0
        )
        setCommentTotal(total)
      } catch {
        if (!activeFlag.current) return
        setComments([])
        setCommentTotal(0)
      }
    },
    [debateId, commentSort]
  )

  useEffect(() => {
    window.localStorage.setItem("respond.comment_sort", commentSort)
  }, [commentSort])

  const handlePostComment = useCallback(
    async (content: string, isReflection: boolean) => {
      if (!requireVerified()) {
        throw new Error(EMAIL_VERIFICATION_REQUIRED_MESSAGE)
      }
      const res = await postDebateComment(debateId, { content, is_reflection: isReflection || undefined })
      setHighlightCommentId(res.data.id)
      const activeFlag = { current: true }
      await loadComments(activeFlag)
    },
    [debateId, loadComments, requireVerified]
  )

  const handlePostReply = useCallback(
    async (parentId: string, content: string) => {
      if (!requireVerified()) {
        throw new Error(EMAIL_VERIFICATION_REQUIRED_MESSAGE)
      }
      const res = await postDebateComment(debateId, { content, parent_id: parentId })
      setHighlightCommentId(res.data.id)
      const activeFlag = { current: true }
      await loadComments(activeFlag)
    },
    [debateId, loadComments, requireVerified]
  )

  const handleToggleCommentVote = useCallback(async (commentId: string) => {
    if (!requireVerified()) return
    try {
      const res = await toggleCommentVote(commentId)
      setComments((prev) => updateCommentVote(prev, res.data.comment_id, res.data.upvote_count, res.data.voted))
    } catch (err) {
      if (isEmailNotVerifiedError(err)) {
        toast.error(EMAIL_VERIFICATION_REQUIRED_MESSAGE)
      } else if (err instanceof Error) {
        toast.error(err.message)
      } else {
        toast.error("Failed to vote.")
      }
    }
  }, [requireVerified])

  const handleEditComment = useCallback(
    async (commentId: string, content: string) => {
      if (!requireVerified()) {
        throw new Error(EMAIL_VERIFICATION_REQUIRED_MESSAGE)
      }
      await updateDebateComment(debateId, commentId, { content })
      const activeFlag = { current: true }
      await loadComments(activeFlag)
    },
    [debateId, loadComments, requireVerified]
  )

  const handleDeleteComment = useCallback(
    async (commentId: string) => {
      if (!requireVerified()) {
        throw new Error(EMAIL_VERIFICATION_REQUIRED_MESSAGE)
      }
      await deleteDebateComment(debateId, commentId)
      const activeFlag = { current: true }
      await loadComments(activeFlag)
    },
    [debateId, loadComments, requireVerified]
  )

  return {
    comments,
    commentSort,
    commentTotal,
    handleDeleteComment,
    handleEditComment,
    handlePostComment,
    handlePostReply,
    handleToggleCommentVote,
    highlightCommentId,
    loadComments,
    setCommentSort,
    setCommentTotal,
    setComments,
    setHighlightCommentId,
  }
}
