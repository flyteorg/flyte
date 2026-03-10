import { Variable } from '@/gen/flyteidl2/core/interface_pb'
import { LiteralType, SimpleType } from '@/gen/flyteidl2/core/types_pb'

export const simpleTypeToString = (type: SimpleType) => {
  switch (type) {
    case SimpleType.NONE:
      return 'none'
    case SimpleType.INTEGER:
      return 'integer'
    case SimpleType.FLOAT:
      return 'float'
    case SimpleType.STRING:
      return 'string'
    case SimpleType.BOOLEAN:
      return 'boolean'
    case SimpleType.DATETIME:
      return 'datetime'
    case SimpleType.DURATION:
      return 'duration'
    case SimpleType.BINARY:
      return 'binary'
    case SimpleType.ERROR:
      return 'error'
    case SimpleType.STRUCT:
      return 'struct'
    default:
      return 'unknown'
  }
}

export const collectionTypeToString = (/*_type: LiteralType*/): string => {
  return `not implemented[]`
}

export const literalTypeToString = (literalType?: LiteralType): string => {
  if (!literalType) return 'unknown'

  const type = literalType.type
  switch (type?.case) {
    case 'simple':
      return simpleTypeToString(type.value)
    case 'collectionType':
      return collectionTypeToString(/*type.value*/)
    case 'mapValueType':
      return `map<string, ${literalTypeToString(type.value)}>`
    case 'schema':
      return 'schema'
    case 'blob':
      return 'blob'
    case 'enumType':
      return 'enum'
    case 'structuredDatasetType':
      return 'sd'
    case 'unionType':
      if (!type.value.variants) return 'union'
      const concatListOfSubTypes = type.value.variants
        .map((variant) => literalTypeToString(variant))
        .join(' | ')
      return `union [${concatListOfSubTypes}]`
    default:
      return 'unknown'
  }
}

export const variableToString = (variable: Variable) => {
  if (!variable) return 'undefined'
  return literalTypeToString(variable.type)
}
