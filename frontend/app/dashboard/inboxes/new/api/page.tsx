"use client"

import { useState } from "react"
import { useRouter } from "next/navigation"
import Link from "next/link"
import { AppSidebar } from "@/components/layout/app-sidebar"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
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
import { Loader2, Bot, Copy, Check } from "lucide-react"
import { useCreateInbox } from "@/lib/api/hooks"

export default function NewAPIInboxPage() {
  const router = useRouter()
  const createMutation = useCreateInbox()
  
  const [name, setName] = useState("")
  const [webhookUrl, setWebhookUrl] = useState("")
  const [autoAssignment, setAutoAssignment] = useState(false)
  const [copied, setCopied] = useState<string | null>(null)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!name.trim()) return

    try {
      await createMutation.mutateAsync({
        name: name.trim(),
        channel_type: "api",
        auto_assignment: autoAssignment,
      })
      router.push("/dashboard/inboxes")
    } catch (error) {
      console.error("Failed to create inbox:", error)
    }
  }

  const handleCopy = (text: string, id: string) => {
    navigator.clipboard.writeText(text)
    setCopied(id)
    setTimeout(() => setCopied(null), 2000)
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
                  <BreadcrumbPage>API</BreadcrumbPage>
                </BreadcrumbItem>
              </BreadcrumbList>
            </Breadcrumb>
          </div>
        </header>

        <div className="flex flex-1 flex-col gap-6 p-4 pt-0">
          <div className="flex items-center gap-4">
            <div className="w-12 h-12 rounded-xl bg-purple-100 flex items-center justify-center">
              <Bot className="h-6 w-6 text-purple-600" />
            </div>
            <div>
              <h1 className="text-2xl font-bold">Configurar Canal API</h1>
              <p className="text-muted-foreground">
                Integre sistemas externos via API REST
              </p>
            </div>
          </div>

          <Card className="bg-purple-50 border-purple-200">
            <CardContent className="pt-6">
              <h3 className="font-semibold text-purple-800 mb-2">O que e um Canal API?</h3>
              <p className="text-sm text-purple-700">
                Use este canal para integrar sistemas externos como CRM, ERP ou aplicativos 
                customizados. Voce podera enviar e receber mensagens programaticamente 
                atraves da nossa API REST.
              </p>
            </CardContent>
          </Card>

          <form onSubmit={handleSubmit} className="space-y-6">
            <Card>
              <CardHeader>
                <CardTitle>Configuracao do Canal</CardTitle>
                <CardDescription>
                  Configure as informacoes basicas do seu canal API
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
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
                    URL que recebera notificacoes de novas mensagens via POST
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
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>Endpoints da API</CardTitle>
                <CardDescription>
                  Apos criar o inbox, use estes endpoints com o ID gerado
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="space-y-2">
                  <Label className="text-xs text-muted-foreground">Enviar mensagem</Label>
                  <div className="flex items-center gap-2">
                    <code className="flex-1 p-3 bg-muted rounded-lg text-sm font-mono">
                      POST /api/v1/conversations/{"{conversation_id}"}/messages
                    </code>
                    <Button 
                      type="button" 
                      variant="outline" 
                      size="icon"
                      onClick={() => handleCopy("POST /api/v1/conversations/{conversation_id}/messages", "send")}
                    >
                      {copied === "send" ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}
                    </Button>
                  </div>
                </div>

                <div className="space-y-2">
                  <Label className="text-xs text-muted-foreground">Criar contato</Label>
                  <div className="flex items-center gap-2">
                    <code className="flex-1 p-3 bg-muted rounded-lg text-sm font-mono">
                      POST /api/v1/contacts
                    </code>
                    <Button 
                      type="button" 
                      variant="outline" 
                      size="icon"
                      onClick={() => handleCopy("POST /api/v1/contacts", "contact")}
                    >
                      {copied === "contact" ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}
                    </Button>
                  </div>
                </div>

                <div className="space-y-2">
                  <Label className="text-xs text-muted-foreground">Listar conversas</Label>
                  <div className="flex items-center gap-2">
                    <code className="flex-1 p-3 bg-muted rounded-lg text-sm font-mono">
                      GET /api/v1/conversations?inbox_id={"{inbox_id}"}
                    </code>
                    <Button 
                      type="button" 
                      variant="outline" 
                      size="icon"
                      onClick={() => handleCopy("GET /api/v1/conversations?inbox_id={inbox_id}", "list")}
                    >
                      {copied === "list" ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}
                    </Button>
                  </div>
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
