"use client"

import { useState } from "react"
import { GripVertical, Pencil, Trash2, Plus, ArrowLeft, Save, X } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Input } from "@/components/ui/input"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import {
  ListProvider,
  ListGroup,
  ListItems,
  ListItem,
  type DragEndEvent,
} from "@/components/kibo-ui/list"
import {
  type FilterItem,
  type FilterRule,
  type FilterField,
  type FilterOperator,
} from "./advanced-filters"

type ManageFiltersDialogProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  allFilters: FilterItem[]
  onEdit: (filter: FilterItem) => void
  onDelete: (id: string) => void
  onReorder: (filters: FilterItem[]) => void
  onSaveNew: (name: string, rules: FilterRule[]) => void
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
  favorite: [{ value: "equals", label: "igual a" }],
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

export function ManageFiltersDialog({
  open,
  onOpenChange,
  allFilters,
  onEdit,
  onDelete,
  onReorder,
  onSaveNew,
}: ManageFiltersDialogProps) {
  const [view, setView] = useState<"list" | "form">("list")
  const [newListName, setNewListName] = useState("")
  const [newListRules, setNewListRules] = useState<FilterRule[]>([])

  const handleDragEnd = (event: DragEndEvent) => {
    const { active, over } = event
    if (!over || active.id === over.id) return

    const oldIndex = allFilters.findIndex((f) => f.id === active.id)
    const newIndex = allFilters.findIndex((f) => f.id === over.id)
    if (oldIndex === -1 || newIndex === -1) return

    const newFilters = [...allFilters]
    const [removed] = newFilters.splice(oldIndex, 1)
    newFilters.splice(newIndex, 0, removed)
    onReorder(newFilters)
  }

  const addRule = () => {
    setNewListRules([
      ...newListRules,
      { id: generateId(), field: "status", operator: "equals", value: undefined },
    ])
  }

  const removeRule = (id: string) => {
    setNewListRules(newListRules.filter((r) => r.id !== id))
  }

  const updateRule = (id: string, updates: Partial<FilterRule>) => {
    setNewListRules(
      newListRules.map((r) => {
        if (r.id !== id) return r
        const updated = { ...r, ...updates }
        if (updates.field && updates.field !== r.field) {
          updated.operator = operatorsByField[updates.field][0].value
          updated.value = undefined
        }
        return updated
      })
    )
  }

  const handleSave = () => {
    if (!newListName.trim()) return
    onSaveNew(newListName.trim(), newListRules)
    setNewListName("")
    setNewListRules([])
    setView("list")
  }

  const handleOpenForm = () => {
    setNewListName("")
    setNewListRules([])
    setView("form")
  }

  const handleBack = () => {
    setView("list")
  }

  const handleOpenChange = (isOpen: boolean) => {
    if (!isOpen) {
      setView("list")
    }
    onOpenChange(isOpen)
  }

  const customFiltersCount = allFilters.filter((f) => f.type === "saved").length

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-sm gap-0 p-0">
        {view === "list" ? (
          <>
            <DialogHeader className="px-4 py-3 border-b">
              <DialogTitle className="text-base">Listas</DialogTitle>
            </DialogHeader>

            <div className="max-h-72 overflow-y-auto">
              <ListProvider onDragEnd={handleDragEnd} className="gap-0">
                <ListGroup id="filters" className="bg-transparent">
                  <ListItems className="gap-1 p-2">
                    {allFilters.map((filter, index) => {
                      const isDefault = filter.type === "default"
                      const rulesCount = filter.rules?.length ?? 0

                      return (
                        <ListItem
                          key={filter.id}
                          id={filter.id}
                          name={filter.name}
                          index={index}
                          parent="filters"
                          className="p-0 shadow-none border-0 bg-transparent hover:bg-muted/50 rounded-md"
                        >
                          <div className="flex items-center w-full gap-2 px-2 py-1.5">
                            <GripVertical className="h-4 w-4 text-muted-foreground/50 shrink-0" />
                            <span className="flex-1 text-sm truncate">{filter.name}</span>
                            {isDefault ? (
                              <Badge variant="outline" className="text-[10px] px-1.5 py-0 font-normal text-muted-foreground">
                                padrão
                              </Badge>
                            ) : (
                              <div className="flex items-center gap-0.5 shrink-0">
                                <span className="text-xs text-muted-foreground mr-1">{rulesCount}</span>
                                <Button
                                  variant="ghost"
                                  size="icon"
                                  onClick={(e) => {
                                    e.stopPropagation()
                                    onEdit(filter)
                                    onOpenChange(false)
                                  }}
                                  className="h-7 w-7"
                                >
                                  <Pencil className="h-3.5 w-3.5" />
                                </Button>
                                <Button
                                  variant="ghost"
                                  size="icon"
                                  onClick={(e) => {
                                    e.stopPropagation()
                                    onDelete(filter.id)
                                  }}
                                  className="h-7 w-7 text-muted-foreground hover:text-destructive hover:bg-destructive/10"
                                >
                                  <Trash2 className="h-3.5 w-3.5" />
                                </Button>
                              </div>
                            )}
                          </div>
                        </ListItem>
                      )
                    })}
                  </ListItems>
                </ListGroup>
              </ListProvider>
            </div>

            <DialogFooter className="px-4 py-3 border-t">
              <div className="flex-1 text-xs text-muted-foreground">
                {customFiltersCount} {customFiltersCount === 1 ? "personalizada" : "personalizadas"}
              </div>
              <Button size="sm" onClick={handleOpenForm}>
                <Plus className="h-4 w-4 mr-1" />
                Nova
              </Button>
            </DialogFooter>
          </>
        ) : (
          <>
            <DialogHeader className="px-4 py-3 border-b flex-row items-center gap-2 space-y-0">
              <Button variant="ghost" size="icon" onClick={handleBack} className="h-8 w-8 shrink-0">
                <ArrowLeft className="h-4 w-4" />
              </Button>
              <DialogTitle className="text-base flex-1">Nova lista</DialogTitle>
            </DialogHeader>

            <div className="p-4 space-y-4">
              <Input
                placeholder="Nome da lista"
                value={newListName}
                onChange={(e) => setNewListName(e.target.value)}
                autoFocus
              />

              <div className="space-y-2 max-h-40 overflow-y-auto">
                {newListRules.map((rule) => {
                  const operators = operatorsByField[rule.field]
                  const values = valuesByField[rule.field]
                  const needsValue = !["present", "notPresent"].includes(rule.operator)

                  return (
                    <div key={rule.id} className="flex items-center gap-1.5">
                      <Select
                        value={rule.field}
                        onValueChange={(v) => updateRule(rule.id, { field: v as FilterField })}
                      >
                        <SelectTrigger className="h-8 flex-1 text-xs">
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
                        value={rule.operator}
                        onValueChange={(v) => updateRule(rule.id, { operator: v as FilterOperator })}
                      >
                        <SelectTrigger className="h-8 flex-1 text-xs">
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
                          value={rule.value}
                          onValueChange={(v) => updateRule(rule.id, { value: v })}
                        >
                          <SelectTrigger className="h-8 flex-1 text-xs">
                            <SelectValue placeholder="..." />
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

                      <Button variant="ghost" size="icon" onClick={() => removeRule(rule.id)} className="h-8 w-8 shrink-0">
                        <X className="h-3.5 w-3.5" />
                      </Button>
                    </div>
                  )
                })}
              </div>

              <Button variant="outline" size="sm" onClick={addRule} className="w-full">
                <Plus className="h-3.5 w-3.5 mr-1" />
                Adicionar filtro
              </Button>
            </div>

            <DialogFooter className="px-4 py-3 border-t">
              <Button variant="outline" size="sm" onClick={handleBack}>
                Cancelar
              </Button>
              <Button size="sm" onClick={handleSave} disabled={!newListName.trim()}>
                <Save className="h-4 w-4 mr-1" />
                Salvar
              </Button>
            </DialogFooter>
          </>
        )}
      </DialogContent>
    </Dialog>
  )
}
