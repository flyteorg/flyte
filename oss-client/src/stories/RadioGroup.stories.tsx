import type { Meta, StoryObj } from '@storybook/nextjs'
import { RadioGroup } from '@/components/RadioGroup'

const meta: Meta<typeof RadioGroup> = {
  title: 'primitives/RadioGroup',
  component: RadioGroup,
  tags: ['autodocs'],
  parameters: {
    docs: {
      description: {
        component: `
The RadioGroup component displays a list of options that users can select from using radio buttons.
It supports custom labels, helper text, and a disabled state.

## Usage

\`\`\`tsx
type OptionId = 'option1' | 'option2' | 'option3'
const [value, setValue] = useState<OptionId>('option1')

<RadioGroup<OptionId>
  id="example-radio-group"
  labelText="Choose a value"
  helpText="Only one can be selected"
  radioOptions={[
    { id: 'option1', title: 'Option 1' },
    { id: 'option2', title: 'Option 2' },
  ]}
  selectedOptionId={value}
  setSelectedOptionId={setValue}
/>
\`\`\`
        `,
      },
    },
  },
  argTypes: {
    disabled: { control: 'boolean' },
    labelText: { control: 'text' },
    helpText: { control: 'text' },
    radioOptions: { table: { disable: true } },
    selectedOptionId: { table: { disable: true } },
    setSelectedOptionId: { table: { disable: true } },
    id: { control: false },
  },
  args: {
    disabled: false,
    labelText: 'Select an option',
    helpText: 'This is some help text.',
    selectedOptionId: 'option1',
    radioOptions: [
      { id: 'option1', title: 'Option One' },
      { id: 'option2', title: 'Option Two' },
      { id: 'option3', title: 'Option Three' },
    ],
    id: 'example-radio-group',
  },
}

export default meta

type Story = StoryObj<typeof RadioGroup>

export const Default: Story = {
  name: 'Default',
  args: {},
}

export const Disabled: Story = {
  name: 'Disabled State',
  args: {
    disabled: true,
  },
}
