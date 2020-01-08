import { dateToTimestamp } from 'common/utils';
import { Core } from 'flyteidl';
import { utc as moment } from 'moment';
import { InputValue } from '../types';
import { allowedDateFormats } from './constants';
import { ConverterInput, InputHelper } from './types';

function parseDate(value: InputValue) {
    return value instanceof Date
        ? value
        : moment(value.toString(), allowedDateFormats).toDate();
}

function toLiteral({ value }: ConverterInput): Core.ILiteral {
    const datetime = dateToTimestamp(parseDate(value));
    return {
        scalar: { primitive: { datetime } }
    };
}

function validate({ value }: ConverterInput) {
    const parsed = parseDate(value);
    if (Number.isNaN(parsed.getTime())) {
        throw new Error('Value is not a valid Date');
    }
}

export const datetimeHelper: InputHelper = {
    toLiteral,
    validate
};
