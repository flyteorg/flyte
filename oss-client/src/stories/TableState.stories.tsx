import { Meta, StoryObj } from '@storybook/nextjs'
import { TableState } from '@/components/Tables'

const meta: Meta<typeof TableState> = {
  title: 'Components/TableState',
  component: TableState,
  tags: ['autodocs'],
  decorators: [
    (Story) => (
      <div className="h-[250px] w-full p-6">
        <Story />
      </div>
    ),
  ],
  parameters: {
    layout: 'fullscreen',
    docs: {
      description: {
        component: `
The TableState reusable component handles data states for tables.

## Features
- Loading state
- Error state
- No data state (eg. no items in a database)
- No results found (filtered search returns no matches)

## Usage
\`\`\`tsx
import { TableState } from '@/components/TableState'

<TableState
  data={query.dataArray}
  dataName="projects"
  isError={query.isError}
  isLoading={query.isLoading}
  searchQuery="flyteProject"
>
  {(data) => <Table data={data} />}
</TableState>
\`\`\`
        `,
      },
    },
  },
  argTypes: {
    data: {
      description:
        'The data list to display in the table. If the list is empty or undefined, the appropriate fallback state will be shown.',
      control: 'object',
    },
    dataLabel: {
      description: 'The label of data list to use in fallback message',
      control: 'text',
    },
    subtitle: {
      description: 'Descriptive text shown in fallback message',
      control: 'text',
    },
    isLoading: {
      description: 'Displays the loading spinner',
      control: 'boolean',
    },
    isError: {
      description: 'Displays an error message',
      control: 'boolean',
    },
    searchQuery: {
      description: `The search query text used to show 'no results' state and used in fallback message`,
      control: 'text',
    },
    content: {
      description: 'Optional custom content displayed below the subtitle',
      control: 'text',
    },
    children: {
      description:
        'A render function that receives the data array and returns the component to display when data is available.',
      control: false,
    },
  },
  args: {
    data: [],
    dataLabel: 'projects',
    subtitle: 'Create a project to see it here',
    isLoading: false,
    isError: true,
    children: () => 'render table',
  },
}

export default meta

type Story = StoryObj<typeof TableState>

export const Error: Story = {
  args: {
    data: undefined,
    dataLabel: 'projects',
    isError: true,
  },
  parameters: {
    docs: {
      description: {
        story: 'Shown when data fetching fails.',
      },
    },
  },
}

export const Loading: Story = {
  args: {
    data: undefined,
    dataLabel: 'projects',
    isLoading: true,
  },
  parameters: {
    docs: {
      description: {
        story: 'Shown while data is being fetched.',
      },
    },
  },
}

export const NoData: Story = {
  name: 'No data in database',
  args: {
    data: [],
    dataLabel: 'tasks',
    isError: false,
    subtitle: 'Create your first task',
  },
  parameters: {
    docs: {
      description: {
        story:
          'Shown when there is no data available and no search filter is applied.',
      },
    },
  },
}

export const NoDataWithSearch: Story = {
  name: 'No data found with search',
  args: {
    data: [],
    dataLabel: 'runs',
    isError: false,
    searchQuery: 'run123',
  },
  parameters: {
    docs: {
      description: {
        story:
          'Shown when a search query is applied but no matching data is found.',
      },
    },
  },
}

export const NoDataWithCustomContent: Story = {
  name: 'With extra content below texts',
  args: {
    data: [],
    dataLabel: 'projects',
    isError: false,
    subtitle: 'Create a project to see it here',
    content: (
      <button className="w-xs rounded bg-green-800 p-2 text-white hover:bg-green-900">
        Create new project
      </button>
    ),
  },
  parameters: {
    docs: {
      description: {
        story:
          'Shows the empty state with an additional custom content below the message.',
      },
    },
  },
}
