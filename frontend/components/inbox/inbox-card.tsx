"use client"

import { Card } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { 
  MoreHorizontal, 
  Power, 
  Trash2, 
  Loader2, 
  QrCode,
  Bot,
} from "lucide-react"
import { SiWhatsapp, SiTelegram } from "@icons-pack/react-simple-icons"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import type { Inbox } from "@/lib/api/types"
import { cn } from "@/lib/utils"

interface InboxCardProps {
  inbox: Inbox
  onConnect?: () => void
  onDisconnect?: () => void
  onDelete?: () => void
  onShowQR?: () => void
  isLoading?: boolean
}

export const channelConfig = {
  whatsapp: { 
    label: "WhatsApp", 
    brandColor: "#25D366",
    bgLight: "bg-green-100",
    icon: SiWhatsapp,
  },
  telegram: { 
    label: "Telegram", 
    brandColor: "#26A5E4",
    bgLight: "bg-blue-100",
    icon: SiTelegram,
  },
  api: { 
    label: "API", 
    brandColor: "#8B5CF6",
    bgLight: "bg-purple-100",
    icon: Bot,
  },
}

export const statusConfig = {
  connected: { label: "Conectado", color: "bg-green-500" },
  disconnected: { label: "Desconectado", color: "bg-gray-400" },
  connecting: { label: "Conectando", color: "bg-blue-500" },
  qr_code: { label: "Aguardando", color: "bg-orange-500" },
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
    <Card className="p-4 hover:shadow-md transition-shadow">
      <div className="flex items-start justify-between">
        <div className="flex items-center gap-3">
          <div className={cn("p-2.5 rounded-lg", channel.bgLight)}>
            <ChannelIcon className="h-5 w-5" color={channel.brandColor} />
          </div>
          <div>
            <h3 className="font-medium">{inbox.name}</h3>
            <p className="text-sm text-muted-foreground">{channel.label}</p>
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
      </div>

      <div className="mt-4 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <span className={cn("h-2 w-2 rounded-full", status.color)} />
          <span className="text-sm text-muted-foreground">
            {isConnecting && <Loader2 className="h-3 w-3 animate-spin inline mr-1" />}
            {status.label}
          </span>
        </div>
        
        {canConnect && needsQR && (
          <Button onClick={handleConnect} size="sm" variant="outline" disabled={isLoading}>
            {isLoading ? <Loader2 className="h-4 w-4 animate-spin" /> : <QrCode className="h-4 w-4" />}
          </Button>
        )}
      </div>
    </Card>
  )
}
