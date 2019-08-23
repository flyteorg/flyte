import * as React from 'react';

import { WaitForData, withRouteParams } from 'components/common';
import { useWorkflow } from 'components/hooks';

import { WorkflowVersionDetailsContent } from './WorkflowVersionDetailsContent';

export interface WorkflowVersionDetailsRouteParams {
    domainId: string;
    projectId: string;
    workflowName: string;
    version: string;
}
export type WorkflowVersionDetailsProps = WorkflowVersionDetailsRouteParams;

/** The view component for Workflow version details page */
export const WorkflowVersionDetailsContainer: React.FC<
    WorkflowVersionDetailsRouteParams
> = ({ domainId, projectId, workflowName, version }) => {
    const workflow = useWorkflow({
        version,
        project: projectId,
        domain: domainId,
        name: workflowName
    });

    return (
        <WaitForData {...workflow}>
            <WorkflowVersionDetailsContent workflow={workflow.value} />
        </WaitForData>
    );
};

export const WorkflowVersionDetails = withRouteParams<
    WorkflowVersionDetailsRouteParams
>(WorkflowVersionDetailsContainer);
