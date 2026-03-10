import { LiteralType } from '@/gen/flyteidl2/core/types_pb'
import { literalTypeToJsonSchema } from '../types'
import { LiteralMap } from '@/gen/flyteidl2/core/literals_pb'
import { keys } from 'lodash'
import { literalToValue } from '../literals'

export const mapHelper = {
  literalTypeToJsonSchema: (type: LiteralType) => {
    return {
      type: 'object',
      additionalProperties: literalTypeToJsonSchema(
        type.type.value as LiteralType,
      ),
    }
  },
  literalToValue: (value: LiteralMap) => {
    return keys(value.literals).reduce((acc, key) => {
      return {
        ...acc,
        [key]: literalToValue(value.literals[key]),
      }
    }, {})
  },
}
