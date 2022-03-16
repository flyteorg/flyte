import { Blob, BlobDimensionality } from 'models/Common/types';
import * as React from 'react';
import { PrintValue } from '../PrintValue';
import { useLiteralStyles } from '../styles';

const dimensionalityStrings: Record<BlobDimensionality, string> = {
  [BlobDimensionality.SINGLE]: 'single',
  [BlobDimensionality.MULTIPART]: 'multi-part',
};

/** Renders type and URI information for a Blob */
export const BlobValue: React.FC<{ blob: Blob }> = ({ blob }) => {
  const literalStyles = useLiteralStyles();
  const {
    metadata: { type },
  } = blob;
  const formatString = type.format ? ` (${type.format})` : '';
  const typeString = `${dimensionalityStrings[type.dimensionality]}${formatString}`;
  return (
    <div className={literalStyles.nestedContainer}>
      <PrintValue label="type" value={typeString} />
      <PrintValue label="uri" value={blob.uri} />
    </div>
  );
};
