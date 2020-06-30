import { log } from 'common/log';
import { useWorkflowExecutionInputs } from 'components/Executions/useWorkflowExecution';
import { FetchableData } from 'components/hooks';
import { Execution, Variable } from 'models';
import { useEffect, useState } from 'react';
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
    const inputs = useWorkflowExecutionInputs(execution);
    const [values, setValues] = useState<LiteralValueMap>(new Map());
    const {
        closure: { workflowId },
        spec: { launchPlan }
    } = execution;

    const value = { launchPlan, values, workflow: workflowId };

    useEffect(() => {
        const { literals } = inputs.value;
        const convertedValues: LiteralValueMap = Object.keys(literals).reduce(
            (out, name) => {
                const workflowInput = workflowInputs[name];
                if (!workflowInput) {
                    log.error(`Unexpected missing workflow input: ${name}`);
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

        setValues(convertedValues);
    }, [inputs.value]);

    return { ...inputs, value };
}
