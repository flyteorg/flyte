'use client'

import { QueryClientProvider } from '@tanstack/react-query'
import { TransportProvider } from '@connectrpc/connect-query'
import { LogsIcon } from '@/components/icons/LogsIcon'
import { Header } from '@/components/Header'
import { Logo } from '@/components/Logo'
import { ArrowPathIcon } from '@heroicons/react/16/solid'
import { Button } from '@/components/Button'
import { finalTransport, queryClient } from '@/lib/apiUtils'

export function errorBoundaryFallbackRenderer(
  error: unknown,
  resetBoundary: () => void,
) {
  console.error('Faro captured error:', error)

  return (
    <QueryClientProvider client={queryClient}>
      <TransportProvider transport={finalTransport}>
        <main className="bg-primary relative flex h-full w-full flex-col">
          <Header logoComponent={<Logo width={18} />} showSearch={false} />

          <div className="flex flex-1 items-center justify-center">
            <div className="flex max-w-[300px] flex-col items-center px-6 text-center">
              <div className="flex items-center justify-center gap-2 text-sm text-zinc-500">
                <LogsIcon />
                An unexpected error occurred
              </div>

              <pre className="mt-4 font-mono text-sm whitespace-pre-wrap text-zinc-300">
                {error instanceof Error ? error.message : String(error)}
              </pre>

              <p className="mt-4 text-sm leading-relaxed text-zinc-500">
                Reloading the page may resolve the issue. If not, please contact
                your administrator or support for assistance.
              </p>

              <Button plain onClick={resetBoundary} className="mt-6">
                <ArrowPathIcon />
                Reload
              </Button>
            </div>
          </div>
        </main>
      </TransportProvider>
    </QueryClientProvider>
  )
}
