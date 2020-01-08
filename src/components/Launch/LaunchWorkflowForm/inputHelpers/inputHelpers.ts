import { ValueError } from 'errors';
import { Core } from 'flyteidl';
import { Literal } from 'models';
import { InputProps, InputTypeDefinition, InputValue } from '../types';
import { literalNone } from './constants';
import { getHelperForInput } from './getHelperForInput';

type ToLiteralParams = Pick<InputProps, 'typeDefinition' | 'value'>;
export function inputToLiteral(input: ToLiteralParams): Core.ILiteral {
    if (input.value == null) {
        return literalNone();
    }
    const { typeDefinition, value } = input;

    const { toLiteral } = getHelperForInput(typeDefinition.type);
    return toLiteral({ value, typeDefinition });
}

export function defaultValueForInputType(
    typeDefinition: InputTypeDefinition
): InputValue | undefined {
    return getHelperForInput(typeDefinition.type).defaultValue;
}

export function literalToInputValue(
    typeDefinition: InputTypeDefinition,
    literal: Core.ILiteral
): InputValue | undefined {
    const { defaultValue, fromLiteral } = getHelperForInput(
        typeDefinition.type
    );

    try {
        return fromLiteral(literal);
    } catch (e) {
        // If something goes wrong (most likely malformed default value input),
        // we'll return the system default value.
        console.debug((e as Error).message);
        return defaultValue;
    }
}

type ValidationParams = Pick<InputProps, 'name' | 'typeDefinition' | 'value'>;
export function validateInput(input: ValidationParams) {
    if (input.value == null) {
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
