import { Link } from '@material-ui/core';
import * as React from 'react';
import { DetailedNodeExecution } from '../types';
import { NodeExecutionsTableState } from './types';

/** Renders a link that, when clicked, will trigger selection of the
 * given NodeExecution.
 */
export const SelectNodeExecutionLink: React.FC<{
    className?: string;
    execution: DetailedNodeExecution;
    linkText: string;
    state: NodeExecutionsTableState;
}> = ({ className, execution, linkText, state }) => {
    const onClick = () => state.setSelectedExecution(execution);
    return (
        <Link
            className={className}
            component="button"
            onClick={onClick}
            variant="body1"
        >
            {linkText}
        </Link>
    );
};
