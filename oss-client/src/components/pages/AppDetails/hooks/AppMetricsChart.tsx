import { MetricsChartProps } from '@/components/Charts'
import { ExecutionMetric } from '@/gen/usage/payload_pb'
import { Timestamp } from '@bufbuild/protobuf/wkt'
import { AdvancedMetricsPlaceholder } from '../../../Charts/AdvancedMetricsPlaceholder'
import { AppMetricConfig } from './transform'

export type AppMetricsChartProps = MetricsChartProps & {
  configs: AppMetricConfig[]
  metrics: ExecutionMetric[]
  appId: string | undefined
  domain: string
  org: string
  project: string
  startTime?: Timestamp
  endTime?: Timestamp
  noDataMessage?: string
}

export const AppMetricsChart = (_props: Omit<AppMetricsChartProps, 'data'>) => {
  return <AdvancedMetricsPlaceholder />
}
