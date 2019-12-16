import { dateToTimestamp, millisecondsToDuration } from 'common/utils';
import { Core } from 'flyteidl';
import * as Long from 'long';
import { utc as moment } from 'moment';
import { InputProps, InputType, InputTypeDefinition } from './types';

interface ConverterInput {
    value: string;
    typeDefinition: InputTypeDefinition;
}

function literalNone(): Core.ILiteral {
    return { scalar: { noneType: {} } };
}

function booleanToLiteral({ value }: ConverterInput): Core.ILiteral {
    return { scalar: { primitive: { boolean: Boolean(value) } } };
}

function stringToLiteral({ value }: ConverterInput): Core.ILiteral {
    return { scalar: { primitive: { stringValue: value } } };
}

function integerToLiteral({ value }: ConverterInput): Core.ILiteral {
    return {
        scalar: { primitive: { integer: Long.fromString(value) } }
    };
}

function floatToLiteral({ value }: ConverterInput): Core.ILiteral {
    return {
        scalar: { primitive: { floatValue: parseFloat(value) } }
    };
}

function dateToLiteral({ value }: ConverterInput): Core.ILiteral {
    const datetime = dateToTimestamp(moment(value).toDate());
    return {
        scalar: { primitive: { datetime } }
    };
}

function durationToLiteral({ value }: ConverterInput): Core.ILiteral {
    const duration = millisecondsToDuration(parseInt(value, 10));
    return {
        scalar: { primitive: { duration } }
    };
}

function parseCollection(list: string, type: InputType) {
    let parsed;
    // We can't rely on the built-in parser, because it will try to read integers as
    // normal numbers and overflow for large values. So we'll split the list manually
    if (type === InputType.Integer) {
        if (list[0] !== '[' || list[list.length - 1] !== ']') {
            throw new Error('Value is not in array format ([])');
        }
        parsed = list
            .substr(1, list.length - 2) // remove the []
            .split(',')
            .map(s => s.trim());
    } else {
        parsed = JSON.parse(list);
        if (!Array.isArray(parsed)) {
            throw new Error('Value did not parse to an array');
        }
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
