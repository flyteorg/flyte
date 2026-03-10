import { LiteralType, SimpleType } from '@/gen/flyteidl2/core/types_pb'
import { noneHelper } from './types/simple/none'
import { integerHelper } from './types/simple/integer'
import { floatHelper } from './types/simple/float'
import { stringHelper } from './types/simple/string'
import { booleanHelper } from './types/simple/boolean'
import { datetimeHelper } from './types/simple/datetime'
import { durationHelper } from './types/simple/duration'
import { binaryHelper } from './types/simple/binary'
import { errorHelper } from './types/simple/error'
import { structHelper } from './types/simple/struct'
import { blobHelper } from './types/blob'
import { unionHelper } from './types/union'
import { collectionHelper } from './types/collection'
import { mapHelper } from './types/map'
import { structuredDatasetHelper } from './types/structuredDataset'
import { VariableMap } from '@/gen/flyteidl2/core/interface_pb'

export type JsonSchemaProperty = {
  title?: string
  description?: string
  type: string
  format?: string
  enum?: (string | number)[]
  default?: unknown
}

export type JsonSchemaObject = {
  title?: string
  description?: string
  type: string
  format?: string
  properties?: Record<string, JsonSchemaProperty | JsonSchemaObject>
  items?: JsonSchemaType
  additionalProperties?: JsonSchemaType
  oneOf?: Array<{
    type: string
    properties: Record<string, JsonSchemaProperty | JsonSchemaObject>
  }>
  default?: unknown
}

export type JsonSchemaType = JsonSchemaObject | JsonSchemaProperty

export type LiteralTypeCase =
  | 'blob'
  | 'simple'
  | 'schema'
  | 'collectionType'
  | 'mapValueType'
  | 'enumType'
  | 'structuredDatasetType'
  | 'unionType'

export const simpleTypeToJsonSchemaType = (type: LiteralType) => {
  const simpleType = type.type.value
  switch (simpleType) {
    case SimpleType.NONE: // 0
      return noneHelper.literalTypeToJsonSchema(type)
    case SimpleType.INTEGER: // 1
      return integerHelper.literalTypeToJsonSchema(type)
    case SimpleType.FLOAT: // 2
      return floatHelper.literalTypeToJsonSchema(type)
    case SimpleType.STRING: {
      // 3
      return stringHelper.literalTypeToJsonSchema(type)
    }
    case SimpleType.BOOLEAN: // 4
      return booleanHelper.literalTypeToJsonSchema(type)
    case SimpleType.DATETIME: // 5
      return datetimeHelper.literalTypeToJsonSchema(type)
    case SimpleType.DURATION: // 6
      return durationHelper.literalTypeToJsonSchema(type) // RJSF allows custom formats
    case SimpleType.BINARY: // 7
      return binaryHelper.literalTypeToJsonSchema(type)
    case SimpleType.ERROR: // 8
      // This is a special case, we can return a string with the error message
      return errorHelper.literalTypeToJsonSchema(type)
    case SimpleType.STRUCT: // 9
      // TODO: Handle structured types
      return structHelper.literalTypeToJsonSchema(type)
  }
}

export const literalTypeToJsonSchema = (type: LiteralType): JsonSchemaType => {
  switch (type.type.case) {
    case 'simple':
      return simpleTypeToJsonSchemaType(type) as JsonSchemaType
    case 'blob':
      return blobHelper.literalTypeToJsonSchema(type)
    case 'collectionType':
      return collectionHelper.literalTypeToJsonSchema(type)
    case 'mapValueType':
      return mapHelper.literalTypeToJsonSchema(type)
    case 'structuredDatasetType':
      return structuredDatasetHelper.literalTypeToJsonSchema(type)
    case 'unionType':
      return unionHelper.literalTypeToJsonSchema(type)
    default:
      return { type: 'object' }
  }
}

export const variableMapToJsonSchemaTypes = (variableMap: VariableMap) => {
  return variableMap.variables?.reduce((acc, entry) => {
    const variable = entry.value

    if (!variable) {
      return acc
    }

    const jsonType = literalTypeToJsonSchema(variable.type!)
    const description = variable.description

    return {
      ...acc,
      [entry.key]: { ...(description ? { description } : {}), ...jsonType },
    }
  }, {} as JsonSchemaObject)
}
