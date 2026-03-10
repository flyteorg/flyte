import type { Meta, StoryObj } from '@storybook/nextjs'
import { DescriptionListWrapper } from '@/components/DescriptionListWrapper'

const meta = {
  component: DescriptionListWrapper,
  parameters: {
    layout: 'padded',
  },
  decorators: [
    (Story) => (
      <div className="w-[800px]">
        <Story />
      </div>
    ),
  ],
  tags: ['autodocs'],
} satisfies Meta<typeof DescriptionListWrapper>

export default meta
type Story = StoryObj<typeof DescriptionListWrapper>

const testdata = {
  launchPlan: {
    resourceType: 'TASK',
    project: 'flytesnacks',
    domain: 'development',
    name: 'simple.run_test',
    version: 'ZgYCqa3pv6G3nzsX6PF2lg',
    org: '',
  },
  inputs: null,
  metadata: {
    mode: 'MANUAL',
    principal: 'dogfood-flyteadmin',
    nesting: 0,
    scheduledAt: null,
    parentNodeExecution: null,
    referenceExecution: null,
    systemMetadata: {
      executionCluster: 'dogfood-1',
      namespace: '',
    },
    artifactIds: [],
    userIdentity: null,
  },
  taskResourceAttributes: {
    defaults: {
      cpu: '2',
      gpu: '0',
      memory: '1000Mi',
      storage: '',
      ephemeralStorage: '0',
    },
    limits: {
      cpu: '4096',
      gpu: '256',
      memory: '2Ti',
      storage: '',
      ephemeralStorage: '0',
    },
  },
}

export const Default: Story = {
  args: {
    rawJson: testdata,
    sections: [
      {
        id: 'launch-plan',
        name: 'Launch Plan',
        items: [
          { name: 'Resource Type', value: testdata.launchPlan.resourceType },
          { name: 'Project', value: testdata.launchPlan.project },
          { name: 'Domain', value: testdata.launchPlan.domain },
          { name: 'Name', value: testdata.launchPlan.name },
          { name: 'Version', value: testdata.launchPlan.version },
        ],
      },
      {
        id: 'metadata',
        name: 'Metadata',
        items: [
          { name: 'Mode', value: testdata.metadata.mode },
          { name: 'Principal', value: testdata.metadata.principal },
          { name: 'Nesting', value: testdata.metadata.nesting },
          {
            name: 'Execution Cluster',
            value: testdata.metadata.systemMetadata.executionCluster,
          },
        ],
      },
      {
        id: 'resource-attributes',
        name: 'Resource Attributes',
        items: [
          {
            name: 'CPU Default',
            value: testdata.taskResourceAttributes.defaults.cpu,
          },
          {
            name: 'GPU Default',
            value: testdata.taskResourceAttributes.defaults.gpu,
          },
          {
            name: 'Memory Default',
            value: testdata.taskResourceAttributes.defaults.memory,
          },
          {
            name: 'CPU Limit',
            value: testdata.taskResourceAttributes.limits.cpu,
          },
          {
            name: 'GPU Limit',
            value: testdata.taskResourceAttributes.limits.gpu,
          },
          {
            name: 'Memory Limit',
            value: testdata.taskResourceAttributes.limits.memory,
          },
        ],
      },
    ],
  },
}

export const WithPrimitiveValues: Story = {
  args: {
    rawJson: testdata,
    sections: [
      {
        id: 'primitive-values',
        name: 'Configuration',
        items: [
          { name: 'Environment name', value: 'env_name' },
          {
            name: 'Image',
            value:
              '356633062068.dkr.ecr.us-east-2.amazonaws.com/union/demo:image-gPA9cR27Fp4p_6_QDoro1w',
          },
          { name: 'Dependencies', value: '' },
          {
            name: 'Secrets',
            value: {
              SNOWFLAKE_NAME: '****',
              SNOWFLAKE_PASSWORD: '***',
            },
          },
        ],
      },
    ],
  },
}

export const WithNestedObjects: Story = {
  args: {
    rawJson: testdata,
    sections: [
      {
        id: 'resources',
        name: 'Resources',
        items: [
          { name: 'CPU Request', value: '2' },
          { name: 'CPU Limit', value: '4096' },
          { name: 'Memory Request', value: '1000Mi' },
          { name: 'Memory Limit', value: '2Ti' },
          { name: 'Storage Request', value: '0' },
          { name: 'Storage Limit', value: '0' },
          { name: 'GPU Request', value: '0' },
          { name: 'GPU Limit', value: '256' },
        ],
      },
      {
        id: 'configuration',
        name: 'Configuration',
        items: [
          { name: 'Timeout', value: '10m' },
          { name: 'Retry', value: '3' },
          { name: 'Priority', value: '0' },
          { name: 'Interruptible', value: 'false' },
          { name: 'OverwriteCache', value: 'false' },
          { name: 'MaxParallelism', value: '25' },
          { name: 'ExecutionCluster', value: 'dogfood-1' },
        ],
      },
    ],
  },
}

export const EmptyValues: Story = {
  args: {
    rawJson: {},
    sections: [
      {
        id: 'empty-values',
        name: 'Empty Values',
        items: [
          { name: 'Empty String', value: '' },
          { name: 'Null Value', value: null },
          { name: 'Empty Object', value: {} },
          { name: 'Empty Array', value: [] },
        ],
      },
    ],
  },
}
