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
import { Checkbox } from "@/components/ui/checkbox"
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
import { 
  Phone, 
  Send, 
  Bot, 
  Mail, 
  Globe, 
  MessageSquare,
  ChevronLeft,
  Check,
  ArrowRight,
  Loader2,
  QrCode,
  Cloud,
  ExternalLink
} from "lucide-react"
import { cn } from "@/lib/utils"
import { useCreateInbox } from "@/lib/api/hooks"

type ChannelType = "whatsapp" | "telegram" | "api" | "email" | "website" | "sms"
type Step = 1 | 2 | 3 | 4

interface Channel {
  type: ChannelType
  name: string
  description: string
  icon: React.ElementType
  available: boolean
}

const channels: Channel[] = [
  { type: "website", name: "Site", description: "Criar um widget de chat ao vivo", icon: Globe, available: false },
  { type: "whatsapp", name: "WhatsApp", description: "Atenda seus clientes no WhatsApp", icon: Phone, available: true },
  { type: "sms", name: "SMS", description: "Integrar o canal SMS com Twilio", icon: MessageSquare, available: false },
  { type: "email", name: "E-Mail", description: "Conectar com Gmail, Outlook", icon: Mail, available: false },
  { type: "api", name: "API", description: "Crie um canal usando nossa API", icon: Bot, available: true },
  { type: "telegram", name: "Telegram", description: "Configure usando o token do bot", icon: Send, available: true },
]

const steps = [
  { number: 1, title: "Escolha o Canal", description: "Escolha o provedor que voce deseja integrar" },
  { number: 2, title: "Criar Caixa de Entrada", description: "Configurar a caixa de entrada" },
  { number: 3, title: "Adicionar Agentes", description: "Adicionar agentes a caixa criada" },
  { number: 4, title: "Entao!", description: "Esta tudo pronto para comecar!" },
]

