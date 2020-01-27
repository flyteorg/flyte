import { FetchableData, useWorkflowExecutionInputs } from 'components/hooks';
import { Execution, Variable } from 'models';
import { useEffect, useState } from 'react';
import { literalToInputValue } from './inputHelpers/inputHelpers';
import { InitialLaunchParameters, InputValueMap } from './types';
import { createInputCacheKey, getInputDefintionForLiteralType } from './utils';

export interface UseExecutionLaunchConfigurationArgs {
    execution: Execution;
    workflowInputs: Record<string, Variable>;
}

export function useExecutionLaunchConfiguration({
    execution,
    workflowInputs
}: UseExecutionLaunchConfigurationArgs): FetchableData<
    InitialLaunchParameters
> {
    const inputs = useWorkflowExecutionInputs(execution);
    const [values, setValues] = useState<InputValueMap>(new Map());
    const {
        closure: { workflowId },
        spec: { launchPlan }
    } = execution;

    const value = { launchPlan, values, workflow: workflowId };

    useEffect(() => {
        const { literals } = inputs.value;
        const convertedValues = Object.keys(literals).reduce((out, name) => {
            const workflowInput = workflowInputs[name];
            if (!workflowInput) {
                // TODO-NOW: Log it? Error?
                return out;
            }
            const typeDefinition = getInputDefintionForLiteralType(
                workflowInput.type
            );
            const inputValue = literalToInputValue(
                typeDefinition,
                literals[name]
            );
            const key = createInputCacheKey(name, typeDefinition);
            if (inputValue !== undefined) {
                out.set(key, inputValue);
            }
            return out;
        }, new Map());

        setValues(convertedValues);
    }, [inputs.value]);

    return { ...inputs, value };
}
