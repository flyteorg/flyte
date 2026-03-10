import { useState } from 'react'
import { MenuItem, PopoverMenu } from '../Popovers'
import { BaseDatePicker } from './BaseDatePicker'
import { useDateRangeParams } from './useDateRangeParams'
import { detectQuickRangeLabel } from './detectLabel'
import { labeledQuickRanges, type QuickRange } from './quickRanges'

export const DatePickerPopover = ({
  labelPrefix = 'Items',
  maxDaysBack,
  quickRanges = labeledQuickRanges,
}: {
  labelPrefix?: string
  maxDaysBack?: number
  quickRanges?: QuickRange[]
  /** Optional maximum number of days back that can be selected */
}) => {
  const [isOpen, setIsOpen] = useState(false)
  const { clearDateRange, dateRange } = useDateRangeParams()
  const handleApplyDates = () => {
    // Note: setDateRange is already called by BaseDatePicker with the quick range label
    setIsOpen(false)
  }
  const handleCancel = () => {
    setIsOpen(false)
  }

  const handleClear = () => {
    clearDateRange()
  }

  const readableRange = detectQuickRangeLabel(
    dateRange,
    quickRanges,
    'All time',
  )

  const label = (
    <div className="font-semibold tracking-[0.25%]">
      {labelPrefix} {readableRange}
    </div>
  )

  const items: MenuItem[] = [
    {
      id: 'date-picker',
      type: 'custom',
      component: (
        <BaseDatePicker
          onApply={handleApplyDates}
          onCancel={handleCancel}
          onClear={handleClear}
          quickRanges={quickRanges}
          maxDaysBack={maxDaysBack}
        />
      ),
    },
  ]
  return (
    <PopoverMenu
      itemCustomClassName="!p-0"
      items={items}
      label={label}
      open={isOpen}
      onOpenChange={setIsOpen}
      size="xs"
      triggerClassName="!px-0 [&>svg]:dark:text-(--system-gray-4)"
      variant="dropdown"
    />
  )
}
