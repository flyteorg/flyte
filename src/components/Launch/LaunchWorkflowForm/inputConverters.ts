import { dateToTimestamp, millisecondsToDuration } from 'common/utils';
import { Core } from 'flyteidl';
import * as Long from 'long';
import * as LossLessJSON from 'lossless-json';
import { utc as moment } from 'moment';
import {
    InputProps,
    InputType,
    InputTypeDefinition,
    InputValue
} from './types';

function losslessReviver(key: string, value: any) {
    if (value && value.isLosslessNumber) {
        // *All* numbers will be converted to LossLessNumber, but we only want
        // to use a Long if it would overflow
        try {
            return value.valueOf();
        } catch {
            return Long.fromString(value.toString());
        }
    }
    return value;
}

function parseJSON(value: string) {
    return LossLessJSON.parse(value, losslessReviver);
}

interface ConverterInput {
    value: InputValue;
    typeDefinition: InputTypeDefinition;
}

function literalNone(): Core.ILiteral {
    return { scalar: { noneType: {} } };
}

/** Converts any of our acceptable values to an actual boolean. */
function parseBoolean(value: InputValue) {
    if (typeof value === 'boolean') {
        return value;
    }
    if (typeof value === 'string') {
        return ['f', 'false', '0'].includes(value.toLowerCase()) ? false : true;
    }
    return !!value;
}

function booleanToLiteral({ value }: ConverterInput): Core.ILiteral {
    return { scalar: { primitive: { boolean: parseBoolean(value) } } };
}

function stringToLiteral({ value }: ConverterInput): Core.ILiteral {
    const stringValue = typeof value === 'string' ? value : value.toString();
    return { scalar: { primitive: { stringValue } } };
}

function integerToLiteral({ value }: ConverterInput): Core.ILiteral {
    const integer =
        value instanceof Long ? value : Long.fromString(value.toString());
    return {
        scalar: { primitive: { integer } }
    };
}

function floatToLiteral({ value }: ConverterInput): Core.ILiteral {
    const floatValue =
        typeof value === 'number' ? value : parseFloat(value.toString());
    return {
        scalar: { primitive: { floatValue } }
    };
}

function dateToLiteral({ value }: ConverterInput): Core.ILiteral {
    const parsed =
        value instanceof Date ? value : moment(value.toString()).toDate();
    const datetime = dateToTimestamp(parsed);
    return {
        scalar: { primitive: { datetime } }
    };
}

function durationToLiteral({ value }: ConverterInput): Core.ILiteral {
    const parsed =
        typeof value === 'number' ? value : parseInt(value.toString(), 10);
    const duration = millisecondsToDuration(parsed);
    return {
        scalar: { primitive: { duration } }
    };
}

function parseCollection(list: string) {
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
    let parsed: any[];
    // If we're processing a nested collection, it may already have been parsed
    if (Array.isArray(value)) {
        parsed = value;
    } else {
        const stringValue =
            typeof value === 'string' ? value : value.toString();
        if (!stringValue.length) {
            return literalNone();
        }
        parsed = parseCollection(stringValue);
        if (!parsed.length) {
            return literalNone();
        }
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
    if (input.value == null) {
        return literalNone();
    }
    const { typeDefinition, value } = input;

    const converter = inputTypeConverters[typeDefinition.type];
    return converter({ value, typeDefinition });
}
