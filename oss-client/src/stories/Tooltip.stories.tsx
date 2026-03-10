'use client'

import { Tooltip } from '@/components/Tooltip'
import type { Meta, StoryObj } from '@storybook/nextjs'
import { ChatBubbleLeftIcon } from '@heroicons/react/24/solid'
import { Placement } from '@floating-ui/react'

type Story = StoryObj<typeof Tooltip>

const placements: Placement[] = ['top', 'left', 'right', 'bottom']

const meta = {
  title: 'components/Tooltip',
  component: Tooltip,
  tags: ['autodocs'],
  argTypes: {
    placement: {
      options: placements,
      mapping: placements.reduce((acc, p) => ({ ...acc, [p]: p }), {}),
      control: { type: 'select' },
      description: 'Tooltip placement',
      table: {
        defaultValue: { summary: 'top' },
        type: { summary: 'placement', detail: placements.join('\n') },
      },
    },
    offsetProp: {
      control: {
        type: 'range',
        min: -100,
        max: 100,
        step: 1,
      },
      table: {
        defaultValue: { summary: '8' },
        type: { summary: 'offsetTop', detail: 'Number' },
      },
    },
    content: {
      description: 'Content inside tooltip',
      control: { type: 'text' },
      table: { type: { summary: 'string | ReactNode' } },
    },
    children: {
      description: 'Fragment that will show tooltip on-hover',
      control: { type: 'text' },
      table: { type: { summary: 'ReactNode' } },
    },
    shouldUseClientPoint: {
      description:
        'Positions the tooltip at a given client point (x, y) from a mouse event',
      control: 'boolean',
    },
  },
  parameters: {
    docs: {
      description: {
        component: `
The \`Tooltip\` is a component that shows some additional information when user hovers over a trigger element.
      `,
      },
    },
  },
} satisfies Meta<typeof Tooltip>
export default meta

export const Default: Story = {
  render: (args) => (
    <div className="bg-component text-sm/5 dark:text-white">
      <Tooltip
        {...args}
        content={args.content ?? 'Message'}
        closeDelay={args.closeDelay ?? 2000}
        offsetProp={args.offsetProp ?? 0}
      >
        {args.children ?? (
          <div className="flex w-max items-center gap-1">
            <ChatBubbleLeftIcon className="h-4 w-4" />
            Hover to see tooltip
          </div>
        )}
      </Tooltip>
    </div>
  ),
}
