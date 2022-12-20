import * as React from 'react';
import { useLiteralStyles } from './styles';

/** Helper for consistent rendering of the label rendered next to a value,
 * including support for themes.
 */
export const ValueLabel: React.FC<{ label: React.ReactNode }> = ({ label }) => {
  const literalStyles = useLiteralStyles();
  return <span className={literalStyles.valueLabel}>{label}:</span>;
};
