import { Core } from 'flyteidl';
import { InputValue } from '../types';
import { primitiveLiteralPaths } from './constants';
import { ConverterInput, InputHelper, InputValidatorParams } from './types';
import { extractLiteralWithCheck } from './utils';

function fromLiteral(literal: Core.ILiteral): InputValue {
    return extractLiteralWithCheck<string>(
        literal,
        primitiveLiteralPaths.scalarString
    );
}

function toLiteral({ value }: ConverterInput): Core.ILiteral {
    const stringValue = typeof value === 'string' ? value : value.toString();
    return { scalar: { primitive: { stringValue } } };
}

function validate({ value }: InputValidatorParams) {
    if (typeof value !== 'string') {
        throw new Error('Value is not a string');
    }
    if (value && value[0] === ' ') {
        throw new Error('Value should not have a leading space');
    }
    if (value && value[value.length - 1] === ' ') {
        throw new Error('Value should not have a trailing space');
    }
}

export const stringHelper: InputHelper = {
    fromLiteral,
    toLiteral,
    validate
};
