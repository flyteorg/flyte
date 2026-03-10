import { StructuredDataset } from '@/gen/flyteidl2/core/literals_pb'
import {
  LiteralType,
  StructuredDatasetType,
} from '@/gen/flyteidl2/core/types_pb'

export const structuredDatasetHelper = {
  literalTypeToJsonSchema: (type: LiteralType) => {
    return {
      type: 'object',
      format: 'structured-dataset',
      properties: {
        uri: { type: 'string' },
        format: {
          type: 'string',
          default: (type.type.value as StructuredDatasetType).format,
        },
      },
    }
  },
  literalToValue: (value: StructuredDataset) => {
    return {
      uri: value?.uri,
      format: value?.metadata?.structuredDatasetType?.format,
    }
  },
}
