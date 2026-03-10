export function isMarkdown(text?: string): boolean {
  if (!text) return false

  const markdownPatterns: RegExp[] = [
    /^#{1,6} /m, // headings (#, ##, ###...)
    /\*\*[^*]+\*\*/, // bold (**text**)
    /\*[^*]+\*/, // italic (*text*)
    /!\[.*?\]\(.*?\)/, // images ![alt](url)
    /\[.*?\]\(.*?\)/, // links [text](url)
    /^> /m, // blockquotes
    /^[-*+] /m, // unordered list
    /^\d+\.\s/m, // ordered list
    /`[^`]+`/, // inline code
    /```[\s\S]*?```/, // code block
    /-{3,}|_{3,}|\*{3,}/, // horizontal rule
    /^\|(.+\|)+/m, // tables
  ]

  return markdownPatterns.some((pattern) => pattern.test(text))
}

/**
 * Valid HTML tag name: letter then optional letters, digits, hyphens (no space after <).
 * Use this to avoid false positives on "a < b" (space after <) or "a <1>" (digit after <).
 */
const VALID_HTML_TAG_NAME = /<([a-zA-Z][a-zA-Z0-9-]*)(\s|>|\/)/

/**
 * Tags that indicate HTML. Includes common structure/formatting tags and tags that are
 * often used in attacks (object, embed, svg, iframe, base, link, meta). We treat content
 * containing these as HTML so it is rendered in a sandboxed iframe rather than risking
 * it being handled as plain text or markdown elsewhere.
 */
const HTML_INDICATOR_TAGS =
  /<(html|head|body|div|span|p|a|img|script|style|table|ul|ol|li|h[1-6]|object|embed|svg|iframe|base|link|meta)[\s>\/]/i

/**
 * Returns true only if the string looks like real HTML (not e.g. "a < b and c > d").
 * Used to decide whether to render in an iframe; false positives hurt performance and UX.
 *
 * - Requires valid HTML tag names (letter after <, no space) to avoid math/comparisons.
 * - For fragment-like content, requires a known HTML tag (including dangerous ones) and a closing tag.
 */
export function isHTML(text?: string): boolean {
  if (!text) return false

  // Unambiguous document structure
  if (
    /<!DOCTYPE\s+html/i.test(text) ||
    /<html[\s>]/i.test(text) ||
    /<\?xml/i.test(text)
  ) {
    return true
  }

  const trimmed = text.trim()

  // Must have at least one opening tag with a valid tag name (no "a < b" or "<1>")
  if (!VALID_HTML_TAG_NAME.test(trimmed)) {
    return false
  }

  // Fragment path: require a known HTML tag (including dangerous/embedding tags) and a closing tag
  if (!HTML_INDICATOR_TAGS.test(trimmed)) {
    return false
  }

  const hasClosingTag = /<\/([a-zA-Z][a-zA-Z0-9-]*)>/i.test(trimmed)
  return hasClosingTag
}
