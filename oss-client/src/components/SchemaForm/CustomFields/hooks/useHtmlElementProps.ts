import { useTypeSafeHandlers } from '@/components/SchemaForm/CustomFields/hooks/useTypeSafeHandlers'
import { useObjectPropertyContext } from '@/components/SchemaForm/CustomFields/Objects/ObjectPropertyContext'
import { skipFalseAttributes } from '@/lib/htmlAttributeUtils.ts'
import { FieldProps } from '@rjsf/utils'
import { useMemo } from 'react'

export const FIELD_NAME_PREFIX = 'root_'

export const useHtmlElementProps = ({
  className,
  required,
  id,
  name,
  disabled,
  readonly,
  rawErrors,
  ...otherProps
}: FieldProps) => {
  const eventHandlers = useTypeSafeHandlers(otherProps)
  const { objectDescriptionId } = useObjectPropertyContext()

  return useMemo(() => {
    // Extract aria-describedby from otherProps if present
    const existingDescribedBy = otherProps['aria-describedby'] as
      | string
      | undefined

    // Combine existing aria-describedby with object description ID
    const describedBy = [existingDescribedBy, objectDescriptionId]
      .filter(Boolean)
      .join(' ')

    return {
      className,
      ...skipFalseAttributes({
        required,
        'aria-required': required,
        'aria-disabled': disabled,
        'aria-readonly': readonly,
        'aria-invalid': !!rawErrors?.length,
        disabled: disabled,
        readOnly: readonly,
      }),
      ...(describedBy && { 'aria-describedby': describedBy }),
      id: id ?? name,
      name: `${FIELD_NAME_PREFIX}${id ?? name}`,
      'data-testid': `input-${id ?? name}`,
      ...eventHandlers,
    }
  }, [
    id,
    className,
    required,
    eventHandlers,
    disabled,
    name,
    readonly,
    rawErrors,
    objectDescriptionId,
    otherProps,
  ])
}
