import { useMemo, useEffect } from 'react'
import { format, subDays, startOfDay } from 'date-fns'
import { DateRange, DayPicker } from 'react-day-picker'
import { Button } from '../Button'
import { labeledQuickRanges, type QuickRange } from './quickRanges'
import { type SideButtonItem } from './SideButtons'
import { useDateRangeParams } from './useDateRangeParams'
import { useDateRangeState } from './useDateRangeState'
import { useDatePickerHandlers } from './useDateHandlers'
import { DateRangeInputs } from './DateRangeInputs'
import { QuickRangeSidebar } from './SideButtons'
import 'react-day-picker/style.css'
import './styles.css'

export function BaseDatePicker({
  onApply,
  onCancel,
  onClear,
  quickRanges = labeledQuickRanges,
  maxDaysBack,
}: {
  onApply: (dateRange: DateRange | undefined) => void
  onCancel: () => void
  onClear: () => void
  quickRanges?: QuickRange[]
  /** Optional maximum number of days back that can be selected */
  maxDaysBack?: number
}) {
  const { dateRange: urlDateRange, setDateRange } = useDateRangeParams()

  const {
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
  } = useDateRangeState(urlDateRange)

  const {
    handleCalendarSelect,
    handleSideButtonClick,
    handleFromInputBlur,
    handleToInputBlur,
  } = useDatePickerHandlers({
    selected,
    setSelected,
    hasTimeComponent,
    setHasTimeComponent,
    hasTimeInRange,
    updateInputsFromRange,
    parseInputDate,
    fromInput,
    setFromInput,
    toInput,
    setToInput,
    setLastClickedQuickKey,
  })

  const sideButtons: SideButtonItem[] = useMemo(() => {
    const btnConfigs = quickRanges.map(({ label, getRange }) => ({
      displayText: label,
      key: label,
      onClick: getRange,
    }))
    // Only apply divider layout for default quick ranges
    // Custom quick ranges are displayed as-is without dividers
    if (quickRanges === labeledQuickRanges) {
      return [
        ...btnConfigs.slice(0, 2),
        { key: 'divider-1', type: 'divider' },
        ...btnConfigs.slice(2, -2),
        { key: 'divider-2', type: 'divider' },
        ...btnConfigs.slice(-2),
      ]
    }
    return btnConfigs
  }, [quickRanges])

  // Initialize from URL
  useEffect(() => {
    if (urlDateRange) {
      const isTimeBased = hasTimeInRange(urlDateRange)
      setSelected(urlDateRange)
      setHasTimeComponent(isTimeBased)
      updateInputsFromRange(urlDateRange, isTimeBased)
    } else {
      setSelected(undefined)
      setHasTimeComponent(false)
      setFromInput('')
      setToInput('')
    }
  }, [
    urlDateRange,
    hasTimeInRange,
    setSelected,
    setHasTimeComponent,
    updateInputsFromRange,
    setFromInput,
    setToInput,
  ])

  const handleApply = () => {
    setDateRange(selected, lastClickedQuickKey)
    onApply(selected)
  }

  const handleCancel = () => {
    const urlHasTime = hasTimeInRange(urlDateRange)
    setSelected(urlDateRange)
    setHasTimeComponent(urlHasTime)
    updateInputsFromRange(urlDateRange, urlHasTime)
    onCancel()
  }

  const handleClear = () => {
    setSelected(undefined)
    setFromInput('')
    setToInput('')
    setHasTimeComponent(false)
    setLastClickedQuickKey('')
    onClear()
  }

  // Validation for maxDaysBack
  const validationError = useMemo(() => {
    if (!maxDaysBack || !selected?.from) return null

    const minAllowedDate = startOfDay(subDays(new Date(), maxDaysBack))
    if (selected.from < minAllowedDate) {
      return `Date filtering is limited to the last ${maxDaysBack} days`
    }
    return null
  }, [maxDaysBack, selected?.from])

  // Hide Clear button when user is prohibited from truly clearing the filter
  const shouldShowClearButton = !maxDaysBack

  return (
    <div className="flex flex-col">
      <div className="flex w-full">
        <div className="flex">
          <QuickRangeSidebar
            lastClickedQuickKey={lastClickedQuickKey}
            sideButtons={sideButtons}
            onSideButtonClick={handleSideButtonClick}
          />
          <div className="flex flex-col px-4 py-4">
            <DateRangeInputs
              fromInput={fromInput}
              toInput={toInput}
              hasTimeComponent={hasTimeComponent}
              onFromInputChange={setFromInput}
              onToInputChange={setToInput}
              onFromInputBlur={handleFromInputBlur}
              onToInputBlur={handleToInputBlur}
            />

            <DayPicker
              animate
              classNames={{
                chevron: 'fill-zinc-500',
                caption_label: 'text-sm font-normal text-(--system-gray-6)',
              }}
              formatters={{
                formatCaption: (month) => format(month, 'MMMM'),
                formatWeekdayName: (weekday) => format(weekday, 'EEEEE'),
              }}
              mode="range"
              numberOfMonths={2}
              navLayout="around"
              pagedNavigation
              selected={selected}
              onSelect={handleCalendarSelect}
            />
          </div>
        </div>
      </div>
      <div className="flex min-h-24 items-center justify-between gap-2 border-t-1 border-(--system-gray-3) p-1">
        {validationError ? (
          <span className="pl-2 text-xs text-(--accent-red)">
            {validationError}
          </span>
        ) : (
          <span />
        )}
        <div className="flex items-center gap-2">
          <Button onClick={handleCancel} outline size="xs" color="zinc">
            Cancel
          </Button>
          {shouldShowClearButton && (
            <Button onClick={handleClear} outline size="xs" color="zinc">
              Clear
            </Button>
          )}
          <Button
            outline
            onClick={handleApply}
            size="xs"
            disabled={!!validationError}
          >
            Apply
          </Button>
        </div>
      </div>
    </div>
  )
}
