import { RequestConfig } from 'models/AdminEntity/types';
import { listIdentifiers } from 'models/Common/api';
import { IdentifierScope, NamedEntityIdentifier, ResourceType } from 'models/Common/types';
import { listLaunchPlans } from 'models/Launch/api';
import { LaunchPlan } from 'models/Launch/types';
import { usePagination } from './usePagination';

/** A hook for fetching a paginated list of launch plans */
export function useLaunchPlans(scope: IdentifierScope, config: RequestConfig) {
  return usePagination<LaunchPlan, IdentifierScope>(
    { ...config, cacheItems: true, fetchArg: scope },
    listLaunchPlans,
  );
}

/** A hook for fetching a paginated list of launch plan ids */
export function useLaunchPlanIds(scope: IdentifierScope, config: RequestConfig) {
  return usePagination<NamedEntityIdentifier, IdentifierScope>(
    { ...config, fetchArg: scope },
    (scope, requestConfig) =>
      listIdentifiers({ scope, type: ResourceType.LAUNCH_PLAN }, requestConfig),
  );
}
