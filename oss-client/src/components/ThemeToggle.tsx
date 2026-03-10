'use client'
import { useEffect, useState } from 'react'
import { useTheme } from 'next-themes'
import useLocalStorage from 'use-local-storage'

function SunIcon(props: React.ComponentPropsWithoutRef<'svg'>) {
  return (
    <svg viewBox="0 0 20 20" fill="none" aria-hidden="true" {...props}>
      <path d="M12.5 10a2.5 2.5 0 1 1-5 0 2.5 2.5 0 0 1 5 0Z" />
      <path
        strokeLinecap="round"
        d="M10 5.5v-1M13.182 6.818l.707-.707M14.5 10h1M13.182 13.182l.707.707M10 15.5v-1M6.11 13.889l.708-.707M4.5 10h1M6.11 6.111l.708.707"
      />
    </svg>
  )
}

function MoonIcon(props: React.ComponentPropsWithoutRef<'svg'>) {
  return (
    <svg viewBox="0 0 20 20" fill="none" aria-hidden="true" {...props}>
      <path d="M15.224 11.724a5.5 5.5 0 0 1-6.949-6.949 5.5 5.5 0 1 0 6.949 6.949Z" />
    </svg>
  )
}

export function ThemeToggle() {
  const [cachedThemePreference, setCachedThemePreference] =
    useLocalStorage<string>('cachedThemePreference', 'dark')
  const { resolvedTheme, setTheme } = useTheme()
  const otherTheme = resolvedTheme === 'dark' ? 'light' : 'dark'
  const [mounted, setMounted] = useState(false)

  const handleClick = () => {
    setTheme(otherTheme)
    setCachedThemePreference(otherTheme)
  }

  // Initialize theme from localStorage on mount
  useEffect(() => {
    setMounted(true)
    if (
      cachedThemePreference &&
      ['dark', 'light', 'system'].includes(cachedThemePreference)
    ) {
      setTheme(cachedThemePreference)
    }
  }, [cachedThemePreference, setTheme])

  return (
    <button
      type="button"
      className="flex size-9 items-center justify-center rounded-md transition hover:bg-zinc-900/5 dark:hover:bg-white/5"
      aria-label={mounted ? `Switch to ${otherTheme} theme` : 'Toggle theme'}
      onClick={handleClick}
    >
      <span className="pointer-fine:hidden absolute size-12" />
      <SunIcon className="hidden size-7 fill-(--system-gray-5) stroke-(--system-gray-5) dark:block" />
      <MoonIcon className="size-7 fill-(--system-gray-1) dark:hidden" />
    </button>
  )
}
