'use client'

import { EnrichedAction } from '@/gen/flyteidl2/workflow/run_definition_pb'
import { create } from 'zustand'
import { flattenTree, GROUP_SEPARATOR } from './flattenTree'
import { ActionId, ActionWithChildren, FlatRunNode } from './types'

// initial state to determine whether to collapse all sidebar items
const AUTOCOLLAPSE_NEW_ITEMS = false

type RunStore = {
  actions: Record<ActionId, ActionWithChildren>
  collapsedItems: Set<ActionId>
  userExpandedItems: Set<ActionId> // Track what user explicitly expanded
  flatItems: FlatRunNode[]
  reset: () => void
  run: EnrichedAction | null
  setRun: (run: EnrichedAction | null) => void
  getAction: (id?: string) => ActionWithChildren | undefined
  toggleCollapsed: (id: ActionId) => void
  toggleCollapseAll: () => void
  autoCollapseNewItems: boolean
  upsertAction: (enrichedActions: EnrichedAction[]) => void
}

// prevent replacing root run more often than necessary by checking properties that we expect to change
function hasRunChanged(
  newRun: EnrichedAction,
  prevRun: EnrichedAction | null,
): boolean {
  if (!prevRun) return true
  return (
    prevRun.action?.status?.phase !== newRun.action?.status?.phase ||
    prevRun.action?.status?.endTime !== newRun.action?.status?.endTime ||
    prevRun.meetsFilter !== newRun.meetsFilter ||
    prevRun.action?.status?.attempts !== newRun.action?.status?.attempts ||
    JSON.stringify(prevRun.childrenPhaseCounts) !==
      JSON.stringify(newRun.childrenPhaseCounts)
  )
}

export const useRunStore = create<RunStore>((set, get) => ({
  actions: {},
  collapsedItems: new Set<ActionId>(),
  userExpandedItems: new Set<ActionId>(),
  flatItems: [],
  autoCollapseNewItems: AUTOCOLLAPSE_NEW_ITEMS,
  reset: () => {
    set({
      actions: {},
      collapsedItems: new Set(),
      userExpandedItems: new Set(),
      flatItems: [],
      autoCollapseNewItems: false,
    })
  },
  run: null,
  setRun: (run) => set({ run: run }),

  getAction: (id?: string) => {
    if (!id) return undefined
    const actions = get().actions
    return actions[id]
  },

  toggleCollapsed: (id) =>
    set((state) => {
      const run = state.run?.action
      if (!run?.id?.name || id === run.id.name) {
        return {}
      }

      const newCollapsed = new Set(state.collapsedItems)
      const newUserExpanded = new Set(state.userExpandedItems)

      if (newCollapsed.has(id)) {
        // User is expanding - track this
        newCollapsed.delete(id)
        newUserExpanded.add(id)
      } else {
        // User is collapsing - remove from user expanded
        newCollapsed.add(id)
        newUserExpanded.delete(id)
      }

      const flatItems = flattenTree(run.id.name, state.actions, newCollapsed)
      return {
        collapsedItems: newCollapsed,
        userExpandedItems: newUserExpanded,
        flatItems,
      }
    }),

  toggleCollapseAll: () =>
    set((state: RunStore) => {
      const run = state.run?.action
      if (!run?.id?.name) {
        return {}
      }

      const shouldCollapseAll = state.collapsedItems.size === 0

      if (shouldCollapseAll) {
        // get all collapsible items except root
        const allCollapsibleIds = state.flatItems
          .filter((item) => {
            if (item.id === run.id?.name) return false
            return (
              item.isGroup ||
              (item.node.children && item.node.children.length > 0) ||
              (item.node.groupChildren &&
                Object.keys(item.node.groupChildren).length > 0)
            )
          })
          .map((item) => item.id)

        const flatItems = flattenTree(
          run.id.name,
          state.actions,
          new Set(allCollapsibleIds),
        )

        return {
          collapsedItems: new Set(allCollapsibleIds),
          autoCollapseNewItems: true, // Enable auto-collapse for new items
          userExpandedItems: new Set(),
          flatItems,
        }
      } else {
        // Expand all
        const flatItems = flattenTree(run.id.name, state.actions, new Set())

        return {
          collapsedItems: new Set(),
          autoCollapseNewItems: false, // Disable auto-collapse
          userExpandedItems: new Set(), // Clear user intentions
          flatItems,
        }
      }
    }),

  upsertAction: (incomingActions: EnrichedAction[]) => {
    const prevActions = get().actions
    const nodes: Record<ActionId, ActionWithChildren> = { ...prevActions }
    const currentCollapsed = new Set(get().collapsedItems)
    const userExpandedItems = get().userExpandedItems
    const autoCollapseNewItems = get().autoCollapseNewItems

    for (const enrichedAction of incomingActions) {
      if (
        !enrichedAction.meetsFilter &&
        enrichedAction.action?.id?.name !== 'a0'
      ) {
        if (enrichedAction?.action?.id?.name) {
          delete nodes[enrichedAction?.action?.id?.name]
        }
        continue
      }

      const id = enrichedAction?.action?.id?.name || ''
      const parentId = enrichedAction?.action?.metadata?.parent
      const group = enrichedAction?.action?.metadata?.group

      const existing = nodes[id]
      const isNewItem = !existing

      const merged: ActionWithChildren = {
        ...existing,
        ...enrichedAction,
        children: existing?.children ?? [],
        groupChildren: existing?.groupChildren ?? {},
      }

      nodes[id] = merged

      // Auto-collapse new items if flag is set
      if (isNewItem && autoCollapseNewItems && parentId) {
        currentCollapsed.add(id)
      }

      // Link to parent if available
      if (parentId) {
        const parent = nodes[parentId] ?? {
          runId: {
            org: '',
            project: '',
            domain: '',
            rootName: '',
            name: parentId,
          },
          inputs: {},
          children: [],
          groupChildren: {},
        }

        if (group) {
          const groupFolderId = `${parentId}${GROUP_SEPARATOR}${group}`
          const isNewGroup =
            !parent.groupChildren[group] ||
            parent.groupChildren[group].length === 0

          if (!parent.groupChildren[group]) {
            parent.groupChildren[group] = []
          }
          if (!parent.groupChildren[group].includes(id)) {
            parent.groupChildren[group].push(id)
          }

          // Auto-collapse new groups unless user explicitly expanded them
          if (
            isNewGroup &&
            autoCollapseNewItems &&
            !userExpandedItems.has(groupFolderId)
          ) {
            currentCollapsed.add(groupFolderId)
          }
        } else {
          if (!parent.children.includes(id)) {
            parent.children.push(id)
          }
        }
        nodes[parentId] = parent
      } else {
        // Handle root run update
        const prevRun = get().run
        if (
          prevRun?.action?.id?.name !== enrichedAction.action?.id?.name ||
          hasRunChanged(enrichedAction, prevRun)
        ) {
          set({ run: enrichedAction })
        }
      }
    }

    const run = get().run
    const rootId = run?.action?.id?.name || Object.keys(nodes)[0] || ''

    // Ensure root is never collapsed
    currentCollapsed.delete(rootId)

    const flatItems = flattenTree(rootId, nodes, currentCollapsed)
    set({
      actions: nodes,
      flatItems,
      collapsedItems: currentCollapsed,
    })
  },
}))
