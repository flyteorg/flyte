import React from 'react'
import type { Meta, StoryObj } from '@storybook/nextjs'
import { Button, styles } from '@/components/Button'
import {
  ArrowRightIcon,
  CloudArrowUpIcon,
  PlusIcon,
} from '@heroicons/react/20/solid'

const sizes = Object.keys(styles.size) as (keyof typeof styles.size)[]
const colors = Object.keys(styles.colors) as (keyof typeof styles.colors)[]

/**
 * The Button component is a versatile UI element that supports multiple variants,
 * sizes, and colors. It can be rendered as either a button or a link, and includes
 * support for icons and different visual styles.
 *
 * ## Design
 *
 * The Button component follows the [Catalyst UI design system](https://www.figma.com/design/fd9kUG1YQRJhhG7xu5Wago/Catalyst-UI-for-Figma--Community-?node-id=102-575&t=ZTzJsURGglxowJ11-0).
 *
 * ## Features
 *
 * - Multiple variants: solid, outline, and plain
 * - Various sizes from xs to xl
 * - Extensive color palette
 * - Icon support with proper spacing
 * - Link support
 * - Loading and disabled states
 * - Touch-friendly hit areas
 *
 * ## Usage
 *
 * ```tsx
 * // Basic usage
 * <Button>Click me</Button>
 *
 * // With icon
 * <Button>
 *   Next step
 *   <ArrowRightIcon data-slot="icon" />
 * </Button>
 *
 * // As link
 * <Button href="/docs">Documentation</Button>
 * ```
 */
const meta: Meta<typeof Button> = {
  title: 'primitives/Button',
  component: Button,
  tags: ['autodocs'],
  parameters: {
    docs: {
      description: {
        component: `
The Button component is a versatile UI element that supports multiple variants,
sizes, and colors. It can be rendered as either a button or a link, and includes
support for icons and different visual styles.

## Features

- Multiple variants: solid, outline, and plain
- Various sizes from xs to xl
- Extensive color palette
- Icon support with proper spacing
- Link support
- Loading and disabled states
- Touch-friendly hit areas

## Usage

\`\`\`tsx
// Basic usage
<Button>Click me</Button>

// With icon
<Button>
  Next step
  <ArrowRightIcon data-slot="icon" />
</Button>

// As link
<Button href="/docs">Documentation</Button>
\`\`\`
        `,
      },
    },
    design: {
      type: 'figma',
      url: 'https://www.figma.com/design/fd9kUG1YQRJhhG7xu5Wago/Catalyst-UI-for-Figma--Community-?node-id=102-575&t=ZTzJsURGglxowJ11-0',
    },
  },
  argTypes: {
    size: {
      description: 'The size of the button',
      control: 'select',
      options: sizes,
      table: {
        type: { summary: sizes.join(' | ') },
        defaultValue: { summary: 'md' },
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
        defaultValue: { summary: 'false' },
      },
    },
    href: {
      description: 'URL to navigate to when clicked (renders as a link)',
      control: 'text',
      table: {
        type: { summary: 'string' },
      },
    },
    children: {
      description: 'The content to display inside the button',
      control: 'text',
      table: {
        type: { summary: 'React.ReactNode' },
      },
    },
  },
  args: {
    children: 'Button text',
    size: 'md',
    color: 'dark/zinc',
  },
}

export default meta

type Story = StoryObj<typeof Button>

export const Default: Story = {
  name: 'Default (Solid)',
  args: {},
}

export const WithIcon: Story = {
  name: 'With Icon',
  args: {
    children: (
      <>
        Next step
        <ArrowRightIcon data-slot="icon" aria-hidden="true" />
      </>
    ),
  },
}

export const IconOnly: Story = {
  name: 'Icon Only',
  args: {
    children: <PlusIcon data-slot="icon" aria-hidden="true" />,
    'aria-label': 'Add item',
  },
}

export const Outline: Story = {
  name: 'Outline',
  args: {
    outline: true,
  },
}

export const OutlineWithIcon: Story = {
  name: 'Outline with Icon',
  args: {
    outline: true,
    children: (
      <>
        Upload
        <CloudArrowUpIcon data-slot="icon" aria-hidden="true" />
      </>
    ),
  },
}

export const Plain: Story = {
  name: 'Plain',
  args: {
    plain: true,
  },
}

export const PlainWithIcon: Story = {
  name: 'Plain with Icon',
  args: {
    plain: true,
    children: (
      <>
        Settings
        <ArrowRightIcon data-slot="icon" aria-hidden="true" />
      </>
    ),
  },
}

export const AsLink: Story = {
  name: 'As Link',
  args: {
    href: '#',
    children: 'Visit documentation',
  },
}

export const Sizes: Story = {
  name: 'All Sizes',
  render: (args) => (
    <div className="flex flex-wrap items-end gap-4">
      {sizes.map((size) => (
        <Button key={size} size={size} color={args.color}>
          Size: {size}
        </Button>
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
          Solid Variants
        </h3>
        <div className="flex flex-wrap gap-4">
          {colors.map((color) => (
            <Button key={color} color={color}>
              {color}
            </Button>
          ))}
        </div>
      </div>
      <div className="flex flex-col gap-2">
        <h3 className="text-sm font-semibold text-zinc-900 dark:text-white">
          Outline Variants
        </h3>
        <div className="flex flex-wrap gap-4">
          {colors.map((color) => (
            <Button key={color} color={color} outline>
              {color}
            </Button>
          ))}
        </div>
      </div>
      <div className="flex flex-col gap-2">
        <h3 className="text-sm font-semibold text-zinc-900 dark:text-white">
          Plain Variants
        </h3>
        <div className="flex flex-wrap gap-4">
          {colors.map((color) => (
            <Button key={color} color={color} plain>
              {color}
            </Button>
          ))}
        </div>
      </div>
    </div>
  ),
}

export const Disabled: Story = {
  name: 'Disabled State',
  args: {
    disabled: true,
    children: 'Disabled button',
  },
}

export const Loading: Story = {
  name: 'Loading State',
  args: {
    disabled: true,
    children: (
      <>
        <svg
          className="mr-3 -ml-1 h-5 w-5 animate-spin text-white"
          xmlns="http://www.w3.org/2000/svg"
          fill="none"
          viewBox="0 0 24 24"
          data-slot="icon"
        >
          <circle
            className="opacity-25"
            cx="12"
            cy="12"
            r="10"
            stroke="currentColor"
            strokeWidth="4"
          />
          <path
            className="opacity-75"
            fill="currentColor"
            d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
          />
        </svg>
        Processing...
      </>
    ),
  },
}
