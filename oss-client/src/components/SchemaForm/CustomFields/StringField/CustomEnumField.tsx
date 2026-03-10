import { FieldProps } from '@rjsf/utils'
import { ChangeEvent, useCallback } from 'react'
import { fromFormattedValueToText } from '../../utils'
import { useHtmlElementProps } from '../hooks'

const CustomEnumField = (fieldProps: FieldProps & { type?: string }) => {
  const elementProps = useHtmlElementProps(fieldProps)
  const textValue = fromFormattedValueToText(
    `${fieldProps?.formData || ''}`,
    fieldProps.schema.format,
  )

  const handleSelectChange = useCallback(
    (ev: ChangeEvent<HTMLSelectElement>) => {
      fieldProps.onChange(ev.target.value)
    },
    [fieldProps],
  )

  const handleSelectFocus = useCallback(
    (ev: React.FocusEvent<HTMLSelectElement>) => {
      fieldProps.onFocus?.(fieldProps.id ?? '', ev.target.value)
    },
    [fieldProps],
  )

  const handleSelectBlur = useCallback(() => {
    fieldProps.onBlur?.(fieldProps.idSchema?.$id ?? '', fieldProps.formData)
  }, [fieldProps])

  // Extract only compatible props for select element
  const {
    onChange: _onChange,
    onFocus: _onFocus,
    onBlur: _onBlur,
    ...selectProps
  } = elementProps

  return (
    <select
      {...selectProps}
      value={textValue}
      onChange={handleSelectChange}
      onFocus={handleSelectFocus}
      onBlur={handleSelectBlur}
    >
      {fieldProps.schema.enum?.map((enumValue) => (
        <option key={String(enumValue)} value={String(enumValue)}>
          {String(enumValue)}
        </option>
      ))}
    </select>
  )
}

export default CustomEnumField
