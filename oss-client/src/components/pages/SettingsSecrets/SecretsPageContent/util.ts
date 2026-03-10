import { Secret } from '@/gen/flyteidl2/secret/definition_pb'
import { toDateFormat } from '@/lib/dateUtils'
import { getSecretScope } from '@/lib/secretUtils'
import { type SecretsTableRow } from './types'

export const formatForTable = (s: Secret): SecretsTableRow => {
  const relativeCreatedAt = toDateFormat({
    timestamp: s.secretMetadata?.createdTime,
    relativeThreshold: 'week',
  })

  return {
    name: s.id?.name,
    scope: getSecretScope(s?.id?.domain, s?.id?.project),
    createdAt: relativeCreatedAt,
    original: s,
    actions: s,
  }
}
