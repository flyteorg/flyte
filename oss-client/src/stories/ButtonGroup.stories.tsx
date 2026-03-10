import ButtonGroup from '@/components/ButtonGroup'
import type { Meta, StoryObj } from '@storybook/nextjs'
import { useState } from 'react'

const meta: Meta<typeof ButtonGroup> = {
  title: 'primitives/ButtonGroup',
  component: ButtonGroup,
  parameters: {
    layout: 'centered',
  },
  tags: ['autodocs'],
}

export default meta
type Story = StoryObj<typeof ButtonGroup>

const viewOptions = [
  { label: 'Pretty', value: 'pretty' },
  { label: 'Raw', value: 'raw' },
]

const timeOptions = [
  { label: 'Days', value: 'days' },
  { label: 'Weeks', value: 'weeks' },
  { label: 'Months', value: 'months' },
  { label: 'Years', value: 'years' },
]

export const Default: Story = {
  args: {
    options: viewOptions,
    value: 'pretty',
    size: 'sm',
    color: 'dark/zinc',
  },
}

export const WithInteraction: Story = {
  render: () => {
    // eslint-disable-next-line react-hooks/rules-of-hooks
    const [value, setValue] = useState('pretty')
    return (
      <ButtonGroup
        options={viewOptions}
        value={value}
        onChange={setValue}
        size="sm"
        color="dark/zinc"
      />
    )
  },
}

export const MultipleOptions: Story = {
  args: {
    options: timeOptions,
    value: 'days',
    size: 'sm',
    color: 'dark/zinc',
  },
}

export const Sizes: Story = {
  render: () => {
    return (
      <div className="flex flex-col gap-4">
        <ButtonGroup
          options={viewOptions}
          value="pretty"
          size="xs"
          color="dark/zinc"
        />
        <ButtonGroup
          options={viewOptions}
          value="pretty"
          size="sm"
          color="dark/zinc"
        />
        <ButtonGroup
          options={viewOptions}
          value="pretty"
          size="md"
          color="dark/zinc"
        />
        <ButtonGroup
          options={viewOptions}
          value="pretty"
          size="lg"
          color="dark/zinc"
        />
        <ButtonGroup
          options={viewOptions}
          value="pretty"
          size="xl"
          color="dark/zinc"
        />
      </div>
    )
  },
}

export const Colors: Story = {
  render: () => {
    return (
      <div className="flex flex-col gap-4">
        <ButtonGroup
          options={viewOptions}
          value="pretty"
          size="sm"
          color="dark/zinc"
        />
        <ButtonGroup
          options={viewOptions}
          value="pretty"
          size="sm"
          color="light"
        />
        <ButtonGroup
          options={viewOptions}
          value="pretty"
          size="sm"
          color="dark/white"
        />
        <ButtonGroup
          options={viewOptions}
          value="pretty"
          size="sm"
          color="dark"
        />
        <ButtonGroup
          options={viewOptions}
          value="pretty"
          size="sm"
          color="white"
        />
        <ButtonGroup
          options={viewOptions}
          value="pretty"
          size="sm"
          color="zinc"
        />
      </div>
    )
  },
}
