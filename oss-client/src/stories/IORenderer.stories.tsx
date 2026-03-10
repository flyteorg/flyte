import { Meta, StoryObj } from '@storybook/nextjs'
import { IORenderer } from '../components/IORenderer/IORenderer'
import type { IORendererProps } from '../components/IORenderer/types'

const meta: Meta<typeof IORenderer> = {
  title: 'components/IORenderer',
  component: IORenderer,
  parameters: {
    layout: 'centered',
    docs: {
      description: {
        component:
          'A component to render workflow I/O. Raw view shows JSON tree; formatted view shows licensed-edition placeholder.',
      },
    },
  },
}

export default meta

type Story = StoryObj<IORendererProps>

const mockJsonSchema: IORendererProps['jsonSchema'] = {
  type: 'object',
  properties: {
    a: {
      type: 'integer',
      title: 'A',
      default: 1,
    },
  },
  required: ['a'],
}

export const Default: Story = {
  args: {
    isLoading: false,
    jsonSchema: mockJsonSchema,
    formData: { a: 1 },
    expandLevel: 8,
  },
  decorators: [
    (Story) => (
      <div className="mx-auto w-[600px]">
        <Story />
      </div>
    ),
  ],
}

export const Loading: Story = {
  args: {
    isLoading: true,
    jsonSchema: mockJsonSchema,
    expandLevel: 8,
  },
  decorators: [
    (Story) => (
      <div className="mx-auto w-[600px]">
        <Story />
      </div>
    ),
  ],
}

export const NoData: Story = {
  args: {
    isLoading: false,
    jsonSchema: { type: 'object', properties: {} },
    noDataMessage: 'No inputs or outputs',
    expandLevel: 8,
  },
  decorators: [
    (Story) => (
      <div className="mx-auto w-[600px]">
        <Story />
      </div>
    ),
  ],
}
