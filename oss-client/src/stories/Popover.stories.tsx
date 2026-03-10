import type { Meta, StoryObj } from '@storybook/react'
import { Popover } from '@/components/Popovers/Popover'

const meta = {
  title: 'Components/Popover',
  component: Popover,
  parameters: {
    layout: 'centered',
    docs: {
      description: {
        component:
          'A flexible popover component built with floating-ui that supports click and hover triggers, multiple placements, and both controlled and uncontrolled state management.',
      },
    },
  },
  tags: ['autodocs'],
  argTypes: {
    placement: {
      control: 'select',
      options: [
        'top',
        'top-start',
        'top-end',
        'right',
        'right-start',
        'right-end',
        'bottom',
        'bottom-start',
        'bottom-end',
        'left',
        'left-start',
        'left-end',
      ],
      description: 'Preferred placement of the popover relative to the trigger',
    },
    trigger: {
      control: 'select',
      options: ['click', 'hover'],
      description: 'How the popover should be triggered',
    },
    offset: {
      control: 'number',
      description: 'Distance between the trigger and popover content',
    },
    disabled: {
      control: 'boolean',
      description: 'Whether the popover is disabled',
    },
    portal: {
      control: 'boolean',
      description: 'Whether to render the popover in a portal',
    },
  },
} satisfies Meta<typeof Popover>

export default meta
type Story = StoryObj<typeof meta>

// Sample content components for reuse
const SimpleTooltip = ({
  title,
  description,
}: {
  title: string
  description: string
}) => (
  <div className="max-w-xs rounded-lg bg-gray-900 p-3 text-white shadow-lg">
    <div className="mb-1 text-sm font-medium">{title}</div>
    <div className="text-xs opacity-90">{description}</div>
  </div>
)

const InfoCard = ({
  title,
  children,
}: {
  title: string
  children: React.ReactNode
}) => (
  <div className="max-w-sm rounded-lg border border-gray-200 bg-white p-4 shadow-lg">
    <h3 className="mb-2 font-semibold text-gray-900">{title}</h3>
    <div className="text-sm text-gray-600">{children}</div>
  </div>
)

const ActionMenu = ({ onAction }: { onAction: (action: string) => void }) => (
  <div className="min-w-40 rounded-lg border border-gray-200 bg-white py-2 shadow-lg">
    {['Edit', 'Duplicate', 'Archive', 'Delete'].map((action) => (
      <button
        key={action}
        onClick={() => onAction(action)}
        className={`w-full px-4 py-2 text-left text-sm transition-colors hover:bg-gray-50 ${
          action === 'Delete' ? 'text-red-600 hover:bg-red-50' : 'text-gray-700'
        }`}
      >
        {action}
      </button>
    ))}
  </div>
)

export const Default: Story = {
  args: {
    placement: 'bottom',
    trigger: 'click',
    offset: 8,
    disabled: false,
    portal: true,
    children: <button>Placeholder</button>,
    content: <div>Placeholder content</div>,
  },
  render: (args) => (
    <div className="p-8">
      <Popover
        {...args}
        content={
          <InfoCard title="Default Popover">
            This is the default popover configuration. It opens on click and
            positions itself below the trigger.
          </InfoCard>
        }
      >
        <button className="rounded-md bg-blue-500 px-4 py-2 text-white transition-colors hover:bg-blue-600">
          Click me
        </button>
      </Popover>
    </div>
  ),
}

// Click trigger examples
export const ClickTrigger = {
  render: () => (
    <div className="space-y-6 p-8">
      <h3 className="mb-4 text-lg font-semibold">Click Trigger Examples</h3>

      <div className="flex flex-wrap gap-4">
        <Popover
          content={
            <InfoCard title="Basic Click">
              Click anywhere outside to dismiss this popover.
            </InfoCard>
          }
          trigger="click"
          placement="bottom"
        >
          <button className="rounded-md bg-blue-500 px-4 py-2 text-white hover:bg-blue-600">
            Basic Click
          </button>
        </Popover>

        <Popover
          content={
            <ActionMenu onAction={(action) => alert(`Selected: ${action}`)} />
          }
          trigger="click"
          placement="bottom-start"
        >
          <button className="flex items-center gap-2 rounded-md bg-gray-600 px-4 py-2 text-white hover:bg-gray-700">
            Actions Menu
            <svg
              className="h-4 w-4"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M19 9l-7 7-7-7"
              />
            </svg>
          </button>
        </Popover>

        <Popover
          content={
            <div className="max-w-xs rounded-lg border border-yellow-200 bg-yellow-50 p-4">
              <div className="flex items-start gap-3">
                <div className="mt-0.5 text-yellow-600">⚠️</div>
                <div>
                  <h4 className="font-medium text-yellow-800">Warning</h4>
                  <p className="mt-1 text-sm text-yellow-700">
                    This action cannot be undone. Are you sure you want to
                    continue?
                  </p>
                </div>
              </div>
            </div>
          }
          trigger="click"
          placement="top"
        >
          <button className="rounded-md bg-red-500 px-4 py-2 text-white hover:bg-red-600">
            Delete Item
          </button>
        </Popover>
      </div>
    </div>
  ),
}

