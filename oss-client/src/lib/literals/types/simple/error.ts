import { LiteralType, Error } from '@/gen/flyteidl2/core/types_pb'

export const errorHelper = {
  literalTypeToJsonSchema: (_type: LiteralType) => {
    return {
      type: 'string',
      format: 'error',
    }
  },
  literalToValue: (value: Error) => {
    return value?.message
  },
}
