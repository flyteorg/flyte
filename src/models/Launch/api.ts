import { Admin } from 'flyteidl';
import { getAdminEntity } from 'models/AdminEntity/AdminEntity';
import { defaultPaginationConfig } from 'models/AdminEntity/constants';
import { RequestConfig } from 'models/AdminEntity/types';
import { endpointPrefixes } from 'models/Common/constants';
import { Identifier, IdentifierScope } from 'models/Common/types';
import { makeIdentifierPath } from 'models/Common/utils';
import { LaunchPlan } from './types';
import { launchPlanListTransformer } from './utils';

/** Fetches a list of `LaunchPlan` records matching the provided `scope` */
export const listLaunchPlans = (scope: IdentifierScope, config?: RequestConfig) =>
  getAdminEntity(
    {
      path: makeIdentifierPath(endpointPrefixes.launchPlan, scope),
      messageType: Admin.LaunchPlanList,
      transform: launchPlanListTransformer,
    },
    { ...defaultPaginationConfig, ...config },
  );

/** Fetches an individual `LaunchPlan` */
export const getLaunchPlan = (id: Identifier, config?: RequestConfig) =>
  getAdminEntity<Admin.LaunchPlan, LaunchPlan>(
    {
      path: makeIdentifierPath(endpointPrefixes.launchPlan, id),
      messageType: Admin.LaunchPlan,
    },
    config,
  );
