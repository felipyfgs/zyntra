"use client"

import { useState } from "react"
import { Plus, X, Save } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { Separator } from "@/components/ui/separator"

export type FilterField = "status" | "type" | "media" | "favorite" | "tags"

export type FilterOperator = "equals" | "notEquals" | "present" | "notPresent"

export type FilterRule = {
  id: string
  field: FilterField
  operator: FilterOperator
  value?: string
}

export type SavedFilter = {
  id: string
  name: string
  rules: FilterRule[]
}

export type FilterItem = {
  id: string
  name: string
  type: "default" | "saved"
  defaultType?: "all" | "unread" | "groups"
  rules?: FilterRule[]
}

export const DEFAULT_FILTERS: FilterItem[] = [
  { id: "default-all", name: "Todos", type: "default", defaultType: "all" },
  { id: "default-unread", name: "Não lidos", type: "default", defaultType: "unread" },
  { id: "default-groups", name: "Grupos", type: "default", defaultType: "groups" },
]

type AdvancedFiltersProps = {
  filters: FilterRule[]
  onFiltersChange: (filters: FilterRule[]) => void
  onApply: () => void
  onCancel: () => void
  onSaveAsList?: (name: string, rules: FilterRule[]) => void
}

const filterFields: { value: FilterField; label: string }[] = [
  { value: "status", label: "Status" },
  { value: "type", label: "Tipo" },
  { value: "media", label: "Mídia" },
  { value: "favorite", label: "Favorito" },
  { value: "tags", label: "Etiquetas" },
]

const operatorsByField: Record<FilterField, { value: FilterOperator; label: string }[]> = {
  status: [
    { value: "equals", label: "igual a" },
    { value: "notEquals", label: "diferente de" },
  ],
  type: [
    { value: "equals", label: "igual a" },
    { value: "notEquals", label: "diferente de" },
  ],
  media: [
    { value: "present", label: "presente" },
    { value: "notPresent", label: "não presente" },
  ],
  favorite: [
    { value: "equals", label: "igual a" },
  ],
  tags: [
    { value: "equals", label: "contém" },
    { value: "notEquals", label: "não contém" },
  ],
}

const valuesByField: Record<FilterField, { value: string; label: string }[] | null> = {
  status: [
    { value: "online", label: "Online" },
    { value: "offline", label: "Offline" },
  ],
  type: [
    { value: "individual", label: "Individual" },
    { value: "group", label: "Grupo" },
  ],
  media: null,
  favorite: [
    { value: "true", label: "Sim" },
    { value: "false", label: "Não" },
  ],
  tags: [
    { value: "vip", label: "VIP" },
    { value: "suporte", label: "Suporte" },
    { value: "vendas", label: "Vendas" },
  ],
}

function generateId() {
  return Math.random().toString(36).substring(2, 9)
}

