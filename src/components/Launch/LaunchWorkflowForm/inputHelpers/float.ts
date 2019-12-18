import { Core } from 'flyteidl';
import { ConverterInput, InputHelper } from './types';

function toLiteral({ value }: ConverterInput): Core.ILiteral {
    const floatValue =
        typeof value === 'number' ? value : parseFloat(value.toString());
    return {
        scalar: { primitive: { floatValue } }
    };
}

function validate({ value }: ConverterInput) {}

export const floatHelper: InputHelper = {
    toLiteral,
    validate
};
