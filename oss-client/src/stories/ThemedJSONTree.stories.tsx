import { Meta, StoryObj } from '@storybook/nextjs'
import ThemedJSONTree, {
  ThemedJSONTreeProps,
} from '../components/ThemedJSONTree'

const meta: Meta<typeof ThemedJSONTree> = {
  title: 'components/ThemedJSONTree',
  component: ThemedJSONTree,
  parameters: {
    layout: 'centered',
    docs: {
      description: {
        component: 'A JSON tree viewer that adapts to dark and light mode.',
      },
    },
  },
}

export default meta

type Story = StoryObj<ThemedJSONTreeProps>

const simpleData = {
  name: 'Alice',
  age: 30,
  isAdmin: false,
  tags: ['user', 'admin'],
}

const nestedData = {
  user: {
    id: 123,
    profile: {
      name: 'Bob',
      email: 'bob@example.com',
      preferences: {
        theme: 'dark',
        notifications: true,
      },
    },
    roles: ['editor', 'reviewer'],
  },
  active: true,
}

export const Simple: Story = {
  args: {
    data: simpleData,
    expandLevel: 2,
  },
}

export const Nested: Story = {
  args: {
    data: nestedData,
    expandLevel: 3,
  },
}
