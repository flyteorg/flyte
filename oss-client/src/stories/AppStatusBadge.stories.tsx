import type { Meta, StoryObj } from '@storybook/react'
import { AppStatusBadge } from '@/components/pages/ListApps/components/AppStatusBadge'
import { Status_DeploymentStatus } from '@/gen/flyteidl2/app/app_definition_pb'

const meta: Meta<typeof AppStatusBadge> = {
  title: 'Components/AppStatusBadge',
  component: AppStatusBadge,
  parameters: {
    layout: 'centered',
    docs: {
      description: {
        component:
          'A badge component that displays the deployment status of an application with corresponding colors and icons.',
      },
    },
  },
  argTypes: {
    status: {
      control: 'select',
      options: Object.values(Status_DeploymentStatus).filter(
        (value) => typeof value === 'number',
      ),
      mapping: {
        Active: Status_DeploymentStatus.ACTIVE,
        Assigned: Status_DeploymentStatus.ASSIGNED,
        Deploying: Status_DeploymentStatus.DEPLOYING,
        Failed: Status_DeploymentStatus.FAILED,
        Pending: Status_DeploymentStatus.PENDING,
        'Scaling Down': Status_DeploymentStatus.SCALING_DOWN,
        'Scaling Up': Status_DeploymentStatus.SCALING_UP,
        Stopped: Status_DeploymentStatus.STOPPED,
        Unassigned: Status_DeploymentStatus.UNASSIGNED,
        Unspecified: Status_DeploymentStatus.UNSPECIFIED,
      },
      description: 'The deployment status to display',
    },
  },
  tags: ['autodocs'],
}

export default meta
type Story = StoryObj<typeof AppStatusBadge>

export const Active: Story = {
  args: {
    status: Status_DeploymentStatus.ACTIVE,
  },
}

export const Assigned: Story = {
  args: {
    status: Status_DeploymentStatus.ASSIGNED,
  },
}

export const Deploying: Story = {
  args: {
    status: Status_DeploymentStatus.DEPLOYING,
  },
}

export const Failed: Story = {
  args: {
    status: Status_DeploymentStatus.FAILED,
  },
}

export const Pending: Story = {
  args: {
    status: Status_DeploymentStatus.PENDING,
  },
}

export const ScalingDown: Story = {
  args: {
    status: Status_DeploymentStatus.SCALING_DOWN,
  },
  parameters: {
    docs: {
      description: {
        story: 'Shows the scaling down status with a two-line text display.',
      },
    },
  },
}

export const ScalingUp: Story = {
  args: {
    status: Status_DeploymentStatus.SCALING_UP,
  },
  parameters: {
    docs: {
      description: {
        story: 'Shows the scaling up status with a two-line text display.',
      },
    },
  },
}

export const Stopped: Story = {
  args: {
    status: Status_DeploymentStatus.STOPPED,
  },
}

export const Unassigned: Story = {
  args: {
    status: Status_DeploymentStatus.UNASSIGNED,
  },
}

export const Unspecified: Story = {
  args: {
    status: Status_DeploymentStatus.UNSPECIFIED,
  },
}

// Story showing all statuses in a grid
export const AllStatuses: Story = {
  render: () => (
    <div className="grid grid-cols-3 gap-4">
      {Object.entries({
        Active: Status_DeploymentStatus.ACTIVE,
        Assigned: Status_DeploymentStatus.ASSIGNED,
        Deploying: Status_DeploymentStatus.DEPLOYING,
        Failed: Status_DeploymentStatus.FAILED,
        Pending: Status_DeploymentStatus.PENDING,
        'Scaling Down': Status_DeploymentStatus.SCALING_DOWN,
        'Scaling Up': Status_DeploymentStatus.SCALING_UP,
        Stopped: Status_DeploymentStatus.STOPPED,
        Unassigned: Status_DeploymentStatus.UNASSIGNED,
        Unspecified: Status_DeploymentStatus.UNSPECIFIED,
      }).map(([label, status]) => (
        <div key={label} className="flex flex-col items-center space-y-2">
          <AppStatusBadge status={status} />
          <span className="text-xs text-gray-600 dark:text-gray-400">
            {label}
          </span>
        </div>
      ))}
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story:
          'Overview of all available deployment statuses and their visual representations.',
      },
    },
  },
}

export const InContext: Story = {
  args: {
    status: Status_DeploymentStatus.ACTIVE,
  },
  render: (args) => (
    <div className="space-y-4">
      <h3 className="text-lg font-semibold">Apps</h3>
      <div className="space-y-2">
        <div className="flex items-center gap-4 rounded border p-3">
          <div className="w-1/5">
            <AppStatusBadge {...args} />
          </div>
          <div>
            <h4 className="font-medium">my-web-app</h4>
            <p className="text-sm text-zinc-500">Production deployment</p>
          </div>
        </div>
        <div className="flex items-center gap-4 rounded border p-3">
          <div className="w-1/5">
            <AppStatusBadge status={Status_DeploymentStatus.DEPLOYING} />
          </div>
          <div>
            <h4 className="font-medium">api-service</h4>
            <p className="text-sm text-zinc-500">Staging deployment</p>
          </div>
        </div>
        <div className="flex items-center gap-4 rounded border p-3">
          <div className="w-1/5">
            <AppStatusBadge status={Status_DeploymentStatus.FAILED} />
          </div>
          <div>
            <h4 className="font-medium">worker-service</h4>
            <p className="text-sm text-zinc-500">Development deployment</p>
          </div>
        </div>
      </div>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story:
          'Shows how the AppStatusBadge appears in a typical application list context.',
      },
    },
    layout: 'padded',
  },
}
