import { Meta, StoryObj } from '@storybook/nextjs'
import { Input, InputGroup } from '@/components/Input'
import {
  MagnifyingGlassIcon,
  EnvelopeIcon,
  LockClosedIcon,
  CalendarIcon,
} from '@heroicons/react/24/outline'

const meta: Meta<typeof Input> = {
  title: 'primitives/Input',
  component: Input,
  parameters: {
    layout: 'centered',
    docs: {
      description: {
        component:
          'A flexible input component that supports various types, states, and icon placements.',
      },
    },
  },
  decorators: [
    (Story) => (
      <div className="w-[288px]">
        <Story />
      </div>
    ),
  ],
  tags: ['autodocs'],
}

export default meta
type Story = StoryObj<typeof Input>

/**
 * Basic text input with placeholder
 */
export const Default: Story = {
  args: {
    type: 'text',
    placeholder: 'Enter text',
  },
}

/**
 * Input with leading icon
 */
export const WithLeadingIcon: Story = {
  render: () => (
    <InputGroup>
      <MagnifyingGlassIcon data-slot="icon" />
      <Input type="search" placeholder="Search" />
    </InputGroup>
  ),
}

/**
 * Input with trailing icon
 */
export const WithTrailingIcon: Story = {
  render: () => (
    <InputGroup>
      <Input type="text" placeholder="Enter text" />
      <CalendarIcon data-slot="icon" />
    </InputGroup>
  ),
}

/**
 * Email input with icon
 */
export const Email: Story = {
  render: () => (
    <InputGroup>
      <EnvelopeIcon data-slot="icon" />
      <Input type="email" placeholder="Email address" />
    </InputGroup>
  ),
}

/**
 * Password input with icon
 */
export const Password: Story = {
  render: () => (
    <InputGroup>
      <LockClosedIcon data-slot="icon" />
      <Input type="password" placeholder="Enter password" />
    </InputGroup>
  ),
}

/**
 * Disabled state
 */
export const Disabled: Story = {
  args: {
    type: 'text',
    placeholder: 'Disabled input',
    disabled: true,
  },
}

/**
 * Invalid state
 */
export const Invalid: Story = {
  render: () => (
    <InputGroup>
      <EnvelopeIcon data-slot="icon" />
      <Input
        type="email"
        placeholder="Email address"
        defaultValue="invalid-email"
        data-invalid
      />
    </InputGroup>
  ),
}

/**
 * Date input with icon
 */
export const Date: Story = {
  render: () => (
    <InputGroup>
      <CalendarIcon data-slot="icon" />
      <Input type="date" />
    </InputGroup>
  ),
}

/**
 * Number input
 */
export const Number: Story = {
  args: {
    type: 'number',
    placeholder: 'Enter number',
    min: 0,
    max: 100,
  },
}
