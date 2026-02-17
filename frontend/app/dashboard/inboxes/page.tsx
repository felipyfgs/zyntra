"use client"

import { useState, useEffect } from "react"
import Link from "next/link"
import { AppSidebar } from "@/components/layout/app-sidebar"
import { InboxCard } from "@/components/inbox/inbox-card"
import { InboxListItem } from "@/components/inbox/inbox-list-item"
import { QRCodeDialog } from "@/components/inbox/qr-code-dialog"
import { Button } from "@/components/ui/button"
import { ToggleGroup, ToggleGroupItem } from "@/components/ui/toggle-group"
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
import { Plus, Inbox, Loader2, LayoutGrid, List } from "lucide-react"
import {
  useInboxes,
  useDeleteInbox,
  useDisconnectInbox,
} from "@/lib/api/hooks"
import type { Inbox as InboxType } from "@/lib/api/types"

type ViewMode = "grid" | "list"

export default function InboxesPage() {
  const [qrDialogOpen, setQrDialogOpen] = useState(false)
  const [selectedInbox, setSelectedInbox] = useState<InboxType | null>(null)
  const [viewMode, setViewMode] = useState<ViewMode>("grid")

  // Load view preference from localStorage
  useEffect(() => {
    const saved = localStorage.getItem("inboxes-view-mode")
    if (saved === "grid" || saved === "list") {
      setViewMode(saved)
    }
  }, [])

  const handleViewChange = (value: string) => {
    if (value === "grid" || value === "list") {
      setViewMode(value)
      localStorage.setItem("inboxes-view-mode", value)
    }
  }

  const { data: inboxes, isLoading } = useInboxes()
  const deleteMutation = useDeleteInbox()
  const disconnectMutation = useDisconnectInbox()

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
        <header className="flex h-16 shrink-0 items-center gap-2">
          <div className="flex items-center gap-2 px-4">
            <SidebarTrigger className="-ml-1" />
            <Separator orientation="vertical" className="mr-2 h-4" />
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
          {/* Header */}
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-2xl font-bold">Inboxes</h1>
              <p className="text-muted-foreground">Gerencie seus canais de atendimento</p>
            </div>
            <div className="flex items-center gap-2">
              <ToggleGroup type="single" value={viewMode} onValueChange={handleViewChange}>
                <ToggleGroupItem value="grid" aria-label="Grid view">
                  <LayoutGrid className="h-4 w-4" />
                </ToggleGroupItem>
                <ToggleGroupItem value="list" aria-label="List view">
                  <List className="h-4 w-4" />
                </ToggleGroupItem>
              </ToggleGroup>
              
              <Button asChild>
                <Link href="/dashboard/inboxes/new">
                  <Plus className="mr-2 h-4 w-4" />
                  Novo Inbox
                </Link>
              </Button>
            </div>
          </div>

          {/* Content */}
          {isLoading ? (
            <div className="flex justify-center py-12">
              <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
            </div>
          ) : !inboxes || inboxes.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-16 text-muted-foreground">
              <Inbox className="h-12 w-12 mb-4" />
              <p className="text-lg font-medium">Nenhum inbox configurado</p>
              <p className="text-sm mb-4">Conecte um canal para comecar a receber mensagens</p>
              <Button asChild>
                <Link href="/dashboard/inboxes/new">
                  <Plus className="mr-2 h-4 w-4" />
                  Criar primeiro inbox
                </Link>
              </Button>
            </div>
          ) : viewMode === "grid" ? (
            <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
              {inboxes.map((inbox) => (
                <InboxCard
                  key={inbox.id}
                  inbox={inbox}
                  onShowQR={() => handleShowQR(inbox)}
                  onDisconnect={() => handleDisconnect(inbox.id)}
                  onDelete={() => handleDeleteInbox(inbox.id)}
                  isLoading={deleteMutation.isPending || disconnectMutation.isPending}
                />
              ))}
            </div>
          ) : (
            <div className="space-y-2">
              {inboxes.map((inbox) => (
                <InboxListItem
                  key={inbox.id}
                  inbox={inbox}
                  onShowQR={() => handleShowQR(inbox)}
                  onDisconnect={() => handleDisconnect(inbox.id)}
                  onDelete={() => handleDeleteInbox(inbox.id)}
                  isLoading={deleteMutation.isPending || disconnectMutation.isPending}
                />
              ))}
            </div>
          )}
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
