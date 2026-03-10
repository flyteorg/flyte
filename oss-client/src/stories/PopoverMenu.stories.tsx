import type { Meta, StoryObj } from '@storybook/react'
import React, { useState } from 'react'
import { PopoverMenu, type MenuItem } from '@/components/Popovers'

const meta = {
  title: 'Components/PopoverMenu',
  component: PopoverMenu,
  parameters: {
    layout: 'centered',
    docs: {
      description: {
        component: `A flexible menu component built on top of \`Popover\` that displays a menu of items triggered by a button or custom element. Supports **three variants**, custom components, and extensive styling options.

## Variants

Each variant changes the trigger appearance and behavior.

- **dropdown** (default) - standard dropdown menu triggered by a button with a label and an optional chevron
- **overflow** - renders an icon-only trigger (vertical or horizontal ellipsis)  
- **filter** - shows selected values directly in the trigger (with optional “and X others”)  

## Menu items

\`PopoverMenu\` accepts an array of \`items: MenuItem[]\`.

#### Supported item types:

**item**
- Standard, default menu option
- Supports \`selected\`, \`disabled\`
- Can render \`icon\` and \`label\`
- Optional \`onClick\` handler

**divider**
- Visual separator between items

**custom**
- Renders custom JSX content
- Use the \`component\` property

## Example

\`\`\`ts
const items: MenuItem[] = [
  { id: 'edit', label: 'Edit', type: 'item', onClick: ()=> {} },
  { id: 'copy', label: 'Copy', type: 'item', disabled: true },
  { id: 'sep-1', type: 'divider' },
  {
    id: 'custom-item',
    type: 'custom',
    component: <CustomContent />,
  },
]
\`\`\`
        `,
      },
    },
  },
  tags: ['autodocs'],
  argTypes: {
    variant: {
      control: 'select',
      options: ['dropdown', 'overflow', 'filter'],
      description: 'Menu trigger style',
    },
    size: {
      control: 'select',
      options: ['sm', 'md', 'lg'],
      description: 'Size of the trigger button',
    },
    outline: {
      control: 'boolean',
      description: `Whether to show border on the trigger in 'dropdown' variant`,
    },
    width: {
      control: 'select',
      options: ['auto', 'trigger'],
      description: 'Width behavior of the menu',
    },
    label: {
      control: 'text',
      description: `Label for trigger element for 'dropdown' and 'filter' variants`,
    },
    placement: {
      control: 'select',
      options: [
        'top',
        'top-start',
        'top-end',
        'right',
        'right-start',
        'right-end',
        'bottom',
        'bottom-start',
        'bottom-end',
        'left',
        'left-start',
        'left-end',
      ],
      description: 'Menu placement relative to trigger',
    },
    disabled: {
      control: 'boolean',
      description: 'Whether the menu is disabled',
    },
    showCheckboxes: {
      control: 'boolean',
      description: 'Whether the checkboxes beside items are visible',
    },
    showChevron: {
      control: 'boolean',
      description: 'Whether the chevron icon in the trigger element is visible',
    },
    closeOnItemClick: {
      control: 'radio',
      options: [true, false, 'default-only'],
      description: 'Whether the menu is closed after an item is clicked',
    },
  },
} satisfies Meta<typeof PopoverMenu>

export default meta
type Story = StoryObj<typeof meta>

// Sample icons for stories
const EditIcon = () => (
  <svg
    className="h-4 w-4"
    fill="none"
    stroke="currentColor"
    viewBox="0 0 24 24"
  >
    <path
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeWidth={2}
      d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
    />
  </svg>
)

const TrashIcon = () => (
  <svg
    className="h-4 w-4"
    fill="none"
    stroke="currentColor"
    viewBox="0 0 24 24"
  >
    <path
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeWidth={2}
      d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
    />
  </svg>
)

const CopyIcon = () => (
  <svg
    className="h-4 w-4"
    fill="none"
    stroke="currentColor"
    viewBox="0 0 24 24"
  >
    <path
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeWidth={2}
      d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"
    />
  </svg>
)

const PlayIcon = () => (
  <svg
    className="h-4 w-4"
    fill="none"
    stroke="currentColor"
    viewBox="0 0 24 24"
  >
    <path
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeWidth={2}
      d="M14.828 14.828a4 4 0 01-5.656 0M9 10h1m4 0h1m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
    />
  </svg>
)