export default function NewInboxWizardPage() {
  const router = useRouter()
  const createMutation = useCreateInbox()
  
  const [currentStep, setCurrentStep] = useState<Step>(1)
  const [selectedChannel, setSelectedChannel] = useState<ChannelType | null>(null)
  
  // Form data
  const [name, setName] = useState("")
  const [greetingMessage, setGreetingMessage] = useState("")
  const [autoAssignment, setAutoAssignment] = useState(false)
  const [botToken, setBotToken] = useState("")
  const [webhookUrl, setWebhookUrl] = useState("")
  
  // Agents (mock for now)
  const [selectedAgents, setSelectedAgents] = useState<string[]>([])
  const mockAgents = [
    { id: "1", name: "Voce (Admin)", email: "admin@example.com" },
  ]

  const handleChannelSelect = (channel: ChannelType) => {
    setSelectedChannel(channel)
    setCurrentStep(2)
  }

  const handleStep2Next = () => {
    if (!name.trim()) return
    setCurrentStep(3)
  }

  const handleStep3Next = () => {
    setCurrentStep(4)
  }

  const handleFinish = async () => {
    if (!selectedChannel || !name.trim()) return

    try {
      await createMutation.mutateAsync({
        name: name.trim(),
        channel_type: selectedChannel as "whatsapp" | "telegram" | "api",
        greeting_message: greetingMessage.trim() || undefined,
        auto_assignment: autoAssignment,
      })
      router.push("/dashboard/inboxes")
    } catch (error) {
      console.error("Failed to create inbox:", error)
    }
  }

  const handleBack = () => {
    if (currentStep === 1) {
      router.push("/dashboard/inboxes")
    } else {
      setCurrentStep((prev) => (prev - 1) as Step)
    }
  }

  const toggleAgent = (agentId: string) => {
    setSelectedAgents(prev => 
      prev.includes(agentId) 
        ? prev.filter(id => id !== agentId)
        : [...prev, agentId]
    )
  }

  return (
    <SidebarProvider>
      <AppSidebar />
      <SidebarInset>
        <header className="flex h-16 shrink-0 items-center gap-2 border-b">
          <div className="flex items-center gap-2 px-4">
            <SidebarTrigger className="-ml-1" />
            <Separator orientation="vertical" className="mr-2 h-4" />
            <button 
              onClick={handleBack}
              className="flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground transition-colors"
            >
              <ChevronLeft className="h-4 w-4" />
              Anterior
            </button>
            <Separator orientation="vertical" className="mx-2 h-4" />
            <span className="font-semibold">Caixas de Entrada</span>
          </div>
        </header>

        <div className="flex flex-1">
          {/* Steps Sidebar */}
          <div className="w-64 border-r p-6 space-y-6">
            {steps.map((step) => (
              <div key={step.number} className="flex gap-3">
                <div className={cn(
                  "flex h-8 w-8 shrink-0 items-center justify-center rounded-full text-sm font-medium",
                  currentStep > step.number 
                    ? "bg-primary text-primary-foreground"
                    : currentStep === step.number
                    ? "bg-primary text-primary-foreground"
                    : "bg-muted text-muted-foreground"
                )}>
                  {currentStep > step.number ? (
                    <Check className="h-4 w-4" />
                  ) : (
                    step.number
                  )}
                </div>
                <div>
                  <p className={cn(
                    "font-medium text-sm",
                    currentStep >= step.number ? "text-foreground" : "text-muted-foreground"
                  )}>
                    {step.title}
                  </p>
                  <p className="text-xs text-muted-foreground mt-0.5">
                    {step.description}
                  </p>
                </div>
              </div>
            ))}
          </div>

          {/* Content */}
          <div className="flex-1 p-6 overflow-auto">
            {/* Step 1: Choose Channel */}
            {currentStep === 1 && (
              <div className="space-y-6">
                <div>
                  <h2 className="text-xl font-semibold">Selecione o tipo de canal</h2>
                  <p className="text-muted-foreground mt-1">
                    Escolha o provedor que voce deseja integrar com o sistema
                  </p>
                </div>

                <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
                  {channels.map((channel) => {
                    const Icon = channel.icon
                    return (
                      <button
                        key={channel.type}
                        onClick={() => channel.available && handleChannelSelect(channel.type)}
                        disabled={!channel.available}
                        className={cn(
                          "relative flex flex-col items-center gap-3 p-6 rounded-lg border-2 text-center transition-all",
                          channel.available
                            ? "hover:border-primary hover:bg-muted/50 cursor-pointer"
                            : "opacity-50 cursor-not-allowed"
                        )}
                      >
                        {!channel.available && (
                          <span className="absolute top-2 right-2 text-[10px] px-2 py-0.5 rounded bg-muted text-muted-foreground">
                            Em breve
                          </span>
                        )}
                        <div className="p-3 rounded-full bg-muted">
                          <Icon className="h-6 w-6" />
                        </div>
                        <div>
                          <p className="font-medium">{channel.name}</p>
                          <p className="text-xs text-muted-foreground mt-1">
                            {channel.description}
                          </p>
                        </div>
                      </button>
                    )
                  })}
                </div>
              </div>
            )}

            {/* Step 2: Configure Inbox */}
            {currentStep === 2 && selectedChannel && (
              <div className="space-y-6 max-w-xl">
                <div>
                  <h2 className="text-xl font-semibold">Configurar Caixa de Entrada</h2>
                  <p className="text-muted-foreground mt-1">
                    {selectedChannel === "whatsapp" && "Configure sua conexao WhatsApp"}
                    {selectedChannel === "telegram" && "Configure seu bot do Telegram"}
                    {selectedChannel === "api" && "Configure seu canal de API"}
                  </p>
                </div>

                {/* WhatsApp specific */}
                {selectedChannel === "whatsapp" && (
                  <div className="p-4 rounded-lg bg-green-50 dark:bg-green-950/30 border border-green-200 dark:border-green-800">
                    <h3 className="font-medium text-green-800 dark:text-green-200 mb-2">
                      Selecione seu provedor de API
                    </h3>
                    <p className="text-sm text-green-700 dark:text-green-300 mb-4">
                      Voce pode se conectar via Meta (Cloud API) ou via QR Code
                    </p>
                    <div className="grid gap-3 md:grid-cols-2">
                      <div className="p-4 rounded-lg border bg-background flex items-start gap-3">
                        <div className="p-2 rounded-full bg-muted">
                          <QrCode className="h-5 w-5" />
                        </div>
                        <div>
                          <p className="font-medium">QR Code</p>
                          <p className="text-xs text-muted-foreground">Conexao rapida via QR</p>
                        </div>
                      </div>
                      <div className="p-4 rounded-lg border bg-background flex items-start gap-3 opacity-50">
                        <div className="p-2 rounded-full bg-muted">
                          <Cloud className="h-5 w-5" />
                        </div>
                        <div>
                          <p className="font-medium">Cloud API</p>
                          <p className="text-xs text-muted-foreground">Em breve</p>
                        </div>
                      </div>
                    </div>
                  </div>
                )}

                {/* Telegram specific */}
                {selectedChannel === "telegram" && (
                  <div className="p-4 rounded-lg bg-blue-50 dark:bg-blue-950/30 border border-blue-200 dark:border-blue-800">
                    <h3 className="font-medium text-blue-800 dark:text-blue-200 mb-2">
                      Como obter o Bot Token
                    </h3>
                    <ol className="text-sm text-blue-700 dark:text-blue-300 space-y-1 list-decimal list-inside">
                      <li>Abra o Telegram e busque @BotFather</li>
                      <li>Envie /newbot e siga as instrucoes</li>
                      <li>Copie o token gerado</li>
                    </ol>
                    <a 
                      href="https://t.me/botfather" 
                      target="_blank"
                      className="inline-flex items-center gap-1 text-sm text-blue-700 dark:text-blue-300 hover:underline mt-3"
                    >
                      Abrir BotFather <ExternalLink className="h-3 w-3" />
                    </a>
                  </div>
                )}

                <div className="space-y-4">
                  <div className="space-y-2">
                    <Label htmlFor="name">Nome da Caixa de Entrada *</Label>
                    <Input
                      id="name"
                      value={name}
                      onChange={(e) => setName(e.target.value)}
                      placeholder="Ex: Atendimento Principal"
                    />
                  </div>

                  {selectedChannel === "telegram" && (
                    <div className="space-y-2">
                      <Label htmlFor="token">Bot Token *</Label>
                      <Input
                        id="token"
                        value={botToken}
                        onChange={(e) => setBotToken(e.target.value)}
                        placeholder="123456789:ABCdefGHIjklMNO"
                      />
                    </div>
                  )}

                  {selectedChannel === "api" && (
                    <div className="space-y-2">
                      <Label htmlFor="webhook">Webhook URL (opcional)</Label>
                      <Input
                        id="webhook"
                        value={webhookUrl}
                        onChange={(e) => setWebhookUrl(e.target.value)}
                        placeholder="https://seu-sistema.com/webhook"
                      />
                    </div>
                  )}

                  <div className="space-y-2">
                    <Label htmlFor="greeting">Mensagem de Boas-vindas</Label>
                    <Textarea
                      id="greeting"
                      value={greetingMessage}
                      onChange={(e) => setGreetingMessage(e.target.value)}
                      placeholder="Ola! Como posso ajudar?"
                      rows={3}
                    />
                  </div>

                  <div className="flex items-center justify-between p-4 rounded-lg border">
                    <div>
                      <Label>Atribuicao Automatica</Label>
                      <p className="text-xs text-muted-foreground">Distribuir conversas automaticamente</p>
                    </div>
                    <Switch checked={autoAssignment} onCheckedChange={setAutoAssignment} />
                  </div>
                </div>

                <div className="flex justify-end">
                  <Button onClick={handleStep2Next} disabled={!name.trim()}>
                    Proximo <ArrowRight className="ml-2 h-4 w-4" />
                  </Button>
                </div>
              </div>
            )}

            {/* Step 3: Add Agents */}
            {currentStep === 3 && (
              <div className="space-y-6 max-w-xl">
                <div>
                  <h2 className="text-xl font-semibold">Adicionar Agentes</h2>
                  <p className="text-muted-foreground mt-1">
                    Selecione os agentes que terao acesso a esta caixa de entrada
                  </p>
                </div>

                <div className="space-y-3">
                  {mockAgents.map((agent) => (
                    <label
                      key={agent.id}
                      className="flex items-center gap-3 p-4 rounded-lg border cursor-pointer hover:bg-muted/50"
                    >
                      <Checkbox
                        checked={selectedAgents.includes(agent.id)}
                        onCheckedChange={() => toggleAgent(agent.id)}
                      />
                      <div>
                        <p className="font-medium">{agent.name}</p>
                        <p className="text-sm text-muted-foreground">{agent.email}</p>
                      </div>
                    </label>
                  ))}
                </div>

                <div className="flex justify-end">
                  <Button onClick={handleStep3Next}>
                    Proximo <ArrowRight className="ml-2 h-4 w-4" />
                  </Button>
                </div>
              </div>
            )}

            {/* Step 4: Finish */}
            {currentStep === 4 && (
              <div className="space-y-6 max-w-xl">
                <div className="text-center py-8">
                  <div className="mx-auto w-16 h-16 rounded-full bg-green-100 dark:bg-green-900 flex items-center justify-center mb-4">
                    <Check className="h-8 w-8 text-green-600 dark:text-green-400" />
                  </div>
                  <h2 className="text-xl font-semibold">Tudo Pronto!</h2>
                  <p className="text-muted-foreground mt-2">
                    Sua caixa de entrada esta configurada e pronta para uso.
                  </p>
                </div>

                <div className="p-4 rounded-lg border space-y-3">
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Canal</span>
                    <span className="font-medium capitalize">{selectedChannel}</span>
                  </div>
                  <Separator />
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Nome</span>
                    <span className="font-medium">{name}</span>
                  </div>
                  {greetingMessage && (
                    <>
                      <Separator />
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">Mensagem</span>
                        <span className="font-medium truncate max-w-[200px]">{greetingMessage}</span>
                      </div>
                    </>
                  )}
                  <Separator />
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Auto-atribuicao</span>
                    <span className="font-medium">{autoAssignment ? "Sim" : "Nao"}</span>
                  </div>
                </div>

                <div className="flex justify-end">
                  <Button onClick={handleFinish} disabled={createMutation.isPending}>
                    {createMutation.isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                    Finalizar
                  </Button>
                </div>
              </div>
            )}
          </div>
        </div>
      </SidebarInset>
    </SidebarProvider>
  )
}
