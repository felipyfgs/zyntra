"use client"

import { useQuery, useMutation, useQueryClient, useInfiniteQuery } from "@tanstack/react-query"
import api from "../client"
import type { Message, SendMessageRequest, PaginationMeta } from "../types"

export const messageKeys = {
  all: ["messages"] as const,
  lists: () => [...messageKeys.all, "list"] as const,
  list: (conversationId: string, filter?: { limit?: number; offset?: number }) =>
    [...messageKeys.lists(), conversationId, filter] as const,
  detail: (id: string) => [...messageKeys.all, "detail", id] as const,
}

export function useMessages(conversationId: string | null, filter?: { limit?: number; offset?: number }) {
  return useQuery({
    queryKey: messageKeys.list(conversationId || "", filter),
    queryFn: async () => {
      if (!conversationId) return { messages: [], meta: undefined }

      const params = new URLSearchParams()
      if (filter?.limit) params.set("limit", filter.limit.toString())
      if (filter?.offset) params.set("offset", filter.offset.toString())

      const query = params.toString()
      const endpoint = `/conversations/${conversationId}/messages${query ? `?${query}` : ""}`
      const response = await api.get<Message[]>(endpoint)

      return {
        messages: response.data || [],
        meta: response.meta as PaginationMeta | undefined,
      }
    },
    enabled: !!conversationId,
  })
}

export function useMessagesInfinite(conversationId: string | null) {
  return useInfiniteQuery({
    queryKey: messageKeys.list(conversationId || ""),
    queryFn: async ({ pageParam = 0 }) => {
      if (!conversationId) {
        return { messages: [], meta: undefined, nextOffset: undefined }
      }

      const params = new URLSearchParams()
      params.set("offset", pageParam.toString())
      params.set("limit", "50")

      const response = await api.get<Message[]>(`/conversations/${conversationId}/messages?${params.toString()}`)

      return {
        messages: response.data || [],
        meta: response.meta as PaginationMeta | undefined,
        nextOffset: (response.data?.length || 0) === 50 ? pageParam + 50 : undefined,
      }
    },
    initialPageParam: 0,
    getNextPageParam: (lastPage) => lastPage.nextOffset,
    enabled: !!conversationId,
  })
}

export function useSendMessage() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({ conversationId, message }: { conversationId: string; message: SendMessageRequest }) => {
      const response = await api.post<Message>(`/conversations/${conversationId}/messages`, message)
      return response.data!
    },
    onSuccess: (newMessage, { conversationId }) => {
      queryClient.setQueryData(
        messageKeys.list(conversationId),
        (old: { messages: Message[]; meta?: PaginationMeta } | undefined) => {
          if (!old) return { messages: [newMessage], meta: undefined }
          return {
            ...old,
            messages: [...old.messages, newMessage],
          }
        }
      )
      queryClient.invalidateQueries({ queryKey: ["conversations", "list"] })
    },
  })
}

export function useUpdateMessageStatus() {
  const queryClient = useQueryClient()

  return (conversationId: string, messageId: string, status: Message["status"]) => {
    queryClient.setQueryData(
      messageKeys.list(conversationId),
      (old: { messages: Message[]; meta?: PaginationMeta } | undefined) => {
        if (!old) return old
        return {
          ...old,
          messages: old.messages.map((msg) => (msg.id === messageId ? { ...msg, status } : msg)),
        }
      }
    )
  }
}

export function useAddMessageToCache() {
  const queryClient = useQueryClient()

  return (conversationId: string, message: Message) => {
    queryClient.setQueryData(
      messageKeys.list(conversationId),
      (old: { messages: Message[]; meta?: PaginationMeta } | undefined) => {
        if (!old) return { messages: [message], meta: undefined }
        if (old.messages.some((m) => m.id === message.id)) {
          return old
        }
        return {
          ...old,
          messages: [...old.messages, message],
        }
      }
    )
    queryClient.invalidateQueries({ queryKey: ["conversations", "list"] })
  }
}
