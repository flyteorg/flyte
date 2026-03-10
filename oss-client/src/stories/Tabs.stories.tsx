'use client'

import type { Meta, StoryObj } from '@storybook/nextjs'
import { Tabs } from '@/components/Tabs'
import { ComponentType, useState } from 'react'

const meta = {
  title: 'components/Tabs',
  component: Tabs,
  tags: ['autodocs'],
  parameters: {
    design: {
      type: 'figma',
      url: 'https://www.figma.com/file/Qj5TqXNolqZWqgbkz6zZH7/Union-Cloud?type=design&node-id=2-1055&mode=design&t=YxXRc8tHpvxu0JFh-4',
    },
    docs: {
      description: {
        component: `
Tabs organize content into different views which can be accessed by selecting the corresponding tab.

## Typography
- Font: Title Small/Font
- Size: 14px
- Weight: 500 (Medium)
- Line height: 20px
- Letter spacing: 0.1px
- Text alignment: Center
- Vertical alignment: Middle

## Spacing
- Horizontal padding: 24px (px-6)
- Bottom margin to indicator: 12px (mb-3)
- Indicator height: 3px with rounded top corners

## Colors
- Selected: Amber 400 (#FFBA31)
- Default: Zinc 400
- Hover: Zinc 200
- Indicator hover: Zinc 700
`,
      },
    },
  },
} satisfies Meta<typeof Tabs>

export default meta
type Story = StoryObj<typeof Tabs>

const DefaultTabs = (args: React.ComponentProps<typeof Tabs>) => {
  const [currentTab, setCurrentTab] = useState('overview')
  return <Tabs {...args} currentTab={currentTab} onClickTab={setCurrentTab} />
}

const LongContentTabs = (args: React.ComponentProps<typeof Tabs>) => {
  const [currentTab, setCurrentTab] = useState('first')
  return <Tabs {...args} currentTab={currentTab} onClickTab={setCurrentTab} />
}

const CustomWidthTabs = (args: React.ComponentProps<typeof Tabs>) => {
  const [currentTab, setCurrentTab] = useState('tab1')
  return <Tabs {...args} currentTab={currentTab} onClickTab={setCurrentTab} />
}

export const Default: Story = {
  render: DefaultTabs,
  args: {
    tabs: [
      { label: 'Run', content: 'Run content', path: 'overview' },
      { label: 'Logs', content: 'Logs content', path: 'logs' },
      {
        label: 'Environment',
        content: 'Environment content',
        path: 'environment',
      },
      { label: 'Access', content: 'Access content', path: 'access' },
    ],
  },
}

export const WithLongContent: Story = {
  render: LongContentTabs,
  args: {
    tabs: [
      {
        label: 'First Tab',
        content: (
          <div className="h-96 bg-zinc-800 p-4">
            Scrollable content goes here
          </div>
        ),
        path: 'overview',
      },
      {
        label: 'Second Tab',
        content: (
          <div className="h-96 bg-zinc-800 p-4">More scrollable content</div>
        ),
        path: 'logs',
      },
    ],
  },
}

export const WithCustomWidth: Story = {
  render: CustomWidthTabs,
  args: {
    tabs: [
      { label: 'Tab 1', content: 'Content 1', path: 'overview' },
      { label: 'Tab 2', content: 'Content 2', path: 'logs' },
      { label: 'Tab 3', content: 'Content 3', path: 'access' },
    ],
  },
  parameters: {
    layout: 'padded',
    docs: {
      description: {
        story:
          'Tabs adapt to the width of their container while maintaining consistent spacing.',
      },
    },
  },
  decorators: [
    (Story: ComponentType) => (
      <div className="w-[600px] bg-zinc-900 p-4">
        <Story />
      </div>
    ),
  ],
}
