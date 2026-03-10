import { Literal, Primitive, Scalar } from '@/gen/flyteidl2/core/literals_pb'
import { getFormatDate } from './dateUtils'

export const primitiveToDisplayValue = (primitive: Primitive): string => {
  switch (primitive.value.case) {
    case 'integer':
    case 'floatValue':
    case 'stringValue':
    case 'boolean':
      return primitive.value.value.toLocaleString()
    case 'datetime':
      return getFormatDate(primitive.value.value)
    case 'duration':
      return 'not implemented'
    default:
      return 'unknown'
  }
}
export const scalarToDisplayValue = (scalar: Scalar) => {
  switch (scalar.value.case) {
    case 'primitive':
      return primitiveToDisplayValue(scalar.value.value)
    case 'blob':
      return 'not implemented'
    case 'binary':
      return 'binary'
    case 'schema':
      return 'not implemented'
  }
}

export const literalToDisplayValue = (literal?: Literal) => {
  if (!literal) {
    return 'undefined'
  }
  switch (literal.value.case) {
    case 'scalar':
      return scalarToDisplayValue(literal.value.value)
    case 'collection':
      // Handle collection types
      return 'not implemented'
    case 'map':
      // Handle map types
      return 'not implemented'
    case 'offloadedMetadata':
      // Handle offloaded metadata types
      return 'not implemented'
    default:
      return 'unknown'
  }
}
