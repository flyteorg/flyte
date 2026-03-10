import { XAxisProps, YAxisProps } from 'recharts'
export interface BaseDataPoint {
  time: number // epoch milliseconds
  [key: string]: string | number
}

export interface GradientConfig {
  id: string
  startColor: string
  endColor: string
  startOpacity?: number
  endOpacity?: number
}

export interface LineConfig {
  dataKey: string
  color: string
  strokeWidth?: number
  name?: string
  tooltipText?: string
  unit?: string
  /** Custom formatter for tooltip values. Returns formatted value and optional unit override. */
  formatValue?: (value: number) => { value: string; unit?: string }
}

export interface AreaConfig extends Omit<LineConfig, 'color'> {
  strokeColor: string
  fill: string
}

export interface MetricsChartProps<T extends BaseDataPoint = BaseDataPoint> {
  data: T[]
  lines?: LineConfig[]
  makeLines?: (d: T[]) => LineConfig[]
  areas?: AreaConfig[]
  gradients?: GradientConfig[]
  height?: number
  className?: string
  yAxisConfig?: YAxisProps
  xAxisConfig?: XAxisProps
  tooltipLabelFormatter?: (value: number) => string
}
