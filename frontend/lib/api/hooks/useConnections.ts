"use client"

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import api from "../client"
import type { Connection, CreateConnectionRequest } from "../types"

// Query keys
export const connectionKeys = {
  all: ["connections"] as const,
  lists: () => [...connectionKeys.all, "list"] as const,
  list: () => [...connectionKeys.lists()] as const,
  details: () => [...connectionKeys.all, "detail"] as const,
  detail: (id: string) => [...connectionKeys.details(), id] as const,
}

// Get list of connections
export function useConnections() {
  return useQuery({
    queryKey: connectionKeys.list(),
    queryFn: async () => {
      const response = await api.get<Connection[]>("/connections")
      return response.data || []
    },
  })
}

// Get single connection
export function useConnection(connectionId: string | null) {
  return useQuery({
    queryKey: connectionKeys.detail(connectionId || ""),
    queryFn: async () => {
      if (!connectionId) return null
      const response = await api.get<Connection>(`/connections/${connectionId}`)
      return response.data!
    },
    enabled: !!connectionId,
    refetchInterval: (query) => {
      // Refetch every 5 seconds if connecting or qr_code status
      const data = query.state.data
      if (data && (data.status === "connecting" || data.status === "qr_code")) {
        return 5000
      }
      return false
    },
  })
}

// Create connection
export function useCreateConnection() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (data: CreateConnectionRequest) => {
      const response = await api.post<Connection>("/connections", data)
      return response.data!
    },
    onSuccess: (newConnection) => {
      queryClient.setQueryData(connectionKeys.list(), (old: Connection[] | undefined) => {
        if (!old) return [newConnection]
        return [...old, newConnection]
      })
    },
  })
}

// Delete connection
export function useDeleteConnection() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (connectionId: string) => {
      await api.delete(`/connections/${connectionId}`)
      return connectionId
    },
    onSuccess: (connectionId) => {
      queryClient.setQueryData(connectionKeys.list(), (old: Connection[] | undefined) => {
        if (!old) return []
        return old.filter((c) => c.id !== connectionId)
      })
      queryClient.removeQueries({ queryKey: connectionKeys.detail(connectionId) })
    },
  })
}

// Connect WhatsApp (start QR code flow)
export function useConnectWhatsApp() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (connectionId: string) => {
      const response = await api.post<{ id: string; status: string; message: string }>(
        `/connections/${connectionId}/connect`
      )
      return response.data!
    },
    onSuccess: (data, connectionId) => {
      // Update connection status
      queryClient.setQueryData(connectionKeys.detail(connectionId), (old: Connection | undefined) => {
        if (!old) return old
        return { ...old, status: data.status as Connection["status"] }
      })
      queryClient.invalidateQueries({ queryKey: connectionKeys.list() })
    },
  })
}

// Disconnect WhatsApp
export function useDisconnectWhatsApp() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (connectionId: string) => {
      const response = await api.post<{ id: string; status: string }>(
        `/connections/${connectionId}/disconnect`
      )
      return response.data!
    },
    onSuccess: (data, connectionId) => {
      queryClient.setQueryData(connectionKeys.detail(connectionId), (old: Connection | undefined) => {
        if (!old) return old
        return { ...old, status: "disconnected" as Connection["status"] }
      })
      queryClient.invalidateQueries({ queryKey: connectionKeys.list() })
    },
  })
}

// Update connection status (from NATS events)
export function useUpdateConnectionStatus() {
  const queryClient = useQueryClient()

  return (connectionId: string, status: Connection["status"], phone?: string) => {
    queryClient.setQueryData(connectionKeys.detail(connectionId), (old: Connection | undefined) => {
      if (!old) return old
      return { ...old, status, phone: phone || old.phone }
    })
    queryClient.invalidateQueries({ queryKey: connectionKeys.list() })
  }
}

// Get QR code via polling (wuzapi pattern - HTTP polling instead of WebSocket)
interface QRCodeResponse {
  id: string
  qr_code: string | null
  status: "waiting" | "ready"
}

export function useQRCode(connectionId: string | null, enabled: boolean = true) {
  return useQuery({
    queryKey: ["qrcode", connectionId],
    queryFn: async (): Promise<QRCodeResponse> => {
      if (!connectionId) {
        return { id: "", qr_code: null, status: "waiting" }
      }
      const response = await api.get<QRCodeResponse>(`/connections/${connectionId}/qr`)
      return response.data ?? { id: connectionId, qr_code: null, status: "waiting" }
    },
    enabled: !!connectionId && enabled,
    refetchInterval: (query) => {
      // Poll every 2 seconds while waiting for QR code
      const data = query.state.data
      if (data && data.status === "waiting") {
        return 2000
      }
      // Poll every 5 seconds while QR code is ready (might need refresh)
      if (data && data.status === "ready") {
        return 5000
      }
      return false
    },
    staleTime: 0, // Always consider data stale for polling
  })
}
