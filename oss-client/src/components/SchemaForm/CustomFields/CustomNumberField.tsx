import { asNumber, FieldProps } from '@rjsf/utils'
import React, { ChangeEvent, useCallback, useMemo } from 'react'
import { useHtmlElementProps } from './hooks'

const CustomNumberField = <T = number | bigint,>(fieldProps: FieldProps<T>) => {
  const elementProps = useHtmlElementProps(fieldProps)
  const convertIfLargeInteger = (
    evTargetValue: string,
  ): ReturnType<typeof asNumber> | bigint => {
    if (asNumber(evTargetValue)?.toString() === evTargetValue) {
      return Number(evTargetValue)
    }
    if (BigInt(evTargetValue).toString() === evTargetValue) {
      return BigInt(evTargetValue)
    }
    return evTargetValue
  }

  const handleOnChange = useCallback(
    (ev: ChangeEvent<HTMLInputElement>) => {
      fieldProps.onChange(
        convertIfLargeInteger(ev.target.value) as unknown as T,
      )
    },
    [fieldProps],
  )
  const handleOnBlur = useCallback(() => {
    fieldProps.onBlur?.(fieldProps.idSchema?.$id ?? '', fieldProps.formData)
  }, [fieldProps])

  const inputValue = useMemo(
    () =>
      [fieldProps.formData, fieldProps.defaultValue, fieldProps.schema.default]
        .filter((v) => v !== undefined)?.[0]
        ?.toString(),
    [fieldProps.formData, fieldProps.defaultValue, fieldProps.schema.default],
  )

  return (
    <input
      {...elementProps}
      type="number"
      onChange={handleOnChange}
      onBlur={handleOnBlur}
      value={inputValue}
    />
  )
}

export default CustomNumberField
