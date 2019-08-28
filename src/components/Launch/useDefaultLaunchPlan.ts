import { useAPIContext } from 'components/data/apiContext';
import { useFetchableData } from 'components/hooks';
import { NotFoundError } from 'errors';
import { LaunchPlan, Workflow, WorkflowId } from 'models';

export function useDefaultLaunchPlan(workflow?: Workflow) {
    const { getLaunchPlan } = useAPIContext();

    const doFetch = async (id?: WorkflowId) => {
        if (!id) {
            throw new Error('No workflow id provided');
        }
        try {
            // Default launch plan is the one whose name/version matches the workflow
            return await getLaunchPlan(id);
        } catch (error) {
            if (error instanceof NotFoundError) {
                console.log('No default launch plan found');
                return undefined;
            }
            throw error;
        }
    };
    return useFetchableData<LaunchPlan | undefined, WorkflowId | undefined>({
        doFetch,
        defaultValue: {} as LaunchPlan,
        debugName: 'useDefaultLaunchPlan',
        autoFetch: !!workflow
    });
}
