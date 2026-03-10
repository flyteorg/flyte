import { LiteralType, UnionType } from '@/gen/flyteidl2/core/types_pb'
import { LiteralTypeCase, literalTypeToJsonSchema } from '../types'
import { Union } from '@/gen/flyteidl2/core/literals_pb'
import { literalToValue } from '../literals'

export const unionHelper = {
  literalTypeToJsonSchema: (type: LiteralType) => {
    return {
      type: 'object',
      format: 'union',
      properties: {
        value: {
          type: 'object',
          oneOf: (type.type.value as UnionType).variants.map((variant) => ({
            type: 'object',
            properties: {
              type: {
                type: 'string',
                enum: [variant.type.case].filter(
                  (v): v is LiteralTypeCase => v !== undefined,
                ),
              },
              value: literalTypeToJsonSchema(variant),
            },
          })),
        },
      },
    }
  },
  literalToValue: (value: Union) => {
    return literalToValue(value.value!)
  },
}
