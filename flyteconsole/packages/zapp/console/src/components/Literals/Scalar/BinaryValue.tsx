import { Binary } from 'models/Common/types';
import * as React from 'react';
import { PrintValue } from '../PrintValue';
import { useLiteralStyles } from '../styles';

/** Renders a Binary literal as a `tag` with an informational message, since
 * we can't really show a byte array in a meaningful way
 */
export const BinaryValue: React.FC<{ binary: Binary }> = ({ binary }) => {
  const literalStyles = useLiteralStyles();
  return (
    <div className={literalStyles.nestedContainer}>
      <PrintValue label="tag" value={binary.tag} />
      <em>(binary data not shown)</em>
    </div>
  );
};
