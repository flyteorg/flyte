'use client'

import dynamic from 'next/dynamic'
import React, { useEffect, useMemo } from 'react'
import { loader } from '@monaco-editor/react'
import { registerFlyteLightTheme, FLYTE_LIGHT_THEME } from '@/utils/monacoThemes'

type CodeViewerProps = {
  content: string
  language: string
  theme: string | undefined
  filePath: string
}

// Dynamically import Monaco Editor to avoid SSR issues and improve initial page load.
// Monaco uses browser-only APIs (Web Workers, DOM) that don't exist in Node.js, so we disable SSR.
// Code-splitting also reduces the initial bundle size.
const MonacoEditor = dynamic(
  async () => (await import('@monaco-editor/react')).default,
  {
    ssr: false,
    loading: () => (
      <div className="h-full w-full bg-zinc-50 p-2 text-xs text-zinc-500 dark:bg-zinc-900 dark:text-zinc-400">
        Loading code viewer…
      </div>
    ),
  },
)

// Configure Monaco to load from self-hosted assets instead of CDN.
// Assets are copied to public/monaco/vs during build (see postbuild:monaco script).
// Using /v2/monaco/vs (with basePath) allows us to control caching via next.config.mjs headers,
// which prevents stale cached assets from causing initialization hangs in production.
let monacoConfigured = false
const ensureMonacoConfigured = () => {
  if (monacoConfigured) return
  loader.config({
    paths: {
      vs: '/v2/monaco/vs',
    },
  })
  monacoConfigured = true
}

export const CodeViewer: React.FC<CodeViewerProps> = ({
  content,
  language,
  theme,
  filePath,
}) => {
  useEffect(() => {
    ensureMonacoConfigured()
  }, [])

  const safeTheme = theme === 'dark' || theme === 'light' ? theme : 'dark'
  const editorTheme = useMemo(
    () => (safeTheme === 'dark' ? 'vs-dark' : FLYTE_LIGHT_THEME),
    [safeTheme],
  )

  return (
    <div className="relative h-full w-full">
      <MonacoEditor
        beforeMount={registerFlyteLightTheme}
        key={filePath}
        height="100%"
        width="100%"
        path={filePath}
        language={language || 'plaintext'}
        defaultValue={content}
        theme={editorTheme}
        options={{
          readOnly: true,
          largeFileOptimizations: true,
          stopRenderingLineAfter: 20000,
          folding: false,
          minimap: { enabled: false },
          lineNumbers: 'on',
          glyphMargin: false,
          lineDecorationsWidth: 10,
          lineNumbersMinChars: 4,
          scrollBeyondLastLine: false,
          wordWrap: 'off',
          renderLineHighlight: 'none',
          overviewRulerLanes: 0,
          hideCursorInOverviewRuler: true,
          scrollbar: { vertical: 'auto', horizontal: 'auto' },
          fontSize: 12,
          quickSuggestions: false,
          suggestOnTriggerCharacters: false,
          acceptSuggestionOnEnter: 'off',
          tabCompletion: 'off',
          wordBasedSuggestions: 'off',
          occurrencesHighlight: 'off',
          selectionHighlight: false,
          renderWhitespace: 'none',
          renderValidationDecorations: 'off',
          hover: { enabled: false },
          links: false,
          colorDecorators: false,
          contextmenu: false,
          mouseWheelZoom: false,
          smoothScrolling: false,
          cursorBlinking: 'solid',
          cursorSmoothCaretAnimation: 'off',
          matchBrackets: 'never',
          codeLens: false,
          parameterHints: { enabled: false },
          maxTokenizationLineLength: 400,
          find: {
            addExtraSpaceOnTop: false,
            autoFindInSelection: 'never',
            seedSearchStringFromSelection: 'never',
          },
        }}
      />
    </div>
  )
}