export function AdvancedFilters({
  filters,
  onFiltersChange,
  onApply,
  onCancel,
  onSaveAsList,
}: AdvancedFiltersProps) {
  const [showSaveInput, setShowSaveInput] = useState(false)
  const [listName, setListName] = useState("")

  const addFilter = () => {
    const newFilter: FilterRule = {
      id: generateId(),
      field: "status",
      operator: "equals",
      value: undefined,
    }
    onFiltersChange([...filters, newFilter])
  }

  const removeFilter = (id: string) => {
    onFiltersChange(filters.filter((f) => f.id !== id))
  }

  const updateFilter = (id: string, updates: Partial<FilterRule>) => {
    onFiltersChange(
      filters.map((f) => {
        if (f.id !== id) return f
        const updated = { ...f, ...updates }
        if (updates.field && updates.field !== f.field) {
          updated.operator = operatorsByField[updates.field][0].value
          updated.value = undefined
        }
        return updated
      })
    )
  }

  const clearFilters = () => {
    onFiltersChange([])
  }

  const handleSaveAsList = () => {
    if (!listName.trim() || !onSaveAsList) return
    onSaveAsList(listName.trim(), filters)
    setListName("")
    setShowSaveInput(false)
  }

  const hasFilters = filters.length > 0

  return (
    <div className="w-80">
      <div className="flex items-center justify-between mb-3">
        <h4 className="font-medium text-sm">Filtros avançados</h4>
        {hasFilters && (
          <Button
            variant="ghost"
            size="sm"
            onClick={clearFilters}
            className="h-auto p-1 text-xs text-muted-foreground"
          >
            <X className="h-3 w-3 mr-1" />
            Limpar
          </Button>
        )}
      </div>

      <Separator className="mb-3" />

      <div className="space-y-2 max-h-64 overflow-y-auto">
        {filters.map((filter) => {
          const operators = operatorsByField[filter.field]
          const values = valuesByField[filter.field]
          const needsValue = !["present", "notPresent"].includes(filter.operator)

          return (
            <div key={filter.id} className="flex items-center gap-1.5">
              <Select
                value={filter.field}
                onValueChange={(value) => updateFilter(filter.id, { field: value as FilterField })}
              >
                <SelectTrigger className="h-8 w-24 text-xs">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {filterFields.map((f) => (
                    <SelectItem key={f.value} value={f.value} className="text-xs">
                      {f.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>

              <Select
                value={filter.operator}
                onValueChange={(value) => updateFilter(filter.id, { operator: value as FilterOperator })}
              >
                <SelectTrigger className="h-8 w-28 text-xs">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {operators.map((op) => (
                    <SelectItem key={op.value} value={op.value} className="text-xs">
                      {op.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>

              {needsValue && values && (
                <Select
                  value={filter.value}
                  onValueChange={(value) => updateFilter(filter.id, { value })}
                >
                  <SelectTrigger className="h-8 w-24 text-xs">
                    <SelectValue placeholder="Valor" />
                  </SelectTrigger>
                  <SelectContent>
                    {values.map((v) => (
                      <SelectItem key={v.value} value={v.value} className="text-xs">
                        {v.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              )}

              <Button
                variant="ghost"
                size="icon"
                onClick={() => removeFilter(filter.id)}
                className="h-8 w-8 shrink-0"
              >
                <X className="h-3.5 w-3.5" />
              </Button>
            </div>
          )
        })}
      </div>

      <Button
        variant="outline"
        size="sm"
        onClick={addFilter}
        className="w-full mt-3 text-xs"
      >
        <Plus className="h-3.5 w-3.5 mr-1.5" />
        Adicionar filtro
      </Button>

      {hasFilters && onSaveAsList && (
        <>
          <Separator className="my-3" />
          {showSaveInput ? (
            <div className="flex gap-2">
              <Input
                placeholder="Nome da lista..."
                value={listName}
                onChange={(e) => setListName(e.target.value)}
                onKeyDown={(e) => e.key === "Enter" && handleSaveAsList()}
                className="h-8 text-xs"
                autoFocus
              />
              <Button
                size="sm"
                onClick={handleSaveAsList}
                disabled={!listName.trim()}
                className="h-8"
              >
                <Save className="h-3.5 w-3.5" />
              </Button>
              <Button
                size="sm"
                variant="ghost"
                onClick={() => setShowSaveInput(false)}
                className="h-8"
              >
                <X className="h-3.5 w-3.5" />
              </Button>
            </div>
          ) : (
            <Button
              variant="secondary"
              size="sm"
              onClick={() => setShowSaveInput(true)}
              className="w-full text-xs"
            >
              <Save className="h-3.5 w-3.5 mr-1.5" />
              Salvar como lista
            </Button>
          )}
        </>
      )}

      <Separator className="my-3" />

      <div className="flex justify-end gap-2">
        <Button variant="outline" size="sm" onClick={onCancel}>
          Cancelar
        </Button>
        <Button size="sm" onClick={onApply}>
          Aplicar
        </Button>
      </div>
    </div>
  )
}
