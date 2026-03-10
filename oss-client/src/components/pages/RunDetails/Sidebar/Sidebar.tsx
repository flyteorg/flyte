'use client'

import { useMeasure } from 'react-use'
import { useLayoutStore } from '../state/LayoutStore'
import { useRunStore } from '../state/RunStore'
import { EmptySidebar } from './EmptySidebar'
import { SidebarControls } from './FilterButtons'
import { getWidthByMode } from './util'
import { VirtualizedSidebarContent } from './VirtualizedSidebarContent'

export const Sidebar = () => {
  const [ref, bounds] = useMeasure<HTMLElement>()
  const actions = useRunStore((s) => s.actions)

  const mode = useLayoutStore((s) => s.mode)
  // const shouldShowSidebar = mode === 'default' || mode === 'full-action-log'

  const width = getWidthByMode(mode)

  return (
    <>
      <aside
        className="flex h-full min-h-0 flex-col gap-3 border-gray-700 bg-(--system-black) p-6"
        data-testid="sidebar"
        ref={ref}
        style={{ width, minWidth: width }}
      >
        <div className="flex min-h-0 flex-col gap-1 py-2">
          <SidebarControls />
        </div>
        {Object.keys(actions).length === 0 && <EmptySidebar />}
        <VirtualizedSidebarContent sidebarWidth={bounds.width} />
      </aside>
    </>
  )
}

export default Sidebar
