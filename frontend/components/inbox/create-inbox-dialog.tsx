"use client"

import { useState } from "react"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { ArrowLeft } from "lucide-react"
import { ChannelSelector, type ChannelType } from "./channel-selector"
import { WhatsAppSetup } from "./whatsapp-setup"
import { TelegramSetup } from "./telegram-setup"
import { APISetup } from "./api-setup"
import { useCreateInbox } from "@/lib/api/hooks"
import type { CreateInboxRequest } from "@/lib/api/types"

interface CreateInboxDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onSuccess?: () => void
}

type Step = "select-channel" | "configure"

export function CreateInboxDialog({ open, onOpenChange, onSuccess }: CreateInboxDialogProps) {
  const [step, setStep] = useState<Step>("select-channel")
  const [selectedChannel, setSelectedChannel] = useState<ChannelType | null>(null)
  const createMutation = useCreateInbox()

  const handleChannelSelect = (channel: ChannelType) => {
    setSelectedChannel(channel)
    setStep("configure")
  }

  const handleBack = () => {
    setStep("select-channel")
    setSelectedChannel(null)
  }

  const handleClose = () => {
    onOpenChange(false)
    setTimeout(() => {
      setStep("select-channel")
      setSelectedChannel(null)
    }, 200)
  }

  const handleCreate = async (data: CreateInboxRequest) => {
    try {
      await createMutation.mutateAsync(data)
      handleClose()
      onSuccess?.()
    } catch (error) {
      console.error("Failed to create inbox:", error)
      throw error
    }
  }

  const getTitle = () => {
    if (step === "select-channel") return "Criar Novo Inbox"
    switch (selectedChannel) {
      case "whatsapp": return "Configurar WhatsApp"
      case "telegram": return "Configurar Telegram"
      case "api": return "Configurar Canal API"
      default: return "Configurar Canal"
    }
  }

  return (
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogContent className="sm:max-w-[600px]">
        <DialogHeader>
          <div className="flex items-center gap-2">
            {step === "configure" && (
              <Button variant="ghost" size="icon" onClick={handleBack} className="h-8 w-8">
                <ArrowLeft className="h-4 w-4" />
              </Button>
            )}
            <DialogTitle>{getTitle()}</DialogTitle>
          </div>
        </DialogHeader>

        {step === "select-channel" && (
          <ChannelSelector onSelect={handleChannelSelect} />
        )}

        {step === "configure" && selectedChannel === "whatsapp" && (
          <WhatsAppSetup 
            onSubmit={handleCreate} 
            onCancel={handleBack}
            isLoading={createMutation.isPending}
          />
        )}

        {step === "configure" && selectedChannel === "telegram" && (
          <TelegramSetup 
            onSubmit={handleCreate} 
            onCancel={handleBack}
            isLoading={createMutation.isPending}
          />
        )}

        {step === "configure" && selectedChannel === "api" && (
          <APISetup 
            onSubmit={handleCreate} 
            onCancel={handleBack}
            isLoading={createMutation.isPending}
          />
        )}
      </DialogContent>
    </Dialog>
  )
}
