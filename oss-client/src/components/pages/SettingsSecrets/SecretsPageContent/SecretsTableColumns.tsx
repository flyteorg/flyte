import { useCopyToClipboard } from '@/components/CopyButton'
import { PopoverMenu } from '@/components/Popovers'
import { Secret } from '@/gen/flyteidl2/secret/definition_pb'
import { createColumnHelper } from '@tanstack/react-table'
import { useMemo } from 'react'
import { type SecretsTableRow } from './types'

export const useListSecretsTableColumns = ({
  onDeleteClick,
}: {
  onDeleteClick: (id: Secret) => void
}) => {
  const helper = createColumnHelper<SecretsTableRow>()
  const { handleCopy } = useCopyToClipboard({})

  return useMemo(() => {
    return [
      helper.accessor('name', {
        header: 'Secret name',
        cell: (info) => (
          <div className="text-xs font-semibold text-zinc-950 dark:text-(--system-white)">
            {info.getValue()}
          </div>
        ),
        minSize: 300,
      }),
      helper.accessor('scope', {
        header: 'Access level',
        cell: (info) => (
          <div className="flex flex-col">
            <span className="text-xs font-medium text-zinc-950 capitalize dark:text-(--system-gray-7)">
              {info.getValue().title}
            </span>
            <span className="text-[11px] font-medium text-zinc-950 capitalize dark:text-(--system-gray-6)">
              {info.getValue().subtitle}
            </span>
          </div>
        ),
        minSize: 280,
        size: 280,
      }),
      helper.accessor('actions', {
        cell: (info) => (
          <div
            className="ml-auto w-max"
            onClick={(e) => {
              // prevent navigation
              e.preventDefault()
              e.stopPropagation()
            }}
          >
            <PopoverMenu
              portal
              variant="overflow"
              itemCustomClassName="!px-0 !py-0"
              items={[
                {
                  id: 'copy-secret-key',
                  type: 'item',
                  label: 'Copy secret key',
                  onClick: (e) =>
                    handleCopy(e, info.getValue()?.id?.name ?? ''),
                },
                {
                  id: 'divider',
                  type: 'divider',
                },
                {
                  id: 'delete-secret',
                  type: 'custom',
                  component: (
                    <button
                      onClick={() => onDeleteClick(info.getValue())}
                      className="ml-0 w-full cursor-pointer px-4.5 py-2 text-start text-sm text-(--accent-red) hover:bg-(--system-gray-3) dark:hover:bg-(--system-gray-3)"
                    >
                      Delete
                    </button>
                  ),
                },
              ]}
            />
          </div>
        ),
        header: '',
        minSize: 50,
        size: 50,
      }),
    ]
  }, [handleCopy, helper, onDeleteClick])
}