// Basic menu items for reuse
const basicMenuItems: MenuItem[] = [
  {
    id: 'edit',
    label: 'Edit',
    icon: <EditIcon />,
    onClick: () => alert('Edit clicked'),
    type: 'item',
  },
  {
    id: 'copy',
    label: 'Copy',
    icon: <CopyIcon />,
    onClick: () => alert('Copy clicked'),
    type: 'item',
  },
  { id: 'sep1', type: 'divider' },
  {
    id: 'delete',
    label: 'Delete',
    icon: <TrashIcon />,
    onClick: () => alert('Delete clicked'),
    type: 'item',
  },
]

// Default story
export const Default: Story = {
  args: {
    items: basicMenuItems,
    variant: 'dropdown',
    label: 'Actions',
    outline: false,
    size: 'md',
    width: 'auto',
    placement: 'bottom-start',
  },
  render: (args) => (
    <div className="p-8">
      <PopoverMenu {...args} />
    </div>
  ),
}

// Dropdown variants
export const DropdownVariants = {
  render: () => (
    <div className="space-y-8 p-8">
      <h3 className="mb-4 text-lg font-semibold">Dropdown Variants</h3>

      <div className="space-y-6">
        {/* Size variations */}
        <div className="space-y-3">
          <h4 className="font-medium text-gray-700">Sizes</h4>
          <div className="flex items-center gap-4">
            <PopoverMenu
              items={basicMenuItems}
              variant="dropdown"
              label="Small"
              size="sm"
              outline={true}
            />
            <PopoverMenu
              items={basicMenuItems}
              variant="dropdown"
              label="Medium"
              size="md"
              outline={true}
            />
            <PopoverMenu
              items={basicMenuItems}
              variant="dropdown"
              label="Large"
              size="lg"
              outline={true}
            />
          </div>
        </div>

        {/* Outline variations */}
        <div className="space-y-3">
          <h4 className="font-medium text-gray-700">Outline Styles</h4>
          <div className="flex items-center gap-4">
            <PopoverMenu
              items={basicMenuItems}
              variant="dropdown"
              label="Outlined"
              outline={true}
            />
            <PopoverMenu
              items={basicMenuItems}
              variant="dropdown"
              label="Solid"
              outline={false}
            />
          </div>
        </div>

        {/* Width variations */}
        <div className="space-y-3">
          <h4 className="font-medium text-gray-700">Width Behaviors</h4>
          <div className="flex items-center gap-4">
            <PopoverMenu
              items={basicMenuItems}
              variant="dropdown"
              label="Auto Width"
              width="auto"
              outline={true}
            />
            <PopoverMenu
              items={basicMenuItems}
              variant="dropdown"
              label="Match Trigger Width"
              width="trigger"
              outline={true}
            />
          </div>
        </div>
      </div>
    </div>
  ),
}

