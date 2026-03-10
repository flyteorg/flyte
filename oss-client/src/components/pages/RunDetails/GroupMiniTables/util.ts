import { getRelativeDate, getRunningTime } from '@/lib/dateUtils'
import { ActionWithChildren } from '../state/types'
import { GroupTableItem } from './types'

const getTaskName = (node?: ActionWithChildren) =>
  node?.action?.metadata?.spec.case === 'task'
    ? node.action.metadata.spec.value.shortName
    : ''

export const transformForTable = (node: ActionWithChildren): GroupTableItem => {
  const duration = getRunningTime({
    endTimestamp: node.action?.status?.endTime,
    timestamp: node.action?.status?.startTime,
    showSubSecondPrecision: true,
  })
  console.log('node', node?.action?.metadata?.spec.value)
  return {
    name: {
      actionId: node.action?.id?.name || '',
      taskName: getTaskName(node),
      phase: node.action?.status?.phase,
    },
    duration,
    startTime: node.action?.status?.startTime
      ? getRelativeDate(node.action?.status?.startTime, 'always')
      : '-',
    original: node,
  }
}
