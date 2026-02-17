"use client"

import { useEffect, useState, useCallback, useRef } from "react"
import { Subscription } from "nats.ws"
import {
  connectNats,
  isConnected,
  subscribeToMessages,
  subscribeToMessageStatus,
  subscribeToConnectionStatus,
  subscribeToQRCode,
  unsubscribe,
  NatsEvent,
  MessageData,
  MessageStatusData,
  ConnectionStatusData,
  QRCodeData,
} from "./client"

// Hook for NATS connection status
export function useNatsConnection() {
  const [connected, setConnected] = useState(false)
  const [error, setError] = useState<Error | null>(null)
  const [connecting, setConnecting] = useState(false)

  const connect = useCallback(async () => {
    if (isConnected()) {
      setConnected(true)
      return
    }

    setConnecting(true)
    setError(null)

    try {
      await connectNats()
      setConnected(true)
    } catch (err) {
      setError(err as Error)
      setConnected(false)
    } finally {
      setConnecting(false)
    }
  }, [])

  useEffect(() => {
    connect()
  }, [connect])

  return { connected, connecting, error, reconnect: connect }
}

// Hook for subscribing to new messages
export function useNatsMessages(connectionId: string | null) {
  const [messages, setMessages] = useState<MessageData[]>([])
  const [latestMessage, setLatestMessage] = useState<MessageData | null>(null)
  const subscriptionRef = useRef<Subscription | null>(null)

  useEffect(() => {
    if (!connectionId || !isConnected()) return

    const handleMessage = (event: NatsEvent<MessageData>) => {
      setLatestMessage(event.data)
      setMessages((prev) => [...prev, event.data])
    }

    subscriptionRef.current = subscribeToMessages(connectionId, handleMessage)

    return () => {
      unsubscribe(subscriptionRef.current)
      subscriptionRef.current = null
    }
  }, [connectionId])

  const clearMessages = useCallback(() => {
    setMessages([])
    setLatestMessage(null)
  }, [])

  return { messages, latestMessage, clearMessages }
}

// Hook for subscribing to message status updates
export function useNatsMessageStatus(connectionId: string | null) {
  const [statusUpdates, setStatusUpdates] = useState<MessageStatusData[]>([])
  const [latestStatus, setLatestStatus] = useState<MessageStatusData | null>(null)
  const subscriptionRef = useRef<Subscription | null>(null)

  useEffect(() => {
    if (!connectionId || !isConnected()) return

    const handleStatus = (event: NatsEvent<MessageStatusData>) => {
      setLatestStatus(event.data)
      setStatusUpdates((prev) => [...prev, event.data])
    }

    subscriptionRef.current = subscribeToMessageStatus(connectionId, handleStatus)

    return () => {
      unsubscribe(subscriptionRef.current)
      subscriptionRef.current = null
    }
  }, [connectionId])

  return { statusUpdates, latestStatus }
}

// Hook for subscribing to connection status
export function useNatsConnectionStatus(connectionId: string | null) {
  const [status, setStatus] = useState<ConnectionStatusData | null>(null)
  const subscriptionRef = useRef<Subscription | null>(null)

  useEffect(() => {
    if (!connectionId || !isConnected()) return

    const handleStatus = (event: NatsEvent<ConnectionStatusData>) => {
      setStatus(event.data)
    }

    subscriptionRef.current = subscribeToConnectionStatus(connectionId, handleStatus)

    return () => {
      unsubscribe(subscriptionRef.current)
      subscriptionRef.current = null
    }
  }, [connectionId])

  return { status }
}

// Hook for subscribing to QR codes
export function useNatsQRCode(connectionId: string | null) {
  const [qrCode, setQRCode] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)
  const subscriptionRef = useRef<Subscription | null>(null)

  useEffect(() => {
    if (!connectionId || !isConnected()) {
      setQRCode(null)
      return
    }

    setLoading(true)

    const handleQR = (event: NatsEvent<QRCodeData>) => {
      setQRCode(event.data.code)
      setLoading(false)
    }

    subscriptionRef.current = subscribeToQRCode(connectionId, handleQR)

    return () => {
      unsubscribe(subscriptionRef.current)
      subscriptionRef.current = null
      setQRCode(null)
    }
  }, [connectionId])

  const clearQRCode = useCallback(() => {
    setQRCode(null)
  }, [])

  return { qrCode, loading, clearQRCode }
}

// Combined hook for all real-time updates for a connection
export function useNatsRealtime(connectionId: string | null) {
  const { connected, connecting, error, reconnect } = useNatsConnection()
  const { messages, latestMessage, clearMessages } = useNatsMessages(connected ? connectionId : null)
  const { statusUpdates, latestStatus } = useNatsMessageStatus(connected ? connectionId : null)
  const { status: connectionStatus } = useNatsConnectionStatus(connected ? connectionId : null)
  const { qrCode, clearQRCode } = useNatsQRCode(connected ? connectionId : null)

  return {
    // Connection state
    connected,
    connecting,
    error,
    reconnect,
    
    // Messages
    messages,
    latestMessage,
    clearMessages,
    
    // Message status
    statusUpdates,
    latestStatus,
    
    // Connection status
    connectionStatus,
    
    // QR Code
    qrCode,
    clearQRCode,
  }
}
