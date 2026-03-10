import { useState } from 'react'
import type { Meta, StoryObj } from '@storybook/nextjs'
import { Switch } from '@/components/Switch'

const meta: Meta<typeof Switch> = {
  title: 'components/Switch',
  component: Switch,
  tags: ['autodocs'],
  parameters: {
    docs: {
      description: {
        component: `
The \`Switch\` component is a styled toggle switch.

## Usage

\`\`\`tsx
const [checked, setChecked] = useState(false)

<Switch checked={checked} onChange={setChecked} />
\`\`\`
        `,
      },
    },
  },
  argTypes: {
    checked: {
      description: 'Controls whether the switch is on or off.',
      control: { type: 'boolean' },
      table: {
        type: { summary: 'boolean' },
        defaultValue: { summary: 'false' },
      },
    },
    onChange: {
      table: {
        disable: true,
      },
    },
    color: {
      description: 'The color scheme of the switch',
      control: 'select',
      options: ['gray', 'green'],
      table: {
        type: { summary: 'gray | green' },
        defaultValue: { summary: 'gray' },
      },
    },
    size: {
      description: 'The size of the switch',
      control: 'select',
      options: ['sm', 'md'],
      defaultValue: 'md',
      table: {
        type: { summary: 'sm | md' },
        defaultValue: { summary: 'md' },
      },
    },
    disabled: {
      description: 'Controls whether the switch is disabled.',
      control: { type: 'boolean' },
      table: {
        type: { summary: 'boolean' },
        defaultValue: { summary: 'false' },
      },
    },
  },
}

export default meta

type Story = StoryObj<typeof Switch>

export const Default: Story = {
  name: 'Default',
  render: (args) => {
    const Wrapper = () => {
      const [checked, setChecked] = useState(args.checked)
      return <Switch {...args} checked={checked} onChange={setChecked} />
    }
    return <Wrapper />
  },
}

export const Colors: Story = {
  name: 'Colors',
  render: (args) => {
    const Wrapper = () => {
      const [checked, setChecked] = useState(args.checked)
      return (
        <div className="flex gap-2">
          <Switch color="gray" checked={checked} onChange={setChecked} />
          <Switch color="green" checked={checked} onChange={setChecked} />
        </div>
      )
    }
    return <Wrapper />
  },
}

export const Sizes: Story = {
  name: 'Sizes',
  render: (args) => {
    const Wrapper = () => {
      const [checked, setChecked] = useState(args.checked)
      return (
        <div className="flex gap-2">
          <Switch size="sm" checked={checked} onChange={setChecked} />
          <Switch size="md" checked={checked} onChange={setChecked} />
        </div>
      )
    }
    return <Wrapper />
  },
}

export const Disabled: Story = {
  name: 'Disabled',
  render: () => {
    const Wrapper = () => {
      return (
        <div className="flex flex-col gap-2">
          <p>Gray</p>
          <div className="flex gap-2">
            <Switch color="gray" checked={false} onChange={() => {}} disabled />
            <Switch color="gray" checked={true} onChange={() => {}} disabled />
          </div>
          <p>Green</p>
          <div className="flex gap-2">
            <Switch
              color="green"
              checked={false}
              onChange={() => {}}
              disabled
            />
            <Switch color="green" checked={true} onChange={() => {}} disabled />
          </div>
        </div>
      )
    }
    return <Wrapper />
  },
}
