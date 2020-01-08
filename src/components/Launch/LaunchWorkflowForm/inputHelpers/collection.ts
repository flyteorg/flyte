import { Core } from 'flyteidl';
import { InputValue } from '../types';
import { literalNone } from './constants';
import { getHelperForInput } from './getHelperForInput';
import { parseJSON } from './parseJson';
import { ConverterInput, InputHelper } from './types';

function parseCollection(list: string) {
    const parsed = parseJSON(list);
    if (!Array.isArray(parsed)) {
        throw new Error('Value did not parse to an array');
    }
    return parsed;
}

function fromLiteral(literal: Core.ILiteral): InputValue {
    // TODO
    return '';
}

function toLiteral({
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

    const helper = getHelperForInput(subtype.type);
    const literals = parsed.map(value =>
        helper.toLiteral({ value, typeDefinition: subtype })
    );

    return {
        collection: {
            literals
        }
    };
}

function validate({ value }: ConverterInput) {
    if (typeof value !== 'string') {
        throw new Error('Value must be a string');
    }

    try {
        const parsed = parseCollection(value);
        if (!Array.isArray(parsed)) {
            throw new Error(`Value parsed to type: ${typeof parsed}`);
        }
    } catch (e) {
        throw new Error(`Failed to parse array: ${e}`);
    }
}

export const collectionHelper: InputHelper = {
    fromLiteral,
    toLiteral,
    validate
};
