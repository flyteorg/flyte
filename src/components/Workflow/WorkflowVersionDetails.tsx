import { withRouteParams } from 'components/common/withRouteParams';
import { EntityDetails } from 'components/Entities/EntityDetails';
import { ResourceIdentifier, ResourceType } from 'models/Common/types';
import * as React from 'react';

export interface WorkflowVersionDetailsRouteParams {
    projectId: string;
    domainId: string;
    workflowName: string;
    workflowVersion: string;
}
export type WorkflowDetailsProps = WorkflowVersionDetailsRouteParams;

/**
 * The view component for the Workflow Versions page
 * @param projectId
 * @param domainId
 * @param workflowName
 */
export const WorkflowVersionDetailsContainer: React.FC<WorkflowVersionDetailsRouteParams> = ({
    projectId,
    domainId,
    workflowName,
    workflowVersion
}) => {
    const id = React.useMemo<ResourceIdentifier>(
        () => ({
            resourceType: ResourceType.WORKFLOW,
            project: projectId,
            domain: domainId,
            name: workflowName,
            version: workflowVersion
        }),
        [projectId, domainId, workflowName, workflowVersion]
    );
    return <EntityDetails id={id} versionView showStaticGraph />;
};

export const WorkflowVersionDetails = withRouteParams<
    WorkflowVersionDetailsRouteParams
>(WorkflowVersionDetailsContainer);
