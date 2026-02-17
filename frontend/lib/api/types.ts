// Inbox types (antigo Connection)
export interface Inbox {
  id: string
  name: string
  channel_type: "whatsapp" | "telegram" | "api"
  greeting_message?: string
  auto_assignment: boolean
  status: "disconnected" | "connecting" | "qr_code" | "connected"
  phone?: string
  created_at: string
  updated_at: string
}

export interface CreateInboxRequest {
  name: string
  channel_type: "whatsapp" | "telegram" | "api"
  greeting_message?: string
  auto_assignment?: boolean
}

// Conversation types (antigo Chat)
export interface Conversation {
  id: string
  inbox_id: string
  contact_id: string
  status: "open" | "pending" | "resolved" | "snoozed"
  priority: "low" | "medium" | "high" | "urgent"
  assignee_id?: string
  is_favorite: boolean
  is_archived: boolean
  unread_count: number
  last_message?: string
  last_message_at?: string
  created_at: string
  updated_at: string
  // Populated fields
  contact?: Contact
  inbox?: Inbox
}

export interface ConversationFilter {
  inbox_id?: string
  status?: Conversation["status"]
  assignee_id?: string
  search?: string
  filter?: "all" | "unread" | "groups" | "favorites" | "archived"
  limit?: number
  offset?: number
}

// Message types
export interface Message {
  id: string
  conversation_id: string
  sender_id?: string
  sender_type: "user" | "contact"
  content_type: "text" | "image" | "video" | "audio" | "document" | "location" | "contact"
  content: string
  private: boolean
  status: "pending" | "sent" | "delivered" | "read" | "failed"
  external_id?: string
  created_at: string
  // Populated
  sender?: User | Contact
  attachments?: Attachment[]
}

export interface Attachment {
  id: string
  message_id: string
  file_type: string
  file_url: string
  file_name?: string
  file_size?: number
}

export interface SendMessageRequest {
  content: string
  content_type?: Message["content_type"]
  private?: boolean
}

// Contact types
export interface Contact {
  id: string
  name?: string
  email?: string
  phone_number?: string
  avatar_url?: string
  custom_attributes?: Record<string, unknown>
  created_at: string
  updated_at: string
  // Contact inboxes
  inboxes?: ContactInbox[]
}

export interface ContactInbox {
  id: string
  contact_id: string
  inbox_id: string
  source_id: string
  created_at: string
}

// Label types
export interface Label {
  id: string
  title: string
  color: string
  description?: string
  created_at: string
}

// User types
export interface User {
  id: string
  name: string
  email: string
  role: string
  avatar_url?: string
  created_at: string
}

export interface TokenPair {
  access_token: string
  refresh_token: string
  expires_at: string
  token_type: string
}

export interface AuthResponse {
  user: User
  tokens: TokenPair
}

export interface LoginRequest {
  email: string
  password: string
}

export interface RegisterRequest {
  name: string
  email: string
  password: string
}

// API Key types
export interface APIKey {
  id: string
  user_id: string
  name: string
  key_prefix: string
  permissions: string[]
  expires_at?: string
  revoked_at?: string
  last_used_at?: string
  created_at: string
}

export interface GeneratedAPIKey extends APIKey {
  key: string
}

// Pagination
export interface PaginationMeta {
  page: number
  per_page: number
  total: number
  total_pages: number
}

// Legacy aliases (para compatibilidade temporária)
export type Connection = Inbox
export type CreateConnectionRequest = CreateInboxRequest
export type Chat = Conversation
export type ChatFilter = ConversationFilter
