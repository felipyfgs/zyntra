"use client"

import { cn } from "@/lib/utils"
import { Check, CheckCheck, Clock } from "lucide-react"

export type MessageStatus = "pending" | "sent" | "delivered" | "read" | "failed"

interface ChatBubbleProps {
  content: string
  timestamp: Date
  isOutgoing: boolean
  status?: MessageStatus
  className?: string
}

function formatTime(date: Date): string {
  return date.toLocaleTimeString("pt-BR", {
    hour: "2-digit",
    minute: "2-digit",
  })
}

function StatusIcon({ status }: { status?: MessageStatus }) {
  switch (status) {
    case "pending":
      return <Clock className="h-3 w-3 text-current opacity-60" />
    case "sent":
      return <Check className="h-3 w-3 text-current opacity-60" />
    case "delivered":
      return <CheckCheck className="h-3 w-3 text-current opacity-60" />
    case "read":
      return <CheckCheck className="h-3 w-3 text-blue-500" />
    case "failed":
      return <span className="text-xs text-red-500">!</span>
    default:
      return null
  }
}

export function ChatBubble({
  content,
  timestamp,
  isOutgoing,
  status,
  className,
}: ChatBubbleProps) {
  return (
    <div
      className={cn(
        "flex w-full",
        isOutgoing ? "justify-end" : "justify-start",
        className
      )}
    >
      <div className="relative max-w-[85%] md:max-w-[70%]">
        {/* Tail */}
        <div
          className={cn(
            "absolute top-0 h-3 w-3",
            isOutgoing
              ? "-right-1.5 text-accent dark:text-primary/20"
              : "-left-1.5 text-card"
          )}
        >
          {isOutgoing ? (
            <svg viewBox="0 0 8 13" className="h-full w-full fill-current">
              <path d="M5.188 1H0v11.193l6.467-8.625C7.526 2.156 6.958 1 5.188 1z" />
            </svg>
          ) : (
            <svg viewBox="0 0 8 13" className="h-full w-full fill-current">
              <path d="M2.812 1H8v11.193L1.533 3.568C.474 2.156 1.042 1 2.812 1z" />
            </svg>
          )}
        </div>

        {/* Bubble */}
        <div
          className={cn(
            "relative rounded-lg px-3 py-2 shadow-sm",
            isOutgoing
              ? "rounded-tr-none bg-accent text-accent-foreground dark:bg-primary/20"
              : "rounded-tl-none bg-card text-card-foreground"
          )}
        >
          {/* Content */}
          <p className="whitespace-pre-wrap break-words text-sm leading-relaxed">
            {content}
          </p>

          {/* Timestamp and Status */}
          <div
            className={cn(
              "mt-1 flex items-center justify-end gap-1",
              "text-[11px] text-muted-foreground"
            )}
          >
            <span>{formatTime(timestamp)}</span>
            {isOutgoing && <StatusIcon status={status} />}
          </div>
        </div>
      </div>
    </div>
  )
}

export function ChatBubbleGroup({
  children,
  className,
}: {
  children: React.ReactNode
  className?: string
}) {
  return (
    <div className={cn("flex flex-col gap-1", className)}>
      {children}
    </div>
  )
}
