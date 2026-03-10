import React from 'react'
import type { Meta, StoryObj } from '@storybook/nextjs'
import { CopyButton } from '@/components/CopyButton'
import { styles } from '@/components/Button'

const sizes = ['xs', 'sm', 'md', 'lg'] as const
const colors = Object.keys(styles.colors) as (keyof typeof styles.colors)[]

/**
 * The CopyButton component is a specialized button for copying text to the clipboard.
 * It provides visual feedback when text is copied and supports various sizes and styles.
 *
 * ## Features
 *
 * - Visual feedback when text is copied
 * - Multiple sizes
 * - Various color schemes
 * - Outline and plain variants
 * - Customizable title
 *
 * ## Usage
 *
 * ```tsx
 * // Basic usage
 * <CopyButton value="Text to copy" />
 *
 * // With custom size and color
 * <CopyButton
 *   value="Text to copy"
 *   size="md"
 *   color="indigo"
 * />
 * ```
 */
const meta: Meta<typeof CopyButton> = {
  title: 'primitives/CopyButton',
  component: CopyButton,
  tags: ['autodocs'],
  parameters: {
    docs: {
      description: {
        component: `
The CopyButton component is a specialized button for copying text to the clipboard.
It provides visual feedback when text is copied and supports various sizes and styles.

## Features

- Visual feedback when text is copied
- Multiple sizes
- Various color schemes
- Outline and plain variants
- Customizable title

## Usage

\`\`\`tsx
// Basic usage
<CopyButton value="Text to copy" />

// With custom size and color
<CopyButton 
  value="Text to copy"
  size="md"
  color="indigo"
/>
\`\`\`
        `,
      },
    },
  },
  argTypes: {
    size: {
      description: 'The size of the button',
      control: 'select',
      options: sizes,
      table: {
        type: { summary: sizes.join(' | ') },
        defaultValue: { summary: 'xs' },
      },
    },
    color: {
      description: 'The color scheme of the button',
      control: 'select',
      options: colors,
      table: {
        type: { summary: colors.join(' | ') },
        defaultValue: { summary: 'dark/zinc' },
      },
    },
    outline: {
      description: 'Whether to render the button in outline style',
      control: { type: 'boolean' },
      table: {
        type: { summary: 'boolean' },
        defaultValue: { summary: 'false' },
      },
    },
    plain: {
      description: 'Whether to render the button in plain style',
      control: { type: 'boolean' },
      table: {
        type: { summary: 'boolean' },
        defaultValue: { summary: 'true' },
      },
    },
    value: {
      description: 'The text to copy to clipboard',
      control: 'text',
      table: {
        type: { summary: 'string' },
      },
    },
    title: {
      description: 'The tooltip text shown on hover',
      control: 'text',
      table: {
        type: { summary: 'string' },
        defaultValue: { summary: 'Copy to clipboard' },
      },
    },
  },
  args: {
    value: 'Text to copy',
    size: 'xs',
    color: 'dark/zinc',
    plain: true,
  },
}

export default meta

type Story = StoryObj<typeof CopyButton>

export const Default: Story = {
  name: 'Default (Plain)',
  args: {},
}

export const Outline: Story = {
  name: 'Outline',
  args: {
    outline: true,
    plain: false,
  },
}

export const Sizes: Story = {
  name: 'All Sizes',
  render: (args) => (
    <div className="flex flex-wrap items-end gap-4">
      {sizes.map((size) => (
        <CopyButton
          key={size}
          size={size}
          color={args.color}
          value={`Size: ${size}`}
        />
      ))}
    </div>
  ),
}

export const Colors: Story = {
  name: 'Color Variants',
  render: () => (
    <div className="flex flex-col gap-8">
      <div className="flex flex-col gap-2">
        <h3 className="text-sm font-semibold text-zinc-900 dark:text-white">
          Plain Variants
        </h3>
        <div className="flex flex-wrap gap-4">
          {colors.map((color) => (
            <CopyButton key={color} color={color} value={`Color: ${color}`} />
          ))}
        </div>
      </div>
      <div className="flex flex-col gap-2">
        <h3 className="text-sm font-semibold text-zinc-900 dark:text-white">
          Outline Variants
        </h3>
        <div className="flex flex-wrap gap-4">
          {colors.map((color) => (
            <CopyButton
              key={color}
              color={color}
              outline
              plain={false}
              value={`Color: ${color}`}
            />
          ))}
        </div>
      </div>
    </div>
  ),
}
