import { LiteralType } from '@/gen/flyteidl2/core/types_pb'
import { Duration } from '@/gen/google/protobuf/duration_pb'

export const durationHelper = {
  literalTypeToJsonSchema: (_type: LiteralType) => {
    return {
      type: 'integer',
      format: 'duration',
    }
  },
  literalToValue: (value: Duration) => {
    return value?.toString()
  },
}
