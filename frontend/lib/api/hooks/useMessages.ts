"use client"

import { useQuery, useMutation, useQueryClient, useInfiniteQuery } from "@tanstack/react-query"
import api from "../client"
import type { Message, MessageFilter, SendMessageRequest, PaginationMeta } from "../types"

// Query keys
export const messageKeys = {
  all: ["messages"] as const,
  lists: () => [...messageKeys.all, "list"] as const,
  list: (chatId: string, filter?: Omit<MessageFilter, "chat_id">) => [...messageKeys.lists(), chatId, filter] as const,
  detail: (id: string) => [...messageKeys.all, "detail", id] as const,
}

// Get messages for a chat
export function useMessages(chatId: string | null, filter?: Omit<MessageFilter, "chat_id">) {
  return useQuery({
    queryKey: messageKeys.list(chatId || "", filter),
    queryFn: async () => {
      if (!chatId) return { messages: [], meta: undefined }

      const params = new URLSearchParams()
      if (filter?.page) params.set("page", filter.page.toString())
      if (filter?.per_page) params.set("per_page", filter.per_page.toString())

      const query = params.toString()
      const endpoint = `/chats/${chatId}/messages${query ? `?${query}` : ""}`
      const response = await api.get<Message[]>(endpoint)

      return {
        messages: response.data || [],
        meta: response.meta as PaginationMeta | undefined,
      }
    },
    enabled: !!chatId,
  })
}

// Get messages with infinite scroll (load more)
export function useMessagesInfinite(chatId: string | null) {
  return useInfiniteQuery({
    queryKey: messageKeys.list(chatId || ""),
    queryFn: async ({ pageParam = 1 }) => {
      if (!chatId) {
        return { messages: [], meta: undefined, nextPage: undefined }
      }

      const params = new URLSearchParams()
      params.set("page", pageParam.toString())
      params.set("per_page", "50")

      const response = await api.get<Message[]>(`/chats/${chatId}/messages?${params.toString()}`)

      return {
        messages: response.data || [],
        meta: response.meta as PaginationMeta | undefined,
        nextPage: response.meta && response.meta.page < response.meta.total_pages
          ? response.meta.page + 1
          : undefined,
      }
    },
    initialPageParam: 1,
    getNextPageParam: (lastPage) => lastPage.nextPage,
    enabled: !!chatId,
  })
}

// Send message
export function useSendMessage() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({ chatId, message }: { chatId: string; message: SendMessageRequest }) => {
      const response = await api.post<Message>(`/chats/${chatId}/messages`, message)
      return response.data!
    },
    onSuccess: (newMessage, { chatId }) => {
      // Add message to cache optimistically
      queryClient.setQueryData(
        messageKeys.list(chatId),
        (old: { messages: Message[]; meta?: PaginationMeta } | undefined) => {
          if (!old) return { messages: [newMessage], meta: undefined }
          return {
            ...old,
            messages: [...old.messages, newMessage],
          }
        }
      )
      // Invalidate chat list to update last message
      queryClient.invalidateQueries({ queryKey: ["chats", "list"] })
    },
  })
}

// Update message status locally (from NATS events)
export function useUpdateMessageStatus() {
  const queryClient = useQueryClient()

  return (chatId: string, messageId: string, status: Message["status"]) => {
    queryClient.setQueryData(
      messageKeys.list(chatId),
      (old: { messages: Message[]; meta?: PaginationMeta } | undefined) => {
        if (!old) return old
        return {
          ...old,
          messages: old.messages.map((msg) =>
            msg.id === messageId ? { ...msg, status } : msg
          ),
        }
      }
    )
  }
}

// Add message to cache (from NATS events)
export function useAddMessageToCache() {
  const queryClient = useQueryClient()

  return (chatId: string, message: Message) => {
    queryClient.setQueryData(
      messageKeys.list(chatId),
      (old: { messages: Message[]; meta?: PaginationMeta } | undefined) => {
        if (!old) return { messages: [message], meta: undefined }
        // Check if message already exists
        if (old.messages.some((m) => m.id === message.id)) {
          return old
        }
        return {
          ...old,
          messages: [...old.messages, message],
        }
      }
    )
    // Invalidate chat list to update last message
    queryClient.invalidateQueries({ queryKey: ["chats", "list"] })
  }
}
