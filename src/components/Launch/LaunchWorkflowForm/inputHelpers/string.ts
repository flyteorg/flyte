import { Core } from 'flyteidl';
import { ConverterInput, InputHelper } from './types';

function toLiteral({ value }: ConverterInput): Core.ILiteral {
    const stringValue = typeof value === 'string' ? value : value.toString();
    return { scalar: { primitive: { stringValue } } };
}

function validate({ value }: ConverterInput) {
    if (typeof value !== 'string') {
        throw new Error('Value is not a string');
    }
}

export const stringHelper: InputHelper = {
    toLiteral,
    validate
};
