import { LiteralType } from '@/gen/flyteidl2/core/types_pb'

export const stringHelper = {
  literalTypeToJsonSchema: (_type: LiteralType) => {
    return {
      type: 'string',
    }
  },
  literalToValue: (value: string) => {
    return value
  },
}
