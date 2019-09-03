import { dateToTimestamp, millisecondsToDuration } from 'common/utils';
import { Core } from 'flyteidl';
import * as Long from 'long';
import { utc as moment } from 'moment';
import { InputType } from './types';

function booleanToLiteral(value: string): Core.ILiteral {
    return { scalar: { primitive: { boolean: Boolean(value) } } };
}

function stringToLiteral(stringValue: string): Core.ILiteral {
    return { scalar: { primitive: { stringValue } } };
}

function integerToLiteral(value: string): Core.ILiteral {
    return {
        scalar: { primitive: { integer: Long.fromString(value) } }
    };
}

function floatToLiteral(value: string): Core.ILiteral {
    return {
        scalar: { primitive: { floatValue: parseFloat(value) } }
    };
}

function dateToLiteral(value: string): Core.ILiteral {
    const datetime = dateToTimestamp(moment(value).toDate());
    return {
        scalar: { primitive: { datetime } }
    };
}

function durationToLiteral(value: string): Core.ILiteral {
    const duration = millisecondsToDuration(parseInt(value, 10));
    return {
        scalar: { primitive: { duration } }
    };
}

const unsupportedInput = () => undefined;

type ConverterFn = (value: string) => Core.ILiteral | undefined;

/** Maps an `InputType` to a function which will convert its value into a `Literal` */
export const inputTypeConverters: Record<InputType, ConverterFn> = {
    [InputType.Binary]: unsupportedInput,
    [InputType.Blob]: unsupportedInput,
    [InputType.Boolean]: booleanToLiteral,
    [InputType.Collection]: unsupportedInput,
    [InputType.Datetime]: dateToLiteral,
    [InputType.Duration]: durationToLiteral,
    [InputType.Error]: unsupportedInput,
    [InputType.Float]: floatToLiteral,
    [InputType.Integer]: integerToLiteral,
    [InputType.Map]: unsupportedInput,
    [InputType.None]: unsupportedInput,
    [InputType.Schema]: unsupportedInput,
    [InputType.String]: stringToLiteral,
    [InputType.Struct]: unsupportedInput,
    [InputType.Unknown]: unsupportedInput
};
