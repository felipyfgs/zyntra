import { connect, NatsConnection, StringCodec, Subscription, JetStreamClient, JetStreamManager } from "nats.ws"

const NATS_WS_URL = process.env.NEXT_PUBLIC_NATS_WS_URL || "ws://localhost:9222"

let natsConnection: NatsConnection | null = null
let jetStream: JetStreamClient | null = null

const sc = StringCodec()

export interface NatsConfig {
  url?: string
  reconnect?: boolean
  maxReconnectAttempts?: number
}

// Event types
export type EventType = "message" | "message_status" | "connection_status" | "qr_code"

export interface NatsEvent<T = unknown> {
  type: EventType
  connection_id: string
  timestamp: string
  data: T
}

export interface MessageData {
  id: string
  chat_jid: string
  sender_jid: string
  content: string
  media_type?: string
  direction: "inbound" | "outbound"
  timestamp: string
}

export interface MessageStatusData {
  message_id: string
  chat_jid: string
  status: "sent" | "delivered" | "read"
}

export interface ConnectionStatusData {
  status: "disconnected" | "connecting" | "qr_code" | "connected"
  phone?: string
  jid?: string
}

export interface QRCodeData {
  code: string
}

// Connect to NATS
export async function connectNats(config?: NatsConfig): Promise<NatsConnection> {
  if (natsConnection && !natsConnection.isClosed()) {
    return natsConnection
  }

  const url = config?.url || NATS_WS_URL

  console.log(`[NATS] Connecting to ${url}...`)

  try {
    natsConnection = await connect({
      servers: url,
      reconnect: config?.reconnect ?? true,
      maxReconnectAttempts: config?.maxReconnectAttempts ?? 10,
    })

    console.log("[NATS] Connected successfully")

    // Setup closed handler
    natsConnection.closed().then((err) => {
      if (err) {
        console.error("[NATS] Connection closed with error:", err)
      } else {
        console.log("[NATS] Connection closed")
      }
      natsConnection = null
      jetStream = null
    })

    // Get JetStream client
    jetStream = natsConnection.jetstream()

    return natsConnection
  } catch (error) {
    console.error("[NATS] Failed to connect:", error)
    throw error
  }
}

// Get NATS connection
export function getNatsConnection(): NatsConnection | null {
  return natsConnection
}

// Get JetStream client
export function getJetStream(): JetStreamClient | null {
  return jetStream
}

// Check if connected
export function isConnected(): boolean {
  return natsConnection !== null && !natsConnection.isClosed()
}

// Disconnect from NATS
export async function disconnectNats(): Promise<void> {
  if (natsConnection) {
    await natsConnection.drain()
    natsConnection = null
    jetStream = null
    console.log("[NATS] Disconnected")
  }
}

// Subscribe to a subject
export function subscribe(subject: string, callback: (data: unknown, subject: string) => void): Subscription | null {
  if (!natsConnection) {
    console.error("[NATS] Not connected")
    return null
  }

  const sub = natsConnection.subscribe(subject)
  
  ;(async () => {
    for await (const msg of sub) {
      try {
        const data = JSON.parse(sc.decode(msg.data))
        callback(data, msg.subject)
      } catch (error) {
        console.error("[NATS] Failed to parse message:", error)
      }
    }
  })()

  return sub
}

// Subscribe to messages for a connection
export function subscribeToMessages(
  connectionId: string, 
  callback: (event: NatsEvent<MessageData>) => void
): Subscription | null {
  const subject = `zyntra.messages.${connectionId}`
  return subscribe(subject, (data) => callback(data as NatsEvent<MessageData>))
}

// Subscribe to message status updates
export function subscribeToMessageStatus(
  connectionId: string,
  callback: (event: NatsEvent<MessageStatusData>) => void
): Subscription | null {
  const subject = `zyntra.messages.${connectionId}.status`
  return subscribe(subject, (data) => callback(data as NatsEvent<MessageStatusData>))
}

// Subscribe to connection status
export function subscribeToConnectionStatus(
  connectionId: string,
  callback: (event: NatsEvent<ConnectionStatusData>) => void
): Subscription | null {
  const subject = `zyntra.connections.${connectionId}`
  return subscribe(subject, (data) => callback(data as NatsEvent<ConnectionStatusData>))
}

// Subscribe to QR codes
export function subscribeToQRCode(
  connectionId: string,
  callback: (event: NatsEvent<QRCodeData>) => void
): Subscription | null {
  const subject = `zyntra.qr.${connectionId}`
  return subscribe(subject, (data) => callback(data as NatsEvent<QRCodeData>))
}

// Unsubscribe
export function unsubscribe(subscription: Subscription | null): void {
  if (subscription) {
    subscription.unsubscribe()
  }
}
