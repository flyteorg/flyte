import { LogLine } from '@/gen/flyteidl2/logs/dataplane/payload_pb'
import { getLogDateString } from '@/lib/dateUtils'

export type LogLevel = 'INFO' | 'ERROR' | 'WARNING'

export const getLevelColor = (level: LogLevel) => {
  switch (level) {
    case 'ERROR':
      return 'bg-red-500'
    case 'WARNING':
      return 'bg-yellow-500'
    case 'INFO':
      return 'bg-blue-500'
    default:
      return 'bg-zinc-500'
  }
}

export const logsToString = (logs: LogLine[]): string => {
  return logs
    .map((line) => `${getLogDateString(line.timestamp)}\t${line.message}`)
    .join('\n')
}
