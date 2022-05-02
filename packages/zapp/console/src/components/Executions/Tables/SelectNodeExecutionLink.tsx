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
  // open the side panel for selected execution's detail
  const onClick = (e: React.MouseEvent<HTMLElement>) => {
    // prevent the parent row body onClick event trigger
    e.stopPropagation();
    // use null in case if there is no execution provided - when it is null will close panel
    state.setSelectedExecution(execution?.id ?? null);
  };

  return (
    <Link component="button" className={className} onClick={onClick} variant="body1">
      {linkText}
    </Link>
  );
};
