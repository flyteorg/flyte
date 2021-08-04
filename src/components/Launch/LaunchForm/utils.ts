import { timestampToDate } from 'common/utils';
import { Core } from 'flyteidl';
import { isObject } from 'lodash';
import { Identifier, LiteralType, Variable } from 'models/Common/types';
import { LaunchPlan } from 'models/Launch/types';
import { Task } from 'models/Task/types';
import { Workflow } from 'models/Workflow/types';
import * as moment from 'moment';
import { simpleTypeToInputType, typeLabels } from './constants';
import { inputToLiteral } from './inputHelpers/inputHelpers';
import { typeIsSupported } from './inputHelpers/utils';
import { LaunchState } from './launchMachine';
import { SearchableSelectorOption } from './SearchableSelector';
import {
    BaseInterpretedLaunchState,
    BlobValue,
    InputProps,
    InputType,
    InputTypeDefinition,
    ParsedInput,
    SearchableVersion
} from './types';

/** Creates a unique cache key for an input based on its name and type.
 * Note: This will not be *globally* unique, only unique within a single workflow
 * definition.
 */
export function createInputCacheKey(
    name: string,
    { type }: InputTypeDefinition
): string {
    return `${name}_${type}`;
}

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

export function getTaskInputs(task: Task): Record<string, Variable> {
    if (!task.closure) {
        return {};
    }
    const { compiledTask } = task.closure;
    if (!compiledTask) {
        return {};
    }
    const { interface: ioInterface } = compiledTask.template;
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

/** Formats a list of  records for use in a `SearchableSelector` */
export function versionsToSearchableSelectorOptions(
    items: SearchableVersion[]
): SearchableSelectorOption<Identifier>[] {
    return items.map<SearchableSelectorOption<Identifier>>((item, index) => ({
        data: item.id,
        id: item.id.version,
        name: item.id.version,
        description:
            item.closure && item.closure.createdAt
                ? moment(timestampToDate(item.closure.createdAt)).format(
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

/** Converts a list of Launch form inputs to values that can be submitted with
 * a CreateExecutionRequest.
 */
export function convertFormInputsToLiterals(
    inputs: InputProps[]
): Record<string, Core.ILiteral> {
    const literals: Record<string, Core.ILiteral> = {};
    return inputs.reduce(
        (out, input) => ({
            ...out,
            [input.name]: inputToLiteral(input)
        }),
        literals
    );
}

/** Converts a `LiteralType` to an `InputTypeDefintion` to assist with rendering
 * a type annotation and converting input values.
 */
export function getInputDefintionForLiteralType(
    literalType: LiteralType
): InputTypeDefinition {
    const result: InputTypeDefinition = {
        literalType,
        type: InputType.Unknown
    };

    if (literalType.blob) {
        result.type = InputType.Blob;
    } else if (literalType.collectionType) {
        result.type = InputType.Collection;
        result.subtype = getInputDefintionForLiteralType(
            literalType.collectionType
        );
    } else if (literalType.mapValueType) {
        result.type = InputType.Map;
        result.subtype = getInputDefintionForLiteralType(
            literalType.mapValueType
        );
    } else if (literalType.schema) {
        result.type = InputType.Schema;
    } else if (literalType.simple) {
        result.type = simpleTypeToInputType[literalType.simple];
    } else if (literalType.enumType) {
        result.type = InputType.Enum;
    }
    return result;
}

export function getLaunchInputId(name: string): string {
    return `launch-input-${name}`;
}

/** Given an array of `ParsedInput`s, returns the subset which are required
 * and of a type which cannot be provided via our form.
 */
export function getUnsupportedRequiredInputs(
    inputs: ParsedInput[]
): ParsedInput[] {
    return inputs.filter(
        input =>
            !typeIsSupported(input.typeDefinition) &&
            input.required &&
            input.initialValue === undefined
    );
}

export function isBlobValue(value: unknown): value is BlobValue {
    return isObject(value);
}

/** Determines if a given launch machine state is one in which a user can provide input values. */
export function isEnterInputsState(state: BaseInterpretedLaunchState): boolean {
    return [
        LaunchState.UNSUPPORTED_INPUTS,
        LaunchState.ENTER_INPUTS,
        LaunchState.VALIDATING_INPUTS,
        LaunchState.INVALID_INPUTS,
        LaunchState.SUBMIT_VALIDATING,
        LaunchState.SUBMITTING,
        LaunchState.SUBMIT_FAILED,
        LaunchState.SUBMIT_SUCCEEDED
    ].some(state.matches);
}
