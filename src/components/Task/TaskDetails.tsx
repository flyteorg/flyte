import { withRouteParams } from 'components/common/withRouteParams';
import { EntityDetails } from 'components/Entities/EntityDetails';
import { ResourceIdentifier, ResourceType } from 'models/Common/types';
import * as React from 'react';

export interface TaskDetailsRouteParams {
  projectId: string;
  domainId: string;
  taskName: string;
}
export type TaskDetailsProps = TaskDetailsRouteParams;

/** The view component for the Task landing page */
export const TaskDetailsContainer: React.FC<TaskDetailsRouteParams> = ({
  projectId,
  domainId,
  taskName,
}) => {
  const id = React.useMemo<ResourceIdentifier>(
    () => ({
      resourceType: ResourceType.TASK,
      project: projectId,
      domain: domainId,
      name: taskName,
    }),
    [projectId, domainId, taskName],
  );
  return <EntityDetails id={id} />;
};

export const TaskDetails = withRouteParams<TaskDetailsRouteParams>(TaskDetailsContainer);
