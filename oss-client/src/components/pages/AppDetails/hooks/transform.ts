import { ExecutionMetric, GetAppMetricsResponse } from '@/gen/usage/payload_pb'
import { JsonObject } from '@bufbuild/protobuf'

export interface AppMetricConfig {
  metricId: ExecutionMetric
  dataKey: string
  name: string
  unit?: string
  transform?: (value: string) => number
  generateDataKey?: (
    labels?: JsonObject | undefined,
    baseKey?: string,
  ) => string
}

export const TRANSFORMATIONS = {
  parseNumber: (value: string): number => {
    return Number(parseFloat(value).toFixed(2))
  },
  parseInt: (value: string): number => {
    return parseInt(value)
  },
  parsePercent: (value: string): number => {
    return Number((Number(value) * 100).toFixed(2))
  },
  // Convert rate (requests/sec) to a more readable number
  parseRate: (value: string): number => {
    return Number(parseFloat(value).toFixed(4))
  },
  // Convert bytes to GB
  bytesToGB: (value: string): number => {
    return Number((parseFloat(value) / (1000 * 1000 * 1000)).toFixed(2))
  },
  // Convert bytes to MB
  bytesToMB: (value: string): number => {
    return Number((parseFloat(value) / (1000 * 1000)).toFixed(2))
  },
}

type PromMetric = [number, string]

type PromSeries = {
  metric: JsonObject
  values?: PromMetric[]
}

type MetricJson = {
  result: PromSeries[]
  resultType: string
}

/**
 * Maps envoy_response_code_class to human-readable HTTP status code ranges
 */
export const STATUS_CODE_CLASS_MAP: Record<string, string> = {
  '1': '1xx',
  '2': '2xx',
  '3': '3xx',
  '4': '4xx',
  '5': '5xx',
}

export const transformAppMetricsData = ({
  apiResponse,
  configs,
}: {
  apiResponse: GetAppMetricsResponse | undefined
  configs: AppMetricConfig[]
}) => {
  if (!apiResponse) return []

  const timeSeriesMap = new Map<number, { [dataKey: string]: number }>()
  const configMap = new Map<number, AppMetricConfig>()

  configs.forEach((config) => {
    configMap.set(config.metricId, config)
  })

  apiResponse.results.forEach((metricResult) => {
    const config = configMap.get(metricResult.metric)
    if (!config) return

    const data = metricResult.data as MetricJson | undefined
    if (!data?.result || !Array.isArray(data.result)) return

    data.result.forEach((resultItem) => {
      if (!resultItem?.values) return

      // Generate a unique dataKey based on the metric labels
      const metric = resultItem.metric
      const dataKey = config.generateDataKey
        ? config.generateDataKey(metric, config.dataKey)
        : config.dataKey

      resultItem.values.forEach((item) => {
        if (Array.isArray(item) && item.length === 2) {
          const [timestamp, value] = item as PromMetric
          const timeKey = timestamp * 1000 // epoch milliseconds

          if (!timeSeriesMap.has(timeKey)) {
            timeSeriesMap.set(timeKey, {})
          }

          const transformedValue = config.transform
            ? config.transform(value)
            : parseFloat(value)

          timeSeriesMap.get(timeKey)![dataKey] = transformedValue
        }
      })
    })
  })

  const result = Array.from(timeSeriesMap.entries())
    .map(([time, data]) => ({
      time,
      ...data,
    }))
    .sort((a, b) => a.time - b.time)

  return result
}
