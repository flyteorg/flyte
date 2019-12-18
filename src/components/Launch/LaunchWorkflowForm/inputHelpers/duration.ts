import { millisecondsToDuration } from 'common/utils';
import { Core } from 'flyteidl';
import { ConverterInput, InputHelper } from './types';

function toLiteral({ value }: ConverterInput): Core.ILiteral {
    const parsed =
        typeof value === 'number' ? value : parseInt(value.toString(), 10);
    const duration = millisecondsToDuration(parsed);
    return {
        scalar: { primitive: { duration } }
    };
}

function validate({ value }: ConverterInput) {}

export const durationHelper: InputHelper = {
    toLiteral,
    validate
};
