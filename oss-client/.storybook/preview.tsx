import React from 'react'
import type { Decorator, Preview } from '@storybook/react'
import '@/styles/tailwind.css'
import { ThemeProvider } from 'next-themes'
import { NuqsAdapter } from 'nuqs/adapters/next/app'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { stringify } from 'safe-stable-stringify'

// Replace JSON.stringify with safe-stable-stringify to handle circular references
// details: https://github.com/storybookjs/storybook/issues/22452
const originalJSONStringify = JSON.stringify
JSON.stringify = function (value, replacer, space) {
  try {
    const result = stringify(value, replacer, space)
    return result || originalJSONStringify(value, replacer, space)
  } catch (error) {
    // Fallback to original if stringify fails
    return originalJSONStringify(value, replacer, space)
  }
}

const withThemeProvider: Decorator = (Story, context) => {
  const queryClient = new QueryClient()

  // Get theme from Storybook's theme parameter
  const theme = context.globals?.theme || 'dark'
  const backgroundColor = theme === 'dark' ? '#1a1a1a' : '#ffffff'

  // Update document background based on theme
  React.useEffect(() => {
    const body = document.body
    const html = document.documentElement
    const backgroundColor = theme === 'dark' ? '#1a1a1a' : '#ffffff'

    // Apply to body and html immediately
    body.style.backgroundColor = backgroundColor
    if (theme === 'dark') {
      html.classList.add('dark')
      html.classList.remove('light')
    } else {
      html.classList.add('light')
      html.classList.remove('dark')
    }

    // Apply background to Storybook's canvas and docs areas with a small delay
    // to ensure DOM elements are available
    const applyBackgrounds = () => {
      const selectors = [
        '[data-testid="storybook-canvas"]',
        '.docs-story',
        '.docs-wrapper',
        '.docs-container',
        '.docs-page',
        '.sb-show-main',
      ]

      selectors.forEach(selector => {
        const elements = document.querySelectorAll(selector)
        elements.forEach(element => {
          if (element instanceof HTMLElement) {
            element.style.backgroundColor = backgroundColor
          }
        })
      })
    }

    // Apply immediately and also after a short delay
    applyBackgrounds()
    const timeoutId = setTimeout(applyBackgrounds, 100)

    return () => clearTimeout(timeoutId)
  }, [theme])

  return (
    <QueryClientProvider client={queryClient}>
      <NuqsAdapter>
        <ThemeProvider
          attribute="class"
          defaultTheme={theme}
          forcedTheme={theme}
          enableSystem={false}
        >
          <div
            className={`${theme === 'dark' ? 'dark' : ''}`}
            style={{
              backgroundColor,
              width: '100%'
            }}
          >
            <Story />
          </div>
        </ThemeProvider>
      </NuqsAdapter>
    </QueryClientProvider>
  )
}

const preview: Preview = {
  decorators: [withThemeProvider],
  parameters: {
    controls: {
      matchers: {
        color: /(background|color)$/i,
        date: /Date$/i,
      },
    },
    nextjs: {
      appDirectory: true, // https://storybook.js.org/docs/get-started/frameworks/nextjs#set-nextjsappdirectory-to-true
    },
    backgrounds: {
      default: 'dark',
      values: [
        { name: 'light', value: '#ffffff' },
        { name: 'dark', value: '#1a1a1a' },
      ],
      grid: {
        disable: true,
      },
    },

  },
  globalTypes: {
    theme: {
      description: 'Global theme for components',
      defaultValue: 'dark',
      toolbar: {
        title: 'Theme',
        icon: 'circlehollow',
        items: [
          { value: 'light', title: 'Light', icon: 'sun' },
          { value: 'dark', title: 'Dark', icon: 'moon' },
        ],
        dynamicTitle: true,
      },
    },
  },
}

export default preview
