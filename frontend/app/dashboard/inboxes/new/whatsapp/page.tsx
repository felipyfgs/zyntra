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
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group"
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
import { Loader2, QrCode, Cloud, Phone } from "lucide-react"
import { useCreateInbox } from "@/lib/api/hooks"

type ConnectionMethod = "qrcode" | "cloud_api"

export default function NewWhatsAppInboxPage() {
  const router = useRouter()
  const createMutation = useCreateInbox()
  
  const [name, setName] = useState("")
  const [method, setMethod] = useState<ConnectionMethod>("qrcode")
  const [greetingMessage, setGreetingMessage] = useState("")
  const [autoAssignment, setAutoAssignment] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!name.trim()) return

    try {
      await createMutation.mutateAsync({
        name: name.trim(),
        channel_type: "whatsapp",
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
                  <BreadcrumbPage>WhatsApp</BreadcrumbPage>
                </BreadcrumbItem>
              </BreadcrumbList>
            </Breadcrumb>
          </div>
        </header>

        <div className="flex flex-1 flex-col gap-6 p-4 pt-0">
          <div className="flex items-center gap-4">
            <div className="w-12 h-12 rounded-xl bg-green-100 flex items-center justify-center">
              <Phone className="h-6 w-6 text-green-600" />
            </div>
            <div>
              <h1 className="text-2xl font-bold">Configurar WhatsApp</h1>
              <p className="text-muted-foreground">
                Conecte seu WhatsApp para receber e responder mensagens
              </p>
            </div>
          </div>

          <form onSubmit={handleSubmit} className="space-y-6">
            <Card>
              <CardHeader>
                <CardTitle>Informacoes Basicas</CardTitle>
                <CardDescription>
                  Configure o nome e as opcoes do seu inbox
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
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

                <div className="space-y-2">
                  <Label htmlFor="greeting">Mensagem de Boas-vindas</Label>
                  <Textarea
                    id="greeting"
                    value={greetingMessage}
                    onChange={(e) => setGreetingMessage(e.target.value)}
                    placeholder="Ola! Seja bem-vindo ao nosso atendimento. Como posso ajudar?"
                    rows={3}
                  />
                  <p className="text-xs text-muted-foreground">
                    Enviada automaticamente quando um novo contato inicia uma conversa
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
                <CardTitle>Metodo de Conexao</CardTitle>
                <CardDescription>
                  Escolha como deseja conectar seu WhatsApp
                </CardDescription>
              </CardHeader>
              <CardContent>
                <RadioGroup value={method} onValueChange={(v) => setMethod(v as ConnectionMethod)} className="space-y-3">
                  <label className="flex items-start space-x-3 p-4 rounded-lg border cursor-pointer hover:bg-muted/50 transition-colors">
                    <RadioGroupItem value="qrcode" className="mt-1" />
                    <div className="flex-1">
                      <div className="flex items-center gap-2 font-medium">
                        <QrCode className="h-4 w-4 text-green-600" />
                        QR Code (Recomendado)
                      </div>
                      <p className="text-sm text-muted-foreground mt-1">
                        Escaneie o QR code com seu celular para conectar. Simples e rapido, 
                        ideal para a maioria dos casos.
                      </p>
                    </div>
                  </label>
                  
                  <label className="flex items-start space-x-3 p-4 rounded-lg border cursor-not-allowed opacity-60">
                    <RadioGroupItem value="cloud_api" className="mt-1" disabled />
                    <div className="flex-1">
                      <div className="flex items-center gap-2 font-medium">
                        <Cloud className="h-4 w-4 text-blue-600" />
                        WhatsApp Cloud API
                        <span className="text-[10px] px-2 py-0.5 rounded-full bg-muted text-muted-foreground">
                          Em breve
                        </span>
                      </div>
                      <p className="text-sm text-muted-foreground mt-1">
                        Conecte via API oficial do Meta. Requer conta Business verificada 
                        e configuracao no Meta Business Suite.
                      </p>
                    </div>
                  </label>
                </RadioGroup>
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
