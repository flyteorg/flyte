import { millisecondsToDuration } from 'common/utils';
import { Core } from 'flyteidl';
import { isValidFloat } from './float';
import { ConverterInput, InputHelper } from './types';

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
    toLiteral,
    validate
};
