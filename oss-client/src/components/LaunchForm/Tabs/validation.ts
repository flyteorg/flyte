/**
 * Validates key-value pairs for context, env vars, and labels.
 * Returns an error message if invalid, null if valid.
 */
export function getKvPairsValidationError(
  pairs: { key?: string; value?: string }[] | undefined,
): string | null {
  if (!pairs || pairs.length === 0) return null
  const hasValueWithEmptyKey = pairs.some(
    (kv) => !(kv.key?.trim() ?? '') && !!(kv.value?.trim() ?? ''),
  )
  if (hasValueWithEmptyKey) {
    return 'Key is required'
  }
  const keyCounts = new Map<string, number>()
  const keysWithSpaces: string[] = []
  pairs.forEach((kv) => {
    const trimmedKey = kv.key?.trim()
    if (trimmedKey) {
      if (/\s/.test(trimmedKey)) keysWithSpaces.push(trimmedKey)
      keyCounts.set(trimmedKey, (keyCounts.get(trimmedKey) || 0) + 1)
    }
  })
  if (keysWithSpaces.length > 0) {
    return `Keys cannot contain spaces: ${keysWithSpaces.join(', ')}`
  }
  const duplicateKeys = Array.from(keyCounts.entries())
    .filter(([, count]) => count > 1)
    .map(([key]) => key)
  if (duplicateKeys.length > 0) {
    return `Duplicate keys found: ${duplicateKeys.join(', ')}`
  }
  return null
}
