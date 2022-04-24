import * as React from 'react';
import { PrintValue } from './PrintValue';

/** Reusable component which prints an informational message about a value type
 * which we don't currently support in the UI
 */
export const UnsupportedType: React.FC<{ label: React.ReactNode }> = ({ label }) => (
  <PrintValue label={label} value={<em>This type is not yet supported.</em>} />
);
