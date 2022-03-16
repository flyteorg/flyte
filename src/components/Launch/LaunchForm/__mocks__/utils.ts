import { Core } from 'flyteidl';
import { BlobDimensionality } from 'models/Common/types';
import { InputType, InputTypeDefinition } from '../types';

export function primitiveLiteral(primitive: Core.IPrimitive): Core.ILiteral {
  return { scalar: { primitive } };
}

export function blobLiteral({
  uri,
  format,
  dimensionality,
}: {
  uri?: string;
  format?: string;
  dimensionality?: BlobDimensionality;
}): Core.ILiteral {
  return {
    scalar: {
      blob: { uri, metadata: { type: { format, dimensionality } } },
    },
  };
}

export function collectionInputTypeDefinition(
  typeDefinition: InputTypeDefinition,
): InputTypeDefinition {
  return {
    literalType: {
      collectionType: typeDefinition.literalType,
    },
    type: InputType.Collection,
    subtype: typeDefinition,
  };
}

export function nestedCollectionInputTypeDefinition(
  typeDefinition: InputTypeDefinition,
): InputTypeDefinition {
  return {
    literalType: {
      collectionType: {
        collectionType: typeDefinition.literalType,
      },
    },
    type: InputType.Collection,
    subtype: {
      literalType: { collectionType: typeDefinition.literalType },
      type: InputType.Collection,
      subtype: typeDefinition,
    },
  };
}
