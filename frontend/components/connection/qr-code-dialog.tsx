"use client"

import { useEffect } from "react"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { useConnectWhatsApp, useQRCode, useConnection } from "@/lib/api/hooks"
import { Loader2, CheckCircle2, Smartphone } from "lucide-react"

interface QRCodeDialogProps {
  connectionId: string | null
  connectionName: string
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function QRCodeDialog({
  connectionId,
  connectionName,
  open,
  onOpenChange,
}: QRCodeDialogProps) {
  const connectMutation = useConnectWhatsApp()
  
  // Poll QR code from API (wuzapi pattern - HTTP polling)
  const { data: qrData, isLoading: qrLoading } = useQRCode(
    open ? connectionId : null,
    open
  )
  
  // Poll connection status
  const { data: connection } = useConnection(open ? connectionId : null)

  // Start connection when dialog opens
  useEffect(() => {
    if (open && connectionId && !connectMutation.isPending) {
      connectMutation.mutate(connectionId)
    }
  }, [open, connectionId])

  // Close dialog when connected
  useEffect(() => {
    if (connection?.status === "connected") {
      setTimeout(() => onOpenChange(false), 1500)
    }
  }, [connection?.status, onOpenChange])

  const isConnecting = connectMutation.isPending || qrLoading
  const isConnected = connection?.status === "connected"
  const qrCode = qrData?.qr_code

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Conectar {connectionName}</DialogTitle>
          <DialogDescription>
            Escaneie o QR code com seu WhatsApp para conectar
          </DialogDescription>
        </DialogHeader>

        <div className="flex flex-col items-center justify-center py-6">
          {isConnected ? (
            <div className="flex flex-col items-center gap-4 text-green-600">
              <CheckCircle2 className="h-16 w-16" />
              <p className="text-lg font-medium">Conectado com sucesso!</p>
              {connection?.phone && (
                <p className="text-sm text-muted-foreground">
                  {connection.phone}
                </p>
              )}
            </div>
          ) : qrCode ? (
            <div className="flex flex-col items-center gap-4">
              {/* QR code is already base64 image from backend (wuzapi pattern) */}
              <div className="rounded-lg border bg-white p-4">
                <img 
                  src={qrCode} 
                  alt="WhatsApp QR Code" 
                  width={256} 
                  height={256}
                  className="w-64 h-64"
                />
              </div>
              <div className="flex items-center gap-2 text-sm text-muted-foreground">
                <Smartphone className="h-4 w-4" />
                <span>Abra o WhatsApp e escaneie o codigo</span>
              </div>
            </div>
          ) : (
            <div className="flex flex-col items-center gap-4">
              <Loader2 className="h-12 w-12 animate-spin text-muted-foreground" />
              <p className="text-sm text-muted-foreground">
                {isConnecting ? "Gerando QR code..." : "Aguardando..."}
              </p>
            </div>
          )}
        </div>

        {connectMutation.isError && (
          <p className="text-sm text-destructive text-center">
            Erro ao conectar. Tente novamente.
          </p>
        )}
      </DialogContent>
    </Dialog>
  )
}
