import { ChildPhaseCounts } from '../types'
import { LayoutMode } from '../state/LayoutStore'

export const getChildCount = (
  childPhaseCounts: ChildPhaseCounts | undefined,
) => {
  let childCount = 0
  if (childPhaseCounts) {
    const phaseCounts = Object.values(childPhaseCounts)
    childCount = phaseCounts.reduce((a, b) => {
      return a + b
    }, 0)
  }
  return childCount
}

const DEFAULT_SIDEBAR_WIDTH = 520

export const getWidthByMode = (mode: LayoutMode): number | string => {
  switch (mode) {
    case 'full-action-log':
      return '100%'
    // case 'mini-action-log': // not-implemented
    //   return 20
    case 'no-action-log':
      return 0
    case 'default':
    default: {
      return DEFAULT_SIDEBAR_WIDTH
    }
  }
}

export const getTimelineWidth = (mode: LayoutMode, sidebarWidth: number) => {
  // We want the timeline to feel generous on wide screens, but to *yield space*
  // to the left-side row content (text, timestamp, badges, menu) as the sidebar
  // narrows — rather than forcing truncation of the label.

  // Minimum usable width for the timeline track (anything smaller is visually noisy)
  const MIN_TIMELINE_PX = 120
  // Hard cap so the timeline never dominates the row on very wide screens
  const MAX_TIMELINE_PX = 1020

  // Approx space needed for left-side content to remain readable.
  // Tune this number based on real UI: label + timestamp + badges + padding.
  const RESERVED_FOR_CONTENT_PX = 500

  const clamp = (n: number, min: number, max: number) =>
    Math.min(max, Math.max(min, n))

  switch (mode) {
    case 'default':
      // Compact timeline in the default sidebar
      return 86

    case 'full-action-log': {
      // Prefer a proportional width on wide screens…
      const preferred = Math.floor(sidebarWidth * 0.7)

      // …but once we start to cut into content width, start shrinking the timeline to preserve
      // space for labels instead of truncating them.
      const available = sidebarWidth - RESERVED_FOR_CONTENT_PX

      // Choose the smaller of the preferred width and what we can afford.
      // Allow it to shrink down to MIN_TIMELINE_PX, and disappear entirely only
      // if the sidebar is extremely narrow.
      const target = Math.min(preferred, available)

      if (target <= 0) return 0
      return clamp(target, MIN_TIMELINE_PX, MAX_TIMELINE_PX)
    }

    case 'no-action-log':
    default:
      return 0
  }
}
