import { useEffect, useState } from 'react'
import { type SystemColor, type AccentColor } from '@/types/colors'

/** Default union color when --union cannot be read (SSR or before paint). Keep in sync with :root in tailwind.css (Flyte theme). */
const UNION_COLOR_DEFAULT = '#6f2aef'

/** Reads the current theme's --union from the document. Only call from effect to avoid getComputedStyle during render. */
function readUnionColorFromDocument(): string {
  if (typeof document === 'undefined') return UNION_COLOR_DEFAULT
  const value = getComputedStyle(document.documentElement)
    .getPropertyValue('--union')
    .trim()
  return value || UNION_COLOR_DEFAULT
}

type Palette = {
  dark: {
    accent: Record<AccentColor, string>
    bg: Record<AccentColor, string>
    system: Record<SystemColor, string>
    union: string
  }
  light: {
    accent: Record<AccentColor, string>
    bg: Record<AccentColor, string>
    system: Record<SystemColor, string>
    union: string
  }
}

export const palette: Palette = {
  dark: {
    accent: {
      green: '#65BD15',
      blue: '#1BA0FF',
      purple: '#C084FC',
      red: '#F43B3E',
      orange: '#FB923C',
      yellow: '#FACC15',
      gray: '#A1A1AA',
    },
    bg: {
      green: '#1A2E05',
      blue: '#172554',
      purple: '#3B0764',
      red: '#450A0A',
      orange: '#431407',
      yellow: '#422006',
      gray: '#27272A',
    },
    system: {
      white: '#FFFFFF',
      'gray-1': '#1D1D1D',
      'gray-2': '#242424',
      'gray-3': '#343434',
      'gray-4': '#3F3F46',
      'gray-5': '#71717A',
      black: '#131313',
    },
    union: UNION_COLOR_DEFAULT,
  },
  light: {
    accent: {
      green: '#57961F',
      blue: '#4791D6',
      purple: '#874FC2',
      red: '#ED3038',
      orange: '#BD5E00',
      yellow: '#8F7308',
      gray: '#6B6B70',
    },
    bg: {
      green: '#D4F0BD',
      blue: '#DBEBFA',
      purple: '#F0D6FF',
      red: '#FFCFC7',
      orange: '#FFE3C7',
      yellow: '#FFE8B0',
      gray: '#E3E3E5',
    },
    system: {
      white: '#131313',
      'gray-1': '#f5f5f5',
      'gray-2': '#ededed',
      'gray-3': '#e6e6e6',
      'gray-4': '#d4d4d4',
      'gray-5': '#8c8c8c',
      black: '#ffffff',
    },
    union: UNION_COLOR_DEFAULT,
  },
}

export const useAccent = (color: AccentColor) => {
  const theme = 'light' // Flyte theme uses light palette
  return palette[theme].accent[color]
}

export const usePalette = () => {
  const theme = 'light' // Flyte theme uses light palette
  const basePalette = palette[theme]

  const [unionColor, setUnionColor] = useState(UNION_COLOR_DEFAULT)
  useEffect(() => {
    setUnionColor(readUnionColorFromDocument())
  }, [])

  return { ...basePalette, union: unionColor }
}

export const useBackground = (color: AccentColor) => {
  const theme = 'light' // Flyte theme uses light palette
  return palette[theme].bg[color]
}

export const useSystemColor = (color: SystemColor) => {
  const theme = 'light' // Flyte theme uses light palette
  return palette[theme].system[color]
}
