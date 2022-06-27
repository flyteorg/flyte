import { withRouteParams } from 'components/common/withRouteParams';
import { EntityDetails } from 'components/Entities/EntityDetails';
import { ResourceIdentifier, ResourceType } from 'models/Common/types';
import * as React from 'react';

export interface LaunchPlanDetailsRouteParams {
  projectId: string;
  domainId: string;
  launchPlanName: string;
}
export type LaunchPlanDetailsProps = LaunchPlanDetailsRouteParams;

/** The view component for the LaunchPlan landing page */
export const LaunchPlanDetailsContainer: React.FC<LaunchPlanDetailsRouteParams> = ({
  projectId,
  domainId,
  launchPlanName,
}) => {
  const id = React.useMemo<ResourceIdentifier>(
    () => ({
      resourceType: ResourceType.LAUNCH_PLAN,
      project: projectId,
      domain: domainId,
      name: launchPlanName,
    }),
    [projectId, domainId, launchPlanName],
  );
  return <EntityDetails id={id} />;
};

export const LaunchPlanDetails = withRouteParams<LaunchPlanDetailsRouteParams>(
  LaunchPlanDetailsContainer,
);
