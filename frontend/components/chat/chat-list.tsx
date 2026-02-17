"use client"

import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { cn } from "@/lib/utils"
import { ScrollArea } from "@/components/ui/scroll-area"

export type Chat = {
  id: string
  name: string
  avatar?: string
  lastMessage: string
  timestamp: Date
  unreadCount?: number
  isOnline?: boolean
  isGroup?: boolean
  isWaiting?: boolean
  hasMedia?: boolean
  isFavorite?: boolean
}

type ChatListProps = {
  chats: Chat[]
  selectedId?: string
  onSelect: (chat: Chat) => void
}

function formatTimestamp(date: Date): string {
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  const minutes = Math.floor(diff / 60000)
  const hours = Math.floor(diff / 3600000)
  const days = Math.floor(diff / 86400000)

  if (minutes < 1) return "now"
  if (minutes < 60) return `${minutes}m`
  if (hours < 24) return `${hours}h`
  if (days < 7) return `${days}d`
  return date.toLocaleDateString()
}

function getInitials(name: string): string {
  return name
    .split(" ")
    .map((n) => n[0])
    .join("")
    .toUpperCase()
    .slice(0, 2)
}

export function ChatList({ chats, selectedId, onSelect }: ChatListProps) {
  return (
    <ScrollArea className="flex-1">
      <div className="flex flex-col">
        {chats.map((chat) => (
          <button
            key={chat.id}
            onClick={() => onSelect(chat)}
            className={cn(
              "flex items-center gap-3 p-3 text-left transition-colors hover:bg-muted/50",
              selectedId === chat.id && "bg-muted"
            )}
          >
            <div className="relative">
              <Avatar className="h-10 w-10">
                <AvatarImage src={chat.avatar} alt={chat.name} />
                <AvatarFallback>{getInitials(chat.name)}</AvatarFallback>
              </Avatar>
              {chat.isOnline && (
                <span className="absolute bottom-0 right-0 h-3 w-3 rounded-full border-2 border-background bg-green-500" />
              )}
            </div>
            <div className="flex-1 overflow-hidden">
              <div className="flex items-center justify-between">
                <span className="font-medium truncate">{chat.name}</span>
                <span className="text-xs text-muted-foreground">
                  {formatTimestamp(chat.timestamp)}
                </span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-sm text-muted-foreground truncate">
                  {chat.lastMessage}
                </span>
                {chat.unreadCount && chat.unreadCount > 0 && (
                  <span className="ml-2 flex h-5 min-w-5 items-center justify-center rounded-full bg-primary px-1.5 text-xs font-medium text-primary-foreground">
                    {chat.unreadCount}
                  </span>
                )}
              </div>
            </div>
          </button>
        ))}
      </div>
    </ScrollArea>
  )
}
