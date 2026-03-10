import { FieldProps } from '@rjsf/utils'
import { ChangeEvent, useCallback } from 'react'

export type UseTypeSafeHandlers = Pick<
  FieldProps,
  'id' | 'onBlur' | 'onFocus' | 'onChange' | 'onError'
>

export const useTypeSafeHandlers = <
  T extends HTMLInputElement | HTMLTextAreaElement,
>({
  id,
  onBlur,
  onFocus,
  onChange,
  onError,
}: UseTypeSafeHandlers) => {
  const typeSafeOnFocus = useCallback(
    (ev: ChangeEvent<T>) => onFocus(id ?? '', ev.target.value),
    [id, onFocus],
  )

  const typeSafeOnBlur = useCallback(
    (ev: ChangeEvent<T>) => onBlur(id ?? '', ev.target.value),
    [id, onBlur],
  )

  const typeSafeOnChange = useCallback(
    (ev: ChangeEvent<T>) => {
      onChange(ev.target.value)
    },
    [onChange],
  )

  return {
    onChange: typeSafeOnChange,
    onError,
    onBlur: typeSafeOnBlur,
    onFocus: typeSafeOnFocus,
  }
}
