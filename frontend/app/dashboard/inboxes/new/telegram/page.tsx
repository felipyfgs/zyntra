"use client"

import { useState } from "react"
import { useRouter } from "next/navigation"
import Link from "next/link"
import { AppSidebar } from "@/components/layout/app-sidebar"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
import { Switch } from "@/components/ui/switch"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb"
import { Separator } from "@/components/ui/separator"
import {
  SidebarInset,
  SidebarProvider,
  SidebarTrigger,
} from "@/components/ui/sidebar"
import { Loader2, Send, ExternalLink } from "lucide-react"
import { useCreateInbox } from "@/lib/api/hooks"

export default function NewTelegramInboxPage() {
  const router = useRouter()
  const createMutation = useCreateInbox()
  
  const [name, setName] = useState("")
  const [botToken, setBotToken] = useState("")
  const [greetingMessage, setGreetingMessage] = useState("")
  const [autoAssignment, setAutoAssignment] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!name.trim()) return

    try {
      await createMutation.mutateAsync({
        name: name.trim(),
        channel_type: "telegram",
        greeting_message: greetingMessage.trim() || undefined,
        auto_assignment: autoAssignment,
      })
      router.push("/dashboard/inboxes")
    } catch (error) {
      console.error("Failed to create inbox:", error)
    }
  }

  return (
    <SidebarProvider>
      <AppSidebar />
      <SidebarInset>
        <header className="flex h-16 shrink-0 items-center gap-2">
          <div className="flex items-center gap-2 px-4">
            <SidebarTrigger className="-ml-1" />
            <Separator orientation="vertical" className="mr-2 h-4" />
            <Breadcrumb>
              <BreadcrumbList>
                <BreadcrumbItem>
                  <BreadcrumbLink href="/dashboard/inboxes">Inboxes</BreadcrumbLink>
                </BreadcrumbItem>
                <BreadcrumbSeparator />
                <BreadcrumbItem>
                  <BreadcrumbLink href="/dashboard/inboxes/new">Novo</BreadcrumbLink>
                </BreadcrumbItem>
                <BreadcrumbSeparator />
                <BreadcrumbItem>
                  <BreadcrumbPage>Telegram</BreadcrumbPage>
                </BreadcrumbItem>
              </BreadcrumbList>
            </Breadcrumb>
          </div>
        </header>

        <div className="flex flex-1 flex-col gap-6 p-4 pt-0">
          <div className="flex items-center gap-4">
            <div className="w-12 h-12 rounded-xl bg-blue-100 flex items-center justify-center">
              <Send className="h-6 w-6 text-blue-600" />
            </div>
            <div>
              <h1 className="text-2xl font-bold">Configurar Telegram</h1>
              <p className="text-muted-foreground">
                Conecte um bot do Telegram para receber mensagens
              </p>
            </div>
          </div>

          <Card className="bg-blue-50 border-blue-200">
            <CardContent className="pt-6">
              <h3 className="font-semibold text-blue-800 mb-3">Como obter o Bot Token</h3>
              <ol className="list-decimal list-inside space-y-2 text-sm text-blue-700">
                <li>Abra o Telegram e busque por <strong>@BotFather</strong></li>
                <li>Envie o comando <code className="bg-blue-100 px-1 rounded">/newbot</code></li>
                <li>Siga as instrucoes para criar seu bot</li>
                <li>Copie o token gerado e cole no campo abaixo</li>
              </ol>
              <a 
                href="https://t.me/botfather" 
                target="_blank" 
                rel="noopener noreferrer"
                className="inline-flex items-center gap-1.5 text-blue-700 hover:text-blue-900 hover:underline mt-4 text-sm font-medium"
              >
                Abrir BotFather no Telegram
                <ExternalLink className="h-3.5 w-3.5" />
              </a>
            </CardContent>
          </Card>

          <form onSubmit={handleSubmit} className="space-y-6">
            <Card>
              <CardHeader>
                <CardTitle>Configuracao do Bot</CardTitle>
                <CardDescription>
                  Insira as informacoes do seu bot do Telegram
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
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
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>Opcoes Adicionais</CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="space-y-2">
                  <Label htmlFor="greeting">Mensagem de Boas-vindas</Label>
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
              </CardContent>
            </Card>

            <div className="flex justify-between pt-4">
              <Button type="button" variant="outline" asChild>
                <Link href="/dashboard/inboxes/new">Voltar</Link>
              </Button>
              <Button type="submit" disabled={createMutation.isPending || !name.trim()}>
                {createMutation.isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                Criar Inbox
              </Button>
            </div>
          </form>
        </div>
      </SidebarInset>
    </SidebarProvider>
  )
}
