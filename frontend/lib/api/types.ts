// Chat types
export interface Chat {
  id: string
  connection_id: string
  jid: string
  name: string
  phone?: string
  avatar_url?: string
  is_group: boolean
  is_favorite: boolean
  is_archived: boolean
  unread_count: number
  last_message?: string
  last_message_at?: string
  created_at: string
  updated_at: string
}

export interface ChatFilter {
  connection_id?: string
  search?: string
  filter?: "all" | "unread" | "groups" | "favorites" | "archived"
  page?: number
  per_page?: number
}

// Message types
export interface Message {
  id: string
  connection_id: string
  chat_jid: string
  sender_jid: string
  sender_name?: string
  direction: "inbound" | "outbound"
  content: string
  media_type?: string
  media_url?: string
  status: "pending" | "sent" | "delivered" | "read"
  timestamp: string
  created_at: string
}

export interface MessageFilter {
  chat_id: string
  page?: number
  per_page?: number
  before?: string
  after?: string
}

export interface SendMessageRequest {
  content: string
  media_url?: string
  media_type?: string
}

// Connection types
export interface Connection {
  id: string
  user_id: string
  name: string
  phone?: string
  jid?: string
  status: "disconnected" | "connecting" | "qr_code" | "connected"
  created_at: string
  updated_at: string
}

export interface CreateConnectionRequest {
  name: string
}

// Contact types
export interface Contact {
  id: string
  connection_id: string
  jid: string
  phone: string
  name?: string
  push_name?: string
  avatar_url?: string
  created_at: string
  updated_at: string
}

// Filter types
export interface FilterRule {
  id: string
  field: string
  operator: string
  value?: string
}

export interface SavedFilter {
  id: string
  user_id: string
  name: string
  rules: FilterRule[]
  position: number
  created_at: string
  updated_at: string
}

// Auth types
export interface User {
  id: string
  name: string
  email: string
  role: string
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
