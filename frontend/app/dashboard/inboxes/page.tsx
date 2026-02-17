"use client"

import { useState } from "react"
import { AppSidebar } from "@/components/layout/app-sidebar"
import { InboxCard } from "@/components/inbox/inbox-card"
import { QRCodeDialog } from "@/components/inbox/qr-code-dialog"
import { CreateInboxDialog } from "@/components/inbox/create-inbox-dialog"
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
import { Plus, Inbox, Wifi, WifiOff, Loader2 } from "lucide-react"
import {
  useInboxes,
  useDeleteInbox,
  useDisconnectInbox,
} from "@/lib/api/hooks"
import type { Inbox as InboxType } from "@/lib/api/types"

export default function InboxesPage() {
  const [createDialogOpen, setCreateDialogOpen] = useState(false)
  const [qrDialogOpen, setQrDialogOpen] = useState(false)
  const [selectedInbox, setSelectedInbox] = useState<InboxType | null>(null)

  // API hooks
  const { data: inboxes, isLoading, refetch } = useInboxes()
  const deleteMutation = useDeleteInbox()
  const disconnectMutation = useDisconnectInbox()

  const connectedCount = inboxes?.filter((i) => i.status === "connected").length || 0
  const disconnectedCount = inboxes?.filter((i) => i.status !== "connected").length || 0

  const activeInboxes = inboxes?.filter((i) => i.status === "connected") || []
  const pendingInboxes = inboxes?.filter((i) => i.status !== "connected") || []

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
            <Button onClick={() => setCreateDialogOpen(true)}>
              <Plus className="mr-2 h-4 w-4" />
              Novo Inbox
            </Button>
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
                    <p className="text-sm mb-4">Conecte um canal para comecar a receber mensagens</p>
                    <Button onClick={() => setCreateDialogOpen(true)}>
                      <Plus className="mr-2 h-4 w-4" />
                      Criar primeiro inbox
                    </Button>
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

      {/* Create Inbox Dialog */}
      <CreateInboxDialog
        open={createDialogOpen}
        onOpenChange={setCreateDialogOpen}
        onSuccess={() => refetch()}
      />

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
