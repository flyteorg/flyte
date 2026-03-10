import { getWindow } from './windowUtils'

export function remToPx(remValue: number) {
  const w = getWindow()
  if (!w) return 16 * remValue
  const rootFontSize = parseFloat(
    w.getComputedStyle(document.documentElement).fontSize,
  )

  return remValue * rootFontSize
}
