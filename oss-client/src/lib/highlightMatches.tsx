const escapeRegExp = (string: string) =>
  string.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')

export const highlightMatches = (text: string, searchQuery: string) => {
  if (!searchQuery || !text) return text
  const parts = text.split(new RegExp(`(${escapeRegExp(searchQuery)})`, 'gi'))
  return parts.map((part, i) =>
    part.toLowerCase() === searchQuery.toLowerCase() ? (
      <mark key={i}>{part}</mark>
    ) : (
      part
    ),
  )
}
