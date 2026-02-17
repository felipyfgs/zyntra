"use client"

import { createContext, useContext, useEffect, useState, ReactNode, useCallback } from "react"
import { NatsConnection } from "nats.ws"
import { connectNats, disconnectNats, isConnected, getNatsConnection } from "@/lib/nats/client"

interface NatsContextValue {
  connection: NatsConnection | null
  connected: boolean
  connecting: boolean
  error: Error | null
  connect: () => Promise<void>
  disconnect: () => Promise<void>
}

const NatsContext = createContext<NatsContextValue | null>(null)

interface NatsProviderProps {
  children: ReactNode
  autoConnect?: boolean
}

export function NatsProvider({ children, autoConnect = true }: NatsProviderProps) {
  const [connection, setConnection] = useState<NatsConnection | null>(null)
  const [connected, setConnected] = useState(false)
  const [connecting, setConnecting] = useState(false)
  const [error, setError] = useState<Error | null>(null)

  const connect = useCallback(async () => {
    if (isConnected()) {
      setConnection(getNatsConnection())
      setConnected(true)
      return
    }

    setConnecting(true)
    setError(null)

    try {
      const nc = await connectNats()
      setConnection(nc)
      setConnected(true)
    } catch (err) {
      console.error("[NatsProvider] Connection failed:", err)
      setError(err as Error)
      setConnected(false)
    } finally {
      setConnecting(false)
    }
  }, [])

  const disconnect = useCallback(async () => {
    await disconnectNats()
    setConnection(null)
    setConnected(false)
  }, [])

  useEffect(() => {
    if (autoConnect) {
      connect()
    }

    return () => {
      // Don't disconnect on unmount to preserve connection
      // disconnect()
    }
  }, [autoConnect, connect])

  return (
    <NatsContext.Provider
      value={{
        connection,
        connected,
        connecting,
        error,
        connect,
        disconnect,
      }}
    >
      {children}
    </NatsContext.Provider>
  )
}

export function useNats() {
  const context = useContext(NatsContext)
  if (!context) {
    throw new Error("useNats must be used within a NatsProvider")
  }
  return context
}
