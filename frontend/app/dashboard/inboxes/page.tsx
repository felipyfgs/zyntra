"use client"

import { useState, useMemo } from "react"
import { AppSidebar } from "@/components/layout/app-sidebar"
import { ConnectionCard, type Connection } from "@/components/connection/connection-card"
import { QRCodeDialog } from "@/components/connection/qr-code-dialog"
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
import { Label } from "@/components/ui/label"
import { Input } from "@/components/ui/input"
import { Plus, Link2, Wifi, WifiOff, Loader2 } from "lucide-react"
import {
  useInboxes,
  useCreateInbox,
  useDeleteInbox,
  useDisconnectInbox,
} from "@/lib/api/hooks"
import type { Inbox as APIInbox } from "@/lib/api/types"

// Convert API inbox to UI connection type
function mapAPIInbox(inbox: APIInbox): Connection {
  return {
    id: inbox.id,
    name: inbox.name,
    platform: inbox.channel_type || "whatsapp",
    status: inbox.status as Connection["status"],
    phone: inbox.phone,
    lastSync: inbox.updated_at ? new Date(inbox.updated_at) : undefined,
  }
}

// Mock data for development fallback
const mockConnections: Connection[] = [
  {
    id: "mock-1",
    name: "Conta Principal",
    platform: "whatsapp",
    status: "disconnected",
  },
]

export default function InboxesPage() {
  const [isDialogOpen, setIsDialogOpen] = useState(false)
  const [newInboxName, setNewInboxName] = useState("")
  const [qrDialogOpen, setQrDialogOpen] = useState(false)
  const [selectedInbox, setSelectedInbox] = useState<Connection | null>(null)

  // API hooks
  const { data: apiInboxes, isLoading, isError } = useInboxes()
  const createMutation = useCreateInbox()
  const deleteMutation = useDeleteInbox()
  const disconnectMutation = useDisconnectInbox()

  // Use API data or fallback to mock
  const connections = useMemo(() => {
    if (apiInboxes && apiInboxes.length > 0) {
      return apiInboxes.map(mapAPIInbox)
    }
    if (isError || isLoading) {
      return mockConnections
    }
    return []
  }, [apiInboxes, isLoading, isError])

  const connectedCount = connections.filter((c) => c.status === "connected").length
  const disconnectedCount = connections.filter((c) => c.status !== "connected").length

  const activeConnections = connections.filter((c) => c.status === "connected")
  const pendingConnections = connections.filter((c) => c.status !== "connected")

  const handleAddInbox = async () => {
    if (!newInboxName.trim()) return

    try {
      await createMutation.mutateAsync({ name: newInboxName, channel_type: "whatsapp" })
      setNewInboxName("")
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

  const handleShowQR = (connection: Connection) => {
    setSelectedInbox(connection)
    setQrDialogOpen(true)
  }

  return (
    <SidebarProvider>
      <AppSidebar />
      <SidebarInset>
        <header className="flex h-16 shrink-0 items-center gap-2 transition-[width,height] ease-linear group-has-data-[collapsible=icon]/sidebar-wrapper:h-12">
          <div className="flex items-center gap-2 px-4">
            <SidebarTrigger className="-ml-1" />
            <Separator
              orientation="vertical"
              className="mr-2 data-[orientation=vertical]:h-4"
            />
            <Breadcrumb>
              <BreadcrumbList>
                <BreadcrumbItem>
                  <BreadcrumbPage>Conexoes</BreadcrumbPage>
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
            <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
              <DialogTrigger asChild>
                <Button>
                  <Plus className="mr-2 h-4 w-4" />
                  Novo Inbox
                </Button>
              </DialogTrigger>
              <DialogContent>
                <DialogHeader>
                  <DialogTitle>Novo Inbox WhatsApp</DialogTitle>
                  <DialogDescription>
                    Crie um novo inbox para conectar seu WhatsApp.
                  </DialogDescription>
                </DialogHeader>
                <div className="grid gap-4 py-4">
                  <div className="grid gap-2">
                    <Label htmlFor="name">Nome do Inbox</Label>
                    <Input
                      id="name"
                      value={newInboxName}
                      onChange={(e) => setNewInboxName(e.target.value)}
                      placeholder="Ex: Atendimento Principal"
                      onKeyDown={(e) => e.key === "Enter" && handleAddInbox()}
                    />
                  </div>
                </div>
                <DialogFooter>
                  <Button variant="outline" onClick={() => setIsDialogOpen(false)}>
                    Cancelar
                  </Button>
                  <Button 
                    onClick={handleAddInbox} 
                    disabled={createMutation.isPending || !newInboxName.trim()}
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
                <CardTitle className="text-sm font-medium">Total</CardTitle>
                <Link2 className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">
                  {isLoading ? <Loader2 className="h-6 w-6 animate-spin" /> : connections.length}
                </div>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Conectadas</CardTitle>
                <Wifi className="h-4 w-4 text-green-500" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{connectedCount}</div>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Desconectadas</CardTitle>
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
              <TabsTrigger value="all">Todas ({connections.length})</TabsTrigger>
              <TabsTrigger value="active">Ativas ({activeConnections.length})</TabsTrigger>
              <TabsTrigger value="pending">Pendentes ({pendingConnections.length})</TabsTrigger>
            </TabsList>
            <TabsContent value="all" className="mt-4">
              <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
                {connections.map((connection) => (
                  <ConnectionCard
                    key={connection.id}
                    connection={connection}
                    onShowQR={() => handleShowQR(connection)}
                    onDisconnect={() => handleDisconnect(connection.id)}
                    onDelete={() => handleDeleteInbox(connection.id)}
                    isLoading={deleteMutation.isPending || disconnectMutation.isPending}
                  />
                ))}
                {connections.length === 0 && !isLoading && (
                  <p className="col-span-full text-center text-muted-foreground py-8">
                    Nenhum inbox. Clique em "Novo Inbox" para comecar.
                  </p>
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
                {activeConnections.map((connection) => (
                  <ConnectionCard
                    key={connection.id}
                    connection={connection}
                    onDisconnect={() => handleDisconnect(connection.id)}
                    onDelete={() => handleDeleteInbox(connection.id)}
                    isLoading={deleteMutation.isPending || disconnectMutation.isPending}
                  />
                ))}
                {activeConnections.length === 0 && (
                  <p className="col-span-full text-center text-muted-foreground py-8">
                    Nenhum inbox ativo.
                  </p>
                )}
              </div>
            </TabsContent>
            <TabsContent value="pending" className="mt-4">
              <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
                {pendingConnections.map((connection) => (
                  <ConnectionCard
                    key={connection.id}
                    connection={connection}
                    onShowQR={() => handleShowQR(connection)}
                    onDelete={() => handleDeleteInbox(connection.id)}
                    isLoading={deleteMutation.isPending}
                  />
                ))}
                {pendingConnections.length === 0 && (
                  <p className="col-span-full text-center text-muted-foreground py-8">
                    Nenhum inbox pendente.
                  </p>
                )}
              </div>
            </TabsContent>
          </Tabs>
        </div>
      </SidebarInset>

      {/* QR Code Dialog */}
      <QRCodeDialog
        connectionId={selectedInbox?.id || null}
        connectionName={selectedInbox?.name || ""}
        open={qrDialogOpen}
        onOpenChange={setQrDialogOpen}
      />
    </SidebarProvider>
  )
}
