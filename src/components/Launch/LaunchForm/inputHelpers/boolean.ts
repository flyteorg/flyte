import { Core } from 'flyteidl';
import { InputValue } from '../types';
import { primitiveLiteralPaths } from './constants';
import { ConverterInput, InputHelper, InputValidatorParams } from './types';
import { extractLiteralWithCheck } from './utils';

/** Checks that a value is an acceptable boolean value. These can be
 * an actual Boolean instance, any variant of ('T', 'F', 'TRUE', 'FALSE') or
 * a 0/1 integer value.
 */
function isValidBoolean(value: InputValue) {
  if (typeof value === 'boolean') {
    return true;
  }
  if (typeof value === 'string') {
    return ['t', 'f', '0', '1', 'true', 'false'].includes(value.toLowerCase());
  }
  if (typeof value === 'number') {
    return value === 1 || value === 0;
  }
  return false;
}

/** Converts any of our acceptable values to an actual boolean. Assumes value
 * has been pre-validated.
 */
function parseBoolean(value: InputValue) {
  if (typeof value === 'boolean') {
    return value;
  }
  if (typeof value === 'string') {
    return ['f', 'false', '0'].includes(value.toLowerCase()) ? false : true;
  }
  return !!value;
}

function toLiteral({ value }: ConverterInput): Core.ILiteral {
  return { scalar: { primitive: { boolean: parseBoolean(value) } } };
}

function fromLiteral(literal: Core.ILiteral): InputValue {
  return extractLiteralWithCheck<boolean>(literal, primitiveLiteralPaths.scalarBoolean);
}

function validate({ value }: InputValidatorParams) {
  if (!isValidBoolean(value)) {
    throw new Error('Value is not a valid boolean');
  }
}

export const booleanHelper: InputHelper = {
  fromLiteral,
  toLiteral,
  validate,
  defaultValue: false,
};
