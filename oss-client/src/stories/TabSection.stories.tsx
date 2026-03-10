import type { Meta, StoryObj } from '@storybook/nextjs'
import { TabSection } from '@/components/TabSection'

const meta: Meta<typeof TabSection> = {
  title: 'Components/TabSection',
  component: TabSection,
  args: {
    heading: 'Example Section',
  },
}
export default meta

type Story = StoryObj<typeof TabSection>

// --- Simple usage with children ---
export const Simple: Story = {
  args: {
    children: (
      <div className="p-4 text-sm text-gray-800 dark:text-gray-200">
        <p>
          This is the <strong>simple TabSection</strong>. It just renders its
          children inside the layout.
        </p>
      </div>
    ),
    copyButtonContent: 'Some text to copy',
  },
}

// --- Toggle usage (raw default, formatted shows placeholder) ---
export const Toggle: Story = {
  args: {
    showRawJsonToggle: true,
    defaultView: 'raw',
    sectionContent: ({ isRawView }) =>
      isRawView ? (
        <pre className="overflow-auto bg-gray-900 p-4 text-xs text-white">
          {`{ "foo": "bar", "count": 42 }`}
        </pre>
      ) : (
        <div className="p-4 text-sm text-(--system-gray-5)">
          Formatted view placeholder
        </div>
      ),
    copyButtonContent: '{ "foo": "bar", "count": 42 }',
  },
}
