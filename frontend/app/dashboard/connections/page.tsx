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
  useConnections,
  useCreateConnection,
  useDeleteConnection,
  useDisconnectWhatsApp,
} from "@/lib/api/hooks"
import type { Connection as APIConnection } from "@/lib/api/types"

// Convert API connection to UI connection type
function mapAPIConnection(apiConn: APIConnection): Connection {
  return {
    id: apiConn.id,
    name: apiConn.name,
    platform: "whatsapp",
    status: apiConn.status as Connection["status"],
    phone: apiConn.phone,
    lastSync: apiConn.updated_at ? new Date(apiConn.updated_at) : undefined,
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

export default function ConnectionsPage() {
  const [isDialogOpen, setIsDialogOpen] = useState(false)
  const [newConnectionName, setNewConnectionName] = useState("")
  const [qrDialogOpen, setQrDialogOpen] = useState(false)
  const [selectedConnection, setSelectedConnection] = useState<Connection | null>(null)

  // API hooks
  const { data: apiConnections, isLoading, isError } = useConnections()
  const createMutation = useCreateConnection()
  const deleteMutation = useDeleteConnection()
  const disconnectMutation = useDisconnectWhatsApp()

  // Use API data or fallback to mock
  const connections = useMemo(() => {
    if (apiConnections && apiConnections.length > 0) {
      return apiConnections.map(mapAPIConnection)
    }
    if (isError || isLoading) {
      return mockConnections
    }
    return []
  }, [apiConnections, isLoading, isError])

  const connectedCount = connections.filter((c) => c.status === "connected").length
  const disconnectedCount = connections.filter((c) => c.status !== "connected").length

  const activeConnections = connections.filter((c) => c.status === "connected")
  const pendingConnections = connections.filter((c) => c.status !== "connected")

  const handleAddConnection = async () => {
    if (!newConnectionName.trim()) return

    try {
      await createMutation.mutateAsync({ name: newConnectionName })
      setNewConnectionName("")
      setIsDialogOpen(false)
    } catch (error) {
      console.error("Failed to create connection:", error)
    }
  }

  const handleDeleteConnection = async (id: string) => {
    try {
      await deleteMutation.mutateAsync(id)
    } catch (error) {
      console.error("Failed to delete connection:", error)
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
    setSelectedConnection(connection)
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
              <h1 className="text-2xl font-bold">Conexoes</h1>
              <p className="text-muted-foreground">Gerencie suas conexoes WhatsApp</p>
            </div>
            <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
              <DialogTrigger asChild>
                <Button>
                  <Plus className="mr-2 h-4 w-4" />
                  Nova Conexao
                </Button>
              </DialogTrigger>
              <DialogContent>
                <DialogHeader>
                  <DialogTitle>Nova Conexao WhatsApp</DialogTitle>
                  <DialogDescription>
                    Crie uma nova conexao para conectar seu WhatsApp.
                  </DialogDescription>
                </DialogHeader>
                <div className="grid gap-4 py-4">
                  <div className="grid gap-2">
                    <Label htmlFor="name">Nome da Conexao</Label>
                    <Input
                      id="name"
                      value={newConnectionName}
                      onChange={(e) => setNewConnectionName(e.target.value)}
                      placeholder="Ex: Atendimento Principal"
                      onKeyDown={(e) => e.key === "Enter" && handleAddConnection()}
                    />
                  </div>
                </div>
                <DialogFooter>
                  <Button variant="outline" onClick={() => setIsDialogOpen(false)}>
                    Cancelar
                  </Button>
                  <Button 
                    onClick={handleAddConnection} 
                    disabled={createMutation.isPending || !newConnectionName.trim()}
                  >
                    {createMutation.isPending && (
                      <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    )}
                    Criar Conexao
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
                    onDelete={() => handleDeleteConnection(connection.id)}
                    isLoading={deleteMutation.isPending || disconnectMutation.isPending}
                  />
                ))}
                {connections.length === 0 && !isLoading && (
                  <p className="col-span-full text-center text-muted-foreground py-8">
                    Nenhuma conexao. Clique em "Nova Conexao" para comecar.
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
                    onDelete={() => handleDeleteConnection(connection.id)}
                    isLoading={deleteMutation.isPending || disconnectMutation.isPending}
                  />
                ))}
                {activeConnections.length === 0 && (
                  <p className="col-span-full text-center text-muted-foreground py-8">
                    Nenhuma conexao ativa.
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
                    onDelete={() => handleDeleteConnection(connection.id)}
                    isLoading={deleteMutation.isPending}
                  />
                ))}
                {pendingConnections.length === 0 && (
                  <p className="col-span-full text-center text-muted-foreground py-8">
                    Nenhuma conexao pendente.
                  </p>
                )}
              </div>
            </TabsContent>
          </Tabs>
        </div>
      </SidebarInset>

      {/* QR Code Dialog */}
      <QRCodeDialog
        connectionId={selectedConnection?.id || null}
        connectionName={selectedConnection?.name || ""}
        open={qrDialogOpen}
        onOpenChange={setQrDialogOpen}
      />
    </SidebarProvider>
  )
}
