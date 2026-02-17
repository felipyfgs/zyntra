"use client"

import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
import { Switch } from "@/components/ui/switch"
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group"
import { Loader2, QrCode, Cloud } from "lucide-react"
import type { CreateInboxRequest } from "@/lib/api/types"

interface WhatsAppSetupProps {
  onSubmit: (data: CreateInboxRequest) => Promise<void>
  onCancel: () => void
  isLoading: boolean
}

type ConnectionMethod = "qrcode" | "cloud_api"

export function WhatsAppSetup({ onSubmit, onCancel, isLoading }: WhatsAppSetupProps) {
  const [name, setName] = useState("")
  const [method, setMethod] = useState<ConnectionMethod>("qrcode")
  const [greetingMessage, setGreetingMessage] = useState("")
  const [autoAssignment, setAutoAssignment] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!name.trim()) return

    await onSubmit({
      name: name.trim(),
      channel_type: "whatsapp",
      greeting_message: greetingMessage.trim() || undefined,
      auto_assignment: autoAssignment,
    })
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-6 py-4">
      <div className="space-y-4">
        <div className="space-y-2">
          <Label htmlFor="name">Nome do Inbox *</Label>
          <Input
            id="name"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="Ex: Atendimento Principal"
            required
          />
        </div>

        <div className="space-y-3">
          <Label>Metodo de Conexao</Label>
          <RadioGroup value={method} onValueChange={(v) => setMethod(v as ConnectionMethod)}>
            <div className="flex items-start space-x-3 p-3 rounded-lg border hover:bg-muted/50 cursor-pointer">
              <RadioGroupItem value="qrcode" id="qrcode" className="mt-1" />
              <div className="flex-1">
                <Label htmlFor="qrcode" className="flex items-center gap-2 cursor-pointer font-medium">
                  <QrCode className="h-4 w-4 text-green-600" />
                  QR Code (Recomendado)
                </Label>
                <p className="text-xs text-muted-foreground mt-1">
                  Escaneie o QR code com seu celular para conectar. Simples e rapido.
                </p>
              </div>
            </div>
            <div className="flex items-start space-x-3 p-3 rounded-lg border hover:bg-muted/50 cursor-pointer opacity-60">
              <RadioGroupItem value="cloud_api" id="cloud_api" className="mt-1" disabled />
              <div className="flex-1">
                <Label htmlFor="cloud_api" className="flex items-center gap-2 cursor-pointer font-medium">
                  <Cloud className="h-4 w-4 text-blue-600" />
                  WhatsApp Cloud API
                  <span className="text-[10px] px-1.5 py-0.5 rounded bg-muted text-muted-foreground">Em breve</span>
                </Label>
                <p className="text-xs text-muted-foreground mt-1">
                  Conecte via API oficial do Meta. Requer conta Business verificada.
                </p>
              </div>
            </div>
          </RadioGroup>
        </div>

        <div className="space-y-2">
          <Label htmlFor="greeting">Mensagem de Boas-vindas (opcional)</Label>
          <Textarea
            id="greeting"
            value={greetingMessage}
            onChange={(e) => setGreetingMessage(e.target.value)}
            placeholder="Ola! Seja bem-vindo ao nosso atendimento. Como posso ajudar?"
            rows={3}
          />
          <p className="text-xs text-muted-foreground">
            Enviada automaticamente quando um novo contato inicia uma conversa.
          </p>
        </div>

        <div className="flex items-center justify-between rounded-lg border p-4">
          <div className="space-y-0.5">
            <Label htmlFor="auto_assignment">Atribuicao Automatica</Label>
            <p className="text-xs text-muted-foreground">
              Distribui conversas entre agentes disponiveis
            </p>
          </div>
          <Switch
            id="auto_assignment"
            checked={autoAssignment}
            onCheckedChange={setAutoAssignment}
          />
        </div>
      </div>

      <div className="flex justify-end gap-3 pt-4 border-t">
        <Button type="button" variant="outline" onClick={onCancel}>
          Voltar
        </Button>
        <Button type="submit" disabled={isLoading || !name.trim()}>
          {isLoading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
          Criar Inbox
        </Button>
      </div>
    </form>
  )
}
