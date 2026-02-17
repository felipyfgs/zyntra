"use client"

import { Button } from "@/components/ui/button"
import { 
  MoreHorizontal, 
  Power, 
  Trash2, 
  Loader2, 
  QrCode,
} from "lucide-react"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import type { Inbox } from "@/lib/api/types"
import { cn } from "@/lib/utils"
import { channelConfig, statusConfig } from "./inbox-card"

interface InboxListItemProps {
  inbox: Inbox
  onConnect?: () => void
  onDisconnect?: () => void
  onDelete?: () => void
  onShowQR?: () => void
  isLoading?: boolean
}

export function InboxListItem({
  inbox,
  onConnect,
  onDisconnect,
  onDelete,
  onShowQR,
  isLoading,
}: InboxListItemProps) {
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
    <div className="flex items-center justify-between p-3 rounded-lg border hover:bg-muted/50 transition-colors">
      <div className="flex items-center gap-3">
        <div className={cn("p-2 rounded-lg", channel.bgLight)}>
          <ChannelIcon className={cn("h-4 w-4", channel.textColor)} />
        </div>
        <div>
          <h3 className="font-medium text-sm">{inbox.name}</h3>
          <p className="text-xs text-muted-foreground">{channel.label}</p>
        </div>
      </div>

      <div className="flex items-center gap-4">
        <div className="flex items-center gap-2">
          <span className={cn("h-2 w-2 rounded-full", status.color)} />
          <span className="text-sm text-muted-foreground">
            {isConnecting && <Loader2 className="h-3 w-3 animate-spin inline mr-1" />}
            {status.label}
          </span>
        </div>

        {canConnect && needsQR && (
          <Button onClick={handleConnect} size="sm" variant="ghost" disabled={isLoading}>
            <QrCode className="h-4 w-4" />
          </Button>
        )}

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
    </div>
  )
}
