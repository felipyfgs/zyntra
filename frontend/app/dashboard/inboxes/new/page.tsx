"use client"

import Link from "next/link"
import { AppSidebar } from "@/components/layout/app-sidebar"
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
import { Phone, Send, Bot, Mail, Globe, MessageSquare } from "lucide-react"
import { cn } from "@/lib/utils"

type ChannelType = "whatsapp" | "telegram" | "api" | "email" | "website" | "sms"

interface Channel {
  type: ChannelType
  name: string
  description: string
  icon: React.ElementType
  color: string
  bgColor: string
  hoverColor: string
  available: boolean
  href: string
}

const channels: Channel[] = [
  {
    type: "whatsapp",
    name: "WhatsApp",
    description: "Conecte seu WhatsApp Business via QR Code ou Cloud API para atender clientes",
    icon: Phone,
    color: "text-green-600",
    bgColor: "bg-green-50",
    hoverColor: "hover:bg-green-100 hover:border-green-300",
    available: true,
    href: "/dashboard/inboxes/new/whatsapp",
  },
  {
    type: "telegram",
    name: "Telegram",
    description: "Conecte um bot do Telegram para receber e responder mensagens",
    icon: Send,
    color: "text-blue-600",
    bgColor: "bg-blue-50",
    hoverColor: "hover:bg-blue-100 hover:border-blue-300",
    available: true,
    href: "/dashboard/inboxes/new/telegram",
  },
  {
    type: "api",
    name: "API Channel",
    description: "Integre sistemas externos via API REST para enviar e receber mensagens",
    icon: Bot,
    color: "text-purple-600",
    bgColor: "bg-purple-50",
    hoverColor: "hover:bg-purple-100 hover:border-purple-300",
    available: true,
    href: "/dashboard/inboxes/new/api",
  },
  {
    type: "email",
    name: "Email",
    description: "Gerencie emails de suporte com IMAP/SMTP ou integracao com Gmail",
    icon: Mail,
    color: "text-orange-600",
    bgColor: "bg-orange-50",
    hoverColor: "hover:bg-orange-100 hover:border-orange-300",
    available: false,
    href: "#",
  },
  {
    type: "website",
    name: "Website Chat",
    description: "Adicione um widget de chat ao vivo no seu website",
    icon: Globe,
    color: "text-cyan-600",
    bgColor: "bg-cyan-50",
    hoverColor: "hover:bg-cyan-100 hover:border-cyan-300",
    available: false,
    href: "#",
  },
  {
    type: "sms",
    name: "SMS",
    description: "Envie e receba SMS via Twilio ou outros provedores",
    icon: MessageSquare,
    color: "text-pink-600",
    bgColor: "bg-pink-50",
    hoverColor: "hover:bg-pink-100 hover:border-pink-300",
    available: false,
    href: "#",
  },
]

export default function NewInboxPage() {
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
                  <BreadcrumbPage>Novo Inbox</BreadcrumbPage>
                </BreadcrumbItem>
              </BreadcrumbList>
            </Breadcrumb>
          </div>
        </header>

        <div className="flex flex-1 flex-col gap-6 p-4 pt-0">
          <div>
            <h1 className="text-2xl font-bold">Criar Novo Inbox</h1>
            <p className="text-muted-foreground mt-1">
              Selecione o tipo de canal que deseja conectar ao seu workspace
            </p>
          </div>

          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            {channels.map((channel) => {
              const Icon = channel.icon
              const Component = channel.available ? Link : "div"
              
              return (
                <Component
                  key={channel.type}
                  href={channel.available ? channel.href : "#"}
                  className={cn(
                    "relative flex flex-col gap-4 p-5 rounded-xl border-2 transition-all",
                    channel.available
                      ? `${channel.bgColor} ${channel.hoverColor} cursor-pointer`
                      : "bg-muted/30 cursor-not-allowed opacity-60 border-transparent"
                  )}
                >
                  {!channel.available && (
                    <span className="absolute top-3 right-3 text-[10px] font-semibold px-2 py-1 rounded-full bg-muted text-muted-foreground">
                      Em breve
                    </span>
                  )}
                  
                  <div className={cn(
                    "w-12 h-12 rounded-xl flex items-center justify-center",
                    channel.available ? "bg-white shadow-sm" : "bg-white/50"
                  )}>
                    <Icon className={cn("h-6 w-6", channel.color)} />
                  </div>
                  
                  <div>
                    <h3 className={cn(
                      "font-semibold text-lg",
                      channel.available ? "text-foreground" : "text-muted-foreground"
                    )}>
                      {channel.name}
                    </h3>
                    <p className="text-sm text-muted-foreground mt-1 leading-relaxed">
                      {channel.description}
                    </p>
                  </div>
                </Component>
              )
            })}
          </div>
        </div>
      </SidebarInset>
    </SidebarProvider>
  )
}
