import { format } from 'date-fns'
import type { DateRange } from 'react-day-picker'

interface LabeledRange {
  filterLabel: string
  label: string
  getRange: () => DateRange
}

export const detectQuickRangeLabel = (
  range: DateRange | undefined,
  quickRanges: LabeledRange[],
  fallbackLabel: string,
): string => {
  if (!range?.from || !range?.to) return fallbackLabel

  const { from, to } = range

  // Infer if this is a time-based range
  const hasTimeComponent =
    from.getHours() !== to.getHours() || from.getMinutes() !== to.getMinutes()

  // Formatter depends on time/no-time
  const fmt = hasTimeComponent ? 'yyyy-MM-dd HH:mm' : 'yyyy-MM-dd'

  const fromStr = format(from, fmt)
  const toStr = format(to, fmt)

  for (const { filterLabel, getRange } of quickRanges) {
    const preset = getRange()
    if (!preset.from || !preset.to) continue

    const pf = format(preset.from, fmt)
    const pt = format(preset.to, fmt)

    if (pf === fromStr && pt === toStr) {
      return filterLabel
    }
  }

  // Default: "MM-dd – MM-dd"
  return `${format(from, 'MM/dd')} – ${format(to, 'MM/dd')}`
}
