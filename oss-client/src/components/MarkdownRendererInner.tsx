'use client'

import { CopyButton } from '@/components/CopyButton'
import type { CodeToHtmlOptions } from '@llm-ui/code'
import {
  allLangs,
  allLangsAlias,
  loadHighlighter,
  useCodeBlockToHtml,
} from '@llm-ui/code'
import parseHtml from 'html-react-parser'
import React, { useMemo } from 'react'
import ReactMarkdown from 'react-markdown'
import remarkBreaks from 'remark-breaks'
import remarkGfm from 'remark-gfm'
import { getSingletonHighlighterCore } from 'shiki/core'
import { bundledLanguagesInfo } from 'shiki/langs'
import { bundledThemes } from 'shiki/themes'
import getWasm from 'shiki/wasm'

let shikiHighlighter: ReturnType<typeof loadHighlighter> | null = null

function getShikiHighlighter() {
  if (!shikiHighlighter) {
    const commonLangIds = [
      'javascript',
      'typescript',
      'python',
      'bash',
      'json',
      'yaml',
      'markdown',
      'html',
      'css',
      'sql',
    ]
    const filteredLangInfo = bundledLanguagesInfo.filter((lang) =>
      commonLangIds.includes(lang.id),
    )
    const commonLangs = allLangs(filteredLangInfo)
    const langAliases = allLangsAlias(filteredLangInfo)
    const themes = [
      bundledThemes['github-dark'],
      bundledThemes['github-light'],
    ].filter(Boolean)

    shikiHighlighter = loadHighlighter(
      getSingletonHighlighterCore({
        langs: commonLangs,
        langAlias: langAliases,
        themes: themes,
        loadWasm: getWasm,
      }),
    )
  }
  return shikiHighlighter
}

interface CodeBlockProps {
  language: string
  code: string
  resolvedTheme?: string
}

const CodeBlock = ({ language, code, resolvedTheme }: CodeBlockProps) => {
  const highlighter = useMemo(() => getShikiHighlighter(), [])
  const codeBlock = useMemo(
    () => `\`\`\`${language}\n${code}\n\`\`\`\n`,
    [language, code],
  )
  const codeToHtmlOptions: CodeToHtmlOptions = {
    theme: resolvedTheme === 'dark' ? 'github-dark' : 'github-light',
  }
  const { html } = useCodeBlockToHtml({
    markdownCodeBlock: codeBlock,
    highlighter: highlighter ?? undefined,
    codeToHtmlOptions,
  })

  if (!html) {
    return (
      <div className="group relative my-3">
        <pre className="max-w-full overflow-x-auto rounded-md px-3 py-2 text-[13px] text-zinc-50">
          <code>{code}</code>
        </pre>
        <div className="absolute top-2 right-2 opacity-0 transition-opacity group-hover:opacity-100">
          <CopyButton value={code} size="sm" />
        </div>
      </div>
    )
  }

  return (
    <div className="group relative my-3">
      <div className="[&_pre]:px-4 [&_pre]:py-3">{parseHtml(html)}</div>
      <div className="absolute top-2 right-2 opacity-0 transition-opacity group-hover:opacity-100">
        <CopyButton value={code} size="sm" />
      </div>
    </div>
  )
}

export interface MarkdownRendererInnerProps {
  text: string
  resolvedTheme?: string
}

/**
 * Inner markdown renderer with Shiki code blocks. Loaded on demand when
 * MarkdownContent is first rendered (dynamic import from MarkdownRenderer.tsx).
 */
export function MarkdownRendererInner({
  text,
  resolvedTheme,
}: MarkdownRendererInnerProps) {
  return (
    <ReactMarkdown
      remarkPlugins={[remarkGfm, remarkBreaks]}
      components={{
        code({
          inline,
          className,
          children,
          ...props
        }: {
          inline?: boolean
          className?: string
          children?: React.ReactNode
        } & React.HTMLAttributes<HTMLElement>) {
          const match = /language-(\w+)/.exec(className || '')
          const rawCode = String(children ?? '').replace(/\n$/, '')

          if (inline || !match) {
            return (
              <code
                className={`rounded bg-(--system-gray-3) px-1.5 py-0.5 text-[12px] text-(--system-white) ${className ?? ''}`}
                {...props}
              >
                {children}
              </code>
            )
          }

          const language = match[1]
          return (
            <CodeBlock
              language={language}
              code={rawCode}
              resolvedTheme={resolvedTheme}
            />
          )
        },
      }}
    >
      {text}
    </ReactMarkdown>
  )
}
