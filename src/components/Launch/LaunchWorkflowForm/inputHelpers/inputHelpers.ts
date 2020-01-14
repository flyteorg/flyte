import { ValueError } from 'errors';
import { Core } from 'flyteidl';
import { Literal } from 'models';
import { InputProps, InputTypeDefinition, InputValue } from '../types';
import { literalNone } from './constants';
import { getHelperForInput } from './getHelperForInput';

type ToLiteralParams = Pick<InputProps, 'typeDefinition' | 'value'>;
/** Converts a type/InputValue combination to a `Core.ILiteral` which can be
 * submitted to Admin for creating an execution.
 */
export function inputToLiteral(input: ToLiteralParams): Core.ILiteral {
    if (input.value == null) {
        return literalNone();
    }
    const { typeDefinition, value } = input;

    const { toLiteral } = getHelperForInput(typeDefinition.type);
    return toLiteral({ value, typeDefinition });
}

/** Generates the default value (if any) for a given type. */
export function defaultValueForInputType(
    typeDefinition: InputTypeDefinition
): InputValue | undefined {
    return getHelperForInput(typeDefinition.type).defaultValue;
}

/** Converts a Core.ILiteral to an InputValue which can be used to populate
 * a form control.
 */
export function literalToInputValue(
    typeDefinition: InputTypeDefinition,
    literal: Core.ILiteral
): InputValue | undefined {
    const { defaultValue, fromLiteral } = getHelperForInput(
        typeDefinition.type
    );

    if (literal.scalar && literal.scalar.noneType) {
        return undefined;
    }

    try {
        return fromLiteral(literal, typeDefinition);
    } catch (e) {
        // If something goes wrong (most likely malformed default value input),
        // we'll return the system default value.
        console.debug((e as Error).message);
        return defaultValue;
    }
}

type ValidationParams = Pick<
    InputProps,
    'name' | 'required' | 'typeDefinition' | 'value'
>;
/** Validates a given InputValue based on rules for the provided type. Returns
 * void if no errors, throws an error otherwise.
 */
export function validateInput(input: ValidationParams) {
    if (input.value == null) {
        if (input.required) {
            throw new ValueError(input.name, 'Value is required');
        }
        return;
    }

    const { typeDefinition, value } = input;

    const { validate } = getHelperForInput(typeDefinition.type);
    try {
        validate({ value, typeDefinition });
    } catch (e) {
        const error = e as Error;
        throw new ValueError(input.name, error.message);
    }
}
