import { durationToMilliseconds, millisecondsToDuration } from 'common/utils';
import { Core, Protobuf } from 'flyteidl';
import { InputValue } from '../types';
import { literalValuePaths } from './constants';
import { isValidFloat } from './float';
import { ConverterInput, InputHelper } from './types';
import { extractLiteralWithCheck } from './utils';

function fromLiteral(literal: Core.ILiteral): InputValue {
    const value = extractLiteralWithCheck<Protobuf.IDuration>(
        literal,
        literalValuePaths.scalarDuration
    );
    return durationToMilliseconds(value);
}

function toLiteral({ value }: ConverterInput): Core.ILiteral {
    const parsed =
        typeof value === 'number' ? value : parseInt(value.toString(), 10);
    const duration = millisecondsToDuration(parsed);
    return {
        scalar: { primitive: { duration } }
    };
}

function validate({ value }: ConverterInput) {
    if (!isValidFloat(value)) {
        throw new Error('Value is not a valid duration');
    }
}

export const durationHelper: InputHelper = {
    fromLiteral,
    toLiteral,
    validate
};
