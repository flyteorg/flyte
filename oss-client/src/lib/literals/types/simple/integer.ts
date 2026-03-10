import { LiteralType } from '@/gen/flyteidl2/core/types_pb'

export const integerHelper = {
  literalTypeToJsonSchema: (_type: LiteralType) => {
    return {
      type: 'integer',
    }
  },
  literalToValue: (value: bigint) => {
    return value?.toString()
  },
}
