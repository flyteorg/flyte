import * as classnames from 'classnames';
import { useCommonStyles } from 'components/common/styles';
import { WorkflowExecutionIdentifier } from 'models/Execution/types';
import * as React from 'react';
import { Link as RouterLink } from 'react-router-dom';
import { Routes } from 'routes/routes';
import { history } from 'routes/history';

/** A simple component to render a link to a specific WorkflowExecution */
export const WorkflowExecutionLink: React.FC<{
    className?: string;
    id: WorkflowExecutionIdentifier;
}> = ({ className, id }) => {
    const commonStyles = useCommonStyles();
    const {
        location: { pathname }
    } = history;
    const fromExecutionNav = pathname.split('/').pop() === 'executions';

    return (
        <RouterLink
            className={classnames(commonStyles.primaryLink, className)}
            to={`${Routes.ExecutionDetails.makeUrl(id)}${
                fromExecutionNav ? '?fromExecutionNav=true' : ''
            }`}
        >
            {id.name}
        </RouterLink>
    );
};
