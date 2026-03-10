import type { Meta, StoryObj } from '@storybook/nextjs'
import { ActionDashboardLinks } from '@/components/pages/RunDetails/ActionDashboardLinks'
import { ActionAttempt } from '@/gen/flyteidl2/workflow/run_definition_pb'
import { TaskLog, TaskLog_LinkType } from '@/gen/flyteidl/core/execution_pb'

/**
 * The ActionDashboardLinks component displays integration links for external dashboards
 * and monitoring tools based on the action attempt's log information.
 *
 * ## Features
 *
 * - Automatically detects and displays dashboard links from log info
 * - Supports various integration types (Spark UI, Ray Dashboard, Weights & Biases, etc.)
 * - Renders appropriate icons for each integration type
 * - Only shows when dashboard links are available
 * - Responsive layout with proper spacing and styling
 *
 * ## Usage
 *
 * ```tsx
 * <ActionDashboardLinks attempt={actionAttempt} />
 * ```
 *
 * The component automatically filters logs by DASHBOARD link type and renders
 * external links with appropriate icons based on the log name.
 */
const meta: Meta<typeof ActionDashboardLinks> = {
  title: 'Components/ActionDashboardLinks',
  component: ActionDashboardLinks,
  tags: ['autodocs'],
  parameters: {
    docs: {
      description: {
        component: `
The ActionDashboardLinks component displays integration links for external dashboards
and monitoring tools based on the action attempt's log information.

## Features

- Automatically detects and displays dashboard links from log info
- Supports various integration types (Spark UI, Ray Dashboard, Weights & Biases, etc.)
- Renders appropriate icons for each integration type
- Only shows when dashboard links are available
- Responsive layout with proper spacing and styling

## Integration Types Supported

- **Spark**: Spark UI and Spark Driver dashboards
- **Ray**: Ray dashboard
- **Weights & Biases**: WandB experiment tracking
- **Neptune**: Neptune experiment tracking
- **Dask**: Dask dashboard
- **PyTorch**: PyTorch TensorBoard
- **TensorFlow**: TensorFlow TensorBoard
- **BigQuery**: BigQuery console
- **Databricks**: Databricks workspace
- **Snowflake**: Snowflake console
- **CloudWatch**: AWS CloudWatch logs
- **Stackdriver**: GCP Stackdriver logs

## Usage

\`\`\`tsx
<ActionDashboardLinks attempt={actionAttempt} />
\`\`\`

The component automatically filters logs by DASHBOARD link type and renders
external links with appropriate icons based on the log name.
        `,
      },
    },
  },
  argTypes: {
    attempt: {
      description: 'The action attempt containing log information',
      control: 'object',
    },
  },
}

export default meta
type Story = StoryObj<typeof ActionDashboardLinks>

// Helper function to create mock ActionAttempt with specific log info
const createMockActionAttempt = (logNames: string[]): ActionAttempt =>
  ({
    phase: 1, // RUNNING
    startTime: { seconds: Date.now() / 1000, nanos: 0 },
    attempt: 1,
    logInfo: logNames.map(
      (name, index) =>
        ({
          uri: `https://example.com/${name.toLowerCase().replace(' ', '-')}-${index}`,
          name,
          messageFormat: 0, // UNKNOWN
          ShowWhilePending: true,
          HideOnceFinished: false,
          linkType: TaskLog_LinkType.DASHBOARD,
        }) as TaskLog,
    ),
    logsAvailable: true,
    cacheStatus: 0, // CACHE_DISABLED
    clusterEvents: [],
    phaseTransitions: [],
    cluster: 'default',
  }) as unknown as ActionAttempt

// Story: All supported integration types
export const AllIntegrationTypes: Story = {
  args: {
    attempt: createMockActionAttempt([
      'Spark UI',
      'Ray Dashboard',
      'WandB',
      'Neptune',
      'Dask',
      'PyTorch',
      'TensorFlow',
      'BigQuery',
      'Databricks',
      'Snowflake',
      'CloudWatch',
      'Stackdriver',
      'Spark Driver',
      'Comet',
    ]),
  },
  parameters: {
    docs: {
      description: {
        story:
          'Demonstrates all supported integration types with their respective icons.',
      },
    },
  },
}
