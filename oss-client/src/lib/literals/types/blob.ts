import {
  BlobType,
  BlobType_BlobDimensionality,
  LiteralType,
} from '@/gen/flyteidl2/core/types_pb'
import { Blob } from '@/gen/flyteidl2/core/literals_pb'
export const blobHelper = {
  literalTypeToJsonSchema: (type: LiteralType) => {
    //  TODO: if there is a data-url do file upload
    /** https://rjsf-team.github.io/react-jsonschema-form/docs/usage/widgets/#accept-option
       * {
            type: 'string',
            format: 'data-url',
          }
       */
    return {
      type: 'object',
      format: 'blob',
      properties: {
        uri: {
          type: 'string',
        },
        dimensionality: {
          type: 'integer',
          enum: Object.values(BlobType_BlobDimensionality).filter(
            (v) => typeof v === 'number',
          ) as number[],
          default: (type?.type?.value as BlobType)?.dimensionality,
        },
        format: {
          type: 'string',
          default: (type?.type?.value as BlobType)?.dimensionality,
        },
      },
    }
  },
  literalToValue: (value: Blob) => {
    return {
      uri: value?.uri,
      dimensionality: value?.metadata?.type?.dimensionality,
      format: value?.metadata?.type?.format,
    }
  },
}
