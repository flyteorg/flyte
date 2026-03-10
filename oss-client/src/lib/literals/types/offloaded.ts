import { LiteralType } from '@/gen/flyteidl2/core/types_pb'
import { LiteralOffloadedMetadata } from '@/gen/flyteidl2/core/literals_pb'

export const offloadedHelper = {
  literalTypeToJsonSchema: (_type: LiteralType) => {
    // TODO: offloaded types need special handling
    return {
      type: 'string',
      format: 'offloaded',
      description:
        'This literal is offloaded to a remote storage system. The value is not available',
    }
  },
  literalToValue: (value: LiteralOffloadedMetadata) => {
    return value.uri
  },
}
