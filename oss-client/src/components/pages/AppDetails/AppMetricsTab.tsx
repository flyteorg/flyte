import { AdvancedMetricsPlaceholder } from '@/components/Charts/AdvancedMetricsPlaceholder'
import {
  DatePickerPopover,
  quickRanges7Days,
  useDateRangeParams,
} from '@/components/DatePicker'
import { TabSection } from '@/components/TabSection'
import { subDays } from 'date-fns'
import { useEffect } from 'react'

export const AppMetricsTab = ({}: { appId: string | undefined }) => {
  const { dateRange, setDateRange } = useDateRangeParams()

  // Sync the DatePicker to "Last 7 days" when no existing date params provided
  useEffect(() => {
    if (!dateRange) {
      const now = new Date()
      setDateRange({ from: subDays(now, 7), to: now }, 'Last 7 days')
    }
  }, [dateRange, setDateRange])

  return (
    <div>
      <div className="sticky top-0 z-20 bg-(--system-gray-2)">
        <DatePickerPopover
          labelPrefix="Requests"
          maxDaysBack={7}
          quickRanges={quickRanges7Days}
        />
      </div>
      <div className="flex flex-col gap-5 [&>*:last-child]:mb-5">
        <div className="flex gap-5">
          <TabSection heading="Replica Count">
            <AdvancedMetricsPlaceholder />
          </TabSection>
        </div>
        <div className="flex gap-5">
          <TabSection heading="Requests">
            <AdvancedMetricsPlaceholder />
          </TabSection>
          <TabSection heading="Responses">
            <AdvancedMetricsPlaceholder />
          </TabSection>
        </div>
        <div className="flex gap-5">
          <TabSection heading="Allocated Memory">
            <AdvancedMetricsPlaceholder />
          </TabSection>
          <TabSection heading="Memory Utilization">
            <AdvancedMetricsPlaceholder />
          </TabSection>
        </div>
        <div className="flex gap-5">
          <TabSection heading="Allocated CPU">
            <AdvancedMetricsPlaceholder />
          </TabSection>
          <TabSection heading="CPU Utilization">
            <AdvancedMetricsPlaceholder />
          </TabSection>
        </div>
        <div className="flex gap-5">
          <TabSection heading="GPU Utilization">
            <AdvancedMetricsPlaceholder />
          </TabSection>
          <TabSection heading="GPU Memory Allocated">
            <AdvancedMetricsPlaceholder />
          </TabSection>
        </div>
        <div className="flex gap-5">
          <TabSection heading="GPU SM Active Cycles">
            <AdvancedMetricsPlaceholder />
          </TabSection>
          <TabSection heading="GPU SM Occupancy">
            <AdvancedMetricsPlaceholder />
          </TabSection>
        </div>
      </div>
    </div>
  )
}
