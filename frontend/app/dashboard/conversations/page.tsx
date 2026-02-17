"use client"

import { useState, useMemo, useCallback, useEffect } from "react"
import { AppSidebar } from "@/components/layout/app-sidebar"
import { Button } from "@/components/ui/button"
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover"
import {
  SidebarInset,
  SidebarProvider,
  SidebarTrigger,
} from "@/components/ui/sidebar"
import { Input } from "@/components/ui/input"
import { ChatList, type Chat } from "@/components/chat/chat-list"
import { ChatArea } from "@/components/chat/chat-area"
import { ChatFilters } from "@/components/chat/chat-filters"
import {
  AdvancedFilters,
  type FilterRule,
  type FilterItem,
  DEFAULT_FILTERS,
} from "@/components/chat/advanced-filters"
import { ManageFiltersDialog } from "@/components/chat/manage-filters-dialog"
import { Search, MessageSquare, Filter, Send, Loader2 } from "lucide-react"
import {
  InputGroup,
  InputGroupAddon,
  InputGroupButton,
  InputGroupInput,
} from "@/components/ui/input-group"
import { cn } from "@/lib/utils"
import { useConversations } from "@/lib/api/hooks"
import type { Conversation as APIConversation, ConversationFilter } from "@/lib/api/types"

// Extended Chat type with inboxId
type ChatWithInbox = Chat & { inboxId?: string }

// Convert API Conversation to local Chat type
function mapAPIConversation(conv: APIConversation): ChatWithInbox {
  return {
    id: conv.id,
    inboxId: conv.inbox_id,
    name: conv.contact?.name || conv.contact?.phone_number || "Unknown",
    avatar: conv.contact?.avatar_url,
    lastMessage: conv.last_message || "",
    timestamp: conv.last_message_at ? new Date(conv.last_message_at) : new Date(conv.created_at),
    unreadCount: conv.unread_count,
    isOnline: false,
    isGroup: false,
    isFavorite: conv.is_favorite,
    hasMedia: false,
    isWaiting: conv.status === "pending",
  }
}

function generateId() {
  return Math.random().toString(36).substring(2, 9)
}

