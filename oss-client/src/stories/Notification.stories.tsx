import {
  dialogPanelClassName,
  NotificationContent,
  NotificationProps,
} from '@/components/Notification'
import {
  NotificationsProvider,
  useNotifications,
} from '@/providers/notifications'
import { Transition } from '@headlessui/react'
import { Meta, StoryObj } from '@storybook/nextjs'
import { Fragment } from 'react'

// Component mimics the behavior of the Dialog & DialogPanel which won't show up properly inside Storybook
const MockNotification = ({
  notificationInfo,
  hideNotification,
}: NotificationProps) => (
  <Transition
    show={!!notificationInfo}
    as={Fragment}
    enter="transform ease-out duration-300 transition"
    enterFrom="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
    enterTo="opacity-100 translate-y-0 sm:scale-100"
    leave="transform ease-in duration-200 transition"
    leaveFrom="opacity-100 translate-y-0 sm:scale-100"
    leaveTo="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
  >
    <div className={dialogPanelClassName}>
      <NotificationContent {...{ notificationInfo, hideNotification }} />
    </div>
  </Transition>
)

const meta: Meta<typeof MockNotification> = {
  title: 'primitives/Notification',
  component: MockNotification,
  tags: ['autodocs'],
  decorators: [
    (Story) => (
      <div className="relative h-[150px] w-full">
        <Story />
      </div>
    ),
  ],
  args: {
    notificationInfo: {
      title: 'Notification message',
      variant: 'success',
      text: 'some extra text',
      hasCloseBtn: true,
    },
    hideNotification: () => console.log('Notification closed'),
  },
  parameters: {
    layout: 'fullscreen',
    docs: {
      description: {
        component: `
The Notification component is used within the NotificationsProvider to display a message after an action.

## Features
- color variants for different types of notifications
- can have additional action button
- closes itself but also allows manual closing

## Usage
\`\`\`tsx
import { useNotifications } from '@/providers/notifications'

const { showNotification, hideNotification } = useNotifications()

// show
showNotification('Notification message', 'success')
// hide
hideNotification()

\`\`\`
        `,
      },
    },
  },
  argTypes: {
    hideNotification: { action: 'hideNotification called' },
    notificationInfo: {
      description:
        'Contains information about the notification - texts and color variant',
      control: 'object',
    },
  },
}

export default meta
type Story = StoryObj<typeof MockNotification>

export const Success: Story = {
  args: {
    notificationInfo: {
      title: 'Operation successful',
      variant: 'success',
    },
    hideNotification: () => console.log('hideNotification called'),
  },
}

export const Error: Story = {
  args: {
    notificationInfo: {
      title: 'Something went wrong',
      variant: 'error',
    },
    hideNotification: () => console.log('hideNotification called'),
  },
}

export const Warning: Story = {
  args: {
    notificationInfo: {
      title: 'Warning info',
      variant: 'warning',
    },
    hideNotification: () => console.log('hideNotification called'),
  },
}

export const Running: Story = {
  args: {
    notificationInfo: {
      title: 'Running info',
      variant: 'running',
    },
    hideNotification: () => console.log('hideNotification called'),
  },
}

export const WithExtraText: Story = {
  name: 'With extra text',
  args: {
    notificationInfo: {
      title: 'Main message',
      variant: 'running',
      text: 'some extra text',
    },
    hideNotification: () => console.log('hideNotification called'),
  },
}

export const WithButton: Story = {
  name: 'With button',
  args: {
    notificationInfo: {
      title: 'Main message',
      variant: 'error',
      text: undefined,
      button: {
        label: 'Click me',
        onClick: () => console.log('Button clicked'),
      },
    },
    hideNotification: () => console.log('hideNotification called'),
  },
}

export const WithExtraTextAndButton: Story = {
  name: 'With extra text and button',
  args: {
    notificationInfo: {
      title: 'Main message',
      variant: 'success',
      text: 'some extra text',
      button: {
        label: 'View details',
        onClick: () => console.log('Button clicked'),
      },
    },
    hideNotification: () => console.log('hideNotification called'),
  },
}

export const WithoutCloseButton: Story = {
  name: "Without 'close' button",
  args: {
    notificationInfo: {
      title: 'Main message',
      variant: 'running',
      button: {
        label: 'View details',
        onClick: () => console.log('Button clicked'),
      },
      hasCloseBtn: false,
    },
    hideNotification: () => console.log('hideNotification called'),
  },
}

export const LongTexts: Story = {
  name: 'Long title and text',
  args: {
    notificationInfo: {
      title:
        'Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.',
      text: 'Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.',
      variant: 'running',
      button: {
        label: 'Show more',
        onClick: () => console.log('Button clicked'),
      },
    },
    hideNotification: () => console.log('hideNotification called'),
  },
}

const DemoOpenCloseBehaviour = () => {
  const { showNotification } = useNotifications()
  return (
    <div className="m-3 flex h-[150px] flex-col gap-4">
      <button
        className="w-xs rounded bg-green-800 p-2 text-white hover:bg-green-900"
        onClick={() => showNotification('Operation successful', 'success')}
      >
        Show Example
      </button>
    </div>
  )
}
export const OpenCloseBehaviour: Story = {
  name: 'Open/close behaviour',
  render: () => (
    <NotificationsProvider>
      <DemoOpenCloseBehaviour />
    </NotificationsProvider>
  ),
}
