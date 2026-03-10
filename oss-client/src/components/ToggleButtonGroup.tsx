import { Button } from '@/components/Button'
import { CodeIcon } from '@/components/icons/CodeIcon'
import { TextFormattedIcon } from '@/components/icons/TextFormattedIcon'
import React from 'react'

export type ToggleButtonGroupProps = {
  isRawView: boolean
  onEnableRaw: VoidFunction
  onDisableRaw: VoidFunction
}

export const ToggleButtonGroup: React.FC<ToggleButtonGroupProps> = ({
  isRawView,
  onEnableRaw,
  onDisableRaw,
}) => {
  const classes = {
    base: `items-center rounded-sm !border-none`,
    active: '!bg-(--system-gray-3) [&_[data-slot=icon]]:!text-(--system-white)',
    inactive:
      '!bg-transparent [&_[data-slot=icon]]:!text-(--system-gray-5) hover:[&_[data-slot=icon]]:!text-(--system-white)',
  }

  return (
    <div className="flex gap-1 rounded-md bg-(--system-black) p-1">
      <Button
        className={`${classes.base} ${isRawView ? classes.inactive : classes.active}`}
        onClick={onDisableRaw}
        outline
        size="xxs"
      >
        <TextFormattedIcon data-slot="icon" />
      </Button>
      <Button
        className={`${classes.base} ${isRawView ? classes.active : classes.inactive}`}
        onClick={onEnableRaw}
        outline
        size="xxs"
      >
        <CodeIcon data-slot="icon" />
      </Button>
    </div>
  )
}
