import {
  addMinutes,
  startOfToday,
  endOfToday,
  startOfYesterday,
  subMinutes,
  subHours,
  subDays,
  subMonths,
} from 'date-fns'
import { DateRange } from 'react-day-picker'

export type QuickRange = {
  filterLabel: string // display only: to create filter label
  label: string // display in quickRange menu but also used as unique key for linking params to date-ranges
  getRange: () => DateRange
}

export const labeledQuickRanges: QuickRange[] = [
  {
    filterLabel: 'today',
    label: 'Today',
    getRange: () => ({ from: startOfToday(), to: endOfToday() }),
  },
  {
    filterLabel: 'yesterday',
    label: 'Yesterday',
    getRange: () => {
      const y = startOfYesterday()
      return { from: y, to: new Date(y.getTime() + 24 * 60 * 60 * 1000 - 1) }
    },
  },
  {
    filterLabel: 'in the last 5 min',
    label: 'Last 5 min',
    getRange: () => ({
      from: subMinutes(new Date(), 5),
      to: new Date(),
    }),
  },
  {
    filterLabel: 'in the last 30 min',
    label: 'Last 30 min',
    getRange: () => ({
      from: subMinutes(new Date(), 30),
      to: new Date(),
    }),
  },
  {
    filterLabel: 'in the last hour',
    label: 'Last 1 hour',
    getRange: () => ({
      from: subHours(new Date(), 1),
      to: new Date(),
    }),
  },
  {
    filterLabel: 'in the last 7 days',
    label: 'Last 7 days',
    getRange: () => ({
      from: subDays(new Date(), 7),
      to: new Date(),
    }),
  },
  {
    filterLabel: 'in the last 30 days',
    label: 'Last 30 days',
    getRange: () => ({
      from: subDays(new Date(), 30),
      to: new Date(),
    }),
  },
  {
    filterLabel: 'in the last 90 days',
    label: 'Last 90 days',
    getRange: () => ({
      from: subDays(new Date(), 90),
      to: new Date(),
    }),
  },
  {
    filterLabel: 'in the last 12 months',
    label: 'Last 12 months',
    getRange: () => ({
      from: subMonths(new Date(), 12),
      to: new Date(),
    }),
  },
  {
    filterLabel: 'last week',
    label: 'Last week',
    getRange: () => {
      const today = new Date()
      const start = subDays(today, today.getDay() + 6)
      const end = subDays(today, today.getDay())
      return { from: start, to: end }
    },
  },
  {
    filterLabel: 'last month',
    label: 'Last month',
    getRange: () => {
      const now = new Date()
      const start = new Date(now.getFullYear(), now.getMonth() - 1, 1)
      const end = new Date(now.getFullYear(), now.getMonth(), 0)
      return { from: start, to: end }
    },
  },
]

/**
 * Quick ranges constrained to a maximum of 30 days.
 * Useful for APIs that have a 30-day lookback limit.
 */
export const quickRanges30Days: QuickRange[] = [
  {
    filterLabel: 'today',
    label: 'Today',
    getRange: () => ({ from: startOfToday(), to: endOfToday() }),
  },
  {
    filterLabel: 'yesterday',
    label: 'Yesterday',
    getRange: () => {
      const y = startOfYesterday()
      return { from: y, to: new Date(y.getTime() + 24 * 60 * 60 * 1000 - 1) }
    },
  },
  {
    filterLabel: 'in the last 5 min',
    label: 'Last 5 min',
    getRange: () => ({
      from: subMinutes(new Date(), 5),
      to: new Date(),
    }),
  },
  {
    filterLabel: 'in the last 30 min',
    label: 'Last 30 min',
    getRange: () => ({
      from: subMinutes(new Date(), 30),
      to: new Date(),
    }),
  },
  {
    filterLabel: 'in the last hour',
    label: 'Last 1 hour',
    getRange: () => ({
      from: subHours(new Date(), 1),
      to: new Date(),
    }),
  },
  {
    filterLabel: 'in the last 7 days',
    label: 'Last 7 days',
    getRange: () => ({
      from: subDays(new Date(), 7),
      to: new Date(),
    }),
  },
  // by querying for a time 15 minutes more recent than exactly 30 days ago, we prevent the api from throwing
  // an error since it only retains exactly 30 days worth of data
  {
    filterLabel: 'in the last 30 days',
    label: 'Last 30 days',
    getRange: () => ({
      from: addMinutes(subDays(new Date(), 30), 15),
      to: new Date(),
    }),
  },
]

export const quickRanges7Days: QuickRange[] = [
  {
    filterLabel: 'today',
    label: 'Today',
    getRange: () => ({ from: startOfToday(), to: endOfToday() }),
  },
  {
    filterLabel: 'yesterday',
    label: 'Yesterday',
    getRange: () => {
      const y = startOfYesterday()
      return { from: y, to: new Date(y.getTime() + 24 * 60 * 60 * 1000 - 1) }
    },
  },
  {
    filterLabel: 'in the last 5 min',
    label: 'Last 5 min',
    getRange: () => ({
      from: subMinutes(new Date(), 5),
      to: new Date(),
    }),
  },
  {
    filterLabel: 'in the last 30 min',
    label: 'Last 30 min',
    getRange: () => ({
      from: subMinutes(new Date(), 30),
      to: new Date(),
    }),
  },
  {
    filterLabel: 'in the last hour',
    label: 'Last 1 hour',
    getRange: () => ({
      from: subHours(new Date(), 1),
      to: new Date(),
    }),
  },
  {
    filterLabel: 'in the last 3 days',
    label: 'Last 3 days',
    getRange: () => ({
      from: subDays(new Date(), 3),
      to: new Date(),
    }),
  },
  {
    filterLabel: 'in the last 7 days',
    label: 'Last 7 days',
    getRange: () => ({
      from: subDays(new Date(), 7),
      to: new Date(),
    }),
  },
]
