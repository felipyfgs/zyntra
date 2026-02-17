"use client"

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { MoreHorizontal, Power, RefreshCw, Trash2, Loader2, QrCode } from "lucide-react"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"

export interface Connection {
  id: string
  name: string
  platform: "whatsapp" | "telegram" | "instagram" | "messenger" | "api"
  status: "connected" | "disconnected" | "pending" | "connecting" | "qr_code"
  phone?: string
  lastSync?: Date
}

interface ConnectionCardProps {
  connection: Connection
  onConnect?: () => void
  onDisconnect?: () => void
  onDelete?: () => void
  onShowQR?: () => void
  isLoading?: boolean
}

const platformConfig = {
  whatsapp: { label: "WhatsApp", color: "bg-green-500" },
  telegram: { label: "Telegram", color: "bg-blue-500" },
  instagram: { label: "Instagram", color: "bg-pink-500" },
  messenger: { label: "Messenger", color: "bg-purple-500" },
  api: { label: "API", color: "bg-gray-500" },
}

const statusConfig = {
  connected: { label: "Conectado", variant: "default" as const, color: "bg-green-500" },
  disconnected: { label: "Desconectado", variant: "secondary" as const, color: "bg-gray-500" },
  pending: { label: "Pendente", variant: "outline" as const, color: "bg-yellow-500" },
  connecting: { label: "Conectando...", variant: "outline" as const, color: "bg-blue-500" },
  qr_code: { label: "Aguardando scan", variant: "outline" as const, color: "bg-orange-500" },
}

export function ConnectionCard({
  connection,
  onConnect,
  onDisconnect,
  onDelete,
  onShowQR,
  isLoading,
}: ConnectionCardProps) {
  const platform = platformConfig[connection.platform]
  const status = statusConfig[connection.status] || statusConfig.disconnected

  const isConnecting = connection.status === "connecting" || connection.status === "qr_code"
  const canConnect = connection.status === "disconnected" || connection.status === "pending"

  const handleConnect = () => {
    if (onShowQR) {
      onShowQR()
    } else if (onConnect) {
      onConnect()
    }
  }

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <div className="flex items-center gap-2">
          <div className={`h-3 w-3 rounded-full ${platform.color}`} />
          <CardTitle className="text-base font-medium">{platform.label}</CardTitle>
        </div>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="icon" className="h-8 w-8" disabled={isLoading}>
              <MoreHorizontal className="h-4 w-4" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            {canConnect && (
              <DropdownMenuItem onClick={handleConnect}>
                <QrCode className="mr-2 h-4 w-4" />
                Conectar
              </DropdownMenuItem>
            )}
            {connection.status === "connected" && (
              <DropdownMenuItem onClick={onDisconnect}>
                <Power className="mr-2 h-4 w-4" />
                Desconectar
              </DropdownMenuItem>
            )}
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
            <span className="text-sm text-muted-foreground">{connection.name}</span>
            <Badge variant={status.variant} className="gap-1">
              {isConnecting && <Loader2 className="h-3 w-3 animate-spin" />}
              <span className={`h-2 w-2 rounded-full ${status.color}`} />
              {status.label}
            </Badge>
          </div>
          {connection.phone && (
            <p className="text-sm text-muted-foreground">{connection.phone}</p>
          )}
          {connection.lastSync && (
            <p className="text-xs text-muted-foreground">
              Ultima sinc: {connection.lastSync.toLocaleString()}
            </p>
          )}
          {canConnect && (
            <Button onClick={handleConnect} className="mt-2" size="sm" disabled={isLoading}>
              {isLoading ? (
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              ) : (
                <QrCode className="mr-2 h-4 w-4" />
              )}
              Conectar WhatsApp
            </Button>
          )}
        </div>
      </CardContent>
    </Card>
  )
}
