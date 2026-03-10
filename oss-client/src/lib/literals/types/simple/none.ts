import { LiteralType } from '@/gen/flyteidl2/core/types_pb'
import { Message } from '@bufbuild/protobuf'

export const noneHelper = {
  literalTypeToJsonSchema: (_type: LiteralType) => {
    return {
      type: 'null',
    }
  },
  literalToValue: (_value: Message<'flyteidl2.core.Void'>) => {
    return null
  },
}
