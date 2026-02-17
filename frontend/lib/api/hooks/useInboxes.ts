"use client"

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import api from "../client"
import type { Inbox, CreateInboxRequest } from "../types"

export const inboxKeys = {
  all: ["inboxes"] as const,
  lists: () => [...inboxKeys.all, "list"] as const,
  list: () => [...inboxKeys.lists()] as const,
  details: () => [...inboxKeys.all, "detail"] as const,
  detail: (id: string) => [...inboxKeys.details(), id] as const,
}

export function useInboxes() {
  return useQuery({
    queryKey: inboxKeys.list(),
    queryFn: async () => {
      const response = await api.get<Inbox[]>("/inboxes")
      return response.data || []
    },
  })
}

export function useInbox(inboxId: string | null) {
  return useQuery({
    queryKey: inboxKeys.detail(inboxId || ""),
    queryFn: async () => {
      if (!inboxId) return null
      const response = await api.get<Inbox>(`/inboxes/${inboxId}`)
      return response.data!
    },
    enabled: !!inboxId,
    refetchInterval: (query) => {
      const data = query.state.data
      if (data && (data.status === "connecting" || data.status === "qr_code")) {
        return 5000
      }
      return false
    },
  })
}

export function useCreateInbox() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (data: CreateInboxRequest) => {
      const response = await api.post<Inbox>("/inboxes", data)
      return response.data!
    },
    onSuccess: (newInbox) => {
      queryClient.setQueryData(inboxKeys.list(), (old: Inbox[] | undefined) => {
        if (!old) return [newInbox]
        return [...old, newInbox]
      })
    },
  })
}

export function useDeleteInbox() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (inboxId: string) => {
      await api.delete(`/inboxes/${inboxId}`)
      return inboxId
    },
    onSuccess: (inboxId) => {
      queryClient.setQueryData(inboxKeys.list(), (old: Inbox[] | undefined) => {
        if (!old) return []
        return old.filter((i) => i.id !== inboxId)
      })
      queryClient.removeQueries({ queryKey: inboxKeys.detail(inboxId) })
    },
  })
}

export function useConnectInbox() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (inboxId: string) => {
      const response = await api.post<{ id: string; status: string; message: string }>(
        `/inboxes/${inboxId}/connect`
      )
      return response.data!
    },
    onSuccess: (data, inboxId) => {
      queryClient.setQueryData(inboxKeys.detail(inboxId), (old: Inbox | undefined) => {
        if (!old) return old
        return { ...old, status: data.status as Inbox["status"] }
      })
      queryClient.invalidateQueries({ queryKey: inboxKeys.list() })
    },
  })
}

export function useDisconnectInbox() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (inboxId: string) => {
      const response = await api.post<{ id: string; status: string }>(
        `/inboxes/${inboxId}/disconnect`
      )
      return response.data!
    },
    onSuccess: (_, inboxId) => {
      queryClient.setQueryData(inboxKeys.detail(inboxId), (old: Inbox | undefined) => {
        if (!old) return old
        return { ...old, status: "disconnected" as Inbox["status"] }
      })
      queryClient.invalidateQueries({ queryKey: inboxKeys.list() })
    },
  })
}

export function useUpdateInboxStatus() {
  const queryClient = useQueryClient()

  return (inboxId: string, status: Inbox["status"], phone?: string) => {
    queryClient.setQueryData(inboxKeys.detail(inboxId), (old: Inbox | undefined) => {
      if (!old) return old
      return { ...old, status, phone: phone || old.phone }
    })
    queryClient.invalidateQueries({ queryKey: inboxKeys.list() })
  }
}

interface QRCodeResponse {
  id: string
  qrcode: string | null
  status: string
}

export function useQRCode(inboxId: string | null, enabled: boolean = true) {
  return useQuery({
    queryKey: ["qrcode", inboxId],
    queryFn: async (): Promise<QRCodeResponse> => {
      if (!inboxId) {
        return { id: "", qrcode: null, status: "waiting" }
      }
      const response = await api.get<QRCodeResponse>(`/inboxes/${inboxId}/qrcode`)
      return response.data ?? { id: inboxId, qrcode: null, status: "waiting" }
    },
    enabled: !!inboxId && enabled,
    refetchInterval: (query) => {
      const data = query.state.data
      if (data && !data.qrcode) {
        return 2000
      }
      if (data && data.qrcode) {
        return 5000
      }
      return false
    },
    staleTime: 0,
  })
}

// Legacy aliases
export const connectionKeys = inboxKeys
export const useConnections = useInboxes
export const useConnection = useInbox
export const useCreateConnection = useCreateInbox
export const useDeleteConnection = useDeleteInbox
export const useConnectWhatsApp = useConnectInbox
export const useDisconnectWhatsApp = useDisconnectInbox
export const useUpdateConnectionStatus = useUpdateInboxStatus
