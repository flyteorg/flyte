import { LiteralType } from '@/gen/flyteidl2/core/types_pb'
import { JsonSchemaType } from '../../types'
import { JsonObject } from '@bufbuild/protobuf'

export const structHelper = {
  literalTypeToJsonSchema: (type: LiteralType) => {
    return type.metadata as JsonSchemaType
  },
  literalToValue: (value: JsonObject) => {
    return value
  },
}
