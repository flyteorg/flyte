import { LiteralType } from '@/gen/flyteidl2/core/types_pb'
import { Timestamp } from '@/gen/google/protobuf/timestamp_pb'

export const datetimeHelper = {
  literalTypeToJsonSchema: (_type: LiteralType) => {
    return {
      type: 'string',
      format: 'date-time',
    }
  },
  literalToValue: (value: Timestamp) => {
    return value?.toString()
  },
}
