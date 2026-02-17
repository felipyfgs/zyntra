"use client"

import { useQuery, useMutation, useQueryClient, useInfiniteQuery } from "@tanstack/react-query"
import api from "../client"
import type { Conversation, ConversationFilter, PaginationMeta } from "../types"

export const conversationKeys = {
  all: ["conversations"] as const,
  lists: () => [...conversationKeys.all, "list"] as const,
  list: (filter?: ConversationFilter) => [...conversationKeys.lists(), filter] as const,
  details: () => [...conversationKeys.all, "detail"] as const,
  detail: (id: string) => [...conversationKeys.details(), id] as const,
}

export function useConversations(filter?: ConversationFilter) {
  return useQuery({
    queryKey: conversationKeys.list(filter),
    queryFn: async () => {
      const params = new URLSearchParams()
      if (filter?.inbox_id) params.set("inbox_id", filter.inbox_id)
      if (filter?.status) params.set("status", filter.status)
      if (filter?.assignee_id) params.set("assignee_id", filter.assignee_id)
      if (filter?.search) params.set("search", filter.search)
      if (filter?.filter === "favorites") params.set("favorite", "true")
      if (filter?.filter === "archived") params.set("archived", "true")
      if (filter?.limit) params.set("limit", filter.limit.toString())
      if (filter?.offset) params.set("offset", filter.offset.toString())

      const query = params.toString()
      const endpoint = `/conversations${query ? `?${query}` : ""}`
      const response = await api.get<Conversation[]>(endpoint)

      return {
        conversations: response.data || [],
        meta: response.meta as PaginationMeta | undefined,
      }
    },
  })
}

export function useConversationsInfinite(filter?: Omit<ConversationFilter, "offset">) {
  return useInfiniteQuery({
    queryKey: conversationKeys.list(filter),
    queryFn: async ({ pageParam = 0 }) => {
      const params = new URLSearchParams()
      if (filter?.inbox_id) params.set("inbox_id", filter.inbox_id)
      if (filter?.status) params.set("status", filter.status)
      if (filter?.search) params.set("search", filter.search)
      if (filter?.filter === "favorites") params.set("favorite", "true")
      if (filter?.filter === "archived") params.set("archived", "true")
      params.set("offset", pageParam.toString())
      params.set("limit", "20")

      const query = params.toString()
      const response = await api.get<Conversation[]>(`/conversations?${query}`)

      return {
        conversations: response.data || [],
        meta: response.meta as PaginationMeta | undefined,
        nextOffset: (response.data?.length || 0) === 20 ? pageParam + 20 : undefined,
      }
    },
    initialPageParam: 0,
    getNextPageParam: (lastPage) => lastPage.nextOffset,
  })
}

export function useConversation(conversationId: string | null) {
  return useQuery({
    queryKey: conversationKeys.detail(conversationId || ""),
    queryFn: async () => {
      if (!conversationId) return null
      const response = await api.get<Conversation>(`/conversations/${conversationId}`)
      return response.data!
    },
    enabled: !!conversationId,
  })
}

export function useUpdateConversation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({
      conversationId,
      updates,
    }: {
      conversationId: string
      updates: Partial<Pick<Conversation, "status" | "priority" | "assignee_id" | "is_favorite" | "is_archived">>
    }) => {
      const response = await api.put<Conversation>(`/conversations/${conversationId}`, updates)
      return response.data!
    },
    onSuccess: (data) => {
      queryClient.setQueryData(conversationKeys.detail(data.id), data)
      queryClient.invalidateQueries({ queryKey: conversationKeys.lists() })
    },
  })
}

export function useMarkConversationAsRead() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (conversationId: string) => {
      const response = await api.post<{ message: string }>(`/conversations/${conversationId}/read`)
      return response.data!
    },
    onSuccess: (_, conversationId) => {
      queryClient.invalidateQueries({ queryKey: conversationKeys.detail(conversationId) })
      queryClient.invalidateQueries({ queryKey: conversationKeys.lists() })
    },
  })
}

export function useToggleFavorite() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (conversationId: string) => {
      const response = await api.post<{ is_favorite: boolean }>(`/conversations/${conversationId}/favorite`)
      return response.data!
    },
    onSuccess: (data, conversationId) => {
      queryClient.setQueryData(conversationKeys.detail(conversationId), (old: Conversation | undefined) => {
        if (!old) return old
        return { ...old, is_favorite: data.is_favorite }
      })
      queryClient.invalidateQueries({ queryKey: conversationKeys.lists() })
    },
  })
}

export function useToggleArchive() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (conversationId: string) => {
      const response = await api.post<{ is_archived: boolean }>(`/conversations/${conversationId}/archive`)
      return response.data!
    },
    onSuccess: (data, conversationId) => {
      queryClient.setQueryData(conversationKeys.detail(conversationId), (old: Conversation | undefined) => {
        if (!old) return old
        return { ...old, is_archived: data.is_archived }
      })
      queryClient.invalidateQueries({ queryKey: conversationKeys.lists() })
    },
  })
}

export function useAssignConversation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({ conversationId, assigneeId }: { conversationId: string; assigneeId: string }) => {
      const response = await api.post<{ message: string }>(`/conversations/${conversationId}/assign`, {
        assignee_id: assigneeId,
      })
      return response.data!
    },
    onSuccess: (_, { conversationId }) => {
      queryClient.invalidateQueries({ queryKey: conversationKeys.detail(conversationId) })
      queryClient.invalidateQueries({ queryKey: conversationKeys.lists() })
    },
  })
}

// Legacy aliases
export const chatKeys = conversationKeys
export const useChats = useConversations
export const useChatsInfinite = useConversationsInfinite
export const useChat = useConversation
export const useUpdateChat = useUpdateConversation
export const useMarkChatAsRead = useMarkConversationAsRead
