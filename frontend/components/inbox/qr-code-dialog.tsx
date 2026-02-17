"use client"

import { useEffect } from "react"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { Loader2, RefreshCw, CheckCircle2, XCircle } from "lucide-react"
import { useConnectInbox, useInbox, useQRCode } from "@/lib/api/hooks"

interface QRCodeDialogProps {
  inboxId: string | null
  inboxName: string
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function QRCodeDialog({
  inboxId,
  inboxName,
  open,
  onOpenChange,
}: QRCodeDialogProps) {
  const connectMutation = useConnectInbox()
  const { data: inbox } = useInbox(inboxId)
  const { data: qrData, isLoading: qrLoading, refetch: refetchQR } = useQRCode(
    inboxId,
    open && inbox?.status !== "connected"
  )

  useEffect(() => {
    if (open && inboxId && inbox?.status === "disconnected") {
      connectMutation.mutate(inboxId)
    }
  }, [open, inboxId, inbox?.status])

  useEffect(() => {
    if (inbox?.status === "connected") {
      const timer = setTimeout(() => onOpenChange(false), 2000)
      return () => clearTimeout(timer)
    }
  }, [inbox?.status, onOpenChange])

  const isConnecting = connectMutation.isPending || qrLoading
  const isConnected = inbox?.status === "connected"
  const qrCode = qrData?.qrcode

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Conectar {inboxName}</DialogTitle>
          <DialogDescription>
            Escaneie o QR code com o WhatsApp do seu celular
          </DialogDescription>
        </DialogHeader>

        <div className="flex flex-col items-center gap-4 py-6">
          {isConnected ? (
            <div className="flex flex-col items-center gap-3 text-green-600">
              <CheckCircle2 className="h-16 w-16" />
              <p className="text-lg font-medium">Conectado com sucesso!</p>
              {inbox?.phone && (
                <p className="text-sm text-muted-foreground">{inbox.phone}</p>
              )}
            </div>
          ) : isConnecting && !qrCode ? (
            <div className="flex flex-col items-center gap-3">
              <Loader2 className="h-16 w-16 animate-spin text-muted-foreground" />
              <p className="text-sm text-muted-foreground">Gerando QR code...</p>
            </div>
          ) : qrCode ? (
            <div className="flex flex-col items-center gap-3">
              <div className="bg-white p-4 rounded-lg shadow-sm">
                <img
                  src={qrCode}
                  alt="QR Code"
                  className="w-64 h-64"
                />
              </div>
              <p className="text-xs text-muted-foreground text-center max-w-[280px]">
                Abra o WhatsApp no celular, va em Dispositivos conectados e escaneie este codigo
              </p>
              <Button
                variant="outline"
                size="sm"
                onClick={() => refetchQR()}
                disabled={qrLoading}
              >
                <RefreshCw className={`mr-2 h-4 w-4 ${qrLoading ? 'animate-spin' : ''}`} />
                Atualizar QR
              </Button>
            </div>
          ) : (
            <div className="flex flex-col items-center gap-3 text-muted-foreground">
              <XCircle className="h-16 w-16" />
              <p className="text-sm">Erro ao gerar QR code</p>
              <Button
                variant="outline"
                size="sm"
                onClick={() => inboxId && connectMutation.mutate(inboxId)}
              >
                <RefreshCw className="mr-2 h-4 w-4" />
                Tentar novamente
              </Button>
            </div>
          )}
        </div>
      </DialogContent>
    </Dialog>
  )
}
