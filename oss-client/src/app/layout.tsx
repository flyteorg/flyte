import { Providers } from '@/app/providers'
import { GoogleTagManager } from '@next/third-parties/google'
import { type Metadata } from 'next'

import '@/styles/tailwind.css'

import { EnvScript, env } from 'next-runtime-env'

export const metadata: Metadata = {
  title: {
    template: '%s | Union.ai',
    default: 'Union.ai',
  },
}

export default async function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  const envLoader = {
    NODE_ENV: env('NODE_ENV'),
    GTM_ENV: env('GTM_ENV'),
    GTM_PREVIEW: env('GTM_PREVIEW'),
    GTM_ENV2: env('NEXT_PUBLIC_GTM_ENV'),
    GIT_SHA: env('GIT_SHA'),
    UNION_PLATFORM_ENV: env('UNION_PLATFORM_ENV'),
    UNION_PLATFORM_REGION: env('UNION_PLATFORM_REGION'),
    UNION_ORG_OVERRIDE: env('NEXT_PUBLIC_UNION_ORG_OVERRIDE'),
  }

  return (
    <html lang="en" className="h-full min-h-full" suppressHydrationWarning>
      {envLoader.NODE_ENV !== 'development' && (
        <GoogleTagManager
          gtmId="GTM-58DDT8M"
          auth={env('GTM_ENV')}
          preview={env('GTM_PREVIEW')}
        />
      )}
      <head>
        <meta charSet="UTF-8"></meta>
        <EnvScript env={envLoader} disableNextScript />
      </head>
      <body className="light:bg-white flex h-full overflow-hidden overscroll-none antialiased">
        <Providers>
          <div className="h-full min-h-0 w-full">{children}</div>
        </Providers>
      </body>
    </html>
  )
}
