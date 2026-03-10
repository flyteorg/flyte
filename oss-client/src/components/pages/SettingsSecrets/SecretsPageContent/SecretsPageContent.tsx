import { type ScrollToIndexFn } from '@/components/Tables'
import { ListSecretsResponse } from '@/gen/flyteidl2/secret/payload_pb'
import { useListSecrets } from '@/hooks/useSecrets'
import { useRouter, useSearchParams } from 'next/navigation'
import { useEffect, useMemo, useRef } from 'react'
import { SettingsTablePageLayout } from '../../SettingsUserManagement/SettingsTablePageLayout'
import { SecretsTable } from '.'

export const SecretsPageContent = ({
  setSearchTerm,
  searchTermInput,
  searchTerm,
  secretsQuery,
}: {
  setSearchTerm: (newValue: string | null) => void
  searchTermInput: string
  searchTerm: string
  secretsQuery: ReturnType<typeof useListSecrets>
}) => {
  // Flatten all pages into a single array of secrets
  const allSecrets = useMemo(
    () =>
      secretsQuery.data?.pages?.flatMap(
        (page: ListSecretsResponse) => page.secrets ?? [],
      ) ?? [],
    [secretsQuery.data?.pages],
  )

  const searchParams = useSearchParams()
  const router = useRouter()
  const scrollToIndexRef = useRef<ScrollToIndexFn | null>(null)
  const hasScrolledRef = useRef(false)

  // Filter secrets by search term on client side
  const filteredSecrets = useMemo(() => {
    if (!searchTerm) return allSecrets

    const searchLower = searchTerm.toLowerCase()
    return allSecrets.filter((s) =>
      (s.id?.name || '').toLowerCase().includes(searchLower),
    )
  }, [allSecrets, searchTerm])

  // Handle scrolling to a specific secret when returning from detail page
  useEffect(() => {
    const scrollToSecretName = searchParams.get('scrollTo')
    if (
      scrollToSecretName &&
      scrollToIndexRef.current &&
      filteredSecrets.length > 0 &&
      !hasScrolledRef.current
    ) {
      // Find the index of the secret with this name
      const secretIndex = filteredSecrets.findIndex(
        (s) => s.id?.name === scrollToSecretName,
      )

      if (secretIndex !== -1) {
        // Scroll to the secret after a short delay to ensure table is rendered
        setTimeout(() => {
          scrollToIndexRef.current?.(secretIndex, {
            align: 'start',
            behavior: 'auto',
          })
          hasScrolledRef.current = true

          // Remove scrollTo parameter from URL after scrolling, but preserve returnTo and tab
          const params = new URLSearchParams(searchParams.toString())
          params.delete('scrollTo')
          const newQueryString = params.toString()
          const newUrl = newQueryString
            ? `/settings/secrets?${newQueryString}`
            : '/settings/secrets'
          router.replace(newUrl)
        }, 100)
      } else {
        // Secret not found, just remove the scrollTo parameter
        const params = new URLSearchParams(searchParams.toString())
        params.delete('scrollTo')
        const newQueryString = params.toString()
        const newUrl = newQueryString
          ? `/settings/secrets?${newQueryString}`
          : '/settings/secrets'
        router.replace(newUrl)
      }
    }
  }, [filteredSecrets, searchParams, router])

  // Reset scroll flag when data changes significantly (e.g., new search)
  useEffect(() => {
    hasScrolledRef.current = false
  }, [searchTerm])

  return (
    <SettingsTablePageLayout
      count={filteredSecrets.length}
      countLabel="Secrets"
      searchPlaceholder="Search secrets"
      searchQuery={searchTerm}
      searchValue={searchTermInput ?? undefined}
      onSearchChange={(e) => setSearchTerm(e.target.value)}
      data={filteredSecrets}
      dataLabel="secrets"
      subtitle="Create a secret to see it displayed here"
      isError={secretsQuery.isError}
      isLoading={secretsQuery.isLoading}
    >
      {(data) => (
        <SecretsTable
          secrets={data}
          secretsQuery={secretsQuery}
          onVirtualizerReady={(scrollToIndex) => {
            scrollToIndexRef.current = scrollToIndex
          }}
        />
      )}
    </SettingsTablePageLayout>
  )
}