// Hover trigger examples
export const HoverTrigger = {
  render: () => (
    <div className="space-y-6 p-8">
      <h3 className="mb-4 text-lg font-semibold">Hover Trigger Examples</h3>

      <div className="flex flex-wrap gap-4">
        <Popover
          content={
            <SimpleTooltip
              title="Hover Tooltip"
              description="This appears on hover with a slight delay"
            />
          }
          trigger="hover"
          placement="top"
        >
          <button className="rounded-md bg-green-500 px-4 py-2 text-white hover:bg-green-600">
            Hover for tooltip
          </button>
        </Popover>

        <Popover
          content={
            <div className="max-w-sm rounded-lg border border-gray-200 bg-white p-4 shadow-lg">
              <div className="mb-3 flex items-center gap-3">
                <div className="flex h-12 w-12 items-center justify-center rounded-full bg-blue-100">
                  <svg
                    className="h-6 w-6 text-blue-600"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                    />
                  </svg>
                </div>
                <div>
                  <h4 className="font-medium text-gray-900">Quick Info</h4>
                  <p className="text-sm text-gray-600">
                    Additional details appear on hover
                  </p>
                </div>
              </div>
              <p className="text-xs text-gray-500">
                Hover interactions are great for supplementary information that
                doesn’t require user action.
              </p>
            </div>
          }
          trigger="hover"
          placement="right"
        >
          <div className="inline-flex cursor-pointer items-center gap-2 rounded-md bg-gray-100 px-4 py-2 hover:bg-gray-200">
            <span>Info Card</span>
            <svg
              className="h-4 w-4 text-gray-500"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
              />
            </svg>
          </div>
        </Popover>

        <Popover
          content={
            <SimpleTooltip
              title="Help"
              description="Press Ctrl+H for keyboard shortcuts"
            />
          }
          trigger="hover"
          placement="bottom"
        >
          <button className="flex h-8 w-8 items-center justify-center rounded-full bg-gray-200 text-gray-600 hover:bg-gray-300">
            ?
          </button>
        </Popover>
      </div>
    </div>
  ),
}

// Placement examples
export const Placements = {
  render: () => (
    <div className="p-16">
      <h3 className="mb-8 text-center text-lg font-semibold">
        Placement Examples
      </h3>

      <div className="mx-auto grid max-w-md grid-cols-3 gap-8">
        {/* Top row */}
        <Popover
          content={
            <SimpleTooltip
              title="Top Start"
              description="top-start placement"
            />
          }
          placement="top-start"
          trigger="hover"
        >
          <button className="rounded bg-purple-500 px-3 py-2 text-sm text-white hover:bg-purple-600">
            top-start
          </button>
        </Popover>
        <Popover
          content={<SimpleTooltip title="Top" description="top placement" />}
          placement="top"
          trigger="hover"
        >
          <button className="rounded bg-purple-500 px-3 py-2 text-sm text-white hover:bg-purple-600">
            top
          </button>
        </Popover>
        <Popover
          content={
            <SimpleTooltip title="Top End" description="top-end placement" />
          }
          placement="top-end"
          trigger="hover"
        >
          <button className="rounded bg-purple-500 px-3 py-2 text-sm text-white hover:bg-purple-600">
            top-end
          </button>
        </Popover>

        {/* Middle row */}
        <Popover
          content={<SimpleTooltip title="Left" description="left placement" />}
          placement="left"
          trigger="hover"
        >
          <button className="rounded bg-purple-500 px-3 py-2 text-sm text-white hover:bg-purple-600">
            left
          </button>
        </Popover>
        <div className="flex items-center justify-center">
          <div className="text-sm text-gray-400">Trigger</div>
        </div>
        <Popover
          content={
            <SimpleTooltip title="Right" description="right placement" />
          }
          placement="right"
          trigger="hover"
        >
          <button className="rounded bg-purple-500 px-3 py-2 text-sm text-white hover:bg-purple-600">
            right
          </button>
        </Popover>

        {/* Bottom row */}
        <Popover
          content={
            <SimpleTooltip
              title="Bottom Start"
              description="bottom-start placement"
            />
          }
          placement="bottom-start"
          trigger="hover"
        >
          <button className="rounded bg-purple-500 px-3 py-2 text-sm text-white hover:bg-purple-600">
            bottom-start
          </button>
        </Popover>
        <Popover
          content={
            <SimpleTooltip title="Bottom" description="bottom placement" />
          }
          placement="bottom"
          trigger="hover"
        >
          <button className="rounded bg-purple-500 px-3 py-2 text-sm text-white hover:bg-purple-600">
            bottom
          </button>
        </Popover>
        <Popover
          content={
            <SimpleTooltip
              title="Bottom End"
              description="bottom-end placement"
            />
          }
          placement="bottom-end"
          trigger="hover"
        >
          <button className="rounded bg-purple-500 px-3 py-2 text-sm text-white hover:bg-purple-600">
            bottom-end
          </button>
        </Popover>
      </div>

      <p className="mt-6 text-center text-sm text-gray-500">
        Hover over each button to see the placement in action
      </p>
    </div>
  ),
}

