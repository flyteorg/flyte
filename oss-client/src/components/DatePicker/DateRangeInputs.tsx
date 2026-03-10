import React from 'react'
import { Input, InputGroup } from '../Input'

interface DateRangeInputsProps {
  fromInput: string
  toInput: string
  hasTimeComponent: boolean
  onFromInputChange: (value: string) => void
  onToInputChange: (value: string) => void
  onFromInputBlur: (e: React.FocusEvent<HTMLInputElement>) => void
  onToInputBlur: (e: React.FocusEvent<HTMLInputElement>) => void
}

export const DateRangeInputs: React.FC<DateRangeInputsProps> = ({
  fromInput,
  toInput,
  hasTimeComponent,
  onFromInputChange,
  onToInputChange,
  onFromInputBlur,
  onToInputBlur,
}) => {
  const placeholder = hasTimeComponent ? 'YYYY-MM-DD HH:MM' : 'YYYY-MM-DD'
  const timeHint = hasTimeComponent && (
    <span className="text-xs opacity-60">(YYYY-MM-DD HH:MM)</span>
  )

  return (
    <div className="flex w-full">
      <div className="w-1/2 px-2">
        <label
          className="text-xs dark:text-(--system-gray-5)"
          htmlFor="from-date"
        >
          From {timeHint}
        </label>
        <InputGroup>
          <Input
            className="bg-transparent dark:bg-white/5"
            hideBorder
            noBackground
            id="from-date"
            onBlur={onFromInputBlur}
            onChange={(e) => onFromInputChange(e.target.value)}
            value={fromInput}
            placeholder={placeholder}
          />
        </InputGroup>
      </div>
      <div className="w-1/2 px-2">
        <label
          className="text-sm dark:text-(--system-gray-5)"
          htmlFor="to-date"
        >
          To {timeHint}
        </label>
        <Input
          hideBorder
          id="to-date"
          className="bg-transparent dark:bg-white/5"
          onBlur={onToInputBlur}
          onChange={(e) => onToInputChange(e.target.value)}
          value={toInput}
          placeholder={placeholder}
        />
      </div>
    </div>
  )
}
