import type { Meta, StoryObj } from '@storybook/nextjs'
import { StatusIcon } from '@/components/StatusIcons/StatusIcon'
import { ActionPhase } from '@/gen/flyteidl2/common/phase_pb'

const phaseInfo = {
  [ActionPhase.UNSPECIFIED]: {
    label: 'Not Started',
    description: 'Initial state before the run begins',
  },
  [ActionPhase.QUEUED]: {
    label: 'Queued',
    description: 'Waiting in queue to be processed',
  },
  [ActionPhase.WAITING_FOR_RESOURCES]: {
    label: 'Waiting for Resources',
    description: 'Waiting for compute resources to become available',
  },
  [ActionPhase.INITIALIZING]: {
    label: 'Initializing',
    description: 'Setting up the run environment',
  },
  [ActionPhase.RUNNING]: { label: 'Running', description: 'Currently executing' },
  [ActionPhase.SUCCEEDED]: {
    label: 'Succeeded',
    description: 'Completed successfully',
  },
  [ActionPhase.FAILED]: { label: 'Failed', description: 'Completed with errors' },
  [ActionPhase.ABORTED]: { label: 'Canceled', description: 'Manually stopped' },
  [ActionPhase.TIMED_OUT]: {
    label: 'Timed Out',
    description: 'Exceeded maximum allowed duration',
  },
} as const

// Create a mapping of labels to Phase values for the select control
const phaseOptions = Object.entries(phaseInfo).reduce<Record<string, ActionPhase>>(
  (acc, [phase, info]) => {
    acc[info.label] = Number(phase) as ActionPhase
    return acc
  },
  {},
)

const meta = {
  title: 'components/StatusIcon',
  component: StatusIcon,
  tags: ['autodocs'],
  argTypes: {
    phase: {
      options: Object.keys(phaseOptions),
      mapping: phaseOptions,
      control: { type: 'select' },
      description: 'Current phase of the run',
      table: {
        defaultValue: { summary: 'Not Started' },
        type: {
          summary: 'ActionPhase',
          detail: Object.entries(phaseInfo)
            .map(([, info]) => `${info.label}: ${info.description}`)
            .join('\n'),
        },
      },
    },
    isActive: {
      control: 'boolean',
      description: 'Whether to show the colored background',
      table: {
        defaultValue: { summary: 'true' },
      },
    },
    isStatic: {
      control: 'boolean',
      description: 'Whether icon is animated or not',
      table: {
        defaultValue: { summary: 'false' },
      },
    },
  },
  parameters: {
    design: {
      type: 'figma',
      url: 'https://www.figma.com/file/Qj5TqXNolqZWqgbkz6zZH7/Union-Cloud?type=design&node-id=2-1055&mode=design&t=YxXRc8tHpvxu0JFh-4',
    },
    docs: {
      description: {
        component: `
Status icons represent different states of a run or task.

## States
${Object.entries(phaseInfo)
  .map(([, info]) => `- ${info.label}: ${info.description}`)
  .join('\n')}

## Variants
Each state has two variants:
- Default: Just the icon
- Active: Icon with a colored background at 12% opacity

## Colors
- Not Started: Zinc
- Queued: Purple
- Running: Blue
- Succeeded: Lime
- Timed Out: Yellow
- Canceled: Orange
- Failed: Red
`,
      },
    },
  },
} satisfies Meta<typeof StatusIcon>

export default meta
type Story = StoryObj<typeof StatusIcon>

export const Default: Story = {
  args: {
    isActive: true,
  },
}

export const AllStates: Story = {
  render: () => (
    <div className="flex max-w-[700px] flex-col gap-8">
      <div className="grid grid-cols-[4fr_1fr_1fr_1fr] gap-x-12">
        <div />
        <div className="text-center text-sm font-medium text-zinc-200">
          Default
        </div>
        <div className="text-center text-sm font-medium text-zinc-200">
          Active
        </div>
        <div className="text-center text-sm font-medium text-zinc-200">
          Static
        </div>
      </div>

      {Object.entries(phaseInfo).map(([phase, info]) => (
        <div key={phase} className="grid grid-cols-[4fr_1fr_1fr_1fr] gap-x-12">
          <div className="flex flex-col items-center">
            <div className="text-center text-sm text-zinc-200">
              {info.label}
            </div>
            <div className="text-center text-xs text-zinc-400">
              {info.description}
            </div>
          </div>
          <div className="flex items-center justify-center">
            <StatusIcon phase={Number(phase)} isActive={false} />
          </div>
          <div className="flex items-center justify-center">
            <StatusIcon phase={Number(phase)} isActive={true} />
          </div>
          <div className="flex items-center justify-center">
            <StatusIcon phase={Number(phase)} isStatic={true} />
          </div>
        </div>
      ))}
    </div>
  ),
}