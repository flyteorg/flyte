import { Link } from '@material-ui/core';
import { NodeExecution } from 'models/Execution/types';
import * as React from 'react';
import { NodeExecutionsTableState } from './types';

/** Renders a link that, when clicked, will trigger selection of the
 * given NodeExecution.
 */
export const SelectNodeExecutionLink: React.FC<{
  className?: string;
  execution: NodeExecution;
  linkText: string;
  state: NodeExecutionsTableState;
}> = ({ className, execution, linkText, state }) => {
  const onClick = () => state.setSelectedExecution(execution.id);
  return (
    <Link component="button" className={className} onClick={onClick} variant="body1">
      {linkText}
    </Link>
  );
};
