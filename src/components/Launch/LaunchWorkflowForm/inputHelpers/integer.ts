import { Core } from 'flyteidl';
import * as Long from 'long';
import { ConverterInput, InputHelper } from './types';

function toLiteral({ value }: ConverterInput): Core.ILiteral {
    const integer =
        value instanceof Long ? value : Long.fromString(value.toString());
    return {
        scalar: { primitive: { integer } }
    };
}

function validate({ value }: ConverterInput) {}

export const integerHelper: InputHelper = {
    toLiteral,
    validate
};
