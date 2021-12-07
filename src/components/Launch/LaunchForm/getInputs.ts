import { sortedObjectEntries } from 'common/utils';
import { LaunchPlan } from 'models/Launch/types';
import { Task } from 'models/Task/types';
import { Workflow } from 'models/Workflow/types';
import { requiredInputSuffix } from './constants';
import { LiteralValueMap, ParsedInput } from './types';
import {
    createInputCacheKey,
    formatLabelWithType,
    getInputDefintionForLiteralType,
    getTaskInputs,
    getWorkflowInputs
} from './utils';

// We use a non-empty string for the description to allow display components
// to depend on the existence of a value
const emptyDescription = ' ';

export function getInputsForWorkflow(
    workflow: Workflow,
    launchPlan: LaunchPlan,
    initialValues: LiteralValueMap = new Map()
): ParsedInput[] {
    if (!launchPlan.closure || !workflow) {
        return [];
    }

    const workflowInputs = getWorkflowInputs(workflow);
    const launchPlanInputs = launchPlan.closure.expectedInputs.parameters;
    return sortedObjectEntries(launchPlanInputs).map(value => {
        const [name, parameter] = value;
        const required = !!parameter.required;
        const workflowInput = workflowInputs[name];
        const description =
            workflowInput && workflowInput.description
                ? workflowInput.description
                : emptyDescription;

        const typeDefinition = getInputDefintionForLiteralType(
            parameter.var.type
        );
        const typeLabel = formatLabelWithType(name, typeDefinition);
        const label = required
            ? `${typeLabel}${requiredInputSuffix}`
            : typeLabel;
        const inputKey = createInputCacheKey(name, typeDefinition);
        const defaultVaue =
            parameter.default != null ? parameter.default : undefined;
        const initialValue = initialValues.has(inputKey)
            ? initialValues.get(inputKey)
            : defaultVaue;

        return {
            description,
            initialValue,
            label,
            name,
            required,
            typeDefinition
        };
    });
}

export function getOutputsForWorkflow(launchPlan: LaunchPlan): string[] {
    if (!launchPlan.closure) {
        return [];
    }

    const launchPlanInputs = launchPlan.closure.expectedOutputs.variables;
    return sortedObjectEntries(launchPlanInputs).map(value => {
        const [name, parameter] = value;

        const typeDefinition = getInputDefintionForLiteralType(parameter.type);
        const typeLabel = formatLabelWithType(name, typeDefinition);

        return typeLabel;
    });
}

export function getInputsForTask(
    task: Task,
    initialValues: LiteralValueMap = new Map()
): ParsedInput[] {
    if (!task) {
        return [];
    }

    const taskInputs = getTaskInputs(task);
    return sortedObjectEntries(taskInputs).map(value => {
        const [name, { description = emptyDescription, type }] = value;
        const typeDefinition = getInputDefintionForLiteralType(type);
        const typeLabel = formatLabelWithType(name, typeDefinition);
        const label = `${typeLabel}${requiredInputSuffix}`;
        const inputKey = createInputCacheKey(name, typeDefinition);
        const initialValue = initialValues.has(inputKey)
            ? initialValues.get(inputKey)
            : undefined;

        return {
            description,
            initialValue,
            label,
            name,
            typeDefinition,
            // Task inputs are always required, as there is no default provided
            // by a parent workflow.
            required: true
        };
    });
}
