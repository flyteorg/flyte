import { Suspense } from 'react'
import type { Meta, StoryObj } from '@storybook/react'
import { DatePickerPopover } from '@/components/DatePicker/DatePickerPopover'
import { NuqsAdapter } from 'nuqs/adapters/react'

const DatePickerWrapper = ({ children }: { children: React.ReactNode }) => (
  <Suspense fallback={<div>Loading...</div>}>
    <NuqsAdapter>{children}</NuqsAdapter>
  </Suspense>
)

const meta: Meta<typeof DatePickerPopover> = {
  title: 'Components/DatePicker/DatePickerPopover',
  component: DatePickerPopover,
  decorators: [
    (Story) => (
      <DatePickerWrapper>
        <div className="p-4">
          <Story />
        </div>
      </DatePickerWrapper>
    ),
  ],
  parameters: {
    layout: 'centered',
    docs: {
      description: {
        component:
          'A popover component that contains a date range picker with quick range buttons and calendar selection.',
      },
    },
  },
  argTypes: {},
  tags: ['autodocs'],
}

export default meta
type Story = StoryObj<typeof DatePickerPopover>

export const Default: Story = {
  args: {},
}
