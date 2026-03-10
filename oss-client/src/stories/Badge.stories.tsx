import React from 'react'
import { Meta, StoryObj } from '@storybook/nextjs'
import { Badge, colors } from '@/components/Badge'

/**
 * The `Badge` component is used to highlight and label content with short text or status indicators.
 * It's commonly used for tags, status indicators, counts, or labels.
 */
const meta: Meta<typeof Badge> = {
  title: 'primitives/Badge',
  component: Badge,
  tags: ['autodocs'],
  parameters: {
    design: {
      type: 'figma',
      url: 'https://www.figma.com/design/fd9kUG1YQRJhhG7xu5Wago/Catalyst-UI-for-Figma--Community-?node-id=105-10747&t=NaAuuoPAN5PKL9Lz-0',
    },
    docs: {
      description: {
        component: `
The Badge component is a versatile element for displaying short pieces of information.

## Features
- Multiple color variants using Tailwind colors
- Consistent padding and rounded corners
- Accessible contrast ratios
- Customizable via className prop

## Usage
\`\`\`tsx
import { Badge } from '@/components/Badge'

// Basic usage
<Badge>Default</Badge>

// With color
<Badge color="emerald">Success</Badge>

// With custom class
<Badge className="uppercase">New</Badge>
\`\`\`
        `,
      },
    },
  },
  argTypes: {
    color: {
      description: 'The color variant of the badge',
      control: 'select',
      options: Object.keys(colors),
      table: {
        type: { summary: Object.keys(colors).join(' | ') },
        defaultValue: { summary: 'zinc' },
      },
    },
    className: {
      description: 'Additional CSS classes to apply',
      control: 'text',
    },
    children: {
      description: 'The content to display inside the badge',
      control: 'text',
    },
  },
}

export default meta

type Story = StoryObj<typeof Badge>

export const Default: Story = {
  args: {
    children: 'Badge',
  },
}

export const Colors: Story = {
  render: () => (
    <div className="flex flex-wrap gap-4">
      {(Object.keys(colors) as Array<keyof typeof colors>).map((color) => (
        <Badge key={color} color={color}>
          {color}
        </Badge>
      ))}
    </div>
  ),
}

export const Status: Story = {
  render: () => (
    <div className="flex flex-wrap gap-4">
      <Badge color="emerald">Active</Badge>
      <Badge color="amber">Pending</Badge>
      <Badge color="red">Failed</Badge>
      <Badge color="zinc">Inactive</Badge>
      <Badge color="blue">In Progress</Badge>
    </div>
  ),
}

export const WithCustomClass: Story = {
  args: {
    children: 'FEATURED',
    color: 'indigo',
    className: 'uppercase font-bold tracking-wider',
  },
}

export const WithLongText: Story = {
  args: {
    children: 'This is a badge with longer text that should wrap nicely',
    className: 'max-w-[200px]',
  },
}

export const Semantic: Story = {
  render: () => (
    <div className="flex flex-wrap gap-4">
      <Badge color="green">Success</Badge>
      <Badge color="yellow">Warning</Badge>
      <Badge color="red">Error</Badge>
      <Badge color="blue">Info</Badge>
      <Badge color="purple">Beta</Badge>
    </div>
  ),
}

export const Counts: Story = {
  render: () => (
    <div className="flex flex-wrap gap-4">
      <Badge color="zinc">Comments (23)</Badge>
      <Badge color="blue">Updates (5)</Badge>
      <Badge color="emerald">Approved (12)</Badge>
      <Badge color="amber">Pending (3)</Badge>
      <Badge color="red">Rejected (2)</Badge>
    </div>
  ),
}
