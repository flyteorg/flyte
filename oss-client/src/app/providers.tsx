'use client'

import { finalTransport, isLocalDev, queryClient } from '@/lib/apiUtils'
import '@/lib/telemetry/initializeTelemetry.tsx'
import { NotificationsProvider } from '@/providers/notifications'
import { TransportProvider } from '@connectrpc/connect-query'
import { FaroErrorBoundary } from '@grafana/faro-react'
import { QueryClientProvider } from '@tanstack/react-query'
import { ReactQueryDevtools } from '@tanstack/react-query-devtools'
import { ThemeProvider } from 'next-themes'
import { NuqsAdapter } from 'nuqs/adapters/next/app'
import { errorBoundaryFallbackRenderer } from './errorPageRenderer'

export function Providers({ children }: { children: React.ReactNode }) {
  return (
    <FaroErrorBoundary fallback={errorBoundaryFallbackRenderer}>
      <ThemeProvider
        attribute="class"
        disableTransitionOnChange
        themes={['flyte']}
        defaultTheme="flyte"
        forcedTheme="flyte"
      >
        <TransportProvider transport={finalTransport}>
          <QueryClientProvider client={queryClient}>
            <NuqsAdapter>
              <NotificationsProvider>
                {children}
                {isLocalDev && (
                  <ReactQueryDevtools
                    initialIsOpen={false}
                    buttonPosition="bottom-right"
                  />
                )}
              </NotificationsProvider>
            </NuqsAdapter>
          </QueryClientProvider>
        </TransportProvider>
      </ThemeProvider>
    </FaroErrorBoundary>
  )
}
