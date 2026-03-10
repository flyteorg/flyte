import {
  Switch as HeadlessSwitch,
  SwitchProps as HeadlessSwitchProps,
} from '@headlessui/react'
import clsx from 'clsx'

const containerClasses = {
  color: {
    gray: 'border-(--system-gray-4) bg-(--bg-gray) data-checked:bg-zinc-600',
    green: clsx(
      'border-(--system-gray-4) bg-(--system-gray-4)',
      'data-checked:border-(--accent-green) data-checked:bg-(--accent-green)',
      'disabled:data-checked:border-(--system-gray-4) disabled:data-checked:bg-(--system-gray-4)',
    ),
  },
  size: { sm: 'h-4', md: 'h-5' },
}
const switcherClasses = {
  color: { gray: 'border-(--system-gray-4)', green: 'border-(--system-white)' },
  size: {
    sm: 'group-data-checked:translate-x-4 size-3',
    md: 'group-data-checked:translate-x-3 size-4',
  },
}

type SwitchProps = HeadlessSwitchProps & {
  color?: 'gray' | 'green'
  size?: 'sm' | 'md'
}

export function Switch({
  checked,
  onChange,
  color = 'gray',
  size = 'md',
  ...props
}: SwitchProps) {
  return (
    <HeadlessSwitch
      checked={checked}
      onChange={onChange}
      onClick={(e) => {
        e.stopPropagation()
      }}
      className={clsx(
        'group relative inline-flex w-8 shrink-0 cursor-pointer rounded-full border p-px transition-colors duration-200 ease-in-out',
        containerClasses.color[color],
        containerClasses.size[size],
      )}
      {...props}
    >
      <span className="sr-only">Use setting</span>
      <span
        aria-hidden="true"
        className={clsx(
          'pointer-events-none inline-block transform rounded-full bg-white ring-0 transition-transform duration-200 ease-in-out',
          switcherClasses.color[color],
          switcherClasses.size[size],
        )}
      />
    </HeadlessSwitch>
  )
}
