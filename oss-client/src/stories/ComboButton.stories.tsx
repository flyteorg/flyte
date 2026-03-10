import React from 'react'
import ComboButton, { type ComboButtonProps } from '../components/ComboButton'

const ComboBoxStory = {
  title: 'Components/ComboButton',
  component: ComboButton,
  argTypes: {
    color: {
      control: { type: 'select' },
      options: [
        'dark/zinc',
        'light',
        'dark/white',
        'dark',
        'white',
        'zinc',
        'indigo',
        'orange',
        'amber',
        'yellow',
        'lime',
        'green',
        'emerald',
        'teal',
        'sky',
        'blue',
        'violet',
        'purple',
        'fuchsia',
        'pink',
        'rose',
        'union',
      ],
      defaultValue: 'orange',
    },
    size: {
      control: { type: 'select' },
      options: ['xs', 'sm', 'md', 'lg', 'xl'],
      defaultValue: 'md',
    },
    options: { control: false },
  },
}

const options = [
  { name: 'Abort Run', onClick: () => console.log('Aborted!') },
  { name: 'Pause Run', onClick: () => console.log('Paused!') },
  { name: 'Retry Run', onClick: () => console.log('Retried!') },
]

export const Default = (args: ComboButtonProps) => (
  <div className="flex min-h-screen items-center justify-center p-8">
    <ComboButton color={args.color} size={args.size} options={options} />
  </div>
)

export default ComboBoxStory
