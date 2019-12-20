import { dateToTimestamp } from 'common/utils';
import { Core } from 'flyteidl';
import { utc as moment } from 'moment';
import { ConverterInput, InputHelper } from './types';

function toLiteral({ value }: ConverterInput): Core.ILiteral {
    const parsed =
        value instanceof Date ? value : moment(value.toString()).toDate();
    const datetime = dateToTimestamp(parsed);
    return {
        scalar: { primitive: { datetime } }
    };
}

function validate({ value }: ConverterInput) {}

export const datetimeHelper: InputHelper = {
    toLiteral,
    validate
};
