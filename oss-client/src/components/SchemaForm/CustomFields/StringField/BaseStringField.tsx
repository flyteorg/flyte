import { FieldProps } from '@rjsf/utils'
import React, {
  ChangeEvent,
  useCallback,
  useEffect,
  useMemo,
  useRef,
} from 'react'
import { useHtmlElementProps } from '../hooks'
import {
  fromFormattedValueToText,
  fromTextToFormattedValue,
  getSchemaFormatFallback,
} from '../../utils'

const LONG_TEXT_THRESHOLD = 50

const BaseStringField = (fieldProps: FieldProps & { type?: string }) => {
  const ref = useRef<HTMLTextAreaElement>(null)
  const elementProps = useHtmlElementProps(fieldProps)
  const textValue = fromFormattedValueToText(
    `${fieldProps?.formData || ''}`,
    fieldProps.schema.format,
  )
  const shouldUseTextarea =
    textValue.length > LONG_TEXT_THRESHOLD || textValue.includes('\n')

  // Check if this is an enum field
  // const isEnum = useMemo(() => {
  //   return (
  //     Array.isArray(fieldProps.schema.enum) && fieldProps.schema.enum.length > 0
  //   )
  // }, [fieldProps.schema.enum])

  useEffect(() => {
    if (ref.current) {
      ref.current.style.height =
        ref.current.scrollHeight +
        (ref.current.offsetHeight - ref.current.clientHeight) +
        'px'
    }
  }, [fieldProps.formData, fieldProps.readonly])

  const fieldType = useMemo(() => {
    switch (getSchemaFormatFallback(fieldProps.schema.format)) {
      case 'date':
        return 'date'
      case 'datetime':
        return 'datetime-local'
      case 'time':
        return 'time'
      default:
        return fieldProps.type ?? 'text'
    }
  }, [fieldProps.schema, fieldProps.type])

  const handleOnChange = useCallback(
    (ev: ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
      const value = fromTextToFormattedValue(
        ev.target.value,
        fieldProps.schema.format,
      )
      // undefined passed to treat the field as empty properly
      fieldProps.onChange(value === '' ? undefined : value)
    },
    [fieldProps],
  )
  const handleOnBlur = useCallback(() => {
    fieldProps.onBlur?.(fieldProps.idSchema?.$id ?? '', fieldProps.formData)
  }, [fieldProps])

  return shouldUseTextarea ? (
    <textarea
      {...elementProps}
      value={textValue}
      ref={ref}
      onChange={handleOnChange}
      onBlur={handleOnBlur}
    />
  ) : (
    <input
      {...elementProps}
      type={fieldType}
      value={textValue}
      onChange={handleOnChange}
      onBlur={handleOnBlur}
    />
  )
}

export default BaseStringField