// Overflow variants
export const OverflowVariants = {
  render: () => (
    <div className="space-y-8 p-8">
      <h3 className="mb-4 text-lg font-semibold">Overflow Menu Variants</h3>

      <div className="space-y-6">
        {/* Size variations */}
        <div className="space-y-3">
          <h4 className="font-medium text-gray-700">Sizes</h4>
          <div className="flex items-center gap-4">
            <PopoverMenu items={basicMenuItems} variant="overflow" size="sm" />
            <PopoverMenu items={basicMenuItems} variant="overflow" size="md" />
            <PopoverMenu items={basicMenuItems} variant="overflow" size="lg" />
          </div>
        </div>

        {/* In context example */}
        <div className="space-y-3">
          <h4 className="font-medium text-gray-700">In Context</h4>
          <div className="rounded-lg bg-gray-100 p-4">
            <div className="space-y-3">
              {['Task 1', 'Task 2', 'Task 3'].map((task, index) => (
                <div
                  key={index}
                  className="flex items-center justify-between rounded border bg-white p-3"
                >
                  <div className="flex items-center gap-3">
                    <div className="h-2 w-2 rounded-full bg-green-400"></div>
                    <span className="text-sm font-medium text-gray-500">
                      {task}
                    </span>
                  </div>
                  <PopoverMenu
                    items={[
                      {
                        id: 'complete',
                        label: 'Mark Complete',
                        onClick: () => alert(`Complete ${task}`),
                        type: 'item',
                      },
                      {
                        id: 'edit',
                        label: 'Edit',
                        onClick: () => alert(`Edit ${task}`),
                        type: 'item',
                      },
                      { id: 'sep1', type: 'divider' },
                      {
                        id: 'delete',
                        label: 'Delete',
                        onClick: () => alert(`Delete ${task}`),
                        type: 'item',
                      },
                    ]}
                    variant="overflow"
                    size="sm"
                    placement="bottom-end"
                  />
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </div>
  ),
}

const FiltersComponent: React.FC = () => {
  const [filters, setFilters] = useState<{ users: string[]; colors: string[] }>(
    { users: [], colors: [] },
  )
  const setFilter = (label: 'users' | 'colors', value: string | undefined) => {
    setFilters((prev) => {
      if (!value) return { ...prev, [label]: [] }
      return prev[label].includes(value)
        ? { ...prev, [label]: prev[label].filter((i) => i !== value) }
        : { ...prev, [label]: [...prev[label], value] }
    })
  }
  const usersItems: MenuItem[] = [
    { id: '1', label: 'John Doe' },
    { id: '2', label: 'John Foe' },
    { id: '3', label: 'Mary Jane' },
    { id: '4', label: 'Lou Lou' },
    { id: '5', label: 'Test Test' },
  ]
  const colorsItems: MenuItem[] = colors.map((c) => ({
    label: c,
    id: c,
  }))
  return (
    <div className="space-y-8 p-8">
      <h3 className="mb-4 text-lg font-semibold">Filter Menu Variants</h3>

      <div className="flex flex-col items-start gap-4">
        <h4 className="font-medium text-gray-500">
          - max displayed selected items: default (3)
          <br />- closing dropdown on item click: false
        </h4>
        <PopoverMenu
          label="Users"
          items={usersItems.map((i) => ({
            ...i,
            selected: filters.users.includes(i.label as string),
            onClick: () => setFilter('users', i.label as string),
          }))}
          variant="filter"
          filterProps={{
            displayedValues: (
              <div className="flex items-center gap-1">
                {filters.users.slice(0, 3).map((s) => (
                  <div
                    key={s}
                    className="flex size-4.5 min-w-4.5 items-center justify-center rounded-full bg-gray-500"
                  >
                    {s.split(' ')[0][0]}
                    {s.split(' ')[1][0]}
                  </div>
                ))}
              </div>
            ),
            valuesCount: filters.users?.length || 0,
            onClearClick: () => setFilter('users', undefined),
          }}
          size="md"
          closeOnItemClick={false}
        />

        <h4 className="mt-4 font-medium text-gray-500">
          - max displayed selected items: 2
          <br />- closing dropdown on item click: true
        </h4>
        <PopoverMenu
          label="Colors"
          items={colorsItems.map((i) => ({
            ...i,
            selected: filters.colors.includes(i.label as string),
            onClick: () => setFilter('colors', i.label as string),
          }))}
          variant="filter"
          filterProps={{
            displayedValues: (
              <div className="flex items-center gap-1">
                {filters.colors.slice(0, 2).map((c) => (
                  <div
                    key={c}
                    className="size-4.5 min-w-4.5 rounded-full"
                    style={{ backgroundColor: c }}
                  />
                ))}
              </div>
            ),
            maxDisplayedValues: 2,
            valuesCount: filters.colors.length || 0,
            onClearClick: () => setFilter('colors', undefined),
          }}
          size="md"
        />
      </div>
    </div>
  )
}
// Filter variants
export const FilterVariants = {
  render: () => <FiltersComponent />,
}
const colors = [
  '#EF4444',
  '#F97316',
  '#EAB308',
  '#22C55E',
  '#3B82F6',
  '#8B5CF6',
]

const CustomComponentsComponent = () => {
  const [selectedColor, setSelectedColor] = useState('#3B82F6')
  const [volume, setVolume] = useState(75)

  const ColorPicker = () => (
    <div className="px-3 py-2">
      <div className="mb-2 text-xs font-medium text-gray-500">Choose Color</div>
      <div className="flex gap-1">
        {colors.map((color) => (
          <button
            key={color}
            onClick={() => setSelectedColor(color)}
            className="h-6 w-6 rounded border-2 border-white shadow-sm transition-transform hover:scale-110"
            style={{ backgroundColor: color }}
          />
        ))}
      </div>
    </div>
  )

  const UserCard = () => (
    <div className="flex items-center gap-3 px-3 py-3 hover:bg-gray-50">
      <div className="flex h-10 w-10 items-center justify-center rounded-full bg-gradient-to-r from-blue-400 to-purple-500 text-sm font-medium text-white">
        JD
      </div>
      <div className="flex-1">
        <div className="text-sm font-medium text-gray-900">John Doe</div>
        <div className="text-xs text-gray-500">john@example.com</div>
      </div>
    </div>
  )

  const VolumeSlider = () => (
    <div className="px-3 py-3">
      <div className="mb-2 flex items-center justify-between">
        <span className="text-xs font-medium text-gray-500">Volume</span>
        <span className="text-xs text-gray-700">{volume}%</span>
      </div>
      <input
        type="range"
        min={0}
        max={100}
        value={volume}
        onChange={(e) => setVolume(parseInt(e.target.value))}
        className="h-2 w-full cursor-pointer appearance-none rounded-lg bg-gray-200"
      />
    </div>
  )

  const colorMenuItems: MenuItem[] = [
    { id: 'color-picker', component: <ColorPicker />, type: 'custom' },
    { type: 'divider', id: 'sep1' },
    {
      id: 'custom',
      label: 'Custom Color...',
      onClick: () => alert('Custom color picker'),
    },
    {
      id: 'reset',
      label: 'Reset',
      onClick: () => setSelectedColor('#3B82F6'),
    },
  ]

  const userMenuItems: MenuItem[] = [
    { id: 'user-card', component: <UserCard />, type: 'custom' },
    { type: 'divider', id: 'sep1' },
    { id: 'profile', label: 'View Profile', onClick: () => alert('Profile') },
    { id: 'settings', label: 'Settings', onClick: () => alert('Settings') },
    { type: 'divider', id: 'sep2' },
    { id: 'logout', label: 'Sign Out', onClick: () => alert('Logout') },
  ]

  const controlMenuItems: MenuItem[] = [
    { id: 'volume', component: <VolumeSlider />, type: 'custom' },
    { type: 'divider', id: 'sep1' },
    { id: 'mute', label: 'Mute', onClick: () => setVolume(0) },
    { id: 'max', label: 'Max Volume', onClick: () => setVolume(100) },
  ]

  return (
    <div className="space-y-8 p-8">
      <h3 className="mb-4 text-lg font-semibold">Custom Component Examples</h3>

      <div className="grid grid-cols-1 gap-6 md:grid-cols-3">
        {/* Color Picker Menu */}
        <div className="space-y-3">
          <h4 className="font-medium text-gray-700">Color Picker</h4>
          <div className="flex items-center gap-3">
            <div
              className="h-6 w-6 rounded border border-gray-200"
              style={{ backgroundColor: selectedColor }}
            />
            <PopoverMenu
              items={colorMenuItems}
              variant="dropdown"
              label="Color"
              outline={true}
              menuClassName="min-w-48"
            />
          </div>
        </div>

        {/* User Menu */}
        <div className="space-y-3">
          <h4 className="font-medium text-gray-700">User Profile Menu</h4>
          <PopoverMenu
            items={userMenuItems}
            variant="dropdown"
            label="Account"
            outline={true}
            menuClassName="min-w-64"
          />
        </div>

        {/* Control Menu */}
        <div className="space-y-3">
          <h4 className="font-medium text-gray-700">Volume Control</h4>
          <PopoverMenu
            items={controlMenuItems}
            variant="overflow"
            size="md"
            menuClassName="min-w-48"
          />
        </div>
      </div>
    </div>
  )
}

// Custom components story
export const CustomComponents = {
  render: () => <CustomComponentsComponent />,
}

// Complex menu items
export const ComplexMenus = {
  render: () => {
    const StatusBadge = ({
      status,
      count,
    }: {
      status: string
      count: number
    }) => (
      <div className="flex items-center justify-between px-3 py-2 hover:bg-gray-50">
        <div className="flex items-center gap-2">
          <div
            className={`h-2 w-2 rounded-full ${
              status === 'active'
                ? 'bg-green-400'
                : status === 'pending'
                  ? 'bg-yellow-400'
                  : 'bg-gray-400'
            }`}
          />
          <span className="text-sm text-gray-700 capitalize">{status}</span>
        </div>
        <span className="rounded-full bg-gray-100 px-2 py-1 text-xs text-gray-600">
          {count}
        </span>
      </div>
    )

    const fileMenuItems: MenuItem[] = [
      {
        id: 'new',
        label: 'New File',
        icon: <EditIcon />,
        onClick: () => alert('New'),
      },
      { id: 'open', label: 'Open...', onClick: () => alert('Open') },
      { type: 'divider', id: 'sep1' },
      {
        id: 'recent-header',
        type: 'custom',
        component: (
          <div className="bg-gray-50 px-3 py-2">
            <div className="text-xs font-medium text-gray-500">
              Recent Files
            </div>
          </div>
        ),
      },
      {
        id: 'recent-1',
        type: 'custom',
        component: (
          <button className="flex w-full items-center gap-2 px-3 py-2 text-left text-sm hover:bg-gray-50">
            <div className="flex h-4 w-4 items-center justify-center rounded bg-blue-100">
              <span className="text-xs text-blue-600">📄</span>
            </div>
            <div className="min-w-0 flex-1">
              <div className="truncate text-sm font-medium text-gray-900">
                Project Proposal.docx
              </div>
              <div className="text-xs text-gray-500">2 hours ago</div>
            </div>
          </button>
        ),
      },
      {
        id: 'recent-2',
        type: 'custom',
        component: (
          <button className="flex w-full items-center gap-2 px-3 py-2 text-left text-sm hover:bg-gray-50">
            <div className="flex h-4 w-4 items-center justify-center rounded bg-green-100">
              <span className="text-xs text-green-600">📊</span>
            </div>
            <div className="min-w-0 flex-1">
              <div className="truncate text-sm font-medium text-gray-900">
                Q4 Report.xlsx
              </div>
              <div className="text-xs text-gray-500">Yesterday</div>
            </div>
          </button>
        ),
      },
      { type: 'divider', id: 'sep2' },
      { id: 'save', label: 'Save', onClick: () => alert('Save') },
      { id: 'save-as', label: 'Save As...', onClick: () => alert('Save As') },
    ]

    const statusMenuItems: MenuItem[] = [
      {
        id: 'active-status',
        type: 'custom',
        component: <StatusBadge status="active" count={12} />,
      },
      {
        id: 'pending-status',
        type: 'custom',
        component: <StatusBadge status="pending" count={5} />,
      },
      {
        id: 'inactive-status',
        type: 'custom',
        component: <StatusBadge status="inactive" count={3} />,
      },
      { type: 'divider', id: 'sep1' },
      {
        id: 'refresh',
        label: 'Refresh Status',
        icon: <PlayIcon />,
        onClick: () => alert('Refresh'),
      },
      { id: 'manage', label: 'Manage...', onClick: () => alert('Manage') },
    ]

    return (
      <div className="space-y-8 p-8">
        <h3 className="mb-4 text-lg font-semibold">Complex Menu Examples</h3>

        <div className="grid grid-cols-1 gap-8 md:grid-cols-2">
          {/* File Menu */}
          <div className="space-y-3">
            <h4 className="font-medium text-gray-700">
              File Menu with Recent Items
            </h4>
            <PopoverMenu
              items={fileMenuItems}
              variant="dropdown"
              label="File"
              outline={true}
              menuClassName="min-w-72"
            />
            <p className="text-sm text-gray-500">
              Complex menu with headers, recent files, and mixed content
            </p>
          </div>

          {/* Status Menu */}
          <div className="space-y-3">
            <h4 className="font-medium text-gray-700">Status Overview Menu</h4>
            <PopoverMenu
              items={statusMenuItems}
              variant="dropdown"
              label="Status"
              outline={true}
              menuClassName="min-w-52"
            />
            <p className="text-sm text-gray-500">
              Interactive status indicators with counts and actions
            </p>
          </div>
        </div>
      </div>
    )
  },
}

// Different placements
export const Placements = {
  render: () => (
    <div className="p-16">
      <h3 className="mb-8 text-center text-lg font-semibold">
        Menu Placements
      </h3>

      <div className="mx-auto grid max-w-md grid-cols-3 gap-8">
        {/* Top row */}
        <PopoverMenu
          items={basicMenuItems}
          placement="top-start"
          label="top-start"
          outline={true}
          size="sm"
        />
        <PopoverMenu
          items={basicMenuItems}
          placement="top"
          label="top"
          outline={true}
          size="sm"
        />
        <PopoverMenu
          items={basicMenuItems}
          placement="top-end"
          label="top-end"
          outline={true}
          size="sm"
        />

        {/* Middle row */}
        <PopoverMenu
          items={basicMenuItems}
          placement="left"
          label="left"
          outline={true}
          size="sm"
        />
        <div className="flex items-center justify-center">
          <div className="text-sm text-gray-400">Menu Trigger</div>
        </div>
        <PopoverMenu
          items={basicMenuItems}
          placement="right"
          label="right"
          outline={true}
          size="sm"
        />

        {/* Bottom row */}
        <PopoverMenu
          items={basicMenuItems}
          placement="bottom-start"
          label="bottom-start"
          outline={true}
          size="sm"
        />
        <PopoverMenu
          items={basicMenuItems}
          placement="bottom"
          label="bottom"
          outline={true}
          size="sm"
        />
        <PopoverMenu
          items={basicMenuItems}
          placement="bottom-end"
          label="bottom-end"
          outline={true}
          size="sm"
        />
      </div>
    </div>
  ),
}

// Custom triggers
export const CustomTriggers = {
  render: () => (
    <div className="space-y-8 p-8">
      <h3 className="mb-4 text-lg font-semibold">Custom Trigger Examples</h3>

      <div className="grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-3">
        {/* Avatar trigger */}
        <div className="space-y-3">
          <h4 className="font-medium text-gray-700">Avatar Menu</h4>
          <PopoverMenu
            items={[
              {
                id: 'profile',
                label: 'View Profile',
                onClick: () => alert('Profile'),
              },
              {
                id: 'settings',
                label: 'Settings',
                onClick: () => alert('Settings'),
              },
              { type: 'divider', id: 'sep1' },
              {
                id: 'logout',
                label: 'Sign Out',
                onClick: () => alert('Logout'),
              },
            ]}
            placement="bottom-end"
          >
            <button className="flex items-center gap-2 rounded-lg p-2 transition-colors hover:bg-gray-100">
              <div className="flex h-8 w-8 items-center justify-center rounded-full bg-gradient-to-r from-purple-400 to-pink-400 text-sm font-medium text-white">
                JD
              </div>
              <span className="text-sm text-gray-700">John Doe</span>
            </button>
          </PopoverMenu>
        </div>

        {/* Icon button trigger */}
        <div className="space-y-3">
          <h4 className="font-medium text-gray-700">Icon Button</h4>
          <PopoverMenu items={basicMenuItems} placement="bottom-start">
            <button className="flex h-10 w-10 items-center justify-center rounded-lg bg-blue-500 text-white transition-colors hover:bg-blue-600">
              <EditIcon />
            </button>
          </PopoverMenu>
        </div>

        {/* Card trigger */}
        <div className="space-y-3">
          <h4 className="font-medium text-gray-700">Card Menu</h4>
          <PopoverMenu
            items={[
              { id: 'edit', label: 'Edit Card', onClick: () => alert('Edit') },
              {
                id: 'duplicate',
                label: 'Duplicate',
                onClick: () => alert('Duplicate'),
              },
              { id: 'move', label: 'Move to...', onClick: () => alert('Move') },
              { type: 'divider', id: 'sep1' },
              {
                id: 'archive',
                label: 'Archive',
                onClick: () => alert('Archive'),
              },
              { id: 'delete', label: 'Delete', onClick: () => alert('Delete') },
            ]}
            placement="bottom-end"
          >
            <div className="cursor-pointer rounded-lg border border-gray-200 bg-white p-4 transition-shadow hover:shadow-md">
              <div className="flex items-start justify-between">
                <div>
                  <h4 className="font-medium text-gray-900">Task Card</h4>
                  <p className="mt-1 text-sm text-gray-500">
                    Click to open menu
                  </p>
                </div>
                <div className="text-gray-400">⋯</div>
              </div>
            </div>
          </PopoverMenu>
        </div>
      </div>
    </div>
  ),
}

// Disabled state
export const DisabledState = {
  render: () => (
    <div className="space-y-6 p-8">
      <h3 className="mb-4 text-lg font-semibold">Disabled State</h3>

      <div className="flex gap-4">
        <PopoverMenu
          items={basicMenuItems}
          variant="dropdown"
          label="Disabled Dropdown"
          outline={true}
          disabled={true}
        />

        <PopoverMenu
          items={basicMenuItems}
          variant="overflow"
          disabled={true}
        />

        <PopoverMenu
          items={basicMenuItems}
          variant="dropdown"
          label="Enabled"
          outline={true}
          disabled={false}
        />
      </div>

      <p className="text-sm text-gray-500">
        Disabled menus don’t respond to interactions and show appropriate visual
        states.
      </p>
    </div>
  ),
}
