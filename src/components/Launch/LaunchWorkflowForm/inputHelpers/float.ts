import { Core } from 'flyteidl';
import { InputValue } from '../types';
import { ConverterInput, InputHelper } from './types';

function toLiteral({ value }: ConverterInput): Core.ILiteral {
    const floatValue =
        typeof value === 'number' ? value : parseFloat(value.toString());
    return {
        scalar: { primitive: { floatValue } }
    };
}

export function isValidFloat(value: InputValue): boolean {
    if (typeof value === 'number') {
        return true;
    }
    if (typeof value === 'string' && !Number.isNaN(Number.parseFloat(value))) {
        return true;
    }
    return false;
}

function validate({ value }: ConverterInput) {
    if (!isValidFloat(value)) {
        throw new Error('Value is not a valid floating point number');
    }
}

export const floatHelper: InputHelper = {
    toLiteral,
    validate
};
