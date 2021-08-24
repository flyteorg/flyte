import { stringifyValue } from 'common/utils';
import { Core, Protobuf } from 'flyteidl';
import { InputValue } from '../types';
import { structPath } from './constants';
import { ConverterInput, InputHelper, InputValidatorParams } from './types';
import { extractLiteralWithCheck } from './utils';

export type PrimitiveType = string | number | boolean | null | object;

function asValueWithKind(value: Protobuf.IValue): Protobuf.Value {
    return value instanceof Protobuf.Value
        ? value
        : Protobuf.Value.create(value);
}

export function protobufValueToPrimitive(
    value: Protobuf.IValue
): PrimitiveType | PrimitiveType[] {
    const valueWithKind = asValueWithKind(value);
    const { kind } = valueWithKind;
    switch (kind) {
        case 'nullValue':
            return null;
        case 'structValue':
            if (valueWithKind.structValue == null) {
                throw new Error('Unexpected empty structValue field');
            }
            return protobufStructToObject(valueWithKind.structValue);
        case 'listValue':
            if (valueWithKind.listValue == null) {
                throw new Error('Unexpected empty listValue field');
            }
            return protobufListToArray(valueWithKind.listValue);
        case undefined:
            throw new Error('Unexpected missing Value.kind');
        default:
            return valueWithKind[kind];
    }
}

function primitiveToProtobufValue(value: any): Protobuf.IValue {
    if (value == null) {
        return { nullValue: Protobuf.NullValue.NULL_VALUE };
    }
    if (Array.isArray(value)) {
        return { listValue: { values: value.map(primitiveToProtobufValue) } };
    }
    switch (typeof value) {
        case 'boolean':
            return { boolValue: !!value };
        case 'number':
            return { numberValue: value };
        case 'string':
            return { stringValue: value };
        case 'object':
            return { structValue: objectToProtobufStruct(value) };
        default:
            throw new Error(`Unsupported value type: ${typeof value} `);
    }
}

function protobufListToArray(list: Protobuf.IListValue): PrimitiveType[] {
    if (!list.values) {
        return [];
    }

    return list.values.map(protobufValueToPrimitive);
}

function protobufStructToObject(struct: Protobuf.IStruct): Dictionary<any> {
    if (struct.fields == null) {
        return {};
    }

    return Object.entries(struct.fields).reduce<Dictionary<any>>(
        (out, [key, value]) => {
            out[key] = protobufValueToPrimitive(value);
            return out;
        },
        {}
    );
}

function objectToProtobufStruct(obj: Dictionary<any>): Protobuf.IStruct {
    const fields = Object.entries(obj).reduce<Record<string, Protobuf.IValue>>(
        (out, [key, value]) => {
            try {
                out[key] = primitiveToProtobufValue(value);
                return out;
            } catch (e) {
                throw new Error(
                    `Failed to convert value ${key} to Protobuf.Value: ${e}`
                );
            }
        },
        {}
    );

    return { fields };
}

function fromLiteral(literal: Core.ILiteral): InputValue {
    const structValue = extractLiteralWithCheck<Protobuf.IStruct>(
        literal,
        structPath
    );

    return stringifyValue(protobufStructToObject(structValue));
}

function toLiteral({ value }: ConverterInput): Core.ILiteral {
    let parsedObject: Dictionary<any>;

    if (typeof value === 'object') {
        parsedObject = value;
    } else {
        const stringValue =
            typeof value === 'string' ? value : value.toString();

        try {
            parsedObject = JSON.parse(stringValue);
            if (typeof parsedObject !== 'object') {
                throw new Error(`Result was of type: ${typeof parsedObject}`);
            }
        } catch (e) {
            throw new Error(`Value did not parse to an object`);
        }
    }

    return { scalar: { generic: objectToProtobufStruct(parsedObject) } };
}

function validate({ value }: InputValidatorParams) {
    if (typeof value !== 'string') {
        throw new Error('Value is not a string');
    }

    try {
        JSON.parse(value);
    } catch (e) {
        throw new Error(`Value did not parse to an object: ${e}`);
    }
}

export const structHelper: InputHelper = {
    fromLiteral,
    toLiteral,
    validate
};
