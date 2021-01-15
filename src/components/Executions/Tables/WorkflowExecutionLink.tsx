import * as classnames from 'classnames';
import { useCommonStyles } from 'components/common/styles';
import { WorkflowExecutionIdentifier } from 'models/Execution/types';
import * as React from 'react';
import { Link as RouterLink } from 'react-router-dom';
import { Routes } from 'routes/routes';

/** A simple component to render a link to a specific WorkflowExecution */
export const WorkflowExecutionLink: React.FC<{
    className?: string;
    id: WorkflowExecutionIdentifier;
}> = ({ className, id }) => {
    const commonStyles = useCommonStyles();
    return (
        <RouterLink
            className={classnames(commonStyles.primaryLink, className)}
            to={Routes.ExecutionDetails.makeUrl(id)}
        >
            {id.name}
        </RouterLink>
    );
};
