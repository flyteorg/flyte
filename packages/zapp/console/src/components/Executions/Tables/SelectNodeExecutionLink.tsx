import { Link } from '@material-ui/core';
import { NodeExecution, NodeExecutionIdentifier } from 'models/Execution/types';
import * as React from 'react';

/** Renders a link that, when clicked, will trigger selection of the
 * given NodeExecution.
 */
export const SelectNodeExecutionLink: React.FC<{
  className?: string;
  execution: NodeExecution;
  linkText: string;
  setSelectedExecution: (selectedExecutionId: NodeExecutionIdentifier | null) => void;
}> = ({ className, execution, linkText, setSelectedExecution }) => {
  // open the side panel for selected execution's detail
  const onClick = (e: React.MouseEvent<HTMLElement>) => {
    // prevent the parent row body onClick event trigger
    e.stopPropagation();
    // use null in case if there is no execution provided - when it is null will close panel
    setSelectedExecution(execution?.id ?? null);
  };

  return (
    <Link component="button" className={className} onClick={onClick} variant="body1">
      {linkText}
    </Link>
  );
};
