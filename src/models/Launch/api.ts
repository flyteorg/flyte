import { Admin } from 'flyteidl';
import { defaultPaginationConfig, RequestConfig } from 'models/AdminEntity';
import { getAdminEntity } from 'models/AdminEntity/AdminEntity';
import {
    endpointPrefixes,
    Identifier,
    IdentifierScope,
    makeIdentifierPath
} from 'models/Common';
import { LaunchPlan } from './types';
import { launchPlanListTransformer } from './utils';

/** Fetches a list of `LaunchPlan` records matching the provided `scope` */
export const listLaunchPlans = (
    scope: IdentifierScope,
    config?: RequestConfig
) =>
    getAdminEntity(
        {
            path: makeIdentifierPath(endpointPrefixes.launchPlan, scope),
            messageType: Admin.LaunchPlanList,
            transform: launchPlanListTransformer
        },
        { ...defaultPaginationConfig, ...config }
    );

/** Fetches an individual `LaunchPlan` */
export const getLaunchPlan = (id: Identifier, config?: RequestConfig) =>
    getAdminEntity<Admin.LaunchPlan, LaunchPlan>(
        {
            path: makeIdentifierPath(endpointPrefixes.launchPlan, id),
            messageType: Admin.Task
        },
        config
    );
