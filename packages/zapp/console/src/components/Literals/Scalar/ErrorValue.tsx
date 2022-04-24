import { Error } from 'models/Common/types';
import * as React from 'react';
import { PrintValue } from '../PrintValue';
import { useLiteralStyles } from '../styles';

/** Renders a Flyte `Error` literal as an object with message and node id */
export const ErrorValue: React.FC<{ error: Error }> = ({ error }) => {
  const literalStyles = useLiteralStyles();
  return (
    <div className={literalStyles.nestedContainer}>
      <PrintValue label="error" value={error.message} />
      <PrintValue label="nodeId" value={error.failedNodeId} />
    </div>
  );
};
