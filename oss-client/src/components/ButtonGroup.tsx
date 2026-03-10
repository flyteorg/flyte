import React from 'react'
import { Button } from './Button'
import clsx from 'clsx'

export interface ButtonGroupOption {
  label: string
  value: string
  onClick?: () => void
}

export interface ButtonGroupProps {
  options: ButtonGroupOption[]
  value?: string
  onChange?: (value: string) => void
  size?: 'xs' | 'sm' | 'md' | 'lg' | 'xl'
  color?: 'dark/zinc' | 'light' | 'dark/white' | 'dark' | 'white' | 'zinc'
  className?: string
}

export const ButtonGroup: React.FC<ButtonGroupProps> = ({
  options,
  value,
  onChange,
  size = 'sm',
  color = 'dark/zinc',
  className,
}) => {
  return (
    <span className={clsx('isolate inline-flex overflow-hidden rounded-lg border border-white/20 bg-black/20', className)}>
      {options.map((option) => (
        <Button
          key={option.value}
          size={size}
          color={color}
          plain
          className={clsx(
            'relative',
            // Remove individual button radius and borders
            'rounded-none border-0',
            // Add vertical separator between buttons
            'border-r border-white/20',
            'last:border-r-0',
            // Add back radius on the ends
            'first:rounded-l-lg last:rounded-r-lg',
            // Custom background for selected state
            value === option.value && 'bg-black/40',
            // Ensure focused items appear on top
            'focus:z-10'
          )}
          onClick={() => {
            onChange?.(option.value)
            option.onClick?.()
          }}
        >
          {option.label}
        </Button>
      ))}
    </span>
  )
}

export default ButtonGroup 
