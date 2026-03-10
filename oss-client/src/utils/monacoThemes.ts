import type { Monaco } from '@monaco-editor/react'

const FLYTE_LIGHT_THEME = 'flyte-light'

/**
 * Registers the Flyte light theme for Monaco editor. Use in beforeMount so the
 * theme is available when the editor mounts. Improves contrast and readability
 * over the default vs-light (e.g. for JSON in the launch form).
 */
export function registerFlyteLightTheme(monaco: Monaco): void {
  monaco.editor.defineTheme(FLYTE_LIGHT_THEME, {
    base: 'vs',
    inherit: true,
    rules: [
      { token: 'string', foreground: '0563c1' },
      { token: 'number', foreground: '0d9488' },
      { token: 'key', foreground: '7c2d12' },
      { token: 'delimiter', foreground: '374151' },
    ],
    colors: {
      'editor.background': '#ffffff',
      'editor.foreground': '#1f2937',
      'editorLineNumber.foreground': '#6b7280',
      'editorCursor.foreground': '#1f2937',
      'editor.selectionBackground': '#dbeafe',
    },
  })
}

export { FLYTE_LIGHT_THEME }
