"use client"

import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Switch } from "@/components/ui/switch"
import { Loader2, Copy, Check } from "lucide-react"
import type { CreateInboxRequest } from "@/lib/api/types"

interface APISetupProps {
  onSubmit: (data: CreateInboxRequest) => Promise<void>
  onCancel: () => void
  isLoading: boolean
}

export function APISetup({ onSubmit, onCancel, isLoading }: APISetupProps) {
  const [name, setName] = useState("")
  const [webhookUrl, setWebhookUrl] = useState("")
  const [autoAssignment, setAutoAssignment] = useState(false)
  const [copied, setCopied] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!name.trim()) return

    await onSubmit({
      name: name.trim(),
      channel_type: "api",
      auto_assignment: autoAssignment,
    })
  }

  const handleCopyEndpoint = () => {
    navigator.clipboard.writeText("POST /api/v1/inboxes/{inbox_id}/messages")
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-6 py-4">
      <div className="space-y-4">
        <div className="p-4 bg-purple-50 dark:bg-purple-950/30 rounded-lg text-sm">
          <p className="font-medium text-purple-700 dark:text-purple-300 mb-2">
            Canal API
          </p>
          <p className="text-purple-600 dark:text-purple-400 text-xs">
            Use este canal para integrar sistemas externos via API REST. 
            Voce podera enviar e receber mensagens programaticamente.
          </p>
        </div>

        <div className="space-y-2">
          <Label htmlFor="name">Nome do Inbox *</Label>
          <Input
            id="name"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="Ex: Integracao CRM"
            required
          />
        </div>

        <div className="space-y-2">
          <Label htmlFor="webhook">Webhook URL (opcional)</Label>
          <Input
            id="webhook"
            value={webhookUrl}
            onChange={(e) => setWebhookUrl(e.target.value)}
            placeholder="https://seu-sistema.com/webhook"
            type="url"
          />
          <p className="text-xs text-muted-foreground">
            URL que recebera notificacoes de novas mensagens (POST)
          </p>
        </div>

        <div className="space-y-2">
          <Label>Endpoint para enviar mensagens</Label>
          <div className="flex items-center gap-2">
            <code className="flex-1 p-2 bg-muted rounded text-xs font-mono">
              POST /api/v1/inboxes/{"{inbox_id}"}/messages
            </code>
            <Button 
              type="button" 
              variant="outline" 
              size="sm"
              onClick={handleCopyEndpoint}
            >
              {copied ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}
            </Button>
          </div>
          <p className="text-xs text-muted-foreground">
            Apos criar o inbox, use este endpoint com o ID gerado
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
