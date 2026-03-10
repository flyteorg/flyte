import { KeyValuePair } from '@/gen/flyteidl2/core/literals_pb'

// Run name limits must match API (RunIdentifier.name in idl/common/identifier.proto).
// Character set is restricted for URL and identifier safety (no spaces or special chars).
export const RUN_NAME_MIN_LENGTH = 1
export const RUN_NAME_MAX_LENGTH = 30

/** Only letters, numbers, hyphens, and underscores (URL-safe, no spaces). */
export const RUN_NAME_PATTERN = /^[a-zA-Z0-9_-]+$/

export const RUN_NAME_PATTERN_DESCRIPTION =
  'Run name can only contain letters, numbers, hyphens, and underscores'

export const isRunNameValid = (name: string): boolean => {
  const trimmed = name.trim()
  if (trimmed.length === 0) return false
  if (
    trimmed.length < RUN_NAME_MIN_LENGTH ||
    trimmed.length > RUN_NAME_MAX_LENGTH
  )
    return false
  return RUN_NAME_PATTERN.test(trimmed)
}

export type KVPair = { key: string; value: string }

export type LaunchFormState = {
  envs?: KVPair[]
  labels?: KVPair[]
  interruptible?: boolean
  overwriteCache: boolean
  runName: string
  // serviceAccount: string
  inputs?: Record<string, unknown>
  formData?: Record<string, unknown>
  context?: KeyValuePair[]
}

export type LaunchFormTab =
  | 'inputs'
  | 'context'
  | 'settings'
  | 'env-vars'
  | 'labels'
  | 'debug'

export type ErrorWithRawMessage = {
  rawMessage?: string
  name?: string
}
