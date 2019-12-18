import { ValueError } from 'errors';
import { Core } from 'flyteidl';
import { InputProps } from '../types';
import { literalNone } from './constants';
import { getHelperForInput } from './getHelperForInput';

export function inputToLiteral(input: InputProps): Core.ILiteral {
    if (input.value == null) {
        return literalNone();
    }
    const { typeDefinition, value } = input;

    const { toLiteral } = getHelperForInput(typeDefinition.type);
    return toLiteral({ value, typeDefinition });
}

export function validateInput(input: InputProps) {
    if (input.value == null) {
        // TODO: If input is required, check that here.
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
