"use client"

import { useState, useEffect, useMemo } from "react"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { Button } from "@/components/ui/button"
import {
  Conversation,
  ConversationContent,
  ConversationScrollButton,
} from "@/components/chat/conversation"
import {
  Message,
  MessageContent,
  MessageResponse,
  MessageActions,
  MessageAction,
} from "@/components/chat/message"
import {
  PromptInput,
  PromptInputTextarea,
  PromptInputFooter,
  PromptInputSubmit,
  PromptInputBody,
} from "@/components/chat/prompt-input"
import { type Chat } from "@/components/chat/chat-list"
import { ChevronLeft, Copy, MoreVertical, Phone, Video, Check, Loader2 } from "lucide-react"
import { useMessages, useSendMessage, useAddMessageToCache } from "@/lib/api/hooks"
import { useNatsMessages, useNatsMessageStatus } from "@/lib/nats/hooks"
import type { Message as APIMessage } from "@/lib/api/types"

type ChatMessage = {
  id: string
  content: string
  sender: "user" | "other"
  timestamp: Date
  status?: "pending" | "sent" | "delivered" | "read"
}

type ChatAreaProps = {
  chat: Chat
  connectionId?: string
  onBack?: () => void
}

// Convert API Message to local ChatMessage type
function mapAPIMessage(apiMsg: APIMessage): ChatMessage {
  return {
    id: apiMsg.id,
    content: apiMsg.content,
    sender: apiMsg.direction === "outbound" ? "user" : "other",
    timestamp: new Date(apiMsg.timestamp),
    status: apiMsg.status,
  }
}

function getInitials(name: string): string {
  return name
    .split(" ")
    .map((n) => n[0])
    .join("")
    .toUpperCase()
    .slice(0, 2)
}



export function ChatArea({ chat, connectionId, onBack }: ChatAreaProps) {
  const [pendingMessages, setPendingMessages] = useState<ChatMessage[]>([])
  const [copiedId, setCopiedId] = useState<string | null>(null)

  // Fetch messages from API
  const { data: apiData, isLoading } = useMessages(chat.id)
  const sendMessageMutation = useSendMessage()
  const addMessageToCache = useAddMessageToCache()

  // Real-time updates via NATS
  const newNatsMessage = useNatsMessages(connectionId || "")
  const messageStatus = useNatsMessageStatus(connectionId || "")

  // Handle new messages from NATS
  useEffect(() => {
    if (newNatsMessage) {
      addMessageToCache(chat.id, newNatsMessage as unknown as APIMessage)
    }
  }, [newNatsMessage, chat.id, addMessageToCache])

  // Combine API messages with pending (optimistic) messages
  const messages = useMemo(() => {
    const apiMessages = apiData?.messages?.map(mapAPIMessage) || []
    // Filter out pending messages that now exist in API data
    const stillPending = pendingMessages.filter(
      (pm) => !apiMessages.some((am) => am.id === pm.id || am.content === pm.content)
    )
    return [...apiMessages, ...stillPending]
  }, [apiData, pendingMessages])

  const handleSend = async (message: { text: string }) => {
    if (!message.text.trim()) return

    // Optimistic update with pending message
    const tempId = `temp-${Date.now()}`
    const newMessage: ChatMessage = {
      id: tempId,
      content: message.text,
      sender: "user",
      timestamp: new Date(),
      status: "pending",
    }
    setPendingMessages((prev) => [...prev, newMessage])

    try {
      await sendMessageMutation.mutateAsync({
        chatId: chat.id,
        message: { content: message.text },
      })
    } catch (error) {
      console.error("Failed to send message:", error)
      // Keep the local message on error for retry
    }
  }

  const handleCopy = async (content: string, id: string) => {
    await navigator.clipboard.writeText(content)
    setCopiedId(id)
    setTimeout(() => setCopiedId(null), 2000)
  }

  return (
    <div className="flex h-full flex-col">
      {/* Header */}
      <div className="flex h-[65px] shrink-0 items-center justify-between border-b px-2 md:px-4">
        <div className="flex items-center gap-2 md:gap-3">
          {onBack && (
            <Button
              variant="ghost"
              size="icon"
              onClick={onBack}
              className="md:hidden shrink-0"
            >
              <ChevronLeft className="h-5 w-5" />
            </Button>
          )}
          <div className="relative">
            <Avatar className="h-10 w-10">
              <AvatarImage src={chat.avatar} alt={chat.name} />
              <AvatarFallback>{getInitials(chat.name)}</AvatarFallback>
            </Avatar>
            {chat.isOnline && (
              <span className="absolute bottom-0 right-0 h-3 w-3 rounded-full border-2 border-background bg-green-500" />
            )}
          </div>
          <div>
            <h2 className="font-semibold">{chat.name}</h2>
            <p className="text-sm text-muted-foreground">
              {chat.isOnline ? "Online" : "Offline"}
            </p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Button variant="ghost" size="icon">
            <Phone className="h-5 w-5" />
          </Button>
          <Button variant="ghost" size="icon">
            <Video className="h-5 w-5" />
          </Button>
          <Button variant="ghost" size="icon">
            <MoreVertical className="h-5 w-5" />
          </Button>
        </div>
      </div>

      {/* Messages */}
      <Conversation className="min-h-0 flex-1 overflow-hidden">
        <ConversationContent className="gap-4 p-4">
          {isLoading ? (
            <div className="flex h-full items-center justify-center">
              <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
            </div>
          ) : messages.length === 0 ? (
            <div className="flex h-full flex-col items-center justify-center gap-2 text-muted-foreground">
              <p className="text-sm">Nenhuma mensagem ainda</p>
              <p className="text-xs">Envie uma mensagem para iniciar a conversa</p>
            </div>
          ) : (
            messages.map((message, index) => (
              <Message
                key={message.id || `msg-${index}`}
                from={message.sender === "user" ? "user" : "assistant"}
              >
                <MessageContent>
                  {message.sender === "other" ? (
                    <MessageResponse>{message.content}</MessageResponse>
                  ) : (
                    <p>{message.content}</p>
                  )}
                </MessageContent>
                <MessageActions className="opacity-0 transition-opacity group-hover:opacity-100">
                  <MessageAction
                    tooltip={copiedId === message.id ? "Copiado!" : "Copiar"}
                    onClick={() => handleCopy(message.content, message.id)}
                  >
                    {copiedId === message.id ? (
                      <Check className="h-4 w-4" />
                    ) : (
                      <Copy className="h-4 w-4" />
                    )}
                  </MessageAction>
                </MessageActions>
              </Message>
            ))
          )}
        </ConversationContent>
        <ConversationScrollButton />
      </Conversation>

      {/* Input */}
      <div className="shrink-0 border-t p-3">
        <PromptInput onSubmit={handleSend} className="mx-auto max-w-3xl">
          <PromptInputBody>
            <PromptInputTextarea placeholder="Digite sua mensagem..." />
          </PromptInputBody>
          <PromptInputFooter>
            <div />
            <PromptInputSubmit />
          </PromptInputFooter>
        </PromptInput>
      </div>
    </div>
  )
}
