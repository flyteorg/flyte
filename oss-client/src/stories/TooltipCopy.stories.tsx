'use client'

import { Tooltip } from '@/components/Tooltip'
import type { Meta, StoryObj } from '@storybook/nextjs'
import { CopyButtonWithTooltip } from '@/components/CopyButtonWithTooltip'

type Story = StoryObj<typeof Tooltip>

const meta = {
  title: 'components/CopyButtonWithTooltip',
  component: CopyButtonWithTooltip,
  tags: ['autodocs'],
  argTypes: {
    value: {
      description: 'Value to copy',
      control: { type: 'text' },
    },
    textInitial: {
      description: 'Text to display in tooltip on hover over button',
      control: { type: 'text' },
    },
    textCopied: {
      description: 'Text to display in tooltip after value is copied',
      control: { type: 'text' },
    },
    icon: {
      description: 'Icon to display in the button and in tooltip',
      options: ['chain', 'copy'],
    },
  },
  parameters: {
    docs: {
      description: {
        component: `
The \`CopyButtonWithTooltip\` component serves as simple button with Tooltip to copy value to clipboard.

## Usage

\`\`\`tsx
<CopyButtonWithTooltip 
  value="12345"
  textInitial="Copy run ID"
  textCopied="Run ID copied to clipboard"
/>
\`\`\`
        `,
      },
    },
  },
} satisfies Meta<typeof CopyButtonWithTooltip>
export default meta

export const Default: Story = {
  render: () => <CopyButtonWithTooltip value="12345" />,
}

export const Icons: Story = {
  render: () => (
    <>
      <CopyButtonWithTooltip
        icon="chain"
        value="12345"
        textInitial="Copy run ID"
        textCopied="Run ID copied to clipboard"
      />
      <CopyButtonWithTooltip
        icon="copy"
        value="12345"
        textInitial="Copy URL"
        textCopied="URL copied to clipboard"
      />
    </>
  ),
}
