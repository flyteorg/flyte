import { FieldProps } from '@rjsf/utils'
import React, { useMemo } from 'react'
import {
  BaseStringField,
  CustomDurationField,
  CustomEnumField,
  DatetimeFieldWithBoolean,
} from './StringField'
import { DatetimeFieldWithBooleanConfig } from './StringField/DatetimeFieldWithBoolean'

const CustomStringField = (fieldProps: FieldProps & { type?: string }) => {
  const isEnum = useMemo(() => {
    return (
      Array.isArray(fieldProps.schema.enum) && fieldProps.schema.enum.length > 0
    )
  }, [fieldProps.schema.enum])

  const isDuration = useMemo(
    () => fieldProps.schema.format === 'duration',
    [fieldProps.schema.format],
  )

  const isDatetime = useMemo(
    () =>
      fieldProps.schema.format === 'date-time' ||
      fieldProps.schema.format === 'datetime',
    [fieldProps.schema.format],
  )

  const hasDatetimeBooleanConfig = useMemo(() => {
    const ctx = fieldProps.formContext as
      | DatetimeFieldWithBooleanConfig
      | undefined
    return typeof ctx?.setSelectedField === 'function'
  }, [fieldProps.formContext])

  // If enum, delegate to CustomEnumField
  if (isEnum) {
    return <CustomEnumField {...fieldProps} />
  }

  // If duration, delegate to CustomDurationField
  if (isDuration) {
    return <CustomDurationField {...fieldProps} />
  }

  // If datetime with boolean config, use DatetimeFieldWithBoolean with checkbox
  if (isDatetime && hasDatetimeBooleanConfig) {
    return <DatetimeFieldWithBoolean {...fieldProps} />
  }

  return <BaseStringField {...fieldProps} />
}

export default CustomStringField
