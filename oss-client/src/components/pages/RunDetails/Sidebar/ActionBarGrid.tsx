import { useLayoutStore } from '../state/LayoutStore'
import { getTimelineWidth } from './util'

// allow space for overflow menu
export const OFFSET = 39

const Line = ({ className }: { className: string }) => (
  <div
    className={`absolute top-0 h-full min-h-0 border-r-1 border-(--system-gray-4) opacity-50 ${className}`}
  />
)

export const ActionBarGrid = ({
  height,
  sidebarWidth,
}: {
  height: number
  sidebarWidth: number
}) => {
  const mode = useLayoutStore((s) => s.mode)
  const width = getTimelineWidth(mode, sidebarWidth)
  return (
    <div
      className="pointer-events-none absolute top-0 z-0 h-full min-h-0"
      style={{ width, right: OFFSET, height }}
    >
      <Line className="left-0" />
      <Line className="left-1/3" />
      <Line className="left-2/3" />
      <Line className="right-0" />
    </div>
  )
}
