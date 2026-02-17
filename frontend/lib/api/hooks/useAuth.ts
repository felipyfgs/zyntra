"use client"

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { useRouter } from "next/navigation"
import api, { setTokens, clearTokens, getAccessToken } from "../client"
import type { User, AuthResponse, LoginRequest, RegisterRequest } from "../types"

// Query keys
export const authKeys = {
  all: ["auth"] as const,
  profile: () => [...authKeys.all, "profile"] as const,
}

// Get current user profile
export function useProfile() {
  return useQuery({
    queryKey: authKeys.profile(),
    queryFn: async () => {
      const response = await api.get<User>("/auth/profile")
      return response.data!
    },
    enabled: !!getAccessToken(),
    retry: false,
  })
}

// Login mutation
export function useLogin() {
  const queryClient = useQueryClient()
  const router = useRouter()

  return useMutation({
    mutationFn: async (credentials: LoginRequest) => {
      const response = await api.post<AuthResponse>("/auth/login", credentials)
      return response.data!
    },
    onSuccess: (data) => {
      setTokens(data.tokens.access_token, data.tokens.refresh_token)
      queryClient.setQueryData(authKeys.profile(), data.user)
      router.push("/conversations")
    },
  })
}

// Register mutation
export function useRegister() {
  const queryClient = useQueryClient()
  const router = useRouter()

  return useMutation({
    mutationFn: async (data: RegisterRequest) => {
      const response = await api.post<AuthResponse>("/auth/register", data)
      return response.data!
    },
    onSuccess: (data) => {
      setTokens(data.tokens.access_token, data.tokens.refresh_token)
      queryClient.setQueryData(authKeys.profile(), data.user)
      router.push("/conversations")
    },
  })
}

// Logout
export function useLogout() {
  const queryClient = useQueryClient()
  const router = useRouter()

  return useMutation({
    mutationFn: async () => {
      clearTokens()
      return true
    },
    onSuccess: () => {
      queryClient.clear()
      router.push("/login")
    },
  })
}

// Change password mutation
export function useChangePassword() {
  return useMutation({
    mutationFn: async (data: { current_password: string; new_password: string }) => {
      const response = await api.put<{ message: string }>("/auth/password", data)
      return response.data!
    },
  })
}

// Check if user is authenticated
export function useIsAuthenticated() {
  const token = getAccessToken()
  return !!token
}
