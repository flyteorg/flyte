/**
 * Formats a number for display:
 * - Returns the number as-is if under 1000
 * - Returns formatted string (e.g., "2.5k", "1k") if 1000
 * - Rounds to 1 decimal place
 */
export function formatNumber(num: number): number | string {
  if (num < 1000) {
    return num
  }

  const thousands = num / 1000

  // Round to 1 decimal place
  const rounded = Math.round(thousands * 10) / 10

  // If it's a whole number, don't show decimal
  if (rounded % 1 === 0) {
    return `${rounded}k`
  }

  return `${rounded}k`
}
