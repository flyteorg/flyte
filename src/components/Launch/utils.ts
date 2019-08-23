import { Admin } from 'flyteidl';
import { LaunchPlan } from 'models';

export function isLaunchPlanActive(launchPlan: LaunchPlan) {
    if (!launchPlan.closure) {
        return false;
    }
    return launchPlan.closure.state === Admin.LaunchPlanState.ACTIVE;
}
