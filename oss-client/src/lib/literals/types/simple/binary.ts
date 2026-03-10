import { Binary } from '@/gen/flyteidl2/core/literals_pb'
import { LiteralType } from '@/gen/flyteidl2/core/types_pb'
import { unpack } from 'msgpackr'

export const binaryHelper = {
  literalTypeToJsonSchema: (_type: LiteralType) => {
    return {
      type: 'string',
      format: 'byte',
    }
  },
  literalToValue: (value: Binary) => {
    if (value.tag === 'msgpack') {
      return value ? unpack(value.value) : undefined
    }
    return 'binary (not shown)'
  },
}
