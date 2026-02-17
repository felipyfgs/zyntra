"use client"

import { useState, useEffect, useMemo } from "react"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { Button } from "@/components/ui/button"
import {
  Conversation,
  ConversationContent,
  ConversationScrollButton,
} from "@/components/chat/conversation"
import { ChatBubble, type MessageStatus } from "@/components/chat/chat-bubble"
import {
  PromptInput,
  PromptInputTextarea,
  PromptInputFooter,
  PromptInputSubmit,
  PromptInputBody,
} from "@/components/chat/prompt-input"
import { type Chat } from "@/components/chat/chat-list"
import { ChevronLeft, MoreVertical, Phone, Video, Loader2 } from "lucide-react"
import { useMessages, useSendMessage, useAddMessageToCache } from "@/lib/api/hooks"
import { useNatsMessages, useNatsMessageStatus } from "@/lib/nats/hooks"
import type { Message as APIMessage } from "@/lib/api/types"

type ChatMessage = {
  id: string
  content: string
  sender: "user" | "other"
  timestamp: Date
  status?: MessageStatus
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
    sender: apiMsg.sender_type === "user" ? "user" : "other",
    timestamp: new Date(apiMsg.created_at),
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
        conversationId: chat.id,
        message: { content: message.text },
      })
    } catch (error) {
      console.error("Failed to send message:", error)
      // Keep the local message on error for retry
    }
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

      {/* Messages - distinct background from chat list with doodle pattern */}
      <Conversation className="min-h-0 flex-1 overflow-hidden bg-stone-100 dark:bg-stone-900/50">
        {/* WhatsApp-style doodle pattern overlay */}
        <div 
          className="pointer-events-none absolute inset-0 opacity-[0.06] dark:opacity-[0.08]"
          style={{
            backgroundImage: `url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='120' height='120' viewBox='0 0 120 120'%3E%3Cg fill='%236b7280' fill-opacity='1'%3E%3C!-- Phone --%3E%3Cpath d='M8 4a2 2 0 0 0-2 2v8a2 2 0 0 0 2 2h4a2 2 0 0 0 2-2V6a2 2 0 0 0-2-2H8zm2 11a1 1 0 1 1 0-2 1 1 0 0 1 0 2z' transform='translate(5,5) rotate(-15,10,10)'/%3E%3C!-- Message bubble --%3E%3Cpath d='M2 2h12a2 2 0 0 1 2 2v8a2 2 0 0 1-2 2H6l-4 3v-3a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2z' transform='translate(70,8) rotate(5,8,8)'/%3E%3C!-- Heart --%3E%3Cpath d='M12 21.35l-1.45-1.32C5.4 15.36 2 12.28 2 8.5 2 5.42 4.42 3 7.5 3c1.74 0 3.41.81 4.5 2.09C13.09 3.81 14.76 3 16.5 3 19.58 3 22 5.42 22 8.5c0 3.78-3.4 6.86-8.55 11.54L12 21.35z' transform='translate(35,2) scale(0.6) rotate(10,12,12)'/%3E%3C!-- Star --%3E%3Cpath d='M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z' transform='translate(90,55) scale(0.55) rotate(-10,12,12)'/%3E%3C!-- Camera --%3E%3Cpath d='M12 10.5a3.5 3.5 0 1 0 0 7 3.5 3.5 0 0 0 0-7zM20 4h-3.17L15 2H9L7.17 4H4a2 2 0 0 0-2 2v12a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2V6a2 2 0 0 0-2-2z' transform='translate(2,60) scale(0.55)'/%3E%3C!-- Clock --%3E%3Ccircle cx='12' cy='12' r='9' fill='none' stroke='%236b7280' stroke-width='2' transform='translate(55,35) scale(0.6)'/%3E%3Cpath d='M12 7v5l3 3' fill='none' stroke='%236b7280' stroke-width='2' stroke-linecap='round' transform='translate(55,35) scale(0.6)'/%3E%3C!-- Music note --%3E%3Cpath d='M12 3v10.55A4 4 0 1 0 14 17V7h4V3h-6z' transform='translate(30,55) scale(0.6) rotate(5,12,12)'/%3E%3C!-- Smile emoji --%3E%3Ccircle cx='12' cy='12' r='10' fill='none' stroke='%236b7280' stroke-width='1.5' transform='translate(60,80) scale(0.6)'/%3E%3Ccircle cx='8' cy='10' r='1.5' transform='translate(60,80) scale(0.6)'/%3E%3Ccircle cx='16' cy='10' r='1.5' transform='translate(60,80) scale(0.6)'/%3E%3Cpath d='M8 14s1.5 2 4 2 4-2 4-2' fill='none' stroke='%236b7280' stroke-width='1.5' stroke-linecap='round' transform='translate(60,80) scale(0.6)'/%3E%3C!-- Thumbs up --%3E%3Cpath d='M2 20h2V9H2v11zm20-6a2 2 0 0 0-2-2h-5.5l.69-3.32.02-.16a1.5 1.5 0 0 0-.44-1.06l-1.06-1.06L7 13.17V20h9.5a2 2 0 0 0 1.84-1.21l2.4-5.6c.07-.17.11-.36.11-.55V14z' transform='translate(85,85) scale(0.5) rotate(-5,12,12)'/%3E%3C!-- Location pin --%3E%3Cpath d='M12 2C8.13 2 5 5.13 5 9c0 5.25 7 13 7 13s7-7.75 7-13c0-3.87-3.13-7-7-7zm0 9.5a2.5 2.5 0 1 1 0-5 2.5 2.5 0 0 1 0 5z' transform='translate(5,85) scale(0.55)'/%3E%3C/g%3E%3C/svg%3E")`,
            backgroundSize: '120px 120px',
          }}
        />
        <ConversationContent className="relative z-10 gap-2 p-4">
          {isLoading ? (
            <div className="flex h-full items-center justify-center">
              <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
            </div>
          ) : messages.length === 0 ? (
            <div className="flex h-full flex-col items-center justify-center gap-2 text-muted-foreground">
              <div className="rounded-lg bg-card/80 px-4 py-2 text-center shadow-sm">
                <p className="text-sm">Nenhuma mensagem ainda</p>
                <p className="text-xs opacity-70">Envie uma mensagem para iniciar a conversa</p>
              </div>
            </div>
          ) : (
            messages.map((message, index) => (
              <ChatBubble
                key={message.id || `msg-${index}`}
                content={message.content}
                timestamp={message.timestamp}
                isOutgoing={message.sender === "user"}
                status={message.status}
              />
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
