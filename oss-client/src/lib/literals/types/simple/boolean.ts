import { LiteralType } from '@/gen/flyteidl2/core/types_pb'

export const booleanHelper = {
  literalTypeToJsonSchema: (_type: LiteralType) => {
    return {
      type: 'boolean',
    }
  },
  literalToValue: (value: boolean) => {
    return value
  },
}
