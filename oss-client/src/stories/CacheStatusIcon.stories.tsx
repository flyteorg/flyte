import { CatalogCacheStatus } from '@/gen/flyteidl/core/catalog_pb'
import { CacheStatusIcon } from '@/components/StatusIcons'
import type { Meta, StoryObj } from '@storybook/nextjs'
import { iconSizeMap } from '@/components/StatusIcons/iconSize'

const statusInfo: Record<
  CatalogCacheStatus,
  { label: string; description?: string; isImplemented?: boolean }
> = {
  [CatalogCacheStatus.CACHE_DISABLED]: {
    label: 'Disabled',
    isImplemented: true,
  },
  [CatalogCacheStatus.CACHE_LOOKUP_FAILURE]: {
    label: 'Lookup failure',
    isImplemented: true,
  },
  [CatalogCacheStatus.CACHE_EVICTED]: { label: 'Evicted' },
  [CatalogCacheStatus.CACHE_HIT]: { label: 'Hit', isImplemented: true },
  [CatalogCacheStatus.CACHE_POPULATED]: {
    label: 'Populated',
    isImplemented: true,
  },
  [CatalogCacheStatus.CACHE_PUT_FAILURE]: {
    label: 'Put failure',
    isImplemented: true,
  },
  [CatalogCacheStatus.CACHE_SKIPPED]: { label: 'Skipped' },
  [CatalogCacheStatus.CACHE_MISS]: { label: 'Miss' },
}

const statusOptions = Object.entries(statusInfo).reduce<
  Record<string, CatalogCacheStatus>
>((acc, [status, info]) => {
  acc[info.label] = Number(status) as CatalogCacheStatus
  return acc
}, {})

const meta = {
  title: 'components/CacheStatusIcon',
  component: CacheStatusIcon,
  tags: ['autodocs'],
  argTypes: {
    size: {
      options: Object.keys(iconSizeMap),
      description: 'Icon size',
      control: { type: 'select' },
    },
    status: {
      options: Object.keys(statusOptions),
      mapping: statusOptions,
      control: { type: 'select' },
      description: 'Current phase of the run',
      table: {
        defaultValue: { summary: 'Not Started' },
        type: {
          summary: 'Status',
          detail: Object.entries(statusInfo)
            .map(
              ([val, info]) =>
                `${info.label}: ${info.description} (${CatalogCacheStatus[Number(val)]})`,
            )
            .join('\n'),
        },
      },
    },
    disableTooltip: {
      control: { type: 'boolean' },
      description: 'Disable tooltip on hover',
      table: {
        defaultValue: { summary: 'undefined' },
      },
    },
  },
  parameters: {
    design: {
      type: 'figma',
      url: 'https://www.figma.com/file/Qj5TqXNolqZWqgbkz6zZH7/Union-Cloud?type=design&node-id=2-1055&mode=design&t=YxXRc8tHpvxu0JFh-4',
    },
    docs: {
      description: {
        component: `
Status icons represent different states of a run or task.

## States
${Object.entries(statusInfo)
  .map(
    ([val, info]) =>
      `- ${info.label} (<i>CatalogCacheStatus.${CatalogCacheStatus[Number(val)]}</i>)${info.description ? ':' : ''} ${info.description ?? ''}`,
  )
  .join('\n')}
`,
      },
    },
  },
} satisfies Meta<typeof CacheStatusIcon>

export default meta
type Story = StoryObj<typeof CacheStatusIcon>

export const Default: Story = {
  args: {
    status: CatalogCacheStatus.CACHE_PUT_FAILURE,
  },
}

export const AllStates: Story = {
  render: () => (
    <div className="flex flex-col gap-8">
      {Object.entries(statusInfo)
        .filter(([, { isImplemented }]) => !!isImplemented)
        .map(([status, { label, description }], key) => (
          <div key={key} className="flex items-center gap-4">
            <div className="flex items-center">
              <CacheStatusIcon status={Number(status)} />
            </div>
            <div className="flex flex-col">
              <div className="text-sm text-zinc-500 dark:text-zinc-200">
                {label} (
                <i>CatalogCacheStatus.{CatalogCacheStatus[Number(status)]}</i>)
              </div>
              <div className="text-xs text-zinc-400">{description}</div>
            </div>
          </div>
        ))}
    </div>
  ),
}
