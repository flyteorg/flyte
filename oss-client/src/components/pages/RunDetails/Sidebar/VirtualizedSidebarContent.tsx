import { useVirtualizer } from '@tanstack/react-virtual'
import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { useSelectedItem } from '../hooks/useSelectedItem'
import { useRunStore } from '../state/RunStore'
import { FlatRunNode } from '../state/types'
import { ActionBarGrid } from './ActionBarGrid'
import { RunNode } from './RunNode'
import { useLayoutStore } from '../state/LayoutStore'
import { getTimelineWidth } from './util'

const ITEM_HEIGHT = 28

/**
 * Given the flat ordered list and the index of the first visible item,
 * returns the ancestor chain that has scrolled above the viewport.
 * Works by walking backward and collecting the nearest item at each
 * decreasing depth level (DFS order guarantees correctness).
 */
function getStickyAncestors(
  flatItems: FlatRunNode[],
  firstVisibleIndex: number,
): FlatRunNode[] {
  const item = flatItems[firstVisibleIndex]
  if (!item || item.depth === 0) return []

  const ancestors: FlatRunNode[] = []
  let targetDepth = item.depth - 1

  for (let i = firstVisibleIndex - 1; i >= 0 && targetDepth >= 0; i--) {
    const candidate = flatItems[i]
    if (candidate.depth === targetDepth) {
      ancestors.unshift(candidate)
      targetDepth--
    }
  }

  return ancestors
}

type VirtualizedSidebarContentProps = {
  sidebarWidth: number
}

export const VirtualizedSidebarContent = ({
  sidebarWidth,
}: VirtualizedSidebarContentProps) => {
  const [hasScrolledIntoView, setHasScrolledIntoView] = useState(false)
  const lastScrolledItemIdRef = useRef<string | null>(null)
  const parentRef = useRef<HTMLDivElement>(null)
  const [scrollTop, setScrollTop] = useState(0)
  const flatItems = useRunStore((s) => s.flatItems)
  const { selectedItem } = useSelectedItem()
  const mode = useLayoutStore((s) => s.mode)
  const timelineWidth = getTimelineWidth(mode, sidebarWidth)
  const shouldShowGrid =
    timelineWidth > 0 && (mode === 'default' || mode === 'full-action-log')

  const rowVirtualizer = useVirtualizer({
    count: flatItems.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => ITEM_HEIGHT,
  })

  const handleScroll = useCallback((e: React.UIEvent<HTMLDivElement>) => {
    setScrollTop(e.currentTarget.scrollTop)
  }, [])

  const stickyAncestors = useMemo(() => {
    const baseIndex = Math.floor(scrollTop / ITEM_HEIGHT)
    const base = getStickyAncestors(flatItems, baseIndex)
    // Do one adjusted pass using (scrollTop + currentStickyHeight) as the
    // effective scroll position. This pins a new ancestor to the bottom of
    // the sticky area as soon as it would be covered, instead of letting it
    // slide all the way behind its parent before snapping into place.
    const stickyHeight = base.length * ITEM_HEIGHT
    if (stickyHeight === 0) return base
    const adjustedIndex = Math.floor((scrollTop + stickyHeight) / ITEM_HEIGHT)
    if (adjustedIndex === baseIndex || adjustedIndex >= flatItems.length)
      return base
    return getStickyAncestors(flatItems, adjustedIndex)
  }, [flatItems, scrollTop])

  // Scroll to selected item when it becomes available in flatItems
  // This handles both initial load and when actions arrive later
  useEffect(() => {
    if (!hasScrolledIntoView && selectedItem?.id && flatItems.length > 0) {
      const targetIndex = flatItems.findIndex(
        (item) => item.id === selectedItem.id,
      )
      if (targetIndex !== -1 && selectedItem.id !== 'a0') {
        rowVirtualizer.scrollToIndex(targetIndex, {
          align: 'center',
        })
        lastScrolledItemIdRef.current = selectedItem.id
        setHasScrolledIntoView(true)
      }
    }
  }, [hasScrolledIntoView, selectedItem?.id, flatItems, rowVirtualizer])

  return (
    <div className="relative min-h-0 flex-1">
      <div
        ref={parentRef}
        onScroll={handleScroll}
        className="absolute inset-0 overflow-y-auto pr-3 [scrollbar-gutter:stable] [&::-webkit-scrollbar]:w-2 [&::-webkit-scrollbar-corner]:bg-transparent [&::-webkit-scrollbar-thumb]:rounded-full [&::-webkit-scrollbar-thumb]:bg-zinc-400 dark:[&::-webkit-scrollbar-thumb]:bg-zinc-600 [&::-webkit-scrollbar-thumb:hover]:bg-zinc-500 dark:[&::-webkit-scrollbar-thumb:hover]:bg-zinc-500 [&::-webkit-scrollbar-track]:bg-transparent"
        style={{
          marginRight: '-20px',
        }}
      >
        {shouldShowGrid && (
          <ActionBarGrid
            height={rowVirtualizer.getTotalSize()}
            sidebarWidth={sidebarWidth}
          />
        )}
        <div
          style={{
            height: rowVirtualizer.getTotalSize(),
            position: 'relative',
          }}
        >
          {rowVirtualizer.getVirtualItems().map((virtualRow) => {
            const item = flatItems[virtualRow.index]
            return (
              <div
                key={item.id}
                style={{
                  position: 'absolute',
                  top: 0,
                  left: 0,
                  width: '100%',
                  transform: `translateY(${virtualRow.start}px)`,
                }}
              >
                <RunNode
                  flatRunNode={item}
                  sidebarWidth={sidebarWidth}
                />
              </div>
            )
          })}
        </div>
      </div>
      {stickyAncestors.length > 0 && (
        <div className="pointer-events-none absolute top-0 right-0 left-0 z-10 pr-0">
          <div className="pointer-events-auto bg-(--system-black)">
            {stickyAncestors.map((ancestor) => (
              <RunNode
                key={`sticky-${ancestor.id}`}
                flatRunNode={ancestor}
                sidebarWidth={sidebarWidth}
              />
            ))}
          </div>
        </div>
      )}
    </div>
  )
}
