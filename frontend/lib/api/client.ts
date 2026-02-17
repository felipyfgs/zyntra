const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api/v1"

// Token storage
let accessToken: string | null = null
let refreshToken: string | null = null

export function setTokens(access: string, refresh: string) {
  accessToken = access
  refreshToken = refresh
  if (typeof window !== "undefined") {
    localStorage.setItem("access_token", access)
    localStorage.setItem("refresh_token", refresh)
  }
}

export function getAccessToken() {
  if (!accessToken && typeof window !== "undefined") {
    accessToken = localStorage.getItem("access_token")
  }
  return accessToken
}

export function getRefreshToken() {
  if (!refreshToken && typeof window !== "undefined") {
    refreshToken = localStorage.getItem("refresh_token")
  }
  return refreshToken
}

export function clearTokens() {
  accessToken = null
  refreshToken = null
  if (typeof window !== "undefined") {
    localStorage.removeItem("access_token")
    localStorage.removeItem("refresh_token")
  }
}

// API Response types
export interface APIResponse<T> {
  success: boolean
  data?: T
  error?: {
    code: string
    message: string
    details?: string
  }
  meta?: {
    page: number
    per_page: number
    total: number
    total_pages: number
  }
}

// API Error class
export class APIError extends Error {
  code: string
  status: number

  constructor(code: string, message: string, status: number) {
    super(message)
    this.code = code
    this.status = status
    this.name = "APIError"
  }
}

// Fetch wrapper with auth
async function fetchWithAuth<T>(
  endpoint: string,
  options: RequestInit = {}
): Promise<APIResponse<T>> {
  const token = getAccessToken()

  const headers: HeadersInit = {
    "Content-Type": "application/json",
    ...options.headers,
  }

  if (token) {
    ;(headers as Record<string, string>)["Authorization"] = `Bearer ${token}`
  }

  const response = await fetch(`${API_BASE_URL}${endpoint}`, {
    ...options,
    headers,
  })

  // Handle 204 No Content
  if (response.status === 204) {
    return { success: true }
  }

  const data: APIResponse<T> = await response.json()

  // Handle token refresh
  if (response.status === 401 && data.error?.code === "EXPIRED_TOKEN") {
    const refreshed = await refreshAccessToken()
    if (refreshed) {
      // Retry the request
      return fetchWithAuth<T>(endpoint, options)
    }
    clearTokens()
    throw new APIError("UNAUTHORIZED", "Session expired", 401)
  }

  if (!data.success && data.error) {
    throw new APIError(data.error.code, data.error.message, response.status)
  }

  return data
}

// Refresh token
async function refreshAccessToken(): Promise<boolean> {
  const refresh = getRefreshToken()
  if (!refresh) return false

  try {
    const response = await fetch(`${API_BASE_URL}/auth/refresh`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ refresh_token: refresh }),
    })

    if (!response.ok) return false

    const data = await response.json()
    if (data.success && data.data?.tokens) {
      setTokens(data.data.tokens.access_token, data.data.tokens.refresh_token)
      return true
    }
    return false
  } catch {
    return false
  }
}

// API methods
export const api = {
  get: <T>(endpoint: string) => fetchWithAuth<T>(endpoint, { method: "GET" }),

  post: <T>(endpoint: string, body?: unknown) =>
    fetchWithAuth<T>(endpoint, {
      method: "POST",
      body: body ? JSON.stringify(body) : undefined,
    }),

  put: <T>(endpoint: string, body?: unknown) =>
    fetchWithAuth<T>(endpoint, {
      method: "PUT",
      body: body ? JSON.stringify(body) : undefined,
    }),

  patch: <T>(endpoint: string, body?: unknown) =>
    fetchWithAuth<T>(endpoint, {
      method: "PATCH",
      body: body ? JSON.stringify(body) : undefined,
    }),

  delete: <T>(endpoint: string) => fetchWithAuth<T>(endpoint, { method: "DELETE" }),
}

export default api
