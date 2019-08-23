import { DumpJSON } from 'components/common';
import { useCommonStyles } from 'components/common/styles';
import { DataError } from 'components/Errors';
import * as React from 'react';
import { DetailedNodeExecution } from '../types';

/** Render the task template for a given NodeExecution */
export const NodeExecutionTaskDetails: React.FC<{
    execution: DetailedNodeExecution;
}> = ({ execution }) => {
    const commonStyles = useCommonStyles();
    const content = execution.taskTemplate ? (
        <DumpJSON value={execution.taskTemplate} />
    ) : (
        <DataError errorTitle="No task information found" />
    );
    return (
        <div className={commonStyles.detailsPanelCard}>
            <div className={commonStyles.detailsPanelCardContent}>
                {content}
            </div>
        </div>
    );
};
