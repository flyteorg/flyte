import { sortedObjectEntries } from 'common/utils';
import { LaunchPlan, Workflow } from 'models';
import { requiredInputSuffix } from './constants';
import { LiteralValueMap, ParsedInput } from './types';
import {
    createInputCacheKey,
    formatLabelWithType,
    getInputDefintionForLiteralType,
    getWorkflowInputs
} from './utils';

// We use a non-empty string for the description to allow display components
// to depend on the existence of a value
const emptyDescription = ' ';

export function getInputs(
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
