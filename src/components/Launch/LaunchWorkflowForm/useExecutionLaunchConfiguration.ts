import { log } from 'common/log';
import { useAPIContext } from 'components/data/apiContext';
import { fetchWorkflowExecutionInputs } from 'components/Executions/useWorkflowExecution';
import { FetchableData, useFetchableData } from 'components/hooks';
import { Execution, Variable } from 'models';
import { InitialLaunchParameters, LiteralValueMap } from './types';
import { createInputCacheKey, getInputDefintionForLiteralType } from './utils';

export interface UseExecutionLaunchConfigurationArgs {
    execution: Execution;
    workflowInputs: Record<string, Variable>;
}

/** Returns a fetchable that will result in a `InitialLaunchParameters` object based on a provided `Execution` and its associated workflow inputs. */
export function useExecutionLaunchConfiguration({
    execution,
    workflowInputs
}: UseExecutionLaunchConfigurationArgs): FetchableData<
    InitialLaunchParameters
> {
    const apiContext = useAPIContext();
    return useFetchableData<
        InitialLaunchParameters,
        UseExecutionLaunchConfigurationArgs
    >(
        {
            debugName: 'ExecutionLaunchConfiguration',
            defaultValue: {} as InitialLaunchParameters,
            doFetch: async ({ execution, workflowInputs }) => {
                const {
                    closure: { workflowId },
                    spec: { launchPlan }
                } = execution;

                const inputs = await fetchWorkflowExecutionInputs(
                    execution,
                    apiContext
                );

                const { literals } = inputs;
                const values: LiteralValueMap = Object.keys(literals).reduce(
                    (out, name) => {
                        const workflowInput = workflowInputs[name];
                        if (!workflowInput) {
                            log.error(
                                `Unexpected missing workflow input: ${name}`
                            );
                            return out;
                        }
                        const typeDefinition = getInputDefintionForLiteralType(
                            workflowInput.type
                        );

                        const key = createInputCacheKey(name, typeDefinition);
                        out.set(key, literals[name]);
                        return out;
                    },
                    new Map()
                );

                return { launchPlan, values, workflow: workflowId };
            }
        },
        { execution, workflowInputs }
    );
}
