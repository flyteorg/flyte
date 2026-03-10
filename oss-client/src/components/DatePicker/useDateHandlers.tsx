import { useCallback } from 'react'
import { DateRange } from 'react-day-picker'
import { format } from 'date-fns'

interface UseDatePickerHandlersProps {
  selected: DateRange | undefined
  setSelected: (range: DateRange | undefined) => void
  hasTimeComponent: boolean
  setHasTimeComponent: (hasTime: boolean) => void
  hasTimeInRange: (range: DateRange | undefined) => boolean
  updateInputsFromRange: (
    range: DateRange | undefined,
    isTimeBased: boolean,
  ) => void
  parseInputDate: (value: string, isFromField: boolean) => Date | null
  fromInput: string
  setFromInput: (input: string) => void
  toInput: string
  setToInput: (input: string) => void
  setLastClickedQuickKey: (quickKeyLabel?: string | undefined) => void
}

export const useDatePickerHandlers = ({
  selected,
  setSelected,
  hasTimeComponent,
  setHasTimeComponent,
  hasTimeInRange,
  updateInputsFromRange,
  parseInputDate,
  setFromInput,
  setToInput,
  setLastClickedQuickKey,
}: UseDatePickerHandlersProps) => {
  const handleCalendarSelect = useCallback(
    (dateRange: DateRange | undefined) => {
      setSelected(dateRange)
      setHasTimeComponent(false)
      updateInputsFromRange(dateRange, false)
      setLastClickedQuickKey()
    },
    [
      setSelected,
      setHasTimeComponent,
      setLastClickedQuickKey,
      updateInputsFromRange,
    ],
  )

  const handleSideButtonClick = useCallback(
    (dateRange: DateRange, quickKeyLabel: string) => {
      const isTimeBased = hasTimeInRange(dateRange)
      setSelected(dateRange)
      setHasTimeComponent(isTimeBased)
      updateInputsFromRange(dateRange, isTimeBased)
      setLastClickedQuickKey(quickKeyLabel)
    },
    [
      setSelected,
      setHasTimeComponent,
      setLastClickedQuickKey,
      hasTimeInRange,
      updateInputsFromRange,
    ],
  )

  const handleFromInputBlur = useCallback(
    (e: React.FocusEvent<HTMLInputElement>) => {
      setLastClickedQuickKey()
      const value = e.target.value
      if (!value) {
        setSelected({ from: undefined, to: selected?.to })
        return
      }
      const date = parseInputDate(value, true)
      if (!date) {
        if (selected?.from) {
          const format_str = hasTimeComponent
            ? 'yyyy-MM-dd HH:mm'
            : 'yyyy-MM-dd'
          setFromInput(format(selected.from, format_str))
        } else {
          setFromInput('')
        }
        return
      }

      const newRange = { from: date, to: selected?.to }
      setSelected(newRange)
      setHasTimeComponent(hasTimeInRange(newRange))
    },
    [
      selected,
      hasTimeComponent,
      parseInputDate,
      setSelected,
      setFromInput,
      setHasTimeComponent,
      setLastClickedQuickKey,
      hasTimeInRange,
    ],
  )

  const handleToInputBlur = useCallback(
    (e: React.FocusEvent<HTMLInputElement>) => {
      setLastClickedQuickKey()
      const value = e.target.value
      if (!value) {
        setSelected({ to: selected?.to, from: undefined })
        return
      }

      const date = parseInputDate(value, false)
      if (!date) {
        if (selected?.to) {
          const format_str = hasTimeComponent
            ? 'yyyy-MM-dd HH:mm'
            : 'yyyy-MM-dd'
          setToInput(format(selected.to, format_str))
        } else {
          setToInput('')
        }
        return
      }

      const newRange = { from: selected?.from, to: date }
      setSelected(newRange)
      setHasTimeComponent(hasTimeInRange(newRange))
    },
    [
      selected,
      hasTimeComponent,
      parseInputDate,
      setSelected,
      setToInput,
      setHasTimeComponent,
      setLastClickedQuickKey,
      hasTimeInRange,
    ],
  )

  return {
    handleCalendarSelect,
    handleSideButtonClick,
    handleFromInputBlur,
    handleToInputBlur,
  }
}
