import { FieldProps } from '@rjsf/utils'
import { useHtmlElementProps } from '../hooks'
import { ChangeEvent, useCallback, useEffect, useMemo, useState } from 'react'
import dayjs from 'dayjs'
import {
  getMillisecondsFromDurationString,
  formatDuration,
  isValidDuration,
  isValidReadableDuration,
  formatIso8601Duration,
} from './durationUtils'
import { isNumber } from 'lodash'

const CustomDurationField = <T = string,>(fieldProps: FieldProps<T>) => {
  const [rawValue, setRawValue] = useState<T | undefined>(undefined)
  const [iso8601Value, setIso8601Value] = useState<T | undefined>(undefined)
  const elementProps = useHtmlElementProps(fieldProps)

  const handleOnBlur = useCallback(
    (ev: ChangeEvent<HTMLInputElement>) => {
      const iso8601Value = formatIso8601Duration(ev.target.value)
      fieldProps.onChange(iso8601Value === '' ? undefined : (iso8601Value as T))
      setIso8601Value(iso8601Value === '' ? undefined : (iso8601Value as T))
    },
    [fieldProps, setIso8601Value],
  )

  useEffect(() => {
    fieldProps.onBlur(fieldProps.idSchema.$id, iso8601Value)
  }, [iso8601Value, fieldProps.idSchema.$id, fieldProps.onBlur, fieldProps])

  const handleOnChange = useCallback(
    (ev: ChangeEvent<HTMLInputElement>) => {
      setRawValue(ev.target.value as T)
    },
    [setRawValue],
  )

  const displayValue = useMemo(() => {
    if (rawValue !== undefined) return `${rawValue}`
    const asString = `${fieldProps.formData || ''}`
    if (!isValidDuration(asString) || isNumber(asString)) return asString
    if (isValidReadableDuration(asString)) return asString.trim()

    const rawMilliseconds = getMillisecondsFromDurationString(asString)
    if (rawMilliseconds === null || rawMilliseconds === undefined) {
      return `${fieldProps.formData}`
    }

    return formatDuration(dayjs.duration(rawMilliseconds!, 'milliseconds'), ' ')
  }, [rawValue, fieldProps.formData])

  return (
    <input
      {...elementProps}
      onBlur={handleOnBlur}
      onChange={handleOnChange}
      type="text"
      value={displayValue}
    />
  )
}

export default CustomDurationField
