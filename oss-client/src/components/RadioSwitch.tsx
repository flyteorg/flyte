import clsx from 'clsx'
import { ReactNode } from 'react'

export type RadioSwitchItem<T> = {
  label: string
  value: T
  icon?: ReactNode
}

export type RadioSwitchProps<T> = {
  items: RadioSwitchItem<T>[]
  currentValue: T
  onChange: (value: T) => void
}

export const RadioSwitch = <T,>({
  items,
  currentValue,
  onChange,
}: RadioSwitchProps<T>) => {
  if (!items.length) return null

  return (
    <div className="flex max-w-min flex-row gap-2 rounded-lg bg-(--system-black) p-1 text-xs/5">
      {items.map(({ label, icon, value }, key) => {
        const isActive = value === currentValue && items.length > 1
        return (
          <div
            key={key}
            className={clsx(
              `flex cursor-pointer flex-row flex-nowrap items-center gap-1 rounded-sm px-4 py-0.5 text-nowrap`,
              isActive
                ? 'bg-(--system-gray-3) text-(--system-white)'
                : `text-(--system-gray-5) hover:text-(--system-white)`,
            )}
            onClick={() => onChange(value)}
          >
            {icon}
            {label}
          </div>
        )
      })}
    </div>
  )
}
