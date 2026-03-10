import type { Meta, StoryObj } from '@storybook/react'
import Drawer, { DrawerProps, drawerSizes } from '@/components/Drawer'
import { useState } from 'react'
import { Button } from '@/components/Button.tsx'
import { Tabs } from '@/components/Tabs'

const sampleTabLabels = ['Test', 'Another one', 'And another']

const meta: Meta<typeof Drawer> = {
  title: 'primitives/Drawer',
  component: Drawer,
  tags: ['autodocs'],
  parameters: {
    tabs: {},
  },
  argTypes: {
    title: {
      description: 'Drawer title',
      type: {
        name: 'string',
        required: true,
      },
    },
    isOpen: {
      description: 'True if Drawer is open',
      control: 'boolean',
      table: {
        defaultValue: { summary: 'false' },
      },
      type: {
        name: 'boolean',
        required: true,
      },
    },
    setIsOpen: {
      description: 'Function called when opened or closed',
      control: false,
      type: {
        name: 'function',
        required: true,
      },
    },
    tabs: {
      description: 'Component with tabs',
      control: false,
    },
    children: {
      control: false,
      description: 'Drawer content',
    },
    size: {
      description: 'Drawer width: tailwind sizes or custom px width',
      control: 'select',
      options: Object.keys(drawerSizes),
      table: {
        defaultValue: { summary: 'md' },
      },
    },
    hasFullscreen: {
      description: 'Whether Drawer has fullscreen button',
      control: 'boolean',
      table: {
        type: { summary: 'boolean' },
        defaultValue: { summary: 'false' },
      },
      type: {
        name: 'boolean',
        required: false,
      },
    },
  },
}
export default meta

type Story = StoryObj<typeof Drawer>

const DrawerContainer: React.FC<DrawerProps & { btnText?: string }> = ({
  btnText,
  ...props
}) => {
  const [isOpen, setIsOpen] = useState<boolean>(false)

  return (
    <>
      <Drawer {...props} isOpen={isOpen} setIsOpen={setIsOpen} />
      <Button
        color="union"
        className="mb-3 text-sm/5"
        onClick={() => setIsOpen((prev) => !prev)}
      >
        Open {btnText}
      </Button>
    </>
  )
}

export const WithTabs: Story = {
  name: 'WithTabs',
  render: (args) => <DrawerContainer {...args} />,
  args: {
    title: 'Default drawer',
    tabs: (
      <Tabs
        currentTab={sampleTabLabels[0]}
        onClickTab={() => {}}
        tabs={sampleTabLabels.map((label) => ({
          label,
          path: label,
          content: <div className="mb-5 h-full w-full p-3">{label}</div>,
        }))}
      />
    ),
  },
}

export const WithChildren: Story = {
  name: 'WithChildren',
  render: (args) => <DrawerContainer {...args} />,
  args: {
    title: 'Drawer with children',
    children: sampleTabLabels.map((label, key) => (
      <div key={key} className="px-2 py-1 text-sm/5 text-zinc-200">
        {label}
      </div>
    )),
  },
}

export const WithFullscreenBtn: Story = {
  name: 'WithFullscreenBtn',
  render: (args) => <DrawerContainer {...args} />,
  args: {
    title: 'Drawer with fullscreen button',
    hasFullscreen: true,
  },
}

export const DrawerSizes: Story = {
  name: 'DrawerSizes',
  render: (args) => (
    <>
      <DrawerContainer
        size={700}
        btnText={'(custom px size, here 700px)'}
        {...args}
      />
      <DrawerContainer size="sm" btnText={'(size "sm")'} {...args} />
    </>
  ),
  args: { title: 'Drawer sizes' },
}
