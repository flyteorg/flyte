import { useCommonStyles } from 'components/common/styles';
import * as React from 'react';
import { useLiteralStyles } from './styles';
import { ValueLabel } from './ValueLabel';

/** Helper for consistent rendering of a label and value pair */
export const PrintValue: React.FC<{
  label: React.ReactNode;
  value: React.ReactNode;
}> = ({ label, value }) => {
  const commonStyles = useCommonStyles();
  const literalStyles = useLiteralStyles();
  return (
    <div className={literalStyles.labelValueContainer}>
      <ValueLabel label={label} />
      <div className={commonStyles.textWrapped}>{value}</div>
    </div>
  );
};
