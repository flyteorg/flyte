import stringHash from '@sindresorhus/string-hash'

// Tailwind colors chosen for accessibility and distinctiveness
const ACCESSIBLE_COLORS = [
  'bg-emerald-600', // Deep green
  'bg-violet-600', // Deep violet
  'bg-cyan-600', // Deep cyan
  'bg-orange-500', // Bright orange
  'bg-indigo-600', // Deep indigo
  'bg-teal-600', // Deep teal
  'bg-rose-600', // Deep rose
  'bg-slate-600', // Deep slate
] as const

/**
 * Generates a consistent color for a given string using a hash function
 * Colors are chosen to be:
 * - Accessible (WCAG 2.1 compliant)
 * - Distinct from each other
 * - Distinct from log level colors (red-500, yellow-500, blue-500)
 */
export const stringToColor = (str?: string): string => {
  const hash = stringHash(str || '')
  return ACCESSIBLE_COLORS[hash % ACCESSIBLE_COLORS.length]
}