// Disabled state
export const DisabledState = {
  render: () => (
    <div className="space-y-6 p-8">
      <h3 className="mb-4 text-lg font-semibold">Disabled State</h3>

      <div className="flex gap-4">
        <Popover
          content={
            <SimpleTooltip
              title="This won't show"
              description="Popover is disabled"
            />
          }
          disabled={true}
          trigger="click"
        >
          <button className="cursor-not-allowed rounded-md bg-gray-400 px-4 py-2 text-white">
            Disabled (Click)
          </button>
        </Popover>

        <Popover
          content={
            <SimpleTooltip
              title="This won't show either"
              description="Popover is disabled"
            />
          }
          disabled={true}
          trigger="hover"
        >
          <button className="cursor-not-allowed rounded-md bg-gray-400 px-4 py-2 text-white">
            Disabled (Hover)
          </button>
        </Popover>

        <Popover
          content={
            <SimpleTooltip
              title="This works!"
              description="Popover is enabled"
            />
          }
          disabled={false}
          trigger="click"
        >
          <button className="rounded-md bg-blue-500 px-4 py-2 text-white hover:bg-blue-600">
            Enabled
          </button>
        </Popover>
      </div>
    </div>
  ),
}

// Advanced styling
export const CustomStyling = {
  render: () => (
    <div className="space-y-6 p-8">
      <h3 className="mb-4 text-lg font-semibold">Custom Styling Examples</h3>

      <div className="flex flex-wrap gap-4">
        <Popover
          content={
            <div className="rounded-xl bg-gradient-to-r from-purple-500 to-pink-500 p-4 text-white shadow-2xl">
              <h4 className="mb-2 font-bold">Gradient Popover</h4>
              <p className="text-sm opacity-90">
                Custom styled with gradients and shadows
              </p>
            </div>
          }
          trigger="click"
          contentClassName="animate-in fade-in-0 zoom-in-95 duration-200"
        >
          <button className="rounded-lg bg-gradient-to-r from-purple-500 to-pink-500 px-4 py-2 text-white hover:from-purple-600 hover:to-pink-600">
            Gradient Style
          </button>
        </Popover>

        <Popover
          content={
            <div className="rounded-md border border-green-400 bg-black p-4 font-mono text-sm text-green-400">
              <div className="mb-2">$ system status</div>
              <div className="text-green-300">✓ All systems operational</div>
              <div className="text-green-300">✓ Database connected</div>
              <div className="text-green-300">✓ API responding</div>
            </div>
          }
          trigger="hover"
        >
          <button className="rounded border border-green-400 bg-black px-4 py-2 font-mono text-green-400 hover:bg-gray-900">
            Terminal Style
          </button>
        </Popover>

        <Popover
          content={
            <div className="max-w-xs rounded-2xl border-0 bg-white p-6 shadow-2xl">
              <div className="text-center">
                <div className="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-blue-100">
                  <svg
                    className="h-8 w-8 text-blue-500"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z"
                    />
                  </svg>
                </div>
                <h4 className="mb-2 font-semibold text-gray-900">Card Style</h4>
                <p className="text-sm text-gray-600">
                  Clean card design with centered content and icons
                </p>
              </div>
            </div>
          }
          trigger="click"
          placement="top"
        >
          <button className="rounded-xl border-2 border-gray-200 bg-white px-4 py-2 text-gray-700 shadow-sm transition-all hover:border-gray-300 hover:shadow-md">
            Card Style
          </button>
        </Popover>
      </div>
    </div>
  ),
}
