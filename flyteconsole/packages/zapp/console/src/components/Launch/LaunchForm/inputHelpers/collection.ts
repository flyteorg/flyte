import { Core } from 'flyteidl';
import { InputTypeDefinition, InputValue } from '../types';
import { literalNone } from './constants';
import { getHelperForInput } from './getHelperForInput';
import { parseJSON } from './parseJson';
import { ConverterInput, InputHelper, InputValidatorParams } from './types';
import { collectionChildToString } from './utils';

const missingSubTypeError = 'Unexpected missing subtype for collection';

function parseCollection(list: string) {
  const parsed = parseJSON(list);
  if (!Array.isArray(parsed)) {
    throw new Error('Value did not parse to an array');
  }
  return parsed;
}

function fromLiteral(literal: Core.ILiteral, { subtype }: InputTypeDefinition): InputValue {
  if (!subtype) {
    throw new Error(missingSubTypeError);
  }
  if (!literal.collection) {
    throw new Error('Collection literal missing `collection` property');
  }
  if (!literal.collection.literals) {
    throw new Error('Collection literal missing `colleciton.literals` property');
  }

  const subTypeHelper = getHelperForInput(subtype.type);
  const values = literal.collection.literals.reduce<string[]>((out, literal) => {
    const value = subTypeHelper.fromLiteral(literal, subtype);
    if (value !== undefined) {
      out.push(collectionChildToString(subtype.type, value));
    }
    return out;
  }, []);
  return `[${values.join(',')}]`;
}

function toLiteral({ value, typeDefinition: { subtype } }: ConverterInput): Core.ILiteral {
  if (!subtype) {
    throw new Error(missingSubTypeError);
  }
  let parsed: any[];
  // If we're processing a nested collection, it may already have been parsed
  if (Array.isArray(value)) {
    parsed = value;
  } else {
    const stringValue = typeof value === 'string' ? value : value.toString();
    if (!stringValue.length) {
      return literalNone();
    }
    parsed = parseCollection(stringValue);
  }

  const helper = getHelperForInput(subtype.type);
  const literals = parsed.map((value) => helper.toLiteral({ value, typeDefinition: subtype }));

  return {
    collection: {
      literals,
    },
  };
}

function validate({ value }: InputValidatorParams) {
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
  validate,
};
