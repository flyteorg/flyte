import * as React from 'react';

import { DumpJSON, WaitForData } from 'components/common';
import { useTaskTemplate } from 'components/hooks';
import { Identifier } from 'models';

interface TaskNodeDetailsProps {
    taskId: Identifier;
}

/** Renders information about a Task node in a workflow graph. */
export const TaskNodeDetails: React.FC<TaskNodeDetailsProps> = ({ taskId }) => {
    const template = useTaskTemplate(taskId);
    return (
        <WaitForData {...template}>
            <DumpJSON value={template.value} />
        </WaitForData>
    );
};
