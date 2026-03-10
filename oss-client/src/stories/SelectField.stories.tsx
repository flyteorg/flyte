import type { Meta, StoryObj } from '@storybook/nextjs'
import { SelectField } from '@/components/SelectField'
import { FormProvider, useForm } from 'react-hook-form'

const mockOptions = [
  { value: 'id-1', label: 'great test item 4' },
  { value: 'id-2', label: 'Great thing' },
  { value: 'id-3', label: 'Project name 1' },
  { value: 'id-4', label: 'Project name 2' },
  { value: 'id-5', label: 'Some new item' },
  {
    value: 'id-6',
    label:
      'test item long long ver long name for a project to test so so long that it is hard to read',
  },
  { value: 'id-7', label: 'Test item 2' },
  { value: 'id-8', label: 'Test item 3' },
  { value: 'id-9', label: 'Union project' },
]

const meta: Meta<typeof SelectField> = {
  title: 'components/SelectField',
  component: SelectField,
  tags: ['autodocs'],
  decorators: [
    (Story) => (
      <div className="h-[350px] w-full">
        <Story />
      </div>
    ),
  ],
  parameters: {
    docs: {
      description: {
        component: `
The \`SelectField\` component is a styled select with a built-in search input.

## Usage
\`\`\`tsx
<SelectField name="projects" labelText="Projects" options=[{ value: "id-1", label: "Project name" }] />
\`\`\`
        `,
      },
    },
  },
  args: {
    options: mockOptions,
    name: 'project',
    labelText: 'Project',
  },
  argTypes: {
    name: {
      description: 'Name of the select field',
      control: false,
    },
    labelText: {
      description: 'Label displayed above the selector',
      control: { type: 'text' },
    },
    options: {
      description: 'List of options available to select',
    },
    allOptionsLabel: {
      description: `Label for option to select all items from selector dropdown. (Selecting it does not return all values, it must be handled separately)`,
      control: { type: 'text' },
    },
  },
}

export default meta

type Story = StoryObj<typeof SelectField>

export const Default: Story = {
  render: (args) => {
    const Wrapper = () => {
      const methods = useForm({
        defaultValues: { [args.name]: '' },
      })
      return (
        <FormProvider {...methods}>
          <SelectField {...args} />
        </FormProvider>
      )
    }
    return <Wrapper />
  },
}

export const AllItemsOption: Story = {
  name: "With 'all items' option",
  render: (args) => {
    const Wrapper = () => {
      const methods = useForm({
        defaultValues: { [args.name]: '' },
      })
      return (
        <FormProvider {...methods}>
          <SelectField {...args} />
        </FormProvider>
      )
    }
    return <Wrapper />
  },
  args: {
    allOptionsLabel: 'All items',
  },
}
