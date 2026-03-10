import { AdvancedMetricsPlaceholder } from '@/components/Charts'
import { TabSection } from '@/components/TabSection'

export const RunMetricsTab = () => {
  return (
    <div className="flex h-full flex-col gap-5 p-8 pt-0">
      <div className="flex gap-5">
        <TabSection heading="Memory Quota">
          <AdvancedMetricsPlaceholder />
        </TabSection>
        <TabSection heading="CPU Cores Quota">
          <AdvancedMetricsPlaceholder />
        </TabSection>
      </div>
      <div className="flex gap-5">
        <TabSection heading="GPU Memory Allocated">
          <AdvancedMetricsPlaceholder />
        </TabSection>
        <TabSection heading="GPU Utilization">
          <AdvancedMetricsPlaceholder />
        </TabSection>
      </div>
      <div className="flex gap-5 pb-8">
        <TabSection heading="SM Active Cycles">
          <AdvancedMetricsPlaceholder />
        </TabSection>
        <TabSection heading="SM Occupancy">
          <AdvancedMetricsPlaceholder />
        </TabSection>
      </div>
    </div>
  )
}
