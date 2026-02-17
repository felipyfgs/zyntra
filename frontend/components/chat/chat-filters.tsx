"use client"

import { Settings2 } from "lucide-react"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { ScrollArea, ScrollBar } from "@/components/ui/scroll-area"
import { cn } from "@/lib/utils"
import { type FilterItem } from "./advanced-filters"

export type ChatFilterType = "all" | "unread" | "groups" | "saved"

type ChatFiltersProps = {
  allFilters: FilterItem[]
  activeFilterId?: string
  onFilterChange: (filterId: string) => void
  unreadCount?: number
  onManageFilters?: () => void
}

export function ChatFilters({
  allFilters,
  activeFilterId,
  onFilterChange,
  unreadCount = 0,
  onManageFilters,
}: ChatFiltersProps) {
  return (
    <div className="border-b">
      <ScrollArea className="w-full whitespace-nowrap">
        <div className="flex items-center gap-2 px-3 py-2">
          {allFilters.map((filter) => {
            const isActive = activeFilterId === filter.id
            const showUnreadBadge = filter.defaultType === "unread" && unreadCount > 0

            return (
              <Button
                key={filter.id}
                onClick={() => onFilterChange(filter.id)}
                variant={isActive ? "default" : "outline"}
                size="sm"
                className={cn(
                  "gap-1.5 shrink-0 rounded-full px-4",
                  isActive && "bg-primary text-primary-foreground"
                )}
              >
                {filter.name}
                {showUnreadBadge && (
                  <Badge
                    variant={isActive ? "secondary" : "default"}
                    className="ml-1 h-5 min-w-5 px-1.5"
                  >
                    {unreadCount}
                  </Badge>
                )}
              </Button>
            )
          })}

          {onManageFilters && (
            <Button
              onClick={onManageFilters}
              variant="ghost"
              size="sm"
              className="shrink-0 rounded-full px-3 text-muted-foreground hover:text-foreground"
            >
              <Settings2 className="h-4 w-4" />
            </Button>
          )}
        </div>
        <ScrollBar orientation="horizontal" />
      </ScrollArea>
    </div>
  )
}