export default function ConversationsPage() {
  const [selectedChat, setSelectedChat] = useState<ChatWithInbox | null>(null)
  const [showChatArea, setShowChatArea] = useState(false)
  const [searchQuery, setSearchQuery] = useState("")
  const [allFilters, setAllFilters] = useState<FilterItem[]>([...DEFAULT_FILTERS])
  const [activeFilterId, setActiveFilterId] = useState<string>("default-all")
  const [filterRules, setFilterRules] = useState<FilterRule[]>([])
  const [tempFilterRules, setTempFilterRules] = useState<FilterRule[]>([])
  const [filtersOpen, setFiltersOpen] = useState(false)
  const [manageFiltersOpen, setManageFiltersOpen] = useState(false)
  const [editingFilter, setEditingFilter] = useState<FilterItem | null>(null)
  const [newChatNumber, setNewChatNumber] = useState("")

  // API filter based on active filter
  const apiFilter = useMemo((): ConversationFilter => {
    const filter: ConversationFilter = {}
    if (searchQuery) filter.search = searchQuery
    
    const activeFilter = allFilters.find((f) => f.id === activeFilterId)
    if (activeFilter?.type === "default") {
      switch (activeFilter.defaultType) {
        case "unread":
          filter.filter = "unread"
          break
        case "groups":
          filter.filter = "groups"
          break
      }
    }
    return filter
  }, [searchQuery, activeFilterId, allFilters])

  // Fetch conversations from API
  const { data: apiData, isLoading } = useConversations(apiFilter)

  // Use API data directly
  const chats = useMemo((): ChatWithInbox[] => {
    if (apiData?.conversations) {
      return apiData.conversations.map(mapAPIConversation)
    }
    return []
  }, [apiData])

  const activeFilter = useMemo(
    () => allFilters.find((f) => f.id === activeFilterId),
    [allFilters, activeFilterId]
  )

  const handleSelectChat = (chat: ChatWithInbox) => {
    setSelectedChat(chat)
    setShowChatArea(true)
  }

  const handleBack = () => {
    setShowChatArea(false)
  }

  const handleOpenFilters = (open: boolean) => {
    if (open) {
      if (editingFilter && editingFilter.type === "saved") {
        setTempFilterRules([...(editingFilter.rules || [])])
      } else {
        setTempFilterRules([...filterRules])
      }
    } else {
      setEditingFilter(null)
    }
    setFiltersOpen(open)
  }

  const handleApplyFilters = () => {
    if (editingFilter && editingFilter.type === "saved") {
      setAllFilters((prev) =>
        prev.map((f) =>
          f.id === editingFilter.id ? { ...f, rules: tempFilterRules } : f
        )
      )
      if (activeFilterId === editingFilter.id) {
        setFilterRules(tempFilterRules)
      }
      setEditingFilter(null)
    } else {
      setFilterRules(tempFilterRules)
      setActiveFilterId("default-all")
    }
    setFiltersOpen(false)
  }

  const handleCancelFilters = () => {
    setTempFilterRules(filterRules)
    setEditingFilter(null)
    setFiltersOpen(false)
  }

  const handleSaveAsList = useCallback((name: string, rules: FilterRule[]) => {
    if (editingFilter && editingFilter.type === "saved") {
      setAllFilters((prev) =>
        prev.map((f) =>
          f.id === editingFilter.id ? { ...f, name, rules: [...rules] } : f
        )
      )
      if (activeFilterId === editingFilter.id) {
        setFilterRules(rules)
      }
      setEditingFilter(null)
    } else {
      const newFilter: FilterItem = {
        id: generateId(),
        name,
        type: "saved",
        rules: [...rules],
      }
      setAllFilters((prev) => [...prev, newFilter])
      setFilterRules(rules)
      setActiveFilterId(newFilter.id)
    }
    setFiltersOpen(false)
  }, [editingFilter, activeFilterId])

  const handleFilterChange = useCallback((filterId: string) => {
    setActiveFilterId(filterId)
    const filter = allFilters.find((f) => f.id === filterId)

    if (filter) {
      if (filter.type === "saved" && filter.rules) {
        setFilterRules(filter.rules)
      } else {
        setFilterRules([])
      }
    }
  }, [allFilters])

  const handleDeleteFilter = useCallback((id: string) => {
    const filter = allFilters.find((f) => f.id === id)
    if (filter?.type === "default") return

    setAllFilters((prev) => prev.filter((f) => f.id !== id))
    if (activeFilterId === id) {
      setActiveFilterId("default-all")
      setFilterRules([])
    }
  }, [allFilters, activeFilterId])

  const handleSaveNewFilter = useCallback((name: string, rules: FilterRule[]) => {
    const newFilter: FilterItem = {
      id: generateId(),
      name,
      type: "saved",
      rules: [...rules],
    }
    setAllFilters((prev) => [...prev, newFilter])
  }, [allFilters, activeFilterId])

  const handleReorderFilters = useCallback((reorderedFilters: FilterItem[]) => {
    setAllFilters(reorderedFilters)
  }, [])

  const handleEditFilter = useCallback((filter: FilterItem) => {
    if (filter.type === "default") return

    setEditingFilter(filter)
    setManageFiltersOpen(false)
    setTempFilterRules([...(filter.rules || [])])
    setFiltersOpen(true)
  }, [])

  const handleAddFilter = useCallback(() => {
    setEditingFilter(null)
    setTempFilterRules([])
    setFiltersOpen(true)
  }, [])

  const handleManageFilters = useCallback(() => {
    setManageFiltersOpen(true)
  }, [])

  const handleStartNewChat = () => {
    if (!newChatNumber.trim()) return
    // TODO: Implement new chat creation via API
    // For now, just clear the input - new chats should be created through the backend
    setNewChatNumber("")
  }

  const hasActiveFilters = filterRules.length > 0

  const unreadCount = useMemo(
    () => chats.filter((chat) => chat.unreadCount && chat.unreadCount > 0).length,
    [chats]
  )

  const applyFilterRules = useCallback((chatList: Chat[], rules: FilterRule[]) => {
    let filtered = chatList

    for (const rule of rules) {
      if (!rule.value && !["present", "notPresent"].includes(rule.operator)) {
        continue
      }

      switch (rule.field) {
        case "status":
          if (rule.operator === "equals") {
            filtered = filtered.filter((chat) =>
              rule.value === "online" ? chat.isOnline : !chat.isOnline
            )
          } else if (rule.operator === "notEquals") {
            filtered = filtered.filter((chat) =>
              rule.value === "online" ? !chat.isOnline : chat.isOnline
            )
          }
          break
        case "type":
          if (rule.operator === "equals") {
            filtered = filtered.filter((chat) =>
              rule.value === "group" ? chat.isGroup : !chat.isGroup
            )
          } else if (rule.operator === "notEquals") {
            filtered = filtered.filter((chat) =>
              rule.value === "group" ? !chat.isGroup : chat.isGroup
            )
          }
          break
        case "media":
          if (rule.operator === "present") {
            filtered = filtered.filter((chat) => chat.hasMedia)
          } else if (rule.operator === "notPresent") {
            filtered = filtered.filter((chat) => !chat.hasMedia)
          }
          break
        case "favorite":
          if (rule.operator === "equals") {
            filtered = filtered.filter((chat) =>
              rule.value === "true" ? chat.isFavorite : !chat.isFavorite
            )
          }
          break
      }
    }

    return filtered
  }, [])

  const filteredChats = useMemo(() => {
    let filtered = chats

    // Search filter
    if (searchQuery) {
      filtered = filtered.filter((chat) =>
        chat.name.toLowerCase().includes(searchQuery.toLowerCase())
      )
    }

    // Default filters
    if (activeFilter?.type === "default") {
      switch (activeFilter.defaultType) {
        case "unread":
          filtered = filtered.filter((chat) => chat.unreadCount && chat.unreadCount > 0)
          break
        case "groups":
          filtered = filtered.filter((chat) => chat.isGroup)
          break
      }
    }

    // Advanced filters (from saved list or manual)
    if (filterRules.length > 0) {
      filtered = applyFilterRules(filtered, filterRules)
    }

    return filtered
  }, [chats, searchQuery, activeFilter, filterRules, applyFilterRules])

  return (
    <SidebarProvider className="h-dvh">
      <AppSidebar />
      <SidebarInset className="flex h-full flex-col overflow-hidden">
        <div className="flex min-h-0 flex-1 rounded-xl border m-2 md:m-4">
          {/* Chat List */}
          <div className={cn(
            "flex w-full md:w-96 shrink-0 flex-col overflow-hidden md:border-r",
            showChatArea && "hidden md:flex"
          )}>
            {/* Header: Search + Actions */}
            <div className="flex h-[65px] shrink-0 items-center gap-2 border-b px-3">
              <SidebarTrigger />
              <div className="relative flex-1">
                <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                <Input
                  placeholder="Buscar..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="pl-9"
                />
              </div>
              <Popover open={filtersOpen} onOpenChange={handleOpenFilters}>
                <PopoverTrigger asChild>
                  <Button
                    variant="outline"
                    size="icon"
                    className={cn(
                      "shrink-0 relative",
                      hasActiveFilters && "border-primary text-primary"
                    )}
                  >
                    <Filter className="h-4 w-4" />
                    {hasActiveFilters && (
                      <span className="absolute -top-1 -right-1 flex h-4 w-4 items-center justify-center rounded-full bg-primary text-[10px] text-primary-foreground">
                        {filterRules.length}
                      </span>
                    )}
                  </Button>
                </PopoverTrigger>
                <PopoverContent align="end" className="w-auto p-4">
                  <AdvancedFilters
                    filters={tempFilterRules}
                    onFiltersChange={setTempFilterRules}
                    onApply={handleApplyFilters}
                    onCancel={handleCancelFilters}
                    onSaveAsList={handleSaveAsList}
                  />
                </PopoverContent>
              </Popover>
            </div>

            {/* Quick Filters */}
            <ChatFilters
              allFilters={allFilters}
              activeFilterId={activeFilterId}
              onFilterChange={handleFilterChange}
              unreadCount={unreadCount}
              onManageFilters={handleManageFilters}
            />

            {/* Chat List */}
            {isLoading ? (
              <div className="flex flex-1 items-center justify-center">
                <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
              </div>
            ) : filteredChats.length === 0 ? (
              <div className="flex flex-1 flex-col items-center justify-center gap-2 text-muted-foreground p-4">
                <MessageSquare className="h-8 w-8" />
                <p className="text-sm text-center">Nenhuma conversa encontrada</p>
              </div>
            ) : (
              <ChatList
                chats={filteredChats}
                selectedId={selectedChat?.id}
                onSelect={handleSelectChat}
              />
            )}

            {/* New Chat Input */}
            <div className="shrink-0 border-t p-3">
              <InputGroup>
                <InputGroupInput
                  placeholder="Digite o número..."
                  value={newChatNumber}
                  onChange={(e) => setNewChatNumber(e.target.value)}
                  onKeyDown={(e) => e.key === "Enter" && handleStartNewChat()}
                />
                <InputGroupAddon align="inline-end">
                  <InputGroupButton
                    size="sm"
                    variant="default"
                    onClick={handleStartNewChat}
                    disabled={!newChatNumber.trim()}
                  >
                    <Send className="h-4 w-4" />
                  </InputGroupButton>
                </InputGroupAddon>
              </InputGroup>
            </div>
          </div>

          {/* Chat Area */}
          <div className={cn(
            "flex min-h-0 min-w-0 flex-1 flex-col overflow-hidden",
            !showChatArea && "hidden md:flex"
          )}>
            {selectedChat ? (
              <ChatArea chat={selectedChat} connectionId={selectedChat.inboxId} onBack={handleBack} />
            ) : (
              <div className="flex h-full flex-col items-center justify-center gap-4 text-muted-foreground">
                <MessageSquare className="h-12 w-12" />
                <p>Selecione um chat para comecar</p>
              </div>
            )}
          </div>
        </div>
      </SidebarInset>

      {/* Manage Filters Dialog */}
      <ManageFiltersDialog
        open={manageFiltersOpen}
        onOpenChange={setManageFiltersOpen}
        allFilters={allFilters}
        onEdit={handleEditFilter}
        onDelete={handleDeleteFilter}
        onReorder={handleReorderFilters}
        onSaveNew={handleSaveNewFilter}
      />
    </SidebarProvider>
  )
}
