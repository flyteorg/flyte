import { timestampToDate } from 'common/utils';
import {
    LaunchPlan,
    Literal,
    LiteralMap,
    LiteralType,
    Variable,
    Workflow,
    WorkflowId
} from 'models';
import * as moment from 'moment';
import { simpleTypeToInputType, typeLabels } from './constants';
import { inputTypeConverters } from './inputConverters';
import { SearchableSelectorOption } from './SearchableSelector';
import { InputProps, InputType, InputTypeDefinition } from './types';

/** Safely retrieves the input mapping stored in a workflow, or an empty
 * logic if any optional property along the chain is undefined.
 */
export function getWorkflowInputs(
    workflow: Workflow
): Record<string, Variable> {
    if (!workflow.closure) {
        return {};
    }
    const { compiledWorkflow } = workflow.closure;
    if (!compiledWorkflow) {
        return {};
    }
    const { interface: ioInterface } = compiledWorkflow.primary.template;
    if (!ioInterface) {
        return {};
    }
    const { inputs } = ioInterface;
    if (!inputs) {
        return {};
    }
    return inputs.variables;
}

/** Returns a formatted string based on an InputTypeDefinition.
 * ex. `string`, `string[]`, `map<string, number>`
 */
export function formatType({ type, subtype }: InputTypeDefinition): string {
    if (type === InputType.Collection) {
        return subtype ? `${formatType(subtype)}[]` : 'collection';
    }
    if (type === InputType.Map) {
        return subtype ? `map<string, ${formatType(subtype)}>` : 'map';
    }
    return typeLabels[type];
}

/** Combines a label string with a formatted type definition to generate a string
 * suitable for displaying in a launch form.
 */
export function formatLabelWithType(label: string, type: InputTypeDefinition) {
    const typeString = formatType(type);
    return `${label}${typeString ? ` (${typeString})` : ''}`;
}

/** Formats a list of `Workflow` records for use in a `SearchableSelector` */
export function workflowsToSearchableSelectorOptions(
    workflows: Workflow[]
): SearchableSelectorOption<WorkflowId>[] {
    return workflows.map<SearchableSelectorOption<WorkflowId>>((wf, index) => ({
        data: wf.id,
        id: wf.id.version,
        name: wf.id.version,
        description:
            wf.closure && wf.closure.createdAt
                ? moment(timestampToDate(wf.closure.createdAt)).format(
                      'DD MMM YYYY'
                  )
                : ''
    }));
}

/** Formats a list of `LaunchPlan` records for use in a `SearchableSelector` */
export function launchPlansToSearchableSelectorOptions(
    launchPlans: LaunchPlan[]
): SearchableSelectorOption<LaunchPlan>[] {
    return launchPlans.map<SearchableSelectorOption<LaunchPlan>>(lp => ({
        data: lp,
        id: lp.id.name,
        name: lp.id.name,
        description: ''
    }));
}

function inputToLiteral(input: InputProps) {
    if (!input.value) {
        return undefined;
    }
    const converter = inputTypeConverters[input.typeDefinition.type];
    const value = input.value.toString();
    return converter(value);
}

/** Converts a list of Launch form inputs to values that can be submitted with
 * a CreateExecutionRequest.
 */
export function convertFormInputsToLiteralMap(
    inputs: InputProps[]
): LiteralMap {
    const literals = inputs.reduce<Record<string, Literal>>((out, input) => {
        const converted = inputToLiteral(input);
        return converted
            ? Object.assign(out, { [input.name]: converted })
            : out;
    }, {});
    return {
        literals
    };
}

export function getInputDefintionForLiteralType(
    literalType: LiteralType
): InputTypeDefinition {
    if (literalType.blob) {
        return { type: InputType.Blob };
    }

    if (literalType.collectionType) {
        return {
            type: InputType.Collection,
            subtype: getInputDefintionForLiteralType(literalType.collectionType)
        };
    }

    if (literalType.mapValueType) {
        return {
            type: InputType.Map,
            subtype: getInputDefintionForLiteralType(literalType.mapValueType)
        };
    }

    if (literalType.schema) {
        return { type: InputType.Schema };
    }

    if (literalType.simple) {
        return { type: simpleTypeToInputType[literalType.simple] };
    }

    return { type: InputType.Unknown };
}
