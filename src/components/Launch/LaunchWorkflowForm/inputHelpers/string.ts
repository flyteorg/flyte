import { Core } from 'flyteidl';
import { InputValue } from '../types';
import { literalValuePaths } from './constants';
import { ConverterInput, InputHelper } from './types';
import { extractLiteralWithCheck } from './utils';

function fromLiteral(literal: Core.ILiteral): InputValue {
    return extractLiteralWithCheck<string>(
        literal,
        literalValuePaths.scalarString
    );
}

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
    fromLiteral,
    toLiteral,
    validate
};
