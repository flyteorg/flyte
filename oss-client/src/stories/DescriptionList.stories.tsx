import type { Meta, StoryObj } from '@storybook/nextjs'
import {
  DescriptionList,
  DescriptionTerm,
  DescriptionDetails,
} from '../components/DescriptionList'
import { Badge } from '../components/Badge'

const meta: Meta<typeof DescriptionList> = {
  component: DescriptionList,
  tags: ['autodocs'],
  parameters: {
    design: {
      type: 'figma',
      url: 'https://www.figma.com/design/fd9kUG1YQRJhhG7xu5Wago/Catalyst-UI?node-id=639-18074&t=r3LaWaIcGhum9ZKb-1',
    },
  },
}

export default meta
type Story = StoryObj<typeof DescriptionList>

export const Basic: Story = {
  render: () => (
    <DescriptionList>
      <DescriptionTerm>Name</DescriptionTerm>
      <DescriptionDetails>John Doe</DescriptionDetails>
      <DescriptionTerm>Email</DescriptionTerm>
      <DescriptionDetails>john.doe@example.com</DescriptionDetails>
      <DescriptionTerm>Role</DescriptionTerm>
      <DescriptionDetails>Software Engineer</DescriptionDetails>
    </DescriptionList>
  ),
}

export const WithCustomStyling: Story = {
  render: () => (
    <DescriptionList className="rounded-lg bg-zinc-50 p-4">
      <DescriptionTerm className="font-semibold">Project</DescriptionTerm>
      <DescriptionDetails className="font-mono">
        flyte-platform
      </DescriptionDetails>
      <DescriptionTerm className="font-semibold">Status</DescriptionTerm>
      <DescriptionDetails className="text-green-600">Active</DescriptionDetails>
      <DescriptionTerm className="font-semibold">Last Updated</DescriptionTerm>
      <DescriptionDetails>2024-04-22</DescriptionDetails>
    </DescriptionList>
  ),
}

export const WithLongContent: Story = {
  render: () => (
    <DescriptionList>
      <DescriptionTerm>Description</DescriptionTerm>
      <DescriptionDetails>
        This is a longer description that might span multiple lines. It
        demonstrates how the DescriptionList component handles content that
        exceeds a single line. The component should maintain proper spacing and
        alignment regardless of content length.
      </DescriptionDetails>
      <DescriptionTerm>Tags</DescriptionTerm>
      <DescriptionDetails>
        <div className="flex flex-wrap gap-2">
          <Badge color="blue">React</Badge>
          <Badge color="indigo">TypeScript</Badge>
          <Badge color="purple">Tailwind</Badge>
        </div>
      </DescriptionDetails>
    </DescriptionList>
  ),
}

export const DarkMode: Story = {
  parameters: {
    themes: {
      defaultTheme: 'dark',
    },
  },
  render: () => (
    <DescriptionList>
      <DescriptionTerm>Theme</DescriptionTerm>
      <DescriptionDetails>Dark Mode</DescriptionDetails>
      <DescriptionTerm>Background</DescriptionTerm>
      <DescriptionDetails>#000000</DescriptionDetails>
      <DescriptionTerm>Text Color</DescriptionTerm>
      <DescriptionDetails>#FFFFFF</DescriptionDetails>
    </DescriptionList>
  ),
}

export const Responsive: Story = {
  parameters: {
    viewport: {
      defaultViewport: 'mobile1',
    },
  },
  render: () => (
    <DescriptionList>
      <DescriptionTerm>Device</DescriptionTerm>
      <DescriptionDetails>Mobile</DescriptionDetails>
      <DescriptionTerm>Screen Size</DescriptionTerm>
      <DescriptionDetails>320px × 568px</DescriptionDetails>
      <DescriptionTerm>Orientation</DescriptionTerm>
      <DescriptionDetails>Portrait</DescriptionDetails>
    </DescriptionList>
  ),
}
