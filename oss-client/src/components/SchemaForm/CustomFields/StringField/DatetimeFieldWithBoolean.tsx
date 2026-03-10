import { FieldProps } from '@rjsf/utils'
import React, { useCallback, useMemo } from 'react'
import BaseStringField from './BaseStringField'

export type DatetimeFieldWithBooleanConfig = {
  /** The name of the currently selected field, or undefined if none */
  selectedField?: string
  /** Callback to set or clear the selected field */
  setSelectedField?: (fieldName: string | undefined) => void
  /** Text to display as the checkbox label (default: "Use as boolean field") */
  checkboxLabel?: string
  /** Text to display when the field is disabled (default: "Set automatically") */
  disabledPlaceholder?: string
}

/**
 * A datetime field that includes a checkbox to mark the field as "selected".
 * When selected, the datetime input is replaced with a placeholder.
 *
 * This can be used for scenarios like:
 * - Marking a datetime input to receive a scheduled kickoff time
 * - Designating a field for automatic value injection
 */
const DatetimeFieldWithBoolean = (
  fieldProps: FieldProps & { type?: string },
) => {
  const { formContext, name } = fieldProps
  const config = formContext as DatetimeFieldWithBooleanConfig | undefined

  const isSelected = useMemo(() => {
    return config?.selectedField === name
  }, [config?.selectedField, name])

  const handleCheckboxChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      if (e.target.checked) {
        config?.setSelectedField?.(name)
      } else {
        config?.setSelectedField?.(undefined)
        fieldProps.onChange(undefined)
      }
    },
    [config, fieldProps, name],
  )

  const checkboxLabel = config?.checkboxLabel ?? 'Use automatic value'
  const disabledPlaceholder = config?.disabledPlaceholder ?? 'Set automatically'

  return (
    <div className="flex h-12 flex-col gap-1">
      {isSelected ? (
        <div className="text-md w-full rounded border border-zinc-300 bg-zinc-100 px-2 py-1 text-zinc-500 dark:border-zinc-700 dark:bg-zinc-800 dark:text-zinc-400">
          {disabledPlaceholder}
        </div>
      ) : (
        <BaseStringField {...fieldProps} />
      )}
      <label
        className="flex cursor-pointer items-center gap-2 text-xs"
        htmlFor={`${name}-checkbox`}
      >
        <input
          id={`${name}-checkbox`}
          type="checkbox"
          checked={isSelected}
          onChange={handleCheckboxChange}
          className="h-4 w-4"
        />
        <span className="dark:text-(--system-gray-5)">{checkboxLabel}</span>
      </label>
    </div>
  )
}

export default DatetimeFieldWithBoolean
