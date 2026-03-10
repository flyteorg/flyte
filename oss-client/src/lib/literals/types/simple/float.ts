import { LiteralType } from '@/gen/flyteidl2/core/types_pb'

export const floatHelper = {
  literalTypeToJsonSchema: (_type: LiteralType) => {
    return {
      type: 'integer', // RJSF does not have a specific type for float, so we use integer
      format: 'float', // RJSF allows custom formats, but we use float for floats
    }
  },
  literalToValue: (value: number) => {
    return value
  },
}
