export type AutomationType =
  | 'cron-schedule'
  | 'fixed-rate'
  | 'webhook'
  | 'artifact'

export type CreateTriggerState = CronTrigger | FixedRateTrigger

export type CronTrigger = BaseCreateTriggerState & {
  automationType: 'cron-schedule'
  cronExpression: string
  timezone: string
}

export type FixedRateTrigger = BaseCreateTriggerState & {
  automationType: 'fixed-rate'
  interval: number
  startTime: number | string // date
  shouldStartImmediately: boolean
}

type BaseCreateTriggerState = {
  name: string
  description?: string
  automationType: AutomationType
  inputs: Record<string, unknown>
  labels?: KVPair[]
  envVars?: KVPair[]
  annotations?: KVPair[]
  interruptible: boolean | undefined
  overwriteCache?: boolean
  maxRunConcurrency?: number
  serviceAccount?: string
  dataOutputLocation?: string
  activeOnCreation?: boolean
  formData?: FormData
  /** Name of the datetime input that should receive the scheduled kickoff time */
  kickoffTimeInputArg?: string
}

export type KVPair = { key: string; value: string }

export type FormData = Record<string, unknown>
