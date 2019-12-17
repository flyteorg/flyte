import { dateToTimestamp, millisecondsToDuration } from 'common/utils';
import { Core } from 'flyteidl';
import * as Long from 'long';
import * as LossLessJSON from 'lossless-json';
import { utc as moment } from 'moment';
import { InputProps, InputType, InputTypeDefinition } from './types';

function lossLessToLongReviver(key: string, value: any) {
    if (value && value.isLosslessNumber) {
        return Long.fromString(value.toString());
    }
    return value;
}

function parseJSON(value: string) {
    return LossLessJSON.parse(value, lossLessToLongReviver);
}

interface ConverterInput<T = never> {
    value: string | T;
    typeDefinition: InputTypeDefinition;
}

function literalNone(): Core.ILiteral {
    return { scalar: { noneType: {} } };
}

function booleanToLiteral({ value }: ConverterInput<boolean>): Core.ILiteral {
    return { scalar: { primitive: { boolean: Boolean(value) } } };
}

function stringToLiteral({
    value: stringValue
}: ConverterInput): Core.ILiteral {
    return { scalar: { primitive: { stringValue } } };
}

function integerToLiteral({ value }: ConverterInput<Long>): Core.ILiteral {
    const integer = value instanceof Long ? value : Long.fromString(value);
    return {
        scalar: { primitive: { integer } }
    };
}

function floatToLiteral({ value }: ConverterInput<number>): Core.ILiteral {
    const floatValue = typeof value === 'number' ? value : parseFloat(value);
    return {
        scalar: { primitive: { floatValue } }
    };
}

function dateToLiteral({ value }: ConverterInput): Core.ILiteral {
    const datetime = dateToTimestamp(moment(value).toDate());
    return {
        scalar: { primitive: { datetime } }
    };
}

function durationToLiteral({ value }: ConverterInput): Core.ILiteral {
    const parsed = typeof value === 'number' ? value : parseInt(value, 10);
    const duration = millisecondsToDuration(parsed);
    return {
        scalar: { primitive: { duration } }
    };
}

function parseCollection(list: string, type: InputType) {
    // If we're processing a nested collection, it may already have been parsed
    if (Array.isArray(list)) {
        return list;
    }

    const parsed = parseJSON(list);
    if (!Array.isArray(parsed)) {
        throw new Error('Value did not parse to an array');
    }
    return parsed;
}

function collectionToLiteral({
    value,
    typeDefinition: { subtype }
}: ConverterInput): Core.ILiteral {
    if (!subtype) {
        throw new Error('Unexpected missing subtype for collection');
    }
    if (!value.length) {
        return literalNone();
    }
    const parsed = parseCollection(value, subtype.type);
    if (!parsed.length) {
        return literalNone();
    }
    const converter = inputTypeConverters[subtype.type];
    const literals = parsed.map(value =>
        converter({ value, typeDefinition: subtype })
    );

    return {
        collection: {
            literals
        }
    };
}

const unsupportedInput = literalNone;
type ConverterFn = (input: ConverterInput) => Core.ILiteral;

/** Maps an `InputType` to a function which will convert its value into a `Literal` */
export const inputTypeConverters: Record<InputType, ConverterFn> = {
    [InputType.Binary]: literalNone,
    [InputType.Blob]: unsupportedInput,
    [InputType.Boolean]: booleanToLiteral,
    [InputType.Collection]: collectionToLiteral,
    [InputType.Datetime]: dateToLiteral,
    [InputType.Duration]: durationToLiteral,
    [InputType.Error]: unsupportedInput,
    [InputType.Float]: floatToLiteral,
    [InputType.Integer]: integerToLiteral,
    [InputType.Map]: unsupportedInput,
    [InputType.None]: literalNone,
    [InputType.Schema]: unsupportedInput,
    [InputType.String]: stringToLiteral,
    [InputType.Struct]: unsupportedInput,
    [InputType.Unknown]: unsupportedInput
};

export function inputToLiteral(input: InputProps): Core.ILiteral {
    if (!input.value) {
        return literalNone();
    }
    const { typeDefinition } = input;
    const value = input.value.toString();

    const converter = inputTypeConverters[typeDefinition.type];
    return converter({ value, typeDefinition });
}
