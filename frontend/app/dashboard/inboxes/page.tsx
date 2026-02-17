"use client"

import { useState, useMemo } from "react"
import { AppSidebar } from "@/components/layout/app-sidebar"
import { InboxCard } from "@/components/inbox/inbox-card"
import { QRCodeDialog } from "@/components/inbox/qr-code-dialog"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbList,
  BreadcrumbPage,
} from "@/components/ui/breadcrumb"
import { Separator } from "@/components/ui/separator"
import {
  SidebarInset,
  SidebarProvider,
  SidebarTrigger,
} from "@/components/ui/sidebar"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { Label } from "@/components/ui/label"
import { Input } from "@/components/ui/input"
import { Textarea } from "@/components/ui/textarea"
import { Switch } from "@/components/ui/switch"
import { Plus, Inbox, Wifi, WifiOff, Loader2, Phone, Send, Bot } from "lucide-react"
import {
  useInboxes,
  useCreateInbox,
  useDeleteInbox,
  useDisconnectInbox,
} from "@/lib/api/hooks"
import type { Inbox as InboxType, CreateInboxRequest } from "@/lib/api/types"

export default function InboxesPage() {
  const [isDialogOpen, setIsDialogOpen] = useState(false)
  const [qrDialogOpen, setQrDialogOpen] = useState(false)
  const [selectedInbox, setSelectedInbox] = useState<InboxType | null>(null)
  
  // Form state
  const [formData, setFormData] = useState<CreateInboxRequest>({
    name: "",
    channel_type: "whatsapp",
    greeting_message: "",
    auto_assignment: false,
  })

  // API hooks
  const { data: inboxes, isLoading, isError } = useInboxes()
  const createMutation = useCreateInbox()
  const deleteMutation = useDeleteInbox()
  const disconnectMutation = useDisconnectInbox()

  const connectedCount = inboxes?.filter((i) => i.status === "connected").length || 0
  const disconnectedCount = inboxes?.filter((i) => i.status !== "connected").length || 0

  const activeInboxes = inboxes?.filter((i) => i.status === "connected") || []
  const pendingInboxes = inboxes?.filter((i) => i.status !== "connected") || []

  const resetForm = () => {
    setFormData({
      name: "",
      channel_type: "whatsapp",
      greeting_message: "",
      auto_assignment: false,
    })
  }

  const handleCreateInbox = async () => {
    if (!formData.name.trim()) return

    try {
      await createMutation.mutateAsync(formData)
      resetForm()
      setIsDialogOpen(false)
    } catch (error) {
      console.error("Failed to create inbox:", error)
    }
  }

  const handleDeleteInbox = async (id: string) => {
    try {
      await deleteMutation.mutateAsync(id)
    } catch (error) {
      console.error("Failed to delete inbox:", error)
    }
  }

  const handleDisconnect = async (id: string) => {
    try {
      await disconnectMutation.mutateAsync(id)
    } catch (error) {
      console.error("Failed to disconnect:", error)
    }
  }

  const handleShowQR = (inbox: InboxType) => {
    setSelectedInbox(inbox)
    setQrDialogOpen(true)
  }

  const channelIcons = {
    whatsapp: <Phone className="h-4 w-4 text-green-500" />,
    telegram: <Send className="h-4 w-4 text-blue-500" />,
    api: <Bot className="h-4 w-4 text-gray-500" />,
  }

  return (
    <SidebarProvider>
      <AppSidebar />
      <SidebarInset>
        <header className="flex h-16 shrink-0 items-center gap-2 transition-[width,height] ease-linear group-has-data-[collapsible=icon]/sidebar-wrapper:h-12">
          <div className="flex items-center gap-2 px-4">
            <SidebarTrigger className="-ml-1" />
            <Separator orientation="vertical" className="mr-2 data-[orientation=vertical]:h-4" />
            <Breadcrumb>
              <BreadcrumbList>
                <BreadcrumbItem>
                  <BreadcrumbPage>Inboxes</BreadcrumbPage>
                </BreadcrumbItem>
              </BreadcrumbList>
            </Breadcrumb>
          </div>
        </header>

        <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-2xl font-bold">Inboxes</h1>
              <p className="text-muted-foreground">Gerencie seus canais de atendimento</p>
            </div>
            <Dialog open={isDialogOpen} onOpenChange={(open) => {
              setIsDialogOpen(open)
              if (!open) resetForm()
            }}>
              <DialogTrigger asChild>
                <Button>
                  <Plus className="mr-2 h-4 w-4" />
                  Novo Inbox
                </Button>
              </DialogTrigger>
              <DialogContent className="sm:max-w-[480px]">
                <DialogHeader>
                  <DialogTitle>Criar Novo Inbox</DialogTitle>
                  <DialogDescription>
                    Configure um novo canal de atendimento para sua equipe.
                  </DialogDescription>
                </DialogHeader>
                <div className="grid gap-4 py-4">
                  <div className="grid gap-2">
                    <Label htmlFor="channel_type">Tipo de Canal</Label>
                    <Select
                      value={formData.channel_type}
                      onValueChange={(value: "whatsapp" | "telegram" | "api") => 
                        setFormData(prev => ({ ...prev, channel_type: value }))
                      }
                    >
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="whatsapp">
                          <div className="flex items-center gap-2">
                            <Phone className="h-4 w-4 text-green-500" />
                            WhatsApp
                          </div>
                        </SelectItem>
                        <SelectItem value="telegram">
                          <div className="flex items-center gap-2">
                            <Send className="h-4 w-4 text-blue-500" />
                            Telegram
                          </div>
                        </SelectItem>
                        <SelectItem value="api">
                          <div className="flex items-center gap-2">
                            <Bot className="h-4 w-4 text-gray-500" />
                            API
                          </div>
                        </SelectItem>
                      </SelectContent>
                    </Select>
                  </div>

                  <div className="grid gap-2">
                    <Label htmlFor="name">Nome do Inbox</Label>
                    <Input
                      id="name"
                      value={formData.name}
                      onChange={(e) => setFormData(prev => ({ ...prev, name: e.target.value }))}
                      placeholder="Ex: Atendimento Principal"
                    />
                  </div>

                  <div className="grid gap-2">
                    <Label htmlFor="greeting">Mensagem de Boas-vindas (opcional)</Label>
                    <Textarea
                      id="greeting"
                      value={formData.greeting_message}
                      onChange={(e) => setFormData(prev => ({ ...prev, greeting_message: e.target.value }))}
                      placeholder="Ola! Seja bem-vindo ao nosso atendimento..."
                      rows={3}
                    />
                    <p className="text-xs text-muted-foreground">
                      Esta mensagem sera enviada automaticamente quando um novo contato iniciar uma conversa.
                    </p>
                  </div>

                  <div className="flex items-center justify-between rounded-lg border p-3">
                    <div className="space-y-0.5">
                      <Label htmlFor="auto_assignment">Atribuicao automatica</Label>
                      <p className="text-xs text-muted-foreground">
                        Distribui conversas automaticamente entre os agentes disponiveis
                      </p>
                    </div>
                    <Switch
                      id="auto_assignment"
                      checked={formData.auto_assignment}
                      onCheckedChange={(checked) => 
                        setFormData(prev => ({ ...prev, auto_assignment: checked }))
                      }
                    />
                  </div>
                </div>
                <DialogFooter>
                  <Button variant="outline" onClick={() => setIsDialogOpen(false)}>
                    Cancelar
                  </Button>
                  <Button 
                    onClick={handleCreateInbox} 
                    disabled={createMutation.isPending || !formData.name.trim()}
                  >
                    {createMutation.isPending && (
                      <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    )}
                    Criar Inbox
                  </Button>
                </DialogFooter>
              </DialogContent>
            </Dialog>
          </div>

          {/* Stats Cards */}
          <div className="grid gap-4 md:grid-cols-3">
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Total de Inboxes</CardTitle>
                <Inbox className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">
                  {isLoading ? <Loader2 className="h-6 w-6 animate-spin" /> : inboxes?.length || 0}
                </div>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Conectados</CardTitle>
                <Wifi className="h-4 w-4 text-green-500" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{connectedCount}</div>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Desconectados</CardTitle>
                <WifiOff className="h-4 w-4 text-red-500" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{disconnectedCount}</div>
              </CardContent>
            </Card>
          </div>

          {/* Tabs */}
          <Tabs defaultValue="all" className="flex-1">
            <TabsList>
              <TabsTrigger value="all">Todos ({inboxes?.length || 0})</TabsTrigger>
              <TabsTrigger value="active">Ativos ({activeInboxes.length})</TabsTrigger>
              <TabsTrigger value="pending">Pendentes ({pendingInboxes.length})</TabsTrigger>
            </TabsList>
            
            <TabsContent value="all" className="mt-4">
              <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
                {inboxes?.map((inbox) => (
                  <InboxCard
                    key={inbox.id}
                    inbox={inbox}
                    onShowQR={() => handleShowQR(inbox)}
                    onDisconnect={() => handleDisconnect(inbox.id)}
                    onDelete={() => handleDeleteInbox(inbox.id)}
                    isLoading={deleteMutation.isPending || disconnectMutation.isPending}
                  />
                ))}
                {(!inboxes || inboxes.length === 0) && !isLoading && (
                  <div className="col-span-full flex flex-col items-center justify-center py-12 text-muted-foreground">
                    <Inbox className="h-12 w-12 mb-4" />
                    <p className="text-lg font-medium">Nenhum inbox configurado</p>
                    <p className="text-sm">Clique em "Novo Inbox" para comecar</p>
                  </div>
                )}
                {isLoading && (
                  <div className="col-span-full flex justify-center py-8">
                    <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
                  </div>
                )}
              </div>
            </TabsContent>

            <TabsContent value="active" className="mt-4">
              <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
                {activeInboxes.map((inbox) => (
                  <InboxCard
                    key={inbox.id}
                    inbox={inbox}
                    onDisconnect={() => handleDisconnect(inbox.id)}
                    onDelete={() => handleDeleteInbox(inbox.id)}
                    isLoading={deleteMutation.isPending || disconnectMutation.isPending}
                  />
                ))}
                {activeInboxes.length === 0 && (
                  <p className="col-span-full text-center text-muted-foreground py-8">
                    Nenhum inbox ativo no momento.
                  </p>
                )}
              </div>
            </TabsContent>

            <TabsContent value="pending" className="mt-4">
              <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
                {pendingInboxes.map((inbox) => (
                  <InboxCard
                    key={inbox.id}
                    inbox={inbox}
                    onShowQR={() => handleShowQR(inbox)}
                    onDelete={() => handleDeleteInbox(inbox.id)}
                    isLoading={deleteMutation.isPending}
                  />
                ))}
                {pendingInboxes.length === 0 && (
                  <p className="col-span-full text-center text-muted-foreground py-8">
                    Nenhum inbox pendente de conexao.
                  </p>
                )}
              </div>
            </TabsContent>
          </Tabs>
        </div>
      </SidebarInset>

      {/* QR Code Dialog */}
      <QRCodeDialog
        inboxId={selectedInbox?.id || null}
        inboxName={selectedInbox?.name || ""}
        open={qrDialogOpen}
        onOpenChange={setQrDialogOpen}
      />
    </SidebarProvider>
  )
}
