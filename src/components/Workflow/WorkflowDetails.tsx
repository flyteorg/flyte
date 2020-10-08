import { withRouteParams } from 'components/common';
import { EntityDetails } from 'components/Entities/EntityDetails';
import { ResourceIdentifier, ResourceType } from 'models';
import * as React from 'react';

export interface WorkflowDetailsRouteParams {
    projectId: string;
    domainId: string;
    workflowName: string;
}
export type WorkflowDetailsProps = WorkflowDetailsRouteParams;

/** The view component for the Workflow landing page */
export const WorkflowDetailsContainer: React.FC<WorkflowDetailsRouteParams> = ({
    projectId,
    domainId,
    workflowName
}) => {
    const id = React.useMemo<ResourceIdentifier>(
        () => ({
            resourceType: ResourceType.WORKFLOW,
            project: projectId,
            domain: domainId,
            name: workflowName
        }),
        [projectId, domainId, workflowName]
    );
    return <EntityDetails id={id} />;
};

export const WorkflowDetails = withRouteParams<WorkflowDetailsRouteParams>(
    WorkflowDetailsContainer
);
