"use client"

import { Phone, Send, Bot, Mail, Globe, MessageSquare } from "lucide-react"
import { cn } from "@/lib/utils"

export type ChannelType = "whatsapp" | "telegram" | "api" | "email" | "website" | "sms"

interface Channel {
  type: ChannelType
  name: string
  description: string
  icon: React.ElementType
  color: string
  bgColor: string
  available: boolean
}

const channels: Channel[] = [
  {
    type: "whatsapp",
    name: "WhatsApp",
    description: "Conecte via QR Code ou Cloud API",
    icon: Phone,
    color: "text-green-600",
    bgColor: "bg-green-100 hover:bg-green-200",
    available: true,
  },
  {
    type: "telegram",
    name: "Telegram",
    description: "Conecte usando Bot Token",
    icon: Send,
    color: "text-blue-600",
    bgColor: "bg-blue-100 hover:bg-blue-200",
    available: true,
  },
  {
    type: "api",
    name: "API Channel",
    description: "Integre via API REST",
    icon: Bot,
    color: "text-purple-600",
    bgColor: "bg-purple-100 hover:bg-purple-200",
    available: true,
  },
  {
    type: "email",
    name: "Email",
    description: "Gerencie emails de suporte",
    icon: Mail,
    color: "text-orange-600",
    bgColor: "bg-orange-100 hover:bg-orange-200",
    available: false,
  },
  {
    type: "website",
    name: "Website Chat",
    description: "Widget de chat para seu site",
    icon: Globe,
    color: "text-cyan-600",
    bgColor: "bg-cyan-100 hover:bg-cyan-200",
    available: false,
  },
  {
    type: "sms",
    name: "SMS",
    description: "Mensagens de texto via Twilio",
    icon: MessageSquare,
    color: "text-pink-600",
    bgColor: "bg-pink-100 hover:bg-pink-200",
    available: false,
  },
]

interface ChannelSelectorProps {
  onSelect: (channel: ChannelType) => void
}

export function ChannelSelector({ onSelect }: ChannelSelectorProps) {
  return (
    <div className="py-4">
      <p className="text-sm text-muted-foreground mb-6">
        Selecione o tipo de canal que deseja conectar ao seu workspace.
      </p>
      
      <div className="grid grid-cols-2 md:grid-cols-3 gap-3">
        {channels.map((channel) => {
          const Icon = channel.icon
          return (
            <button
              key={channel.type}
              onClick={() => channel.available && onSelect(channel.type)}
              disabled={!channel.available}
              className={cn(
                "relative flex flex-col items-center gap-3 p-4 rounded-lg border-2 border-transparent transition-all",
                channel.available
                  ? `${channel.bgColor} cursor-pointer hover:border-${channel.color.replace('text-', '')}`
                  : "bg-muted/50 cursor-not-allowed opacity-60"
              )}
            >
              {!channel.available && (
                <span className="absolute top-2 right-2 text-[10px] font-medium px-1.5 py-0.5 rounded bg-muted-foreground/20 text-muted-foreground">
                  Em breve
                </span>
              )}
              <div className={cn(
                "p-3 rounded-full",
                channel.available ? "bg-white/80" : "bg-white/50"
              )}>
                <Icon className={cn("h-6 w-6", channel.color)} />
              </div>
              <div className="text-center">
                <p className={cn(
                  "font-medium text-sm",
                  channel.available ? "text-foreground" : "text-muted-foreground"
                )}>
                  {channel.name}
                </p>
                <p className="text-xs text-muted-foreground mt-0.5 line-clamp-2">
                  {channel.description}
                </p>
              </div>
            </button>
          )
        })}
      </div>
    </div>
  )
}
