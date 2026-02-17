"use client"

import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
import { Switch } from "@/components/ui/switch"
import { Loader2, ExternalLink } from "lucide-react"
import type { CreateInboxRequest } from "@/lib/api/types"

interface TelegramSetupProps {
  onSubmit: (data: CreateInboxRequest) => Promise<void>
  onCancel: () => void
  isLoading: boolean
}

export function TelegramSetup({ onSubmit, onCancel, isLoading }: TelegramSetupProps) {
  const [name, setName] = useState("")
  const [botToken, setBotToken] = useState("")
  const [greetingMessage, setGreetingMessage] = useState("")
  const [autoAssignment, setAutoAssignment] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!name.trim()) return

    await onSubmit({
      name: name.trim(),
      channel_type: "telegram",
      greeting_message: greetingMessage.trim() || undefined,
      auto_assignment: autoAssignment,
    })
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-6 py-4">
      <div className="space-y-4">
        <div className="p-4 bg-blue-50 dark:bg-blue-950/30 rounded-lg text-sm">
          <p className="font-medium text-blue-700 dark:text-blue-300 mb-2">
            Como obter o Bot Token:
          </p>
          <ol className="list-decimal list-inside space-y-1 text-blue-600 dark:text-blue-400 text-xs">
            <li>Abra o Telegram e busque por @BotFather</li>
            <li>Envie /newbot e siga as instrucoes</li>
            <li>Copie o token gerado e cole abaixo</li>
          </ol>
          <a 
            href="https://t.me/botfather" 
            target="_blank" 
            rel="noopener noreferrer"
            className="inline-flex items-center gap-1 text-blue-700 dark:text-blue-300 hover:underline mt-2 text-xs font-medium"
          >
            Abrir BotFather <ExternalLink className="h-3 w-3" />
          </a>
        </div>

        <div className="space-y-2">
          <Label htmlFor="name">Nome do Inbox *</Label>
          <Input
            id="name"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="Ex: Suporte Telegram"
            required
          />
        </div>

        <div className="space-y-2">
          <Label htmlFor="token">Bot Token *</Label>
          <Input
            id="token"
            value={botToken}
            onChange={(e) => setBotToken(e.target.value)}
            placeholder="123456789:ABCdefGHIjklMNOpqrsTUVwxyz"
            required
          />
          <p className="text-xs text-muted-foreground">
            Token fornecido pelo @BotFather ao criar seu bot
          </p>
        </div>

        <div className="space-y-2">
          <Label htmlFor="greeting">Mensagem de Boas-vindas (opcional)</Label>
          <Textarea
            id="greeting"
            value={greetingMessage}
            onChange={(e) => setGreetingMessage(e.target.value)}
            placeholder="Ola! Bem-vindo ao nosso bot de atendimento."
            rows={3}
          />
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
