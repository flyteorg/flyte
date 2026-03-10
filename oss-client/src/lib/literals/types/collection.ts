import { LiteralType } from '@/gen/flyteidl2/core/types_pb'
import { literalTypeToJsonSchema } from '../types'
import { LiteralCollection } from '@/gen/flyteidl2/core/literals_pb'
import { literalToValue } from '../literals'

export const collectionHelper = {
  literalTypeToJsonSchema: (type: LiteralType) => {
    return {
      type: 'array',
      items: literalTypeToJsonSchema(type.type.value as LiteralType),
    }
  },
  literalToValue: (value: LiteralCollection) => {
    return value.literals.map((literal) => {
      return literalToValue(literal)
    })
  },
}
