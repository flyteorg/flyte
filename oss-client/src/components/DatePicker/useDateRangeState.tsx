import { useState, useCallback } from 'react'
import { DateRange } from 'react-day-picker'
import { format, isValid, parseISO, startOfDay, endOfDay } from 'date-fns'

export const useDateRangeState = (initialRange?: DateRange) => {
  const [selected, setSelected] = useState<DateRange | undefined>(initialRange)
  const [fromInput, setFromInput] = useState<string>('')
  const [toInput, setToInput] = useState<string>('')
  const [hasTimeComponent, setHasTimeComponent] = useState<boolean>(false)
  const [lastClickedQuickKey, setLastClickedQuickKey] = useState<string>()

  const hasTimeInRange = useCallback(
    (range: DateRange | undefined): boolean => {
      if (!range?.from) return false

      const isTimeBasedRange =
        range.from.getHours() !== 0 ||
        range.from.getMinutes() !== 0 ||
        range.from.getSeconds() !== 0 ||
        range.from.getMilliseconds() !== 0 ||
        (range.to &&
          (range.to.getHours() !== 23 ||
            range.to.getMinutes() !== 59 ||
            range.to.getSeconds() !== 59))

      return !!isTimeBasedRange
    },
    [],
  )

  const updateInputsFromRange = useCallback(
    (range: DateRange | undefined, isTimeBased: boolean) => {
      if (range?.from) {
        const format_str = isTimeBased ? 'yyyy-MM-dd HH:mm' : 'yyyy-MM-dd'
        setFromInput(format(range.from, format_str))
      } else {
        setFromInput('')
      }

      if (range?.to) {
        const format_str = isTimeBased ? 'yyyy-MM-dd HH:mm' : 'yyyy-MM-dd'
        setToInput(format(range.to, format_str))
      } else {
        setToInput('')
      }
    },
    [],
  )

  const parseInputDate = useCallback(
    (value: string, isFromField: boolean): Date | null => {
      if (!value) return null

      let date = parseISO(value)
      if (!isValid(date)) {
        date = parseISO(`${value}T00:00:00`)
      }
      if (!isValid(date)) return null

      if (!value.includes(' ') && !value.includes('T')) {
        return isFromField ? startOfDay(date) : endOfDay(date)
      }

      return date
    },
    [],
  )

  return {
    selected,
    setSelected,
    fromInput,
    setFromInput,
    toInput,
    setToInput,
    hasTimeComponent,
    setHasTimeComponent,
    hasTimeInRange,
    updateInputsFromRange,
    parseInputDate,
    lastClickedQuickKey,
    setLastClickedQuickKey,
  }
}
