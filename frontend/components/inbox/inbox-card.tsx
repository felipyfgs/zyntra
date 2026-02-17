"use client"

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { 
  MoreHorizontal, 
  Power, 
  Trash2, 
  Loader2, 
  QrCode,
  MessageSquare,
  Phone,
  Bot,
  Send
} from "lucide-react"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import type { Inbox } from "@/lib/api/types"

interface InboxCardProps {
  inbox: Inbox
  onConnect?: () => void
  onDisconnect?: () => void
  onDelete?: () => void
  onShowQR?: () => void
  isLoading?: boolean
}

const channelConfig = {
  whatsapp: { 
    label: "WhatsApp", 
    color: "bg-green-500",
    icon: Phone,
  },
  telegram: { 
    label: "Telegram", 
    color: "bg-blue-500",
    icon: Send,
  },
  api: { 
    label: "API", 
    color: "bg-gray-500",
    icon: Bot,
  },
}

const statusConfig = {
  connected: { label: "Conectado", variant: "default" as const, color: "bg-green-500" },
  disconnected: { label: "Desconectado", variant: "secondary" as const, color: "bg-gray-500" },
  connecting: { label: "Conectando...", variant: "outline" as const, color: "bg-blue-500" },
  qr_code: { label: "Aguardando QR", variant: "outline" as const, color: "bg-orange-500" },
}

export function InboxCard({
  inbox,
  onConnect,
  onDisconnect,
  onDelete,
  onShowQR,
  isLoading,
}: InboxCardProps) {
  const channel = channelConfig[inbox.channel_type] || channelConfig.whatsapp
  const status = statusConfig[inbox.status] || statusConfig.disconnected
  const ChannelIcon = channel.icon

  const isConnecting = inbox.status === "connecting" || inbox.status === "qr_code"
  const canConnect = inbox.status === "disconnected"
  const needsQR = inbox.channel_type === "whatsapp" || inbox.channel_type === "telegram"

  const handleConnect = () => {
    if (needsQR && onShowQR) {
      onShowQR()
    } else if (onConnect) {
      onConnect()
    }
  }

  return (
    <Card className="relative overflow-hidden">
      <div className={`absolute top-0 left-0 w-1 h-full ${channel.color}`} />
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <div className="flex items-center gap-2">
          <div className={`p-1.5 rounded-md ${channel.color}/10`}>
            <ChannelIcon className={`h-4 w-4 text-${channel.color.replace('bg-', '')}`} />
          </div>
          <div>
            <CardTitle className="text-base font-medium">{inbox.name}</CardTitle>
            <span className="text-xs text-muted-foreground">{channel.label}</span>
          </div>
        </div>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="icon" className="h-8 w-8" disabled={isLoading}>
              <MoreHorizontal className="h-4 w-4" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            {canConnect && needsQR && (
              <DropdownMenuItem onClick={handleConnect}>
                <QrCode className="mr-2 h-4 w-4" />
                Conectar
              </DropdownMenuItem>
            )}
            {inbox.status === "connected" && (
              <DropdownMenuItem onClick={onDisconnect}>
                <Power className="mr-2 h-4 w-4" />
                Desconectar
              </DropdownMenuItem>
            )}
            <DropdownMenuSeparator />
            <DropdownMenuItem onClick={onDelete} className="text-destructive">
              <Trash2 className="mr-2 h-4 w-4" />
              Remover
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </CardHeader>
      <CardContent>
        <div className="flex flex-col gap-3">
          <div className="flex items-center justify-between">
            <Badge variant={status.variant} className="gap-1.5">
              {isConnecting && <Loader2 className="h-3 w-3 animate-spin" />}
              <span className={`h-2 w-2 rounded-full ${status.color}`} />
              {status.label}
            </Badge>
            {inbox.auto_assignment && (
              <Badge variant="outline" className="text-xs">
                Auto-assign
              </Badge>
            )}
          </div>

          {inbox.phone && (
            <div className="flex items-center gap-2 text-sm text-muted-foreground">
              <Phone className="h-3.5 w-3.5" />
              {inbox.phone}
            </div>
          )}

          {inbox.greeting_message && (
            <div className="flex items-start gap-2 text-xs text-muted-foreground bg-muted/50 p-2 rounded-md">
              <MessageSquare className="h-3.5 w-3.5 mt-0.5 shrink-0" />
              <span className="line-clamp-2">{inbox.greeting_message}</span>
            </div>
          )}

          {canConnect && needsQR && (
            <Button onClick={handleConnect} className="mt-2" size="sm" disabled={isLoading}>
              {isLoading ? (
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              ) : (
                <QrCode className="mr-2 h-4 w-4" />
              )}
              Conectar {channel.label}
            </Button>
          )}

          {inbox.channel_type === "api" && inbox.status === "disconnected" && (
            <Button onClick={onConnect} className="mt-2" size="sm" variant="outline" disabled={isLoading}>
              <Bot className="mr-2 h-4 w-4" />
              Configurar API
            </Button>
          )}
        </div>
      </CardContent>
    </Card>
  )
}
