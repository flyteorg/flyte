import { sortedObjectEntries } from 'common/utils';
import { LaunchPlan, Workflow } from 'models';
import {
    defaultValueForInputType,
    literalToInputValue
} from './inputHelpers/inputHelpers';
import { ParsedInput } from './types';
import {
    formatLabelWithType,
    getInputDefintionForLiteralType,
    getWorkflowInputs
} from './utils';

// We use a non-empty string for the description to allow display components
// to depend on the existence of a value
const emptyDescription = ' ';

export function getInputs(
    workflow: Workflow,
    launchPlan: LaunchPlan
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
        const label = required ? `${typeLabel}*` : typeLabel;

        const defaultValue =
            parameter.default != null
                ? literalToInputValue(typeDefinition, parameter.default)
                : defaultValueForInputType(typeDefinition);

        return {
            defaultValue,
            description,
            label,
            name,
            required,
            typeDefinition
        };
    });
}
