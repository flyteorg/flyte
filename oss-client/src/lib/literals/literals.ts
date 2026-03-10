import {
  Literal,
  LiteralCollection,
  Primitive,
  Scalar,
} from '@/gen/flyteidl2/core/literals_pb'
import { NamedLiteral } from '@/gen/flyteidl2/task/common_pb'
import { noneHelper } from './types/simple/none'
import { integerHelper } from './types/simple/integer'
import { floatHelper } from './types/simple/float'
import { stringHelper } from './types/simple/string'
import { booleanHelper } from './types/simple/boolean'
import { datetimeHelper } from './types/simple/datetime'
import { durationHelper } from './types/simple/duration'
import { blobHelper } from './types/blob'
import { binaryHelper } from './types/simple/binary'
import { errorHelper } from './types/simple/error'
import { structHelper } from './types/simple/struct'
import { structuredDatasetHelper } from './types/structuredDataset'
import { unionHelper } from './types/union'
import { collectionHelper } from './types/collection'
import { mapHelper } from './types/map'
import { offloadedHelper } from './types/offloaded'

export const isNamedLiteral = (
  literal: Literal | NamedLiteral,
): literal is NamedLiteral => {
  return (literal as NamedLiteral).name !== undefined
}

export const primitiveToValue = (primitive: Primitive) => {
  switch (primitive.value.case) {
    case 'integer':
      integerHelper.literalToValue(primitive.value.value)
    case 'floatValue':
      return floatHelper.literalToValue(primitive.value.value as number)
    case 'stringValue':
      return stringHelper.literalToValue(primitive.value.value)
    case 'boolean':
      return booleanHelper.literalToValue(primitive.value.value)
    case 'datetime':
      return datetimeHelper.literalToValue(primitive.value.value)
    case 'duration':
      return durationHelper.literalToValue(primitive.value.value)
    default:
      throw new Error(`Unsupported primitive type: ${primitive.value.case}`)
  }
}

export const scalarToValue = (scalar: Scalar) => {
  switch (scalar.value.case) {
    case 'primitive':
      return primitiveToValue(scalar.value.value)
    case 'blob':
      return blobHelper.literalToValue(scalar.value.value)
    case 'binary':
      return binaryHelper.literalToValue(scalar.value.value)
    case 'error':
      return errorHelper.literalToValue(scalar.value.value)
    case 'noneType':
      return noneHelper.literalToValue(scalar.value.value)
    case 'generic':
      return structHelper.literalToValue(scalar.value.value)
    case 'structuredDataset':
      return structuredDatasetHelper.literalToValue(scalar.value.value)
    case 'union':
      return unionHelper.literalToValue(scalar.value.value)
  }
}
export const literalToValue = (literal: Literal): unknown => {
  switch (literal?.value.case) {
    case 'scalar':
      return scalarToValue(literal?.value.value)
    case 'collection':
      return collectionHelper.literalToValue(
        literal?.value.value as LiteralCollection,
      )
    case 'map':
      return mapHelper.literalToValue(literal?.value.value)
    case 'offloadedMetadata':
      return offloadedHelper.literalToValue(literal?.value.value)
    default: {
      return undefined
    }
  }
}
