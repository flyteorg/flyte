import { Core } from 'flyteidl';
import { InputValue } from '../types';
import { schemaUriPath } from './constants';
import { ConverterInput, InputHelper, InputValidatorParams } from './types';
import { extractLiteralWithCheck } from './utils';

function fromLiteral(literal: Core.ILiteral): InputValue {
  return extractLiteralWithCheck<string>(literal, schemaUriPath);
}

function toLiteral({ typeDefinition, value }: ConverterInput): Core.ILiteral {
  const uri = typeof value === 'string' ? value : value.toString();
  // Note: schema type may be undefined if user is working with a generic schema.
  const {
    literalType: { schema: type },
  } = typeDefinition;
  return { scalar: { schema: { type, uri } } };
}

function validate({ value }: InputValidatorParams) {
  if (typeof value !== 'string') {
    throw new Error('Value is not a string');
  }
}

export const schemaHelper: InputHelper = {
  fromLiteral,
  toLiteral,
  validate,
};
