import { DeleteSecretConfirmDialog } from '@/components/DeleteSecretConfirmDialog'
import {
  TableFooterLoadMore,
  VirtualizedTable,
  type ScrollToIndexFn,
} from '@/components/Tables'
import { useListSecrets } from '@/hooks/useSecrets'
import { Secret } from '@/gen/flyteidl2/secret/definition_pb'
import { useSearchParams } from 'next/navigation'
import { useCallback, useMemo, useState } from 'react'
import { useListSecretsTableColumns } from './SecretsTableColumns'
import { type SecretsTableRow } from './types'
import { formatForTable } from './util'

interface SecretsTableProps {
  secrets?: Secret[]
  secretsQuery?: ReturnType<typeof useListSecrets>
  onVirtualizerReady?: (scrollToIndex: ScrollToIndexFn) => void
}

export const SecretsTable = ({
  secrets,
  secretsQuery,
  onVirtualizerReady,
}: SecretsTableProps) => {
  const searchParams = useSearchParams()
  const returnTo = searchParams.get('returnTo')

  const [isConfirmationDialogOpen, setConfirmationDialogOpen] = useState(false)
  const [secretToDelete, setSecretToDelete] = useState<Secret | undefined>(
    undefined,
  )

  const onDeleteClick = useCallback((secret: Secret) => {
    setConfirmationDialogOpen(true)
    setSecretToDelete(secret)
  }, [])

  const columns = useListSecretsTableColumns({ onDeleteClick })
  const formattedSecrets = useMemo(() => {
    return secrets?.map(formatForTable)
  }, [secrets])

  return (
    <>
      <VirtualizedTable<SecretsTableRow>
        columns={columns}
        data={formattedSecrets || []}
        footerRow={<TableFooterLoadMore query={secretsQuery} label="secrets" />}
        rowHeight={60}
        overscan={10}
        getRowHref={(secretRow) => {
          const domain = secretRow.original.id?.domain
          const project = secretRow.original.id?.project
          const name = secretRow.original.id?.name
          const domainParam = domain ? `/domain/${domain}` : ''
          const projectParam = project ? `/project/${project}` : ''
          const baseUrl =
            `/settings/secrets/${name || ''}` + domainParam + projectParam
          if (returnTo) {
            return `${baseUrl}?returnTo=${encodeURIComponent(returnTo)}`
          }
          return baseUrl
        }}
        onVirtualizerReady={onVirtualizerReady}
      />
      <DeleteSecretConfirmDialog
        secret={secretToDelete}
        isOpen={isConfirmationDialogOpen}
        setIsOpen={setConfirmationDialogOpen}
      />
    </>
  )
}
