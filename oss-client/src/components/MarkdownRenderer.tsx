'use client'

import clsx from 'clsx'
import dynamic from 'next/dynamic'
import { useTheme } from 'next-themes'
import { memo } from 'react'

const MarkdownRendererInner = dynamic(
  () =>
    import('./MarkdownRendererInner').then((mod) => ({
      default: mod.MarkdownRendererInner,
    })),
  { ssr: false },
)

interface MarkdownContentProps {
  text: string
  className?: string
}

/**
 * Wrapper component to render markdown content with proper styling.
 * Shiki, ReactMarkdown, and code-block deps are loaded on first use (dynamic import).
 * Memoized so that when the parent re-renders we don't re-render as long as text is unchanged.
 */
export const MarkdownContent = memo(function MarkdownContent({
  text,
  className,
}: MarkdownContentProps) {
  const { resolvedTheme } = useTheme()

  return (
    <div
      className={clsx(
        'prose-sm prose max-w-none break-words dark:prose-invert prose-headings:font-semibold prose-p:my-2 prose-ol:my-2 prose-ul:my-2 prose-li:my-1 [&_code]:break-all [&_pre]:overflow-x-auto [&_pre]:break-all [&_pre]:whitespace-pre-wrap [&_pre_code]:bg-transparent [&_pre_code]:p-0 [&_pre_code]:shadow-none',
        className,
      )}
    >
      <MarkdownRendererInner text={text} resolvedTheme={resolvedTheme} />
    </div>
  )
})
