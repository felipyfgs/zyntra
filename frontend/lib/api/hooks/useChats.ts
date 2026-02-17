"use client"

import { useQuery, useMutation, useQueryClient, useInfiniteQuery } from "@tanstack/react-query"
import api from "../client"
import type { Chat, ChatFilter, PaginationMeta } from "../types"

// Query keys
export const chatKeys = {
  all: ["chats"] as const,
  lists: () => [...chatKeys.all, "list"] as const,
  list: (filter?: ChatFilter) => [...chatKeys.lists(), filter] as const,
  details: () => [...chatKeys.all, "detail"] as const,
  detail: (id: string) => [...chatKeys.details(), id] as const,
}

// Get list of chats
export function useChats(filter?: ChatFilter) {
  return useQuery({
    queryKey: chatKeys.list(filter),
    queryFn: async () => {
      const params = new URLSearchParams()
      if (filter?.connection_id) params.set("connection_id", filter.connection_id)
      if (filter?.search) params.set("search", filter.search)
      if (filter?.filter) params.set("filter", filter.filter)
      if (filter?.page) params.set("page", filter.page.toString())
      if (filter?.per_page) params.set("per_page", filter.per_page.toString())

      const query = params.toString()
      const endpoint = `/chats${query ? `?${query}` : ""}`
      const response = await api.get<Chat[]>(endpoint)
      
      return {
        chats: response.data || [],
        meta: response.meta as PaginationMeta | undefined,
      }
    },
  })
}

// Get infinite list of chats (for lazy loading)
export function useChatsInfinite(filter?: Omit<ChatFilter, "page">) {
  return useInfiniteQuery({
    queryKey: chatKeys.list(filter),
    queryFn: async ({ pageParam = 1 }) => {
      const params = new URLSearchParams()
      if (filter?.connection_id) params.set("connection_id", filter.connection_id)
      if (filter?.search) params.set("search", filter.search)
      if (filter?.filter) params.set("filter", filter.filter)
      params.set("page", pageParam.toString())
      params.set("per_page", "20")

      const query = params.toString()
      const response = await api.get<Chat[]>(`/chats?${query}`)

      return {
        chats: response.data || [],
        meta: response.meta as PaginationMeta | undefined,
        nextPage: response.meta && response.meta.page < response.meta.total_pages 
          ? response.meta.page + 1 
          : undefined,
      }
    },
    initialPageParam: 1,
    getNextPageParam: (lastPage) => lastPage.nextPage,
  })
}

// Get single chat
export function useChat(chatId: string | null) {
  return useQuery({
    queryKey: chatKeys.detail(chatId || ""),
    queryFn: async () => {
      if (!chatId) return null
      const response = await api.get<Chat>(`/chats/${chatId}`)
      return response.data!
    },
    enabled: !!chatId,
  })
}

// Update chat (favorite, archive)
export function useUpdateChat() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({ chatId, updates }: { chatId: string; updates: Partial<Pick<Chat, "is_favorite" | "is_archived">> }) => {
      const response = await api.put<Chat>(`/chats/${chatId}`, updates)
      return response.data!
    },
    onSuccess: (data) => {
      // Update specific chat in cache
      queryClient.setQueryData(chatKeys.detail(data.id), data)
      // Invalidate chat lists to refresh
      queryClient.invalidateQueries({ queryKey: chatKeys.lists() })
    },
  })
}

// Mark chat as read
export function useMarkChatAsRead() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (chatId: string) => {
      const response = await api.post<{ message: string }>(`/chats/${chatId}/read`)
      return response.data!
    },
    onSuccess: (_, chatId) => {
      // Invalidate chat to refresh unread count
      queryClient.invalidateQueries({ queryKey: chatKeys.detail(chatId) })
      queryClient.invalidateQueries({ queryKey: chatKeys.lists() })
    },
  })
}
